package migrator

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ilydyu/migrator/internal/config"
	"github.com/ilydyu/migrator/internal/migration"
	"github.com/jackc/pgx/v5"
)

type Migrator struct {
	db     *pgx.Conn
	config config.Config
}

func NewMigrator(cfg config.Config) *Migrator {
	connString := os.Getenv("DATABASE_URL")

	if connString == "" {
		connString = fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	}

	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	migrator := Migrator{
		db:     conn,
		config: cfg,
	}

	return &migrator
}

func (m *Migrator) Init() {
	ctx := context.Background()
	tx, err := m.db.Begin(ctx)

	defer tx.Rollback(ctx)

	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_history (
		version         VARCHAR(255) PRIMARY KEY,
		script_name     VARCHAR(255) NOT NULL,
		executed_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		execution_time  INTERVAL NOT NULL
		);
	`)

	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_lock (
			id              INTEGER PRIMARY KEY DEFAULT 1,
			locked          BOOLEAN DEFAULT FALSE,
			locked_at       TIMESTAMP,
			locked_by       VARCHAR(255)
		);
	`)

	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO schema_lock (id, locked) VALUES (1, FALSE)
		ON CONFLICT (id) DO NOTHING;
	`)

	if err != nil {
		log.Fatal(err)
	}

	err = tx.Commit(ctx)

	if err != nil {
		log.Fatal(err)
	}
}

func (m *Migrator) Create(table string) {
	timestamp := time.Now().UTC().Format("20060102150405")

	upFile := fmt.Sprintf("db/migrations/%s_%s.up.sql", timestamp, table)
	file, err := os.Create(upFile)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	downFile := fmt.Sprintf("db/migrations/%s_%s.down.sql", timestamp, table)

	file, err = os.Create(downFile)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	fmt.Printf("Create up migration: %s\n", upFile)
	fmt.Printf("Create down migration: %s\n", downFile)
}

func (m *Migrator) UpDry() {
	needUp := m.getUnApplyiedMigrations(`select version, script_name, executed_at, execution_time from schema_history`, func(name string) bool {
		return !strings.Contains(name, "down")
	})

	for _, title := range needUp {
		fmt.Println(title)
	}
}

func (m *Migrator) Up() {
	ctx := context.Background()
	tx, err := m.db.Begin(ctx)

	if err != nil {
		log.Fatal(err)
	}

	defer tx.Rollback(ctx)

	m.acquireLock(ctx, tx)

	defer m.releaseLock(ctx, tx)

	needUp := m.getUnApplyiedMigrations(`select version, script_name, executed_at, execution_time from schema_history`, func(name string) bool {
		return !strings.Contains(name, "down")
	})

	for _, f := range needUp {
		now := time.Now()
		data, err := os.ReadFile("db/migrations/" + f)

		if err != nil {
			log.Fatal(err)
		}

		if len(data) == 0 {
			log.Fatal("Migration file is empty, write sql first.")
		}

		_, err = tx.Exec(ctx, string(data))

		if err != nil {
			log.Fatal(err)
		}

		version, _, _ := strings.Cut(f, "_")
		elapsed := time.Since(now)

		_, err = tx.Exec(ctx, `insert into schema_history (version, script_name, executed_at, execution_time) values ($1, $2, $3, $4)`, version, f, now, elapsed)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Migration %s applied in %s\n", f, elapsed)
	}
}

func (m *Migrator) DownSteps(steps int) {
	ctx := context.Background()
	tx, err := m.db.Begin(ctx)

	if err != nil {
		log.Fatal(err)
	}

	defer tx.Rollback(ctx)

	m.acquireLock(ctx, tx)

	defer m.releaseLock(ctx, tx)

	needDown := m.getUnApplyiedMigrations(`select version, script_name, executed_at, execution_time from schema_history order by execution_time desc limit `+strconv.Itoa(steps), func(name string) bool {
		return !strings.Contains(name, "up")
	})

	needDown = needDown[len(needDown)-steps:]

	for _, f := range needDown {
		data, err := os.ReadFile("db/migrations/" + f)

		if err != nil {
			log.Fatal(err)
		}

		if len(data) == 0 {
			log.Fatal("Migration file is empty, write sql first.")
		}

		_, err = tx.Exec(ctx, string(data))

		if err != nil {
			log.Fatal(err)
		}

		version, _, _ := strings.Cut(f, "_")

		_, err = tx.Exec(ctx, `delete from schema_history where version = $1`, version)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Migration %s rollback\n", f)
	}
}

func (m *Migrator) Down() {
	ctx := context.Background()
	tx, err := m.db.Begin(ctx)

	if err != nil {
		log.Fatal(err)
	}

	defer tx.Rollback(ctx)

	m.acquireLock(ctx, tx)

	defer m.releaseLock(ctx, tx)

	needDown := m.getUnApplyiedMigrations(`select version, script_name, executed_at, execution_time from schema_history order by execution_time desc limit 1`, func(name string) bool {
		return !strings.Contains(name, "up")
	})

	needDown = needDown[len(needDown)-1:]

	for _, f := range needDown {
		data, err := os.ReadFile("db/migrations/" + f)

		if err != nil {
			log.Fatal(err)
		}

		if len(data) == 0 {
			log.Fatal("Migration file is empty, write sql first.")
		}

		_, err = tx.Exec(ctx, string(data))

		if err != nil {
			log.Fatal(err)
		}

		version, _, _ := strings.Cut(f, "_")

		_, err = tx.Exec(ctx, `delete from schema_history where version = $1`, version)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Migration %s rollback\n", f)
	}
}

func (m *Migrator) getUnApplyiedMigrations(sql string, check func(name string) bool) []string {
	migrations := m.getMigrations(sql)

	h := map[string]struct{}{}

	for _, mig := range migrations {
		h[mig.ScriptName] = struct{}{}
	}

	need := []string{}

	files, err := os.ReadDir("db/migrations")

	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		_, exists := h[f.Name()]

		if !exists && check(f.Name()) {
			need = append(need, f.Name())
		}
	}

	slices.Sort(need)

	return need
}

func (m *Migrator) getMigrations(sql string) []migration.Migration {
	migrations := []migration.Migration{}
	ctx := context.Background()
	rows, err := m.db.Query(ctx, sql)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var migr migration.Migration

		err := rows.Scan(&migr.Version, &migr.ScriptName, &migr.ExecutedAt, &migr.ExecutionTime)

		if err != nil {
			log.Fatal(err)
		}

		migrations = append(migrations, migr)
	}

	err = rows.Err()

	if err != nil {
		log.Fatal(err)
	}

	return migrations
}

func (m *Migrator) History() {
	migrations := m.getMigrations(`select version, script_name, executed_at, execution_time from schema_history`)
	for _, migr := range migrations {
		fmt.Printf("%s was applied in %s, it took %v\n", migr.ScriptName, migr.ExecutedAt.Format("2006-01-02 15:04:05"), migr.ExecutionTime)
	}
}

func (m *Migrator) acquireLock(ctx context.Context, tx pgx.Tx) {
	result, err := tx.Exec(ctx, `
        UPDATE schema_lock 
        SET locked = TRUE, 
            locked_at = NOW(), 
            locked_by = $1
        WHERE id = 1 AND locked = FALSE
    `, m.config.Database.Host)

	if err != nil {
		log.Fatalf("failed to acquire lock: %v\n", err)
	}

	if result.RowsAffected() == 0 {
		log.Fatalf("could not acquire lock: already locked by another process\n")
	}
}

func (m *Migrator) releaseLock(ctx context.Context, tx pgx.Tx) {
	_, err := tx.Exec(ctx, `
        UPDATE schema_lock 
        SET locked = FALSE, 
            locked_at = NULL, 
            locked_by = NULL
        WHERE id = 1
    `)

	if err != nil {
		log.Fatalf("failed to release lock: %v\n", err)
	}

	err = tx.Commit(ctx)

	if err != nil {
		log.Fatal(err)
	}
}

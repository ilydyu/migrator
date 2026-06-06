package commander

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/ilydyu/migrator/internal/migrator"
)

type Commander struct {
	migrator *migrator.Migrator
}

func NewCommander(m *migrator.Migrator) *Commander {
	return &Commander{migrator: m}
}

func (c *Commander) Command(args []string) {
	if len(args) <= 1 {
		printHelp()
		return
	}

	switch args[1] {
	case "init":
		c.migrator.Init()
	case "create":
		if len(args) <= 2 {
			log.Fatal("Invalid number of parameters, check help")
		}

		c.migrator.Create(args[2])
	case "up":
		if len(args) > 2 {
			if !strings.Contains(args[2], "dry") {
				log.Fatal("Invalid usage, check help")
			}

			c.migrator.UpDry()
		} else {
			c.migrator.Up()
		}
	case "down":
		if len(args) > 2 {
			if !strings.Contains(args[2], "step") && !strings.Contains(args[2], "=") {
				log.Fatal("Invalid usage, check help")
			}

			s := strings.Split(args[2], "=")[1]

			steps, err := strconv.Atoi(s)

			if err != nil {
				log.Fatal("Invalid usage, check help")
			}

			c.migrator.DownSteps(steps)
		} else {
			c.migrator.Down()
		}
	case "history":
		c.migrator.History()
	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Println("migrator init - create directory, shema_history and schema_lock table")
	fmt.Println("migrator create [migration name] - create migration. Example: migrator create create_users")
	fmt.Println("migrator up - apply your migrations")
	fmt.Println("migrator up dry - show what migration should be apply in the future")
	fmt.Println("migrator down - rollback your last migration")
	fmt.Println("migrator down step=[step] - rollback your migrations from end to start, where [step] - it is a number of migrations. Example: migrator down step=2")
	fmt.Println("migrator history - show history")
}

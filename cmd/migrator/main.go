package main

import (
	"os"

	"github.com/ilydyu/migrator/internal/commander"
	"github.com/ilydyu/migrator/internal/config"
	"github.com/ilydyu/migrator/internal/migrator"
)

func main() {
	cfg := config.NewConfig()
	m := migrator.NewMigrator(cfg)
	comm := commander.NewCommander(m)
	comm.Command(os.Args)
}

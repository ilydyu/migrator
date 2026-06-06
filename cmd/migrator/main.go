package main

import (
	"os"

	"github.com/ilydyu/migrator/internal/commander"
	"github.com/ilydyu/migrator/internal/config"
)

func main() {
	cfg := config.NewMockConfig()
	comm := commander.NewCommander(cfg)
	comm.Command(os.Args)
}

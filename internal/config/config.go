package config

import (
	"fmt"
	"log"
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Database struct {
		Name     string `yaml:"name"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"database"`
}

func NewMockConfig() Config {
	return Config{
		Database: struct {
			Name     string "yaml:\"name\""
			Host     string "yaml:\"host\""
			Port     int    "yaml:\"port\""
			Username string "yaml:\"username\""
			Password string "yaml:\"password\""
		}{
			Name:     "postgres",
			Host:     "localhost",
			Port:     5432,
			Username: "postgres",
			Password: "postgres",
		},
	}
}

func NewConfig() Config {
	cfg := NewMockConfig()
	_, err := os.Stat("db/config.yaml")

	if err == nil {

		data, err := os.ReadFile("db/config.yaml")

		if err != nil {
			log.Fatal(err)
		}

		err = yaml.Unmarshal(data, &cfg)

		if err != nil {
			fmt.Println(1)
			log.Fatal(err)
		}

	} else {
		err = os.MkdirAll("db", 0755)

		if err != nil {
			log.Fatal(err)
		}

		err = os.MkdirAll("db/migrations", 0755)

		if err != nil {
			log.Fatal(err)
		}

		bytes, err := yaml.Marshal(cfg)

		fmt.Println(2)
		if err != nil {
			log.Fatal(err)
		}

		os.WriteFile("db/config.yaml", bytes, 0644)
	}

	return cfg
}

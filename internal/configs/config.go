package configs

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	DatabaseURL string
	HasuraURL   string
}

func LoadConfig() *Config {
	c := &Config{}

	c.DatabaseURL = os.Getenv("DATABASE_URL")
	c.HasuraURL = os.Getenv("HASURA_URL")

	return c
}

package configs

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	DatabaseURL string
	Hasura      struct {
		URL    string
		Secret string
		Source string // default: "postgres"
	}
}

func LoadConfig() *Config {
	c := &Config{}

	c.DatabaseURL = os.Getenv("DATABASE_URL")

	c.Hasura.URL = os.Getenv("HASURA_URL")
	c.Hasura.Secret = os.Getenv("HASURA_SECRET")
	c.Hasura.Source = os.Getenv("HASURA_SOURCE")

	return c
}

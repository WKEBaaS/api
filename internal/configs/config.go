package configs

import (
	"crypto/rsa"
	"fmt"
	"os"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	DatabaseURL string
	Hasura      struct {
		URL    string
		Secret string
		Source string // default: "postgres"
	}
	Keycloak struct {
		ClientId     string
		ClientSecret string
		Issuer       string
		RedirectURL  string
	}
	Jwt struct {
		PrivateKey *rsa.PrivateKey
		PublicKey  *rsa.PublicKey
		Issuer     string
		ExpireIn   int
	}
}

func LoadConfig() *Config {
	c := &Config{}

	c.DatabaseURL = os.Getenv("DATABASE_URL")

	c.Hasura.URL = os.Getenv("HASURA_URL")
	c.Hasura.Secret = os.Getenv("HASURA_SECRET")
	c.Hasura.Source = os.Getenv("HASURA_SOURCE")

	c.Keycloak.ClientId = os.Getenv("KEYCLOAK_CLIENT_ID")
	c.Keycloak.ClientSecret = os.Getenv("KEYCLOAK_CLIENT_SECRET")
	c.Keycloak.Issuer = os.Getenv("KEYCLOAK_ISSUER")
	c.Keycloak.RedirectURL = os.Getenv("KEYCLOAK_REDIRECT_URL")

	var err error
	c.Jwt.PrivateKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(os.Getenv("JWT_PRIVATE_KEY")))
	if err != nil {
		panic(fmt.Sprintf("Failed to parse private key %s", err))
	}
	c.Jwt.PublicKey, err = jwt.ParseRSAPublicKeyFromPEM([]byte(os.Getenv("JWT_PUBLIC_KEY")))
	if err != nil {
		panic(fmt.Sprintf("Failed to parse public key %s", err))
	}
	c.Jwt.ExpireIn = func() int {
		valueStr := os.Getenv("JWT_EXPIRE_IN")
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
		return 3600
	}()
	c.Jwt.Issuer = os.Getenv("JWT_ISSUER")
	return c
}

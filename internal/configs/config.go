package configs

import (
	"fmt"
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
	"github.com/lestrrat-go/jwx/v3/jwk"
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
	JWK struct {
		PrivateKey jwk.Key
		PublicKey  jwk.Key
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
	c.JWK.PrivateKey, err = jwk.ParseKey([]byte(os.Getenv("JWT_PRIVATE_KEY")))
	if err != nil {
		panic(fmt.Sprintf("Failed to parse private key %s", err))
	}

	c.JWK.PublicKey, err = jwk.PublicKeyOf(c.JWK.PrivateKey)
	if err != nil {
		panic(fmt.Sprintf("Failed to get public key %s", err))
	}

	c.JWK.ExpireIn = func() int {
		valueStr := os.Getenv("JWT_EXPIRE_IN")
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
		// Default to 7 days
		return 604800
	}()
	c.JWK.Issuer = os.Getenv("JWT_ISSUER")
	return c
}

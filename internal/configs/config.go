package configs

import (
	"log"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
)

type Config struct {
	DatabaseURL string
	Keycloak    struct {
		ClientId     string
		ClientSecret string
		Issuer       string
		RedirectURL  string
		LogoutURL    string
	}
	JWK struct {
		PrivateKey jwk.Key
		PublicKey  jwk.Key
		Algorithm  jwa.KeyAlgorithm
		Issuer     string
		ExpireIn   time.Duration
	}
	Kube struct {
		ConfigPath                    string
		ProjectsNamespace             string
		ProjectsWildcardTLSSecretName string
	}

	PROJECTS_HOST string
}

func LoadConfig() *Config {
	c := &Config{}

	c.DatabaseURL = os.Getenv("DATABASE_URL")

	c.Keycloak.ClientId = os.Getenv("KEYCLOAK_CLIENT_ID")
	c.Keycloak.ClientSecret = os.Getenv("KEYCLOAK_CLIENT_SECRET")
	c.Keycloak.Issuer = os.Getenv("KEYCLOAK_ISSUER")
	c.Keycloak.RedirectURL = os.Getenv("KEYCLOAK_REDIRECT_URL")

	logoutUrl, err := url.Parse(c.Keycloak.Issuer + "/protocol/openid-connect/logout")
	if err != nil {
		slog.Error("Failed to parse logout URL", "error", err)
		panic(err)
	}
	logoutUrlQuery := logoutUrl.Query()
	logoutUrlQuery.Set("client_id", c.Keycloak.ClientId)
	logoutUrl.RawQuery = logoutUrlQuery.Encode()
	c.Keycloak.LogoutURL = logoutUrl.String()

	c.JWK.PrivateKey, err = jwk.ParseKey([]byte(os.Getenv("JWK_PRIVATE_KEY")))
	if err != nil {
		log.Fatalf("Failed to parse private key %s", err)
	}

	c.JWK.PublicKey, err = jwk.PublicKeyOf(c.JWK.PrivateKey)
	if err != nil {
		log.Fatalf("Failed to get public key %s", err)
	}

	var ok bool
	c.JWK.Algorithm, ok = c.JWK.PrivateKey.Algorithm()
	if !ok {
		panic("Failed to get algorithm from JWK private key")
	}

	c.JWK.ExpireIn = func() time.Duration {
		valueStr := os.Getenv("JWT_EXPIRE_IN")
		if value, err := strconv.Atoi(valueStr); err == nil {
			return time.Duration(value) * time.Second
		}
		// Default to 7 days
		return 7 * 24 * time.Hour
	}()
	c.JWK.Issuer = os.Getenv("JWT_ISSUER")

	c.Kube.ConfigPath = os.Getenv("KUBE_CONFIG_PATH")
	c.Kube.ProjectsNamespace = os.Getenv("KUBE_PROJECTS_NAMESPACE")
	c.Kube.ProjectsWildcardTLSSecretName = os.Getenv("KUBE_PROJECTS_WILDCARD_TLS_SECRET_NAME")

	c.PROJECTS_HOST = os.Getenv("PROJECTS_HOST")

	return c
}

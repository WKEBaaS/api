package repo

import (
	"i3s-service/internal/configs"
	"net/http"

	// "github.com/hasura/go-graphql-client"
	"github.com/Khan/genqlient/graphql"
)

type Repository struct {
	client graphql.Client
}

type AuthedTransport struct {
	hasuraURL    string
	hasuraSecret string
	wrapped      http.RoundTripper
}

func InitRepository(config *configs.Config) *Repository {
	repo := &Repository{}

	httpClient := &http.Client{
		Transport: &AuthedTransport{
			hasuraURL:    config.Hasura.URL,
			hasuraSecret: config.Hasura.Secret,
			wrapped:      http.DefaultTransport,
		},
	}

	client := graphql.NewClient(config.Hasura.URL+"/v1/graphql", httpClient)

	repo.client = client

	return repo
}

func (t *AuthedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("x-hasura-admin-secret", t.hasuraSecret)
	req.Header.Set("x-hasura-role", "admin")
	return t.wrapped.RoundTrip(req)
}

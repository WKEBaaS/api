package repo

import (
	"context"
	"encoding/json"
	"log"
	"os"
)

func (r *Repository) GetUserIDByIdentity(ctx context.Context, provider string, providerID string) (*string, error) {
	var query = `query MyQuery {
	  identities(where: {provider: {_eq: $provider}, provider_id: {_eq: $provider_id}}) {
	    user_id
	  }
	}`

	var resp struct {
		Identities []struct {
			UserID string `json:"user_id"`
		} `json:"identities"`
	}

	variables := map[string]any{
		"provider":    provider,
		"provider_id": providerID,
	}

	err := r.client.Query(ctx, query, &resp, variables)
	if err != nil {
		return nil, err
	}

	print(query)

	return nil, nil
}

func print(v interface{}) {
	w := json.NewEncoder(os.Stdout)
	w.SetIndent("", "\t")
	err := w.Encode(v)
	if err != nil {
		log.Fatalf("failed to encode: %v", err)
		panic(err)
	}
}

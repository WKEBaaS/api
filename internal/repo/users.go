package repo

import (
	"context"
	"encoding/json"
	"log"
	"os"
)

func (r *Repository) GetUserIDByIdentity(ctx context.Context, provider string, providerID string) (*string, error) {
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

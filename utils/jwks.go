package utils

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"log/slog"

	"github.com/lestrrat-go/jwx/v3/jwk"
)

func NewEd25519JWK(ctx context.Context) (jwk.Key, *ed25519.PrivateKey, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to generate Ed25519 key:", "error", err)
		return nil, nil, errors.New("failed to generate Ed25519 key")
	}

	// Create JWK from the public key
	key, err := jwk.Import(publicKey)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create JWK from public key:", "error", err)
		return nil, nil, errors.New("failed to create JWK from public key")
	}

	key.Set(jwk.AlgorithmKey, "EdDSA")

	return key, &privateKey, nil
}

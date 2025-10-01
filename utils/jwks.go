package utils

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/lestrrat-go/jwx/v3/jwk"
)

func NewEd25519JWK(ctx context.Context) (jwk.Key, jwk.Key, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to generate Ed25519 key:", "error", err)
		return nil, nil, errors.New("failed to generate Ed25519 key")
	}

	// Create JWK from the publicJWK publicJWK
	publicJWK, err := jwk.Import(publicKey)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create JWK from public key:", "error", err)
		return nil, nil, errors.New("failed to create JWK from public key")
	}

	privateJWK, err := jwk.Import(privateKey)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create JWK from private key:", "error", err)
		return nil, nil, errors.New("failed to create JWK from private key")
	}

	publicJWK.Set(jwk.AlgorithmKey, "EdDSA")

	return publicJWK, privateJWK, nil
}

func NewEd25519JWKStringified(ctx context.Context) (string, string, error) {
	jwkKey, privateKey, err := NewEd25519JWK(ctx)
	if err != nil {
		return "", "", err
	}

	publicJSON, err := json.Marshal(jwkKey)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to marshal JWK to JSON:", "error", err)
		return "", "", errors.New("failed to marshal JWK to JSON")
	}

	privateJSON, err := json.Marshal(privateKey)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to marshal private JWK to JSON:", "error", err)
		return "", "", errors.New("failed to marshal private JWK to JSON")
	}

	return string(publicJSON), string(privateJSON), nil
}

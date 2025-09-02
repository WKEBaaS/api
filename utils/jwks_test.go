package utils

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
)

func TestNewEd25519JWK(t *testing.T) {
	// Set up logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	ctx := context.Background()

	t.Log("=== Testing NewEd25519JWK Function ===")

	// Call the function
	key, privateKey, err := NewEd25519JWK(ctx)
	if err != nil {
		t.Fatalf("NewEd25519JWK failed: %v", err)
	}

	d, _ := json.Marshal(key)
	t.Log("Public JWK Key:", string(d))

	// Verify the JWK key is not nil
	if key == nil {
		t.Fatal("Expected JWK key to be non-nil")
	}

	// Verify private key length (Ed25519 private keys are 64 bytes)
	if len(privateKey) != ed25519.PrivateKeySize {
		t.Errorf("Expected private key size to be %d, got %d", ed25519.PrivateKeySize, len(privateKey))
	}

	t.Log("✓ Basic validation passed")
}

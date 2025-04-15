package services

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

func (s *Service) AuthCallback(ctx context.Context, code string, nonce string, config *oauth2.Config, verifier *oidc.IDTokenVerifier) error {

	oauth2Token, err := config.Exchange(ctx, code)
	if err != nil {
		return err
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return fmt.Errorf("no id_token field in oauth2 token")
	}

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return fmt.Errorf("failed to verify ID token: %w", err)
	}

	if idToken.Nonce != nonce {
		return fmt.Errorf("nonce did not match")
	}

	var idTokenClaims struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Hasura struct {
			XHasuraDefaultRole  string `json:"x-hasura-default-role"`
			XHasuraAllowedRoles string `json:"x-hasura-allowed-roles"`
		} `json:"https://hasura.io/jwt/claims"`
	}
	if err := idToken.Claims(&idTokenClaims); err != nil {
		return fmt.Errorf("failed to parse ID token claims: %w", err)
	}

	_ = jwt.MapClaims{
		"email":  idTokenClaims.Email,
		"name":   idTokenClaims.Name,
		"client": config.ClientID,
		"iss":    s.config.Jwt.Issuer,
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(time.Duration(s.config.Jwt.ExpireIn) * time.Second).Unix(),
		"sub":    idToken.Subject,
		"https://hasura.io/jwt/claims": map[string]any{
			"x-hasura-default-role":  idTokenClaims.Hasura.XHasuraDefaultRole,
			"x-hasura-allowed-roles": idTokenClaims.Hasura.XHasuraAllowedRoles,
		},
	}

	return nil
}

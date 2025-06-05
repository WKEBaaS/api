package services

import (
	"baas-api/internal/configs"
	"baas-api/internal/dto"
	"baas-api/internal/repo"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/lestrrat-go/jwx/v3/jwt"
	"golang.org/x/oauth2"
	"gorm.io/datatypes"
)

var (
	ErrNoIDTokenField             = errors.New("no id_token field in oauth2 token")
	ErrFailedToVerifyIDToken      = errors.New("failed to verify ID token")
	ErrNonceMismatch              = errors.New("nonce mismatch")
	ErrFailedToParseIDTokenClaims = errors.New("failed to parse ID token claims")
	ErrFailedToBuildJWTToken      = errors.New("failed to build JWT token")
	ErrFailedToSignJWTToken       = errors.New("failed to sign JWT token")
)

type AuthService interface {
	AuthCallback(ctx context.Context, in *dto.AuthCallbackInput, secure bool, oauth2Config *oauth2.Config, verifier *oidc.IDTokenVerifier) (*dto.AuthCallbackOutput, error)
}

type authService struct {
	entiryRepo repo.EntityRepository
	userRepo   repo.UserRepository
	config     *configs.Config
}

func NewAuthService(config *configs.Config, entiryRepo repo.EntityRepository, userRepo repo.UserRepository) AuthService {
	return &authService{
		entiryRepo: entiryRepo,
		userRepo:   userRepo,
		config:     config,
	}
}

func (s *authService) AuthCallback(ctx context.Context, in *dto.AuthCallbackInput, secure bool, oauth2Config *oauth2.Config, verifier *oidc.IDTokenVerifier) (*dto.AuthCallbackOutput, error) {
	oauth2Token, err := oauth2Config.Exchange(ctx, in.Code)
	if err != nil {
		return nil, err
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, ErrNoIDTokenField
	}

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to verify ID token", "error", err)
		return nil, ErrFailedToVerifyIDToken
	}

	if idToken.Nonce != in.NonceCookie {
		return nil, ErrNonceMismatch
	}

	var idTokenClaims struct {
		DisplayName string `json:"name"`
		Email       string `json:"email"`
		Username    string `json:"preferred_username"`
	}
	if err := idToken.Claims(&idTokenClaims); err != nil {
		slog.ErrorContext(ctx, "Failed to parse ID token claims", "error", err)
		return nil, ErrFailedToParseIDTokenClaims
	}

	userID, exist, err := s.userRepo.GetUserIDByProviderAndID(ctx, "keycloak", idToken.Subject)
	if err != nil {
		return nil, err
	}

	// If user does not exist, create a new user
	if !exist {
		userEntityID, err := s.entiryRepo.GetEntityByChineseName(ctx, "使用者")
		if err != nil {
			return nil, err
		}

		id, err := s.userRepo.CreateUserFromIdentity(ctx, &repo.CreateUserFromIdentityInput{
			UserEntityID: userEntityID.ID,
			Name:         idTokenClaims.DisplayName,
			Email:        &idTokenClaims.Email,
			Username:     idTokenClaims.Username,
			Provider:     "keycloak",
			ProviderID:   idToken.Subject,
			IdentityData: datatypes.JSON(fmt.Appendf(nil, `{"email":"%s"}`, idTokenClaims.Email)),
		})
		if err != nil {
			return nil, err
		}
		userID = id
	}

	tok, err := jwt.NewBuilder().
		Subject(*userID).
		Issuer(s.config.JWK.Issuer).
		IssuedAt(time.Now()).
		Expiration(time.Now().Add(s.config.JWK.ExpireIn)).
		Claim("email", idTokenClaims.Email).
		Claim("displayName", idTokenClaims.DisplayName).
		Claim("username", idTokenClaims.Username).
		Build()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to build JWT token", "error", err)
		return nil, ErrFailedToBuildJWTToken
	}

	signedToken, err := jwt.Sign(tok, jwt.WithKey(s.config.JWK.Algorithm, s.config.JWK.PrivateKey))
	if err != nil {
		slog.ErrorContext(ctx, "Failed to sign JWT token", "error", err)
		return nil, ErrFailedToSignJWTToken
	}

	out := &dto.AuthCallbackOutput{}
	out.Status = http.StatusFound
	out.Body.Ok = true
	out.TokenCookie = &http.Cookie{Name: "token", Value: string(signedToken), Path: "/", HttpOnly: true, Secure: secure}
	out.Url = in.RedirectURL

	return out, nil
}

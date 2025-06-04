package services

import (
	"baas-api/internal/configs"
	"baas-api/internal/dto"
	"baas-api/internal/repo"
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
	"gorm.io/datatypes"
)

var (
	ErrNoIDTokenField             = errors.New("no id_token field in oauth2 token")
	ErrFailedToVerifyIDToken      = errors.New("failed to verify ID token")
	ErrNonceMismatch              = errors.New("nonce mismatch")
	ErrFailedToParseIDTokenClaims = errors.New("failed to parse ID token claims")
)

type AuthService interface {
	AuthCallback(ctx context.Context, in *dto.AuthCallbackInput, oauth2Config *oauth2.Config, verifier *oidc.IDTokenVerifier) (*dto.AuthCallbackOutput, error)
}

type authService struct {
	entiryRepo  repo.EntityRepository
	userRepo    repo.UserRepository
	jwtIssuer   string
	jwtExpireIn time.Duration
}

func NewAuthService(config *configs.Config, entiryRepo repo.EntityRepository, userRepo repo.UserRepository) AuthService {
	return &authService{
		entiryRepo:  entiryRepo,
		userRepo:    userRepo,
		jwtIssuer:   config.JWK.Issuer,
		jwtExpireIn: config.JWK.ExpireIn,
	}
}

func (s *authService) AuthCallback(ctx context.Context, in *dto.AuthCallbackInput, oauth2Config *oauth2.Config, verifier *oidc.IDTokenVerifier) (*dto.AuthCallbackOutput, error) {
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
		Name     string `json:"name"`
		Email    string `json:"email"`
		Username string `json:"preferred_username"`
	}
	if err := idToken.Claims(&idTokenClaims); err != nil {
		slog.ErrorContext(ctx, "Failed to parse ID token claims", "error", err)
		return nil, ErrFailedToParseIDTokenClaims
	}

	log.Printf("ID Token Claims: %+v", idTokenClaims)

	exist, err := s.userRepo.CheckUserExistsByProviderAndID(ctx, "keycloak", idToken.Subject)
	if err != nil {
		return nil, err
	}

	// If user does not exist, create a new user
	var userID string
	if !exist {
		userEntityID, err := s.entiryRepo.GetEntityByChineseName(ctx, "使用者")
		if err != nil {
			return nil, err
		}

		id, err := s.userRepo.CreateUserFromIdentity(ctx, &repo.CreateUserFromIdentityInput{
			UserEntityID: userEntityID.ID,
			Name:         idTokenClaims.Name,
			Email:        &idTokenClaims.Email,
			Username:     idTokenClaims.Username,
			Provider:     "keycloak",
			ProviderID:   idToken.Subject,
			IdentityData: datatypes.JSON([]byte(fmt.Sprintf(`{"email":"%s"}`, idTokenClaims.Email))),
		})
		if err != nil {
			return nil, err
		}
		userID = *id
	}

	log.Printf("User ID: %s", userID)

	// tok, err := jwt.NewBuilder().
	// 	Issuer(s.jwtIssuer).
	// 	IssuedAt(time.Now()).
	// 	Expiration(time.Now().Add(s.jwtExpireIn)).
	// 	Claim("email", idTokenClaims.Email).
	// 	Claim("name", idTokenClaims.Name).
	// 	Build()

	// _ = jwt.MapClaims{
	// 	"email":  idTokenClaims.Email,
	// 	"name":   idTokenClaims.Name,
	// 	"client": config.ClientID,
	// 	"iss":    s.jwtIssuer,
	// 	"iat":    time.Now().Unix(),
	// 	"exp":    time.Now().Add(s.jwtExpireIn).Unix(),
	// 	"sub":    idToken.Subject,
	// }

	return nil, nil
}

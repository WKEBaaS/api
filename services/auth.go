package services

// import (
// 	"baas-api/config"
// 	"baas-api/internal/dto"
// 	"baas-api/internal/repo"
// 	"context"
// 	"errors"
// 	"fmt"
// 	"log/slog"
// 	"time"
//
// 	"github.com/coreos/go-oidc"
// 	"github.com/lestrrat-go/jwx/v3/jwt"
// 	"golang.org/x/oauth2"
// 	"gorm.io/datatypes"
// )
//
// var (
// 	ErrNoIDTokenField             = errors.New("no id_token field in oauth2 token")
// 	ErrFailedToVerifyIDToken      = errors.New("failed to verify ID token")
// 	ErrNonceMismatch              = errors.New("nonce mismatch")
// 	ErrFailedToParseIDTokenClaims = errors.New("failed to parse ID token claims")
// 	ErrFailedToBuildJWTToken      = errors.New("failed to build JWT token")
// 	ErrFailedToSignJWTToken       = errors.New("failed to sign JWT token")
// )
//
// type AuthService interface {
// 	AuthCallback(ctx context.Context, in *dto.AuthCallbackInput, oauth2Config *oauth2.Config, verifier *oidc.IDTokenVerifier) ([]byte, error)
// }
//
// type authService struct {
// 	entityRepo repo.EntityRepository
// 	userRepo   repo.UserRepository
// 	config     *config.Config
// }
//
// func NewAuthService(config *config.Config, ep repo.EntityRepository, up repo.UserRepository) AuthService {
// 	return &authService{
// 		entityRepo: ep,
// 		userRepo:   up,
// 		config:     config,
// 	}
// }
//
// func (s *authService) AuthCallback(ctx context.Context, in *dto.AuthCallbackInput, oauth2Config *oauth2.Config, verifier *oidc.IDTokenVerifier) ([]byte, error) {
// 	oauth2Token, err := oauth2Config.Exchange(ctx, in.Code)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
// 	if !ok {
// 		return nil, ErrNoIDTokenField
// 	}
//
// 	idToken, err := verifier.Verify(ctx, rawIDToken)
// 	if err != nil {
// 		slog.ErrorContext(ctx, "Failed to verify ID token", "error", err)
// 		return nil, ErrFailedToVerifyIDToken
// 	}
//
// 	if idToken.Nonce != in.NonceCookie {
// 		return nil, ErrNonceMismatch
// 	}
//
// 	var idTokenClaims struct {
// 		DisplayName string   `json:"name"`
// 		Email       *string  `json:"email"`
// 		Username    string   `json:"preferred_username"`
// 		BaaSRoles   []string `json:"baas_roles"`
// 	}
// 	if err := idToken.Claims(&idTokenClaims); err != nil {
// 		slog.ErrorContext(ctx, "Failed to parse ID token claims", "error", err)
// 		return nil, ErrFailedToParseIDTokenClaims
// 	}
//
// 	userID, exist, err := s.userRepo.GetIDByProviderAndID(ctx, "keycloak", idToken.Subject)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// If user does not exist, create a new user
// 	if !exist {
// 		userEntity, err := s.entityRepo.GetByChineseName(ctx, "使用者")
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		var identityData datatypes.JSON
// 		if idTokenClaims.Email != nil {
// 			identityData = datatypes.JSON(fmt.Appendf(nil, `{"email":"%s"}`, *idTokenClaims.Email))
// 		}
//
// 		id, err := s.userRepo.CreateFromIdentity(ctx, &repo.CreateUserFromIdentityInput{
// 			UserEntityID: userEntity.ID,
// 			Name:         idTokenClaims.DisplayName,
// 			Email:        idTokenClaims.Email,
// 			Username:     idTokenClaims.Username,
// 			Provider:     "keycloak",
// 			ProviderID:   idToken.Subject,
// 			IdentityData: identityData,
// 		})
// 		if err != nil {
// 			return nil, err
// 		}
// 		userID = id
// 	}
//
// 	tok, err := jwt.NewBuilder().
// 		Subject(*userID).
// 		Issuer(s.config.JWK.Issuer).
// 		IssuedAt(time.Now()).
// 		Expiration(time.Now().Add(s.config.JWK.ExpireIn)).
// 		Claim("email", idTokenClaims.Email).
// 		Claim("displayName", idTokenClaims.DisplayName).
// 		Claim("username", idTokenClaims.Username).
// 		Claim("baas_roles", idTokenClaims.BaaSRoles).
// 		Build()
// 	if err != nil {
// 		slog.ErrorContext(ctx, "Failed to build JWT token", "error", err)
// 		return nil, ErrFailedToBuildJWTToken
// 	}
//
// 	signedToken, err := jwt.Sign(tok, jwt.WithKey(s.config.JWK.Algorithm, s.config.JWK.PrivateKey))
// 	if err != nil {
// 		slog.ErrorContext(ctx, "Failed to sign JWT token", "error", err)
// 		return nil, ErrFailedToSignJWTToken
// 	}
//
// 	return signedToken, nil
// }

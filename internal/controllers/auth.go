package controllers

import (
	"baas-api/internal/configs"
	"baas-api/internal/dto"
	"baas-api/internal/services"
	"context"
	"log/slog"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type AuthController interface {
	RegisterAuthAPIs(api huma.API)
	authMiddleware(ctx huma.Context, next func(huma.Context))
	authLogin(ctx context.Context, input *dto.AuthLoginInput) (*dto.AuthLoginOutput, error)
	authCallback(ctx context.Context, input *dto.AuthCallbackInput) (*dto.AuthCallbackOutput, error)
}

type authController struct {
	authService  services.AuthService
	oauth2Config oauth2.Config
	provider     *oidc.Provider
	verifier     *oidc.IDTokenVerifier
}

func NewAuthController(config *configs.Config, authService services.AuthService) AuthController {
	provider, err := oidc.NewProvider(context.Background(), config.Keycloak.Issuer)
	if err != nil {
		slog.Error("Failed to create OIDC provider", "error", err)
		panic(err)
	}

	oauth2Config := oauth2.Config{
		ClientID:     config.Keycloak.ClientId,
		ClientSecret: config.Keycloak.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  config.Keycloak.RedirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: config.Keycloak.ClientId})

	return &authController{
		authService:  authService,
		oauth2Config: oauth2Config,
		provider:     provider,
		verifier:     verifier,
	}
}

func (c *authController) RegisterAuthAPIs(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "auth-login",
		Method:      http.MethodGet,
		Path:        "/auth/login",
		Summary:     "Login with Keycloak",
		Description: "Login with Keycloak",
		Tags:        []string{"Auth"},

		Middlewares: huma.Middlewares{c.authMiddleware},
	}, c.authLogin)

	huma.Register(api, huma.Operation{
		OperationID: "auth-callback",
		Method:      http.MethodGet,
		Path:        "/auth/callback",
		Summary:     "Callback",
		Description: "Callback",
		Tags:        []string{"Auth"},
	}, c.authCallback)
}

func (c *authController) authMiddleware(ctx huma.Context, next func(huma.Context)) {
	ctx = huma.WithValue(ctx, "TLS", ctx.TLS() != nil)
	next(ctx)
}

func (c *authController) authLogin(ctx context.Context, input *dto.AuthLoginInput) (*dto.AuthLoginOutput, error) {
	resp := &dto.AuthLoginOutput{}
	state := uuid.New().String()
	nonce := uuid.New().String()
	tls, ok := ctx.Value("TLS").(bool)
	if !ok {
		return nil, huma.Error500InternalServerError("context value TLS not found ")
	}
	resp.RedirectCookie = &http.Cookie{Name: "redirect_url", Value: input.RedirectURL, Path: "/", HttpOnly: true, Secure: tls}
	resp.StateCookie = &http.Cookie{Name: "state", Value: state, Path: "/", HttpOnly: true, Secure: tls}
	resp.NonceCookie = &http.Cookie{Name: "nonce", Value: nonce, Path: "/", HttpOnly: true, Secure: tls}
	resp.Status = http.StatusFound
	resp.Url = c.oauth2Config.AuthCodeURL(state, oidc.Nonce(nonce))
	return resp, nil
}

func (c *authController) authCallback(ctx context.Context, in *dto.AuthCallbackInput) (*dto.AuthCallbackOutput, error) {
	c.authService.AuthCallback(ctx, in, &c.oauth2Config, c.verifier)

	resp := &dto.AuthCallbackOutput{}
	resp.Status = http.StatusOK
	resp.Body.Message = "Authentication successful"
	return resp, nil
}

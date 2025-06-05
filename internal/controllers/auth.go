package controllers

import (
	"baas-api/internal/configs"
	"baas-api/internal/dto"
	"baas-api/internal/services"
	"context"
	"log/slog"
	"net/http"
	"time"

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
	authLogout(ctx context.Context, input *dto.AuthLogoutInput) (*dto.AuthLogoutOutput, error)
}

type authController struct {
	logoutUrl    string
	authService  services.AuthService
	oauth2Config oauth2.Config
	provider     *oidc.Provider
	verifier     *oidc.IDTokenVerifier
	config       *configs.Config
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
		config:       config,
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
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, c.authCallback)

	huma.Register(api, huma.Operation{
		OperationID: "auth-logout",
		Method:      http.MethodGet,
		Path:        "/auth/logout",
		Summary:     "Logout",
		Description: "Logout from Keycloak",
		Tags:        []string{"Auth"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, c.authLogout)
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

	if input.RedirectURL == nil {
		home := c.config.BaaS.Home.String()
		input.RedirectURL = &home
	}

	resp.RedirectCookie = &http.Cookie{Name: "redirect_url", Value: *input.RedirectURL, Path: "/", HttpOnly: true, Secure: tls}
	resp.StateCookie = &http.Cookie{Name: "state", Value: state, Path: "/", HttpOnly: true, Secure: tls}
	resp.NonceCookie = &http.Cookie{Name: "nonce", Value: nonce, Path: "/", HttpOnly: true, Secure: tls}
	resp.Status = http.StatusFound
	resp.Url = c.oauth2Config.AuthCodeURL(state, oidc.Nonce(nonce))
	return resp, nil
}

func (c *authController) authCallback(ctx context.Context, in *dto.AuthCallbackInput) (*dto.AuthCallbackOutput, error) {
	tls, ok := ctx.Value("TLS").(bool)
	if !ok {
		return nil, huma.Error500InternalServerError("context value TLS not found ")
	}

	out, err := c.authService.AuthCallback(ctx, in, tls, &c.oauth2Config, c.verifier)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (c *authController) authLogout(ctx context.Context, input *dto.AuthLogoutInput) (*dto.AuthLogoutOutput, error) {
	tls, ok := ctx.Value("TLS").(bool)
	if !ok {
		return nil, huma.Error500InternalServerError("context value TLS not found ")
	}

	out := &dto.AuthLogoutOutput{}

	// delete the token cookie
	out.TokenCookie = &http.Cookie{Name: "token", Value: "", Path: "/", HttpOnly: true, Secure: tls, Expires: time.Unix(0, 0)}

	out.Status = http.StatusFound
	out.Url = c.config.Keycloak.LogoutURL

	return out, nil
}

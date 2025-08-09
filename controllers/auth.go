package controllers

// import (
// 	"baas-api/config"
// 	"baas-api/internal/controllers/middlewares"
// 	"baas-api/internal/dto"
// 	"baas-api/internal/services"
// 	"context"
// 	"log/slog"
// 	"net/http"
// 	"time"
//
// 	"github.com/coreos/go-oidc"
// 	"github.com/danielgtaylor/huma/v2"
// 	"github.com/google/uuid"
// 	"golang.org/x/oauth2"
// )
//
// type AuthController interface {
// 	RegisterAuthAPIs(api huma.API)
// 	authLogin(ctx context.Context, input *dto.AuthLoginInput) (*dto.AuthLoginOutput, error)
// 	authCallback(ctx context.Context, input *dto.AuthCallbackInput) (*dto.AuthCallbackOutput, error)
// 	authLogout(ctx context.Context, input *dto.AuthLogoutInput) (*dto.AuthLogoutOutput, error)
// }
//
// type authController struct {
// 	logoutUrl    string
// 	authService  services.AuthService
// 	oauth2Config oauth2.Config
// 	provider     *oidc.Provider
// 	verifier     *oidc.IDTokenVerifier
// 	config       *config.Config
// }
//
// func NewAuthController(config *config.Config, authService services.AuthService) AuthController {
// 	provider, err := oidc.NewProvider(context.Background(), config.OIDC.Issuer)
// 	if err != nil {
// 		slog.Error("Failed to create OIDC provider", "error", err)
// 		panic(err)
// 	}
//
// 	oauth2Config := oauth2.Config{
// 		ClientID:     config.OIDC.ClientID,
// 		ClientSecret: config.OIDC.ClientSecret,
// 		Endpoint:     provider.Endpoint(),
// 		RedirectURL:  config.OIDC.RedirectURL,
// 		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
// 	}
//
// 	verifier := provider.Verifier(&oidc.Config{ClientID: config.OIDC.ClientID})
//
// 	return &authController{
// 		authService:  authService,
// 		oauth2Config: oauth2Config,
// 		provider:     provider,
// 		verifier:     verifier,
// 		config:       config,
// 	}
// }
//
// func (c *authController) RegisterAuthAPIs(api huma.API) {
// 	huma.Register(api, huma.Operation{
// 		OperationID: "auth-login",
// 		Method:      http.MethodGet,
// 		Path:        "/auth/login",
// 		Summary:     "Login with Keycloak",
// 		Description: "Login with Keycloak",
// 		Tags:        []string{"Auth"},
// 		Middlewares: huma.Middlewares{middlewares.TLSMiddleware},
// 	}, c.authLogin)
//
// 	huma.Register(api, huma.Operation{
// 		OperationID: "auth-callback",
// 		Method:      http.MethodGet,
// 		Path:        "/auth/callback",
// 		Summary:     "Callback",
// 		Description: "Callback",
// 		Tags:        []string{"Auth"},
// 		Middlewares: huma.Middlewares{middlewares.TLSMiddleware},
// 	}, c.authCallback)
//
// 	huma.Register(api, huma.Operation{
// 		OperationID: "auth-logout",
// 		Method:      http.MethodGet,
// 		Path:        "/auth/logout",
// 		Summary:     "Logout",
// 		Description: "Logout from Keycloak",
// 		Tags:        []string{"Auth"},
// 		Middlewares: huma.Middlewares{middlewares.TLSMiddleware},
// 	}, c.authLogout)
// }
//
// func (c *authController) authLogin(ctx context.Context, in *dto.AuthLoginInput) (*dto.AuthLoginOutput, error) {
// 	resp := &dto.AuthLoginOutput{}
// 	state := uuid.New().String()
// 	nonce := uuid.New().String()
// 	tls, ok := ctx.Value("TLS").(bool)
// 	if !ok {
// 		return nil, huma.Error500InternalServerError("context value TLS not found ")
// 	}
//
// 	resp.RedirectCookie = &http.Cookie{Name: "redirect_url", Value: in.RedirectURL, Path: "/", HttpOnly: true, Secure: tls}
// 	resp.StateCookie = &http.Cookie{Name: "state", Value: state, Path: "/", HttpOnly: true, Secure: tls}
// 	resp.NonceCookie = &http.Cookie{Name: "nonce", Value: nonce, Path: "/", HttpOnly: true, Secure: tls}
// 	resp.Status = http.StatusFound
// 	resp.Url = c.oauth2Config.AuthCodeURL(state, oidc.Nonce(nonce))
// 	return resp, nil
// }
//
// func (c *authController) authCallback(ctx context.Context, in *dto.AuthCallbackInput) (*dto.AuthCallbackOutput, error) {
// 	tls, ok := ctx.Value("TLS").(bool)
// 	if !ok {
// 		return nil, huma.Error500InternalServerError("context value TLS not found ")
// 	}
//
// 	signedToken, err := c.authService.AuthCallback(ctx, in, &c.oauth2Config, c.verifier)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	out := &dto.AuthCallbackOutput{}
// 	out.Status = http.StatusFound
// 	out.Body.Ok = true
// 	out.TokenCookie = &http.Cookie{Name: "token", Value: string(signedToken), Path: "/", HttpOnly: true, Secure: tls}
// 	out.Url = in.RedirectURL
//
// 	return out, nil
// }
//
// func (c *authController) authLogout(ctx context.Context, input *dto.AuthLogoutInput) (*dto.AuthLogoutOutput, error) {
// 	tls, ok := ctx.Value("TLS").(bool)
// 	if !ok {
// 		return nil, huma.Error500InternalServerError("context value TLS not found ")
// 	}
//
// 	out := &dto.AuthLogoutOutput{}
//
// 	// delete the token cookie
// 	out.TokenCookie = &http.Cookie{Name: "token", Value: "", Path: "/", HttpOnly: true, Secure: tls, Expires: time.Unix(0, 0)}
//
// 	out.Status = http.StatusFound
// 	out.Url = *c.config.OIDC.LogoutURL
//
// 	// add the post_logout_redirect_uri if provided
// 	if input.PostLogoutRedirectURI != "" {
// 		query := out.Url.Query()
// 		query.Set("post_logout_redirect_uri", input.PostLogoutRedirectURI)
// 		out.Url.RawQuery = query.Encode()
// 	}
//
// 	return out, nil
// }

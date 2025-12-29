// Package middlewares
//
// BaaS API Auth Middleware
package middlewares

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"baas-api/internal/config"

	"github.com/danielgtaylor/huma/v2"
	"github.com/samber/do/v2"
)

type Session struct {
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	Token     string    `json:"token"`
	UpdatedAt time.Time `json:"updatedAt"`
	UserID    string    `json:"userId"`
	ID        string    `json:"id,omitempty"`
	IPAddress string    `json:"ipAddress,omitempty"`
	UserAgent string    `json:"userAgent,omitempty"`
}

// User represents a user in the system
type User struct {
	CreatedAt     time.Time `json:"createdAt"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"emailVerified"`
	Name          string    `json:"name"`
	UpdatedAt     time.Time `json:"updatedAt"`
	ID            string    `json:"id,omitempty"`
	Image         string    `json:"image,omitempty"`
}

type GetSessionResponse struct {
	Session Session `json:"session"`
	User    User    `json:"user"`
}

type AuthMiddleware func(huma.Context, func(huma.Context))

func NewAuthMiddleware(i do.Injector) (AuthMiddleware, error) {
	api := do.MustInvoke[huma.API](i)
	config := do.MustInvoke[*config.Config](i)

	return func(ctx huma.Context, next func(huma.Context)) {
		getSessionURL := config.Auth.URL.JoinPath("/api/auth/get-session")
		req, err := http.NewRequest(http.MethodGet, getSessionURL.String(), nil)
		if err != nil {
			slog.Error("Failed to create request for session", "error", err)
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "Failed to create request for session")
			return
		}

		cookies := huma.ReadCookies(ctx)
		slog.Debug("AuthMiddleware cookies", "cookies", cookies)
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}

		req.Header.Set("X-Forwarded-Host", ctx.Host())
		slog.Debug("AuthMiddleware request", "url", req.URL.String(), "headers", req.Header)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			slog.Error("Failed to get session", "error", err)
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "Failed to get session")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			slog.Info("Invalid session", "status", resp.StatusCode, "message", string(body))
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Invalid session")
			return
		}

		sessionResp := &GetSessionResponse{}
		if err := json.NewDecoder(resp.Body).Decode(&sessionResp); err != nil {
			slog.Error("Failed to decode session response", "error", err)
		}

		if sessionResp == nil {
			slog.Error("Session response is nil")
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Session not found")
			return
		}

		jwt := resp.Header.Get("Set-Auth-Jwt")

		ctx = huma.WithValue(ctx, "session", sessionResp.Session)
		ctx = huma.WithValue(ctx, "user", sessionResp.User)
		ctx = huma.WithValue(ctx, "jwt", jwt)
		next(ctx)
	}, nil
}

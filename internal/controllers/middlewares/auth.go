package middlewares

import (
	"baas-api/internal/configs"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

// NewAuthMiddleWare creates a middleware that checks if the user has the required roles
// to access the API endpoint based on the JWT token provided in the Authorization header.
//
// Note: This function created middleware will update the user's info in the context with
// the "UserID" key if the token is valid.
func NewAuthMiddleWare(api huma.API, config *configs.Config, authSchema string) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		var anyOfNeededRoles []string
		isAuthorizationRequired := false
		for _, opScheme := range ctx.Operation().Security {
			var ok bool
			if anyOfNeededRoles, ok = opScheme[authSchema]; ok {
				isAuthorizationRequired = true
				break
			}
		}

		if !isAuthorizationRequired {
			next(ctx)
			return
		}

		token := strings.TrimPrefix(ctx.Header("Authorization"), "Bearer ")
		if len(token) == 0 {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
			return
		}

		parsed, err := jwt.ParseString(token,
			jwt.WithKey(config.JWK.Algorithm, config.JWK.PublicKey),
			jwt.WithValidate(true),
			jwt.WithIssuer(config.JWK.Issuer),
		)
		if err != nil {
			slog.WarnContext(ctx.Context(), "Failed to parse JWT token", "error", err)
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Update the context with the user ID from the JWT token
		sub, ok := parsed.Subject()
		if !ok || sub == "" {
			slog.WarnContext(ctx.Context(), "JWT token does not contain a valid subject", "token", token)
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
			return
		}
		ctx = huma.WithValue(ctx, "UserID", sub)

		var roles []any
		err = parsed.Get("baas_roles", &roles)
		if err != nil {
			slog.WarnContext(ctx.Context(), "Failed to get roles from JWT token", "error", err)
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
			return
		}

		for _, role := range roles {
			if roleStr, ok := role.(string); ok {
				if slices.Contains(anyOfNeededRoles, roleStr) {
					slog.DebugContext(ctx.Context(), "User has required role", "role", roleStr)
					next(ctx)
					return
				}
			}
		}

		slog.WarnContext(ctx.Context(), "User does not have any required roles", "requiredRoles", anyOfNeededRoles, "userRoles", roles)
		huma.WriteErr(api, ctx, http.StatusForbidden, "Forbidden: insufficient permissions")
	}
}

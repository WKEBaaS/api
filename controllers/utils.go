// Package controllers
package controllers

import (
	"baas-api/controllers/middlewares"
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
)

func GetSessionFromContext(ctx context.Context) (*middlewares.Session, error) {
	session, ok := ctx.Value("session").(middlewares.Session)
	if !ok {
		slog.ErrorContext(ctx, "Session not found in context")
		return nil, huma.Error401Unauthorized("Session not found in context")
	}
	return &session, nil
}

func GetJWTFromContext(ctx context.Context) (string, error) {
	jwt, ok := ctx.Value("jwt").(string)
	if !ok {
		slog.ErrorContext(ctx, "JWT not found in context")
		return "", huma.Error401Unauthorized("JWT not found in context")
	}
	return jwt, nil
}

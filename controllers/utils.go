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

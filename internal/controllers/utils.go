package controllers

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
)

func GetUserIDFromContext(ctx context.Context) (*string, error) {
	userID, ok := ctx.Value("UserID").(string)
	if !ok || userID == "" {
		slog.WarnContext(ctx, "User ID not found in context")
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	return &userID, nil
}

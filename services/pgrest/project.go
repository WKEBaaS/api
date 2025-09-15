package pgrest

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/supabase-community/postgrest-go"
)

type CreateProjectOutput struct {
	ID         string `json:"id"`
	Ref        string `json:"ref"`
	AuthSecret string `json:"auth_secret"`
}

func (s *PgRestService) CreateProject(ctx context.Context, jwt string, name string, description string) (*CreateProjectOutput, error) {
	pgrst := postgrest.NewClient(s.config.PgREST.URL.String(), "api", nil)
	pgrst.SetAuthToken(jwt)
	resp := pgrst.Rpc("create_project", "", map[string]any{
		"name":        name,
		"description": description,
	})

	if pgrst.ClientError != nil {
		slog.ErrorContext(ctx, "Failed to call create_project RPC", "error", pgrst.ClientError)
		return nil, huma.Error500InternalServerError("Failed to call create_project RPC")
	}

	var project []CreateProjectOutput
	if err := json.Unmarshal([]byte(resp), &project); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal create_project response", "error", err)
		return nil, huma.Error500InternalServerError("Failed to unmarshal create_project response")
	}

	if len(project) == 0 {
		slog.ErrorContext(ctx, "create_project returned no project")
		return nil, huma.Error500InternalServerError("create_project returned no project")
	}

	return &project[0], nil
}

func (s *PgRestService) DeleteProject(ctx context.Context, jwt string, id string) (*string, error) {
	pgrst := postgrest.NewClient(s.config.PgREST.URL.String(), "api", nil)
	pgrst.SetAuthToken(jwt)
	resp := pgrst.Rpc("delete_project", "", map[string]any{
		"project_id": id,
	})

	if pgrst.ClientError != nil {
		slog.ErrorContext(ctx, "Failed to call delete_project RPC", "error", pgrst.ClientError)
		return nil, huma.Error500InternalServerError("Failed to call delete_project RPC")
	}

	var ref string
	if err := json.Unmarshal([]byte(resp), &ref); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal delete_project response", "error", err)
		return nil, huma.Error500InternalServerError("Failed to unmarshal delete_project response")
	}

	if ref == "" {
		slog.ErrorContext(ctx, "delete_project returned no reference")
		return nil, huma.Error500InternalServerError("delete_project returned no reference")
	}

	slog.InfoContext(ctx, "Project deleted", "project_id", id, "reference", ref, "response", resp)

	return &ref, nil
}

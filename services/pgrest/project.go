package pgrest

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/supabase-community/postgrest-go"
)

type CreateProjectOutput struct {
	ID                string `json:"id"`
	Ref               string `json:"ref"`
	AuthSecret        string `json:"auth_secret"`
	S3Bucket          string `json:"s3_bucket"`
	S3AccessKeyID     string `json:"s3_access_key_id"`
	S3SecretAccessKey string `json:"s3_secret_access_key"`
}

type DeleteProjectOutput struct {
	Ref           string `json:"ref"`
	S3Bucket      string `json:"s3_bucket"`
	S3AccessKeyID string `json:"s3_access_key_id"`
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
		if pgErr, _ := s.UnmarshalPgRestError([]byte(resp)); pgErr != nil {
			slog.ErrorContext(ctx, "delete_project error", "code", pgErr.Code, "message", pgErr.Message, "detail", pgErr.Detail, "hint", pgErr.Hint)
			return nil, huma.Error500InternalServerError("delete_project error: " + pgErr.Message)
		}

		slog.ErrorContext(ctx, "Failed to unmarshal create_project response", "error", err)
		return nil, huma.Error500InternalServerError("Failed to unmarshal create_project response")
	}

	if len(project) == 0 {
		slog.ErrorContext(ctx, "create_project returned no project")
		return nil, huma.Error500InternalServerError("create_project returned no project")
	}

	return &project[0], nil
}

func (s *PgRestService) DeleteProject(ctx context.Context, jwt string, id string) (*DeleteProjectOutput, error) {
	pgrst := postgrest.NewClient(s.config.PgREST.URL.String(), "api", nil)
	pgrst.SetAuthToken(jwt)
	resp := pgrst.Rpc("delete_project", "", map[string]any{
		"project_id": id,
	})

	if pgrst.ClientError != nil {
		slog.ErrorContext(ctx, "Failed to call delete_project RPC", "error", pgrst.ClientError)
		return nil, huma.Error500InternalServerError("Failed to call delete_project RPC")
	}

	var out []DeleteProjectOutput
	if err := json.Unmarshal([]byte(resp), &out); err != nil {
		if pgErr, _ := s.UnmarshalPgRestError([]byte(resp)); pgErr != nil {
			slog.ErrorContext(ctx, "delete_project error", "code", pgErr.Code, "message", pgErr.Message, "detail", pgErr.Detail, "hint", pgErr.Hint)
			return nil, huma.Error500InternalServerError("delete_project error: " + pgErr.Message)
		}

		slog.ErrorContext(ctx, "Failed to unmarshal delete_project response", "error", err)
		return nil, huma.Error500InternalServerError("Failed to unmarshal delete_project response")
	}

	if len(out) == 0 {
		slog.ErrorContext(ctx, "delete_project returned no project")
		return nil, huma.Error500InternalServerError("delete_project returned no project")
	}

	return &out[0], nil
}

package pgrest

import (
	"context"
	"encoding/json"
	"log/slog"

	"baas-api/internal/dto"

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

func (s *service) CreateProject(ctx context.Context, jwt string, name string, description string) (*CreateProjectOutput, error) {
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

func (s *service) DeleteProject(ctx context.Context, jwt string, id string) (*DeleteProjectOutput, error) {
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

type UpdateProjectPayload struct {
	ID             string   `json:"p_id"`
	Name           *string  `json:"p_name"`
	Description    *string  `json:"p_description"`
	TrustedOrigins []string `json:"p_trusted_origins"`
	ProxyURL       *string  `json:"p_proxy_url"`
}

type UpdateProjectOutput struct {
	Ref string `json:"ref"`
}

func (s *service) UpdateProject(ctx context.Context, jwt string, payload UpdateProjectPayload) (*UpdateProjectOutput, error) {
	pgrst := postgrest.NewClient(s.config.PgREST.URL.String(), "api", nil)
	pgrst.SetAuthToken(jwt)
	resp := pgrst.Rpc("update_project", "", payload)

	if pgrst.ClientError != nil {
		slog.ErrorContext(ctx, "Failed to call delete_project RPC", "error", pgrst.ClientError)
		return nil, huma.Error500InternalServerError("Failed to call update_project RPC")
	}

	var out UpdateProjectOutput
	if err := json.Unmarshal([]byte(resp), &out); err != nil {
		if pgErr, _ := s.UnmarshalPgRestError([]byte(resp)); pgErr != nil {
			slog.ErrorContext(ctx, "update_project error", "code", pgErr.Code, "message", pgErr.Message, "detail", pgErr.Detail, "hint", pgErr.Hint)
			return nil, huma.Error500InternalServerError("update_project error: " + pgErr.Message)
		}

		slog.ErrorContext(ctx, "Failed to unmarshal update_project response", "error", err)
		return nil, huma.Error500InternalServerError("Failed to unmarshal update_project response")
	}

	return &out, nil
}

type CreateOrUpdateAuthProviderPayload struct {
	ProjectID string                      `json:"project_id"`
	Providers map[string]dto.AuthProvider `json:"providers"`
}

func (s *service) CreateOrUpdateAuthProvider(ctx context.Context, jwt string, payload CreateOrUpdateAuthProviderPayload) error {
	pgrst := postgrest.NewClient(s.config.PgREST.URL.String(), "api", nil)
	pgrst.SetAuthToken(jwt)
	pgrst.Rpc("create_or_update_auth_providers", "", map[string]CreateOrUpdateAuthProviderPayload{
		"payload": payload,
	})

	if pgrst.ClientError != nil {
		slog.ErrorContext(ctx, "Failed to call create_or_update_auth_providers RPC", "error", pgrst.ClientError)
		return huma.Error500InternalServerError("Failed to call create_or_update_auth_providers RPC")
	}

	return nil
}

func (s *service) CheckProjectPermission(ctx context.Context, jwt string, projectID string) error {
	pgrst := postgrest.NewClient(s.config.PgREST.URL.String(), "api", nil)
	pgrst.SetAuthToken(jwt)
	resp := pgrst.Rpc("check_project_permission", "", map[string]any{
		"p_project_id": projectID,
	})
	if pgrst.ClientError != nil {
		if pgErr, _ := s.UnmarshalPgRestError([]byte(resp)); pgErr != nil {
			return pgErr.ToHumaError()
		}
		slog.ErrorContext(ctx, "Failed to call check_project_permission RPC", "error", pgrst.ClientError)
		return huma.Error500InternalServerError("Failed to call check_project_permission RPC")
	}
	return nil
}

func (s *service) CheckProjectPermissionByRef(ctx context.Context, jwt string, projectRef string) error {
	pgrst := postgrest.NewClient(s.config.PgREST.URL.String(), "api", nil)
	pgrst.SetAuthToken(jwt)
	resp := pgrst.Rpc("check_project_permission_by_ref", "", map[string]any{
		"p_project_ref": projectRef,
	})
	if pgrst.ClientError != nil {
		if pgErr, _ := s.UnmarshalPgRestError([]byte(resp)); pgErr != nil {
			return pgErr.ToHumaError()
		}
		slog.ErrorContext(ctx, "Failed to call check_project_permission RPC", "error", pgrst.ClientError)
		return huma.Error500InternalServerError("Failed to call check_project_permission RPC")
	}
	return nil
}

func (s *service) CreateClassFunction(ctx context.Context, jwt string, in *dto.CreateClassFunctionInput) error {
	pgrst := postgrest.NewClient(s.config.PgREST.URL.String(), "api", nil)
	pgrst.SetAuthToken(jwt)
	resp := pgrst.Rpc("new_create_class_function", "", map[string]any{
		"p_project_id":    in.Body.ProjectID,
		"p_name":          in.Body.Name,
		"p_version":       in.Body.Version,
		"p_description":   in.Body.Description,
		"p_authenticated": in.Body.Authenticated,
		"p_root_node":     in.Body.RootNode,
		"p_nodes":         in.Body.Node,
	})

	if pgrst.ClientError != nil {
		if pgErr, _ := s.UnmarshalPgRestError([]byte(resp)); pgErr != nil {
			slog.ErrorContext(ctx, "create_class_function error", "code", pgErr.Code, "message", pgErr.Message, "detail", pgErr.Detail, "hint", pgErr.Hint)
			return pgErr.ToHumaError()
		}

		slog.ErrorContext(ctx, "Failed to call create_class_function RPC", "error", pgrst.ClientError)
		return huma.Error500InternalServerError("Failed to call create_class_function RPC")
	}
	return nil
}

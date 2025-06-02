package hasura

import (
	"baas-api/internal/configs"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HasuraService struct {
	config *configs.Config
}

func InitHasuraService(config *configs.Config) *HasuraService {
	service := &HasuraService{}
	service.config = config

	return service
}

type HasuraResponse struct {
	Error string `json:"error"`
	Path  string `json:"path"`
	Code  string `json:"code"`
}

func (s *HasuraService) PostMetadata(data any) error {
	client := &http.Client{Timeout: 10 * time.Second}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		s.config.Hasura.URL+"/v1/metadata",
		bytes.NewBuffer(jsonData))

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Hasura-Admin-Secret", s.config.Hasura.Secret)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		hasuraResp := &HasuraResponse{}
		if err := json.Unmarshal(body, &hasuraResp); err != nil {
			return fmt.Errorf( //nolint: goerr113
				"status_code: %d\nresponse: %s",
				resp.StatusCode,
				body,
			)
		}

		if hasuraResp.Code == "already-tracked" || hasuraResp.Code == "already-exists" {
			return nil
		}

		return fmt.Errorf("failed to post request: status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

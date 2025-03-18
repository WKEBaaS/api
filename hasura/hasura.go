package hasura

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"i3s-service/internal/configs"
	"io"
	"net/http"
	"time"
)

type HasuraMetadata struct {
	Type string `json:"type"`
	Args struct {
		Source string `json:"source"`
		Table  struct {
			Schema string `json:"schema"`
			Name   string `json:"name"`
		} `json:"table"`
		Configuration struct {
			CustomName        string            `json:"custom_name"`
			CustomColumnNames map[string]string `json:"custom_column_names"`
			CustomRootFields  struct {
				Select          string `json:"select"`
				SelectByPk      string `json:"select_by_pk"`     //nolint: tagliatelle
				SelectAggregate string `json:"select_aggregate"` //nolint: tagliatelle
				Insert          string `json:"insert"`
				InsertOne       string `json:"insert_one"` //nolint: tagliatelle
				Update          string `json:"update"`
				UpdateByPk      string `json:"update_by_pk"` //nolint: tagliatelle
				Delete          string `json:"delete"`
				DeleteByPk      string `json:"delete_by_pk"` //nolint: tagliatelle
			} `json:"custom_root_fields"`
		} `json:"configuration"`
	} `json:"args"`
}

func PostHasuraMetadata(hasuraURL string, hasuraSecret string, data any) error {
	client := &http.Client{Timeout: 10 * time.Second}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		hasuraURL+"/v1/metadata",
		bytes.NewBuffer(jsonData))

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Hasura-Admin-Secret", hasuraSecret)

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

		return fmt.Errorf("failed to post request: status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

func PostI3SMetadata(config *configs.Config) error {

	if err := postAuthUserMetadata(config); err != nil {
		return fmt.Errorf("failed to post auth.users metadata: %w", err)
	}

	return nil
}

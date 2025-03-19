package hasura

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HasuraResponse struct {
	Error string `json:"error"`
	Path  string `json:"path"`
	Code  string `json:"code"`
}

type HasuraTrackTableMetadata struct {
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

func (s *HasuraService) PostTrackTableMetadata(data any) error {
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

func (s *HasuraService) PostTrackTableMetadataWithTableName(schema string, tableSingular string, tablePlural string) error {
	meta := &HasuraTrackTableMetadata{}

	meta.Type = "pg_track_table"
	meta.Args.Source = s.config.Hasura.Source

	meta.Args.Table.Schema = schema
	meta.Args.Table.Name = tablePlural
	meta.Args.Configuration.CustomName = tablePlural

	customRootFields := &meta.Args.Configuration.CustomRootFields

	customRootFields.Select = tablePlural
	customRootFields.SelectByPk = tableSingular
	customRootFields.SelectAggregate = fmt.Sprintf("%s_aggregate", tablePlural)
	customRootFields.Insert = fmt.Sprintf("insert_%s", tablePlural)
	customRootFields.InsertOne = fmt.Sprintf("insert_%s", tableSingular)
	customRootFields.Update = fmt.Sprintf("update_%s", tablePlural)
	customRootFields.UpdateByPk = fmt.Sprintf("update_%s", tableSingular)
	customRootFields.Delete = fmt.Sprintf("delete_%s", tablePlural)
	customRootFields.DeleteByPk = fmt.Sprintf("delete_%s", tableSingular)

	return s.PostTrackTableMetadata(meta)
}

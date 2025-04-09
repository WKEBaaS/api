package hasura

type HasuraCreateFunctionPermissionMetadata struct {
	Type string `json:"type"`
	Args struct {
		Source   string `json:"source"`
		Function struct {
			Schema string `json:"schema"`
			Name   string `json:"name"`
		} `json:"function"`
		Role    string  `json:"role"`
		Comment *string `json:"comment"`
	} `json:"args"`
}

type HasuraCreatePermissionMetadata struct {
	Type string `json:"type"`
	Args struct {
		Source string `json:"source"`
		Table  struct {
			Schema string `json:"schema"`
			Name   string `json:"name"`
		} `json:"table"`
		Role       string         `json:"role"`
		Permission map[string]any `json:"permission"`
		Comment    *string        `json:"comment"`
	} `json:"args"`
}

func (s *HasuraService) CreateFunctionPermission(schema string, functionName string, role string, comment *string) error {
	meta := &HasuraCreateFunctionPermissionMetadata{}

	meta.Type = "pg_create_function_permission"
	meta.Args.Source = s.config.Hasura.Source

	meta.Args.Function.Schema = schema
	meta.Args.Function.Name = functionName
	meta.Args.Role = role
	meta.Args.Comment = comment

	if err := s.PostMetadata(meta); err != nil {
		return err
	}

	return nil
}

func (s *HasuraService) CreatePermission(metaType string, schema string, tableName string, role string, permission map[string]any, comment *string) error {
	meta := &HasuraCreatePermissionMetadata{}

	meta.Type = metaType
	meta.Args.Source = s.config.Hasura.Source

	meta.Args.Table.Schema = schema
	meta.Args.Table.Name = tableName
	meta.Args.Role = role
	meta.Args.Permission = permission
	meta.Args.Comment = comment

	if err := s.PostMetadata(meta); err != nil {
		return err
	}

	return nil
}

func (s *HasuraService) CreateSelectWithoutCheckPermission(schema string, tableName string, role string, comment *string) error {
	permission := map[string]any{
		"columns": "*",
		"filter":  map[string]any{},
	}

	return s.CreatePermission("pg_create_select_permission", schema, tableName, role, permission, comment)
}

package hasura

import "fmt"

type HasuraCreateRelationshipMetadata struct {
	Type string `json:"type"`
	Args struct {
		Source string `json:"source"`
		Table  struct {
			Schema string `json:"schema"`
			Name   string `json:"name"`
		} `json:"table"`
		Name  string `json:"name"`
		Using struct {
			ForeignKeyConstraintOn string `json:"foreign_key_constraint_on"`
		} `json:"using"`
		Comment *string `json:"comment"`
	} `json:"args"`
}

type HasuraCreateArrayRelationshipMetadata struct {
	Type string `json:"type"`
	Args struct {
		Source string `json:"source"`
		Table  struct {
			Schema string `json:"schema"`
			Name   string `json:"name"`
		} `json:"table"`
		Name  string `json:"name"`
		Using struct {
			ForeignKeyConstraintOn struct {
				Table struct {
					Schema string `json:"schema"`
					Name   string `json:"name"`
				} `json:"table"`
				Columns []string `json:"columns"`
			} `json:"foreign_key_constraint_on"`
		} `json:"using"`
		Comment *string `json:"comment"`
	} `json:"args"`
}

func (s *HasuraService) CreateRelationship(schema string, tableName string, relationshipName string, foreignKeyConstraintOn string, comment *string) error {
	meta := &HasuraCreateRelationshipMetadata{}

	meta.Type = "pg_create_object_relationship"
	meta.Args.Source = s.config.Hasura.Source

	meta.Args.Table.Schema = schema
	meta.Args.Table.Name = tableName
	meta.Args.Name = relationshipName
	meta.Args.Using.ForeignKeyConstraintOn = foreignKeyConstraintOn
	meta.Args.Comment = comment

	if err := s.PostMetadata(meta); err != nil {
		return fmt.Errorf("failed to create relationship: %w", err)
	}

	return nil
}

func (s *HasuraService) CreateArrayRelationship(schema string, tableName string, relationshipName string, foreignKeyConstraintOnTableSchema string, foreignKeyConstraintOnTableName string, foreignKeyConstraintOnColumns []string, comment *string) error {
	meta := &HasuraCreateArrayRelationshipMetadata{}

	meta.Type = "pg_create_array_relationship"
	meta.Args.Source = s.config.Hasura.Source

	meta.Args.Table.Schema = schema
	meta.Args.Table.Name = tableName
	meta.Args.Name = relationshipName
	meta.Args.Using.ForeignKeyConstraintOn.Table.Schema = foreignKeyConstraintOnTableSchema
	meta.Args.Using.ForeignKeyConstraintOn.Table.Name = foreignKeyConstraintOnTableName
	meta.Args.Using.ForeignKeyConstraintOn.Columns = foreignKeyConstraintOnColumns
	meta.Args.Comment = comment

	if err := s.PostMetadata(meta); err != nil {
		return fmt.Errorf("failed to create array relationship: %w", err)
	}

	return nil
}

package hasura

import (
	"fmt"

	"github.com/samber/lo"
)

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
				Select          *string `json:"select"`
				SelectByPk      *string `json:"select_by_pk"`     //nolint: tagliatelle
				SelectAggregate *string `json:"select_aggregate"` //nolint: tagliatelle
				Insert          *string `json:"insert"`
				InsertOne       *string `json:"insert_one"` //nolint: tagliatelle
				Update          *string `json:"update"`
				UpdateByPk      *string `json:"update_by_pk"` //nolint: tagliatelle
				Delete          *string `json:"delete"`
				DeleteByPk      *string `json:"delete_by_pk"` //nolint: tagliatelle
			} `json:"custom_root_fields"`
		} `json:"configuration"`
	} `json:"args"`
}

// TrackTable tracks a table in Hasura with the given schema and table names.
//
// Parameters:
// - schema: The schema name where the table resides.
// - tableName: The plural form of the table name, used for multiple record operations.
// - tableSingularName: The singular form of the table name, used for single record operations.
//
// The function requires both singular and plural forms of the table name to correctly set up
// custom root fields in Hasura. These fields are used to generate GraphQL queries and mutations
// that follow a consistent naming convention, making it easier to interact with the API.
func (s *HasuraService) TrackTable(schema string, tableName string, tableSingularName *string) error {
	meta := &HasuraTrackTableMetadata{}

	meta.Type = "pg_track_table"
	meta.Args.Source = s.config.Hasura.Source

	meta.Args.Table.Schema = schema
	meta.Args.Table.Name = tableName
	meta.Args.Configuration.CustomName = tableName

	customRootFields := &meta.Args.Configuration.CustomRootFields

	customRootFields.Select = lo.ToPtr(tableName)
	customRootFields.SelectByPk = tableSingularName
	customRootFields.SelectAggregate = lo.ToPtr(fmt.Sprintf("%s_aggregate", tableName))
	customRootFields.Insert = lo.ToPtr(fmt.Sprintf("insert_%s", tableName))
	customRootFields.Update = lo.ToPtr(fmt.Sprintf("update_%s", tableName))
	customRootFields.Delete = lo.ToPtr(fmt.Sprintf("delete_%s", tableName))

	if tableSingularName != nil {
		customRootFields.InsertOne = lo.ToPtr(fmt.Sprintf("insert_%s", *tableSingularName))
		customRootFields.UpdateByPk = lo.ToPtr(fmt.Sprintf("update_%s", *tableSingularName))
		customRootFields.DeleteByPk = lo.ToPtr(fmt.Sprintf("delete_%s", *tableSingularName))
	}

	return s.PostMetadata(meta)
}

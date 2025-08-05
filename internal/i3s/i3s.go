package i3s

import (
	"baas-api/config"
)

type I3S struct {
	config *config.Config
	// service *services.Service
}

func NewI3S(config *config.Config) *I3S {
	return &I3S{
		config: config,
		// service: service,
	}
}

// func (i3s *I3S) PostMetadata() error {
// 	tables := []struct {
// 		Schema       string
// 		Name         string
// 		SingularName *string
// 	}{
// 		{"auth", "users", lo.ToPtr("user")},
// 		{"auth", "identities", lo.ToPtr("identity")},
// 		{"auth", "sessions", lo.ToPtr("session")},
// 		{"auth", "audit_log_entries", lo.ToPtr("audit_log_entry")},
// 		{"auth", "roles", lo.ToPtr("role")},
// 		{"auth", "user_roles", lo.ToPtr("user_role")},
// 		{"dbo", "classes", lo.ToPtr("class")},
// 		{"dbo", "permissions", lo.ToPtr("permission")},
// 		{"dbo", "permission_enum", nil},
// 		{"dbo", "objects", lo.ToPtr("object")},
// 		{"dbo", "inheritances", lo.ToPtr("inheritance")},
// 		{"dbo", "co", nil},
// 		{"api", "check_class_permission_result", nil},
// 		{"api", "class_result", nil},
// 	}
//
// 	for _, table := range tables {
// 		// if err := i3s.service.Hasura.TrackTable(table.Schema, table.Name, table.SingularName); err != nil {
// 		// slog.Error("failed to track table", "schema", table.Schema, "name", table.Name, "error", err)
// 		// }
// 	}
//
// 	relationships := []struct {
// 		Schema                 string
// 		TableName              string
// 		RelationshipName       string
// 		ForeignKeyConstraintOn string
// 		Comment                *string
// 	}{
// 		{"auth", "user_roles", "role", "role_id", nil},
// 		{"dbo", "classes", "owner", "owner_id", nil},
// 		{"dbo", "co", "object", "oid", nil},
// 		{"dbo", "permissions", "class", "class_id", nil},
// 	}
//
// 	for _, r := range relationships {
// 		// if err := i3s.service.Hasura.CreateRelationship(r.Schema, r.TableName, r.RelationshipName, r.ForeignKeyConstraintOn, r.Comment); err != nil {
// 		// slog.Error("failed to create relationship", "schema", r.Schema, "table", r.TableName, "relationship", r.RelationshipName, "error", err)
// 		// }
// 	}
//
// 	arrayRelationships := []struct {
// 		Schema                string
// 		TableName             string
// 		RelationshipName      string
// 		ForeignKeyTableSchema string
// 		ForeignKeyTableName   string
// 		ForeignKeyColumns     []string
// 		Comment               *string
// 	}{
// 		{"dbo", "classes", "children", "dbo", "inheritances", []string{"pcid"}, nil},
// 		{"dbo", "classes", "parent", "dbo", "inheritances", []string{"ccid"}, nil},
// 		{"dbo", "classes", "co", "dbo", "co", []string{"cid"}, nil},
// 		{"dbo", "classes", "permissions", "dbo", "permissions", []string{"class_id"}, nil},
// 	}
//
// 	for _, r := range arrayRelationships {
// 		// if err := i3s.service.Hasura.CreateArrayRelationship(r.Schema, r.TableName, r.RelationshipName, r.ForeignKeyTableSchema, r.ForeignKeyTableName, r.ForeignKeyColumns, r.Comment); err != nil {
// 		// 	slog.Error("failed to create array relationship", "schema", r.Schema, "table", r.TableName, "relationship", r.RelationshipName, "error", err)
// 		// }
// 	}
//
// 	functions := []struct {
// 		Schema       string
// 		FunctionName string
// 		SessionArg   *string
// 		ExposedAs    *string
// 		Comment      *string
// 	}{
// 		{"api", "check_class_permission", lo.ToPtr("hasura_session"), nil, lo.ToPtr("Check user's class permission")},
// 		{"api", "insert_class", lo.ToPtr("hasura_session"), lo.ToPtr("mutation"), lo.ToPtr("Insert a new class")},
// 		{"api", "delete_class", lo.ToPtr("hasura_session"), lo.ToPtr("mutation"), lo.ToPtr("Delete the class")},
// 	}
//
// 	for _, f := range functions {
// 		// if err := i3s.service.Hasura.TrackFunction(f.Schema, f.FunctionName, f.SessionArg, f.ExposedAs, f.Comment); err != nil {
// 		// 	slog.Error("failed to track function", "schema", f.Schema, "function", f.FunctionName, "error", err)
// 		// }
// 	}
//
// 	// Permissions
// 	selectPermissions := []struct {
// 		Schema    string
// 		TableName string
// 		Role      string
// 		Comment   *string
// 	}{
// 		{"api", "check_class_permission_result", "user", nil},
// 		{"api", "class_result", "user", nil},
// 	}
//
// 	for _, p := range selectPermissions {
// 		// if err := i3s.service.Hasura.CreateSelectWithoutCheckPermission(p.Schema, p.TableName, p.Role, p.Comment); err != nil {
// 		// 	slog.Error("failed to create select permission", "schema", p.Schema, "table", p.TableName, "role", p.Role, "error", err)
// 		// }
// 	}
//
// 	functionPermissions := []struct {
// 		Schema       string
// 		FunctionName string
// 		Role         string
// 		Comment      *string
// 	}{
// 		{"api", "check_class_permission", "user", nil},
// 		{"api", "insert_class", "user", nil},
// 		{"api", "delete_class", "user", nil},
// 	}
//
// 	for _, p := range functionPermissions {
// 		// if err := i3s.service.Hasura.CreateFunctionPermission(p.Schema, p.FunctionName, p.Role, p.Comment); err != nil {
// 		// 	slog.Error("failed to create function permission", "schema", p.Schema, "function", p.FunctionName, "role", p.Role, "error", err)
// 		// }
// 	}
//
// 	return nil
// }

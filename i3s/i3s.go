package i3s

import (
	"i3s-service/internal/configs"
	"i3s-service/internal/services"

	"github.com/samber/lo"
)

type I3S struct {
	config  *configs.Config
	service *services.Service
}

func InitI3S(config *configs.Config, service *services.Service) *I3S {
	return &I3S{
		config:  config,
		service: service,
	}
}

func (i3s *I3S) PostMetadata() error {
	var tables = []struct {
		Schema       string
		Name         string
		SingularName *string
	}{
		{"auth", "users", lo.ToPtr("user")},
		{"auth", "identities", lo.ToPtr("identity")},
		{"auth", "sessions", lo.ToPtr("session")},
		{"auth", "audit_log_entries", lo.ToPtr("audit_log_entry")},
		{"auth", "roles", lo.ToPtr("role")},
		{"auth", "user_roles", lo.ToPtr("user_role")},
		{"dbo", "classes", lo.ToPtr("class")},
		{"dbo", "objects", lo.ToPtr("object")},
		{"dbo", "inheritances", lo.ToPtr("inheritance")},
		{"dbo", "co", nil},
		{"api", "class_permission_enum", nil},
		{"api", "check_class_permission_result", nil},
		// {"dbo", "text_result", nil},
	}

	for _, table := range tables {
		if err := i3s.service.Hasura.TrackTable(table.Schema, table.Name, table.SingularName); err != nil {
			return err
		}
	}

	var relationships = []struct {
		Schema                 string
		TableName              string
		RelationshipName       string
		ForeignKeyConstraintOn string
		Comment                *string
	}{
		{"auth", "user_roles", "role", "role_id", nil},
		{"dbo", "classes", "owner", "owner_id", nil},
		{"dbo", "co", "object", "oid", nil},
	}

	for _, r := range relationships {
		if err := i3s.service.Hasura.CreateRelationship(r.Schema, r.TableName, r.RelationshipName, r.ForeignKeyConstraintOn, r.Comment); err != nil {
			return err
		}
	}

	var arrayRelationships = []struct {
		Schema                string
		TableName             string
		RelationshipName      string
		ForeignKeyTableSchema string
		ForeignKeyTableName   string
		ForeignKeyColumns     []string
		Comment               *string
	}{
		// {"auth", "users", "roles", "auth", "user_roles", []string{"user_id"}, nil},
		// {"auth", "roles", "users", "auth", "users", []string{"role_id"}, nil},
		{"dbo", "classes", "children", "dbo", "inheritances", []string{"pcid"}, nil},
		{"dbo", "classes", "parent", "dbo", "inheritances", []string{"ccid"}, nil},
		{"dbo", "classes", "co", "dbo", "co", []string{"cid"}, nil},
	}

	for _, r := range arrayRelationships {
		if err := i3s.service.Hasura.CreateArrayRelationship(r.Schema, r.TableName, r.RelationshipName, r.ForeignKeyTableSchema, r.ForeignKeyTableName, r.ForeignKeyColumns, r.Comment); err != nil {
			return err
		}
	}

	var functions = []struct {
		Schema       string
		FunctionName string
		SessionArg   *string
		ExposedAs    *string
		Comment      *string
	}{
		// {"dbo", "fn_insert_class", "hasura_session", lo.ToPtr("mutation"), lo.ToPtr("Insert a new class")},
		// {"dbo", "get_session_role", "hasura_session", nil, lo.ToPtr("Get session role")},
		{"api", "check_class_permission", lo.ToPtr("hasura_session"), nil, lo.ToPtr("Check user's class permission")},
	}

	for _, f := range functions {
		if err := i3s.service.Hasura.TrackFunction(f.Schema, f.FunctionName, f.SessionArg, f.ExposedAs, f.Comment); err != nil {
			return err
		}
	}

	return nil
}

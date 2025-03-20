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
		{"dbo", "classes", lo.ToPtr("class")},
		{"dbo", "objects", lo.ToPtr("object")},
		{"dbo", "inheritances", lo.ToPtr("inheritance")},
		{"dbo", "co", nil},
	}

	for _, table := range tables {
		if err := i3s.service.Hasura.PostTrackTableMetadataWithTableName(table.Schema, table.Name, table.SingularName); err != nil {
			return err
		}
	}

	return nil
}

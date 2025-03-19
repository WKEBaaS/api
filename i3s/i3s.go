package i3s

import (
	"fmt"
	"i3s-service/internal/configs"
	"i3s-service/internal/services"
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
	if err := i3s.service.Hasura.PostTrackTableMetadataWithTableName("auth", "user", "users"); err != nil {
		return fmt.Errorf("failed to post auth.users metadata: %w", err)
	}

	if err := i3s.service.Hasura.PostTrackTableMetadataWithTableName("dbo", "class", "classes"); err != nil {
		return fmt.Errorf("failed to post auth.roles metadata: %w", err)
	}
	return nil
}

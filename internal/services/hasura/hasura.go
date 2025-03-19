package hasura

import "i3s-service/internal/configs"

type HasuraService struct {
	config *configs.Config
}

func InitHasuraService(config *configs.Config) *HasuraService {
	service := &HasuraService{}
	service.config = config

	return service
}

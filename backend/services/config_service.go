package services

import (
	"developer-toolbox/backend/models"
	"developer-toolbox/backend/storage"
)

type ConfigService struct {
	repository *storage.ConfigRepository
}

func NewConfigService(repository *storage.ConfigRepository) *ConfigService {
	return &ConfigService{repository: repository}
}

func (s *ConfigService) Load() (models.Settings, error) {
	return s.repository.Load()
}

func (s *ConfigService) Save(settings models.Settings) error {
	return s.repository.Save(settings)
}

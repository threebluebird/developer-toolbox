package storage

import (
	"developer-toolbox/backend/models"
	"encoding/json"
	"errors"
	"os"
)

type ConfigRepository struct {
	storage Storage
	key     string
}

func NewConfigRepository(storage Storage) *ConfigRepository {
	return &ConfigRepository{storage: storage, key: "settings.json"}
}

func (r *ConfigRepository) Load() (models.Settings, error) {
	data, err := r.storage.Get(r.key)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return models.Settings{}, nil
		}
		return models.Settings{}, err
	}

	var settings models.Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return models.Settings{}, err
	}
	return settings, nil
}

func (r *ConfigRepository) Save(settings models.Settings) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return r.storage.Set(r.key, data)
}

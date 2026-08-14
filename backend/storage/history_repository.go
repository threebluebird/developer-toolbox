package storage

import (
	"developer-toolbox/backend/models"
	"encoding/json"
	"errors"
	"os"
)

type HistoryRepository struct {
	storage Storage
	key     string
	limit   int
}

func NewHistoryRepository(storage Storage) *HistoryRepository {
	return &HistoryRepository{storage: storage, key: "history.json", limit: 50}
}

func (r *HistoryRepository) List() ([]models.History, error) {
	data, err := r.storage.Get(r.key)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var entries []models.History
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (r *HistoryRepository) Save(entries []models.History) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return r.storage.Set(r.key, data)
}

func (r *HistoryRepository) Add(entry models.History) error {
	entries, err := r.List()
	if err != nil {
		return err
	}

	// 新记录放在最前面，前端可直接展示最近使用；同时限制最多 50 条。
	entries = append([]models.History{entry}, entries...)
	if len(entries) > r.limit {
		entries = entries[:r.limit]
	}

	return r.Save(entries)
}

func (r *HistoryRepository) Clear() error { return r.Save([]models.History{}) }

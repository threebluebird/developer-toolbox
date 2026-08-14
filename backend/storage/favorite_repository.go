package storage

import (
	"developer-toolbox/backend/models"
	"encoding/json"
	"errors"
	"os"
)

type FavoriteRepository struct {
	storage Storage
	key     string
}

func NewFavoriteRepository(storage Storage) *FavoriteRepository {
	return &FavoriteRepository{storage: storage, key: "favorites.json"}
}

func (r *FavoriteRepository) List() ([]models.Favorite, error) {
	data, err := r.storage.Get(r.key)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var favorites []models.Favorite
	if err := json.Unmarshal(data, &favorites); err != nil {
		return nil, err
	}
	return favorites, nil
}

func (r *FavoriteRepository) Save(favorites []models.Favorite) error {
	data, err := json.MarshalIndent(favorites, "", "  ")
	if err != nil {
		return err
	}
	return r.storage.Set(r.key, data)
}

func (r *FavoriteRepository) Add(toolID string) error {
	return r.AddItem(models.Favorite{ToolID: toolID, Kind: "tool"})
}

func (r *FavoriteRepository) AddItem(item models.Favorite) error {
	favorites, err := r.List()
	if err != nil {
		return err
	}

	for _, existing := range favorites {
		// 收藏操作保持幂等，重复添加同一工具不会生成重复记录。
		if item.ID != "" && existing.ID == item.ID {
			return nil
		}
		if item.Kind == "tool" && existing.Kind == "tool" && existing.ToolID == item.ToolID {
			return nil
		}
	}

	favorites = append(favorites, item)
	return r.Save(favorites)
}

func (r *FavoriteRepository) Remove(toolID string) error {
	favorites, err := r.List()
	if err != nil {
		return err
	}

	filtered := make([]models.Favorite, 0, len(favorites))
	for _, item := range favorites {
		// Removing a tool favorite must not remove saved inputs/requests for it.
		if item.ToolID != toolID || (item.Kind != "" && item.Kind != "tool") {
			filtered = append(filtered, item)
		}
	}

	return r.Save(filtered)
}

func (r *FavoriteRepository) RemoveItem(id string) error {
	favorites, err := r.List()
	if err != nil {
		return err
	}
	filtered := make([]models.Favorite, 0, len(favorites))
	for _, item := range favorites {
		if item.ID != id {
			filtered = append(filtered, item)
		}
	}
	return r.Save(filtered)
}

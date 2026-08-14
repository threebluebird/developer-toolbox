package services

import (
	"developer-toolbox/backend/models"
	"developer-toolbox/backend/storage"
)

type FavoriteService struct {
	repository *storage.FavoriteRepository
}

func NewFavoriteService(repository *storage.FavoriteRepository) *FavoriteService {
	return &FavoriteService{repository: repository}
}

func (s *FavoriteService) List() ([]models.Favorite, error) {
	return s.repository.List()
}

func (s *FavoriteService) Add(toolID string) error {
	return s.repository.Add(toolID)
}

func (s *FavoriteService) Remove(toolID string) error {
	return s.repository.Remove(toolID)
}

func (s *FavoriteService) AddItem(item models.Favorite) error { return s.repository.AddItem(item) }
func (s *FavoriteService) RemoveItem(id string) error         { return s.repository.RemoveItem(id) }

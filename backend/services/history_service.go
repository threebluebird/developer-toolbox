package services

import (
	"developer-toolbox/backend/models"
	"developer-toolbox/backend/storage"
)

type HistoryService struct {
	repository *storage.HistoryRepository
}

func NewHistoryService(repository *storage.HistoryRepository) *HistoryService {
	return &HistoryService{repository: repository}
}

func (s *HistoryService) List() ([]models.History, error) {
	return s.repository.List()
}

func (s *HistoryService) Add(entry models.History) error {
	return s.repository.Add(entry)
}

func (s *HistoryService) Clear() error { return s.repository.Clear() }

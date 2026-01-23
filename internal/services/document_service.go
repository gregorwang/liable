package services

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
)

type DocumentService struct {
	repo *repository.DocumentRepository
}

func NewDocumentService() *DocumentService {
	return &DocumentService{
		repo: repository.NewDocumentRepository(),
	}
}

func (s *DocumentService) ListDocuments() ([]models.SystemDocument, error) {
	return s.repo.ListDocuments()
}

func (s *DocumentService) UpdateDocument(key, content string, userID int) (*models.SystemDocument, error) {
	return s.repo.UpdateDocument(key, content, userID)
}

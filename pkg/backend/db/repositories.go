package db

import (
	"go.vxn.dev/littr/pkg/models"
)

type PollRepositoryInterface interface {
	GetAll() ([]*models.Poll, error)
	GetByID(pollID string) (*models.Poll, error)
	Save(poll *models.Poll) error
	Update(poll *models.Poll) error
	Delete(pollID string) error
}

type PostRepositoryInterface interface {
	GetAll() ([]*models.Post, error)
	GetByID(pollID string) (*models.Post, error)
	Save(poll *models.Post) error
	Update(poll *models.Post) error
	Delete(pollID string) error
}

//
//  Repositories
//

type Repositories struct {
	PollRepository PollRepositoryInterface
	PostRepository PostRepositoryInterface
}

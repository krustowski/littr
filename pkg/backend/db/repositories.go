package db

import (
	"go.vxn.dev/littr/pkg/models"
)

//
//  Repository interfaces
//

type PollRepositoryInterface interface {
	GetAll(interface{}) (*map[string]models.Poll, error)
	GetByID(pollID string) (*models.Poll, error)
	Save(poll *models.Poll) error
	//Update(poll *models.Poll) error
	Delete(pollID string) error
}

type PostRepositoryInterface interface {
	GetAll(interface{}) (*map[string]models.Post, error)
	GetByID(postID string) (*models.Post, error)
	Save(post *models.Post) error
	//Update(post *models.Post) error
	Delete(postID string) error
}

type UserRepositoryInterface interface {
	GetAll(interface{}) (*map[string]models.User, error)
	GetByID(userID string) (*models.User, error)
	Save(user *models.User) error
	//Update(user *models.User) error
	Delete(userID string) error
}

//
//  Repositories
//

type Repositories struct {
	PollRepository PollRepositoryInterface
	PostRepository PostRepositoryInterface
	UserRepository UserRepositoryInterface
}

var Storage *Repositories

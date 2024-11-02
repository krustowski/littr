package models

//
//  Repository interfaces
//

type PollRepositoryInterface interface {
	GetAll(interface{}) (*map[string]Poll, error)
	GetByID(pollID string) (*Poll, error)
	Save(poll *Poll) error
	Delete(pollID string) error
}

type PostRepositoryInterface interface {
	GetAll(interface{}) (*map[string]Post, error)
	GetByID(postID string) (*Post, error)
	Save(post *Post) error
	Delete(postID string) error
}

type UserRepositoryInterface interface {
	GetAll(interface{}) (*map[string]User, error)
	GetByID(userID string) (*User, error)
	Save(user *User) error
	Delete(userID string) error
}

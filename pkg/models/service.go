package models

import (
	"context"
)

//
//  Service interfaces
//

type AuthServiceInterface interface {
	Auth(ctx context.Context, user interface{}) (*User, []string, error)
	Logout(ctx context.Context) error
}

type PollServiceInterface interface {
	Create(ctx context.Context, poll *Poll) error
	Update(ctx context.Context, poll *Poll) error
	Delete(ctx context.Context, pollID string) error
	FindAll(ctx context.Context) (*map[string]Poll, *User, error)
	FindByID(ctx context.Context, pollID string) (*Poll, *User, error)
}

type StatServiceInterface interface {
	Calculate(ctx context.Context) (*map[string]int64, *map[string]UserStat, *map[string]User, error)
}

type TokenServiceInterface interface {
	Create(ctx context.Context, user *User) ([]string, error)
	Delete(ctx context.Context, tokenID string) error
	FindByID(ctx context.Context, tokenID string) (*Token, error)
}

type UserServiceInterface interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, request interface{}) error
	Delete(ctx context.Context, userID string) error
	FindAll(ctx context.Context) (*map[string]User, error)
	FindByID(ctx context.Context, userID string) (*User, error)
}

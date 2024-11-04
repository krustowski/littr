package models

import (
	"context"
)

//
//  Service interfaces
//

type PollServiceInterface interface {
	Create(ctx context.Context, poll *Poll) error
	Update(ctx context.Context, poll *Poll) error
	Delete(ctx context.Context, pollID string) error
	FindAll(ctx context.Context) (*map[string]Poll, *User, error)
	FindByID(ctx context.Context, pollID string) (*Poll, *User, error)
}

type UserServiceInterface interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, request interface{}) error
	Delete(ctx context.Context, userID string) error
	FindAll(ctx context.Context) (*map[string]User, error)
	FindByID(ctx context.Context, userID string) (*User, error)
}

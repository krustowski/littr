package models

import (
	"context"

	gomail "github.com/wneessen/go-mail"
)

//
//  Service interfaces
//

type AuthServiceInterface interface {
	Auth(ctx context.Context, user interface{}) (*User, []string, error)
	Logout(ctx context.Context) error
}

type MailServiceInterface interface {
	ComposeMail(payload interface{}) (*gomail.Msg, error)
	SendMail(msg *gomail.Msg) error
}

type NotificationServiceInterface interface {
	SendNotification(ctx context.Context, postID string) error
}

type PollServiceInterface interface {
	Create(ctx context.Context, createRequest interface{}) error
	Update(ctx context.Context, updateRequest interface{}) error
	Delete(ctx context.Context, pollID string) error
	FindAll(ctx context.Context) (*map[string]Poll, *User, error)
	FindByID(ctx context.Context, pollID string) (*Poll, *User, error)
}

type PostServiceInterface interface {
	Create(ctx context.Context, post *Post) error
	Update(ctx context.Context, post *Post) error
	Delete(ctx context.Context, postID string) error
	FindAll(ctx context.Context) (*map[string]Post, *User, error)
	//FindPage(ctx context.Context, opts interface{}) (*map[string]Post, *map[string]User, error)
	FindByID(ctx context.Context, postID string) (*Post, *User, error)
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
	Create(ctx context.Context, createRequest interface{}) error
	Subscribe(ctx context.Context, device *Device) error
	Unsubscribe(ctx context.Context, uuid string) error
	Activate(ctx context.Context, userID string) error
	Update(ctx context.Context, updateRequest interface{}) error
	UpdateAvatar(ctx context.Context, updateRequest interface{}) (*string, error)
	UpdateSubscriptionTags(ctx context.Context, uuid string, tags []string) error
	ProcessPassphraseRequest(ctx context.Context, updateRequest interface{}) error
	Delete(ctx context.Context, userID string) error
	FindAll(ctx context.Context) (*map[string]User, error)
	FindByID(ctx context.Context, userID string) (*User, error)
	FindPostsByID(ctx context.Context, userID string) (*map[string]Post, *map[string]User, error)
}

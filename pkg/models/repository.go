package models

//
//  Repository interfaces
//

type PollRepositoryInterface interface {
	GetAll() (*map[string]Poll, error)
	GetPage(opts interface{}) (*map[string]Poll, error)
	GetByID(pollID string) (*Poll, error)
	Save(poll *Poll) error
	Delete(pollID string) error
}

type PostRepositoryInterface interface {
	GetAll() (*map[string]Post, error)
	GetPage(opts interface{}) (*map[string]Post, *map[string]User, error)
	GetByID(postID string) (*Post, error)
	Save(post *Post) error
	Delete(postID string) error
}

type RequestRepositoryInterface interface {
	GetByID(reqID string) (*Request, error)
	Save(req *Request) error
	Delete(reqID string) error
}

type SubscriptionRepositoryInterface interface {
	//GetAll() *map[string][]Device
	GetByUserID(userID string) (*[]Device, error)
	Save(userID string, sub *[]Device) error
	Delete(userID string) error
}

type TokenRepositoryInterface interface {
	GetAll() (*map[string]Token, error)
	GetByID(tokenID string) (*Token, error)
	Save(token *Token) error
	Delete(tokenID string) error
}

type UserRepositoryInterface interface {
	GetAll() (*map[string]User, error)
	GetPage(opts interface{}) (*map[string]User, error)
	GetByID(userID string) (*User, error)
	Save(user *User) error
	Delete(userID string) error
}

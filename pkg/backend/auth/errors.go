package auth

import "errors"

var (
	errAuthFailed    = errors.New("wrong credentials entered, or such user does not exist")
	errInvalidInput  = errors.New("ivalid input: cannot assert type AuthUser")
	errNotActivated  = errors.New("user has not been activated yet, check your mail inbox")
	errTokenDeletion = errors.New("could not delete associated token")
)

var (
	msgSessionTerminated = "session terminated, void cookies provided"
)

package users

import "errors"

var ErrUserRequestDecodingFailed = errors.New("could not decode the user request payload")

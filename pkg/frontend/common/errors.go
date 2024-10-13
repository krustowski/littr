package common

const (
	// toast types
	TTYPE_ERR = "error"
	TTYPE_INFO = "info"
	TTYPE_SUCCESS = "success"
)

const (
	// generic error messages on FE
	ERR_CANNOT_REACH_BE = "cannot reach the server"
	ERR_CANNOT_GET_DATA = "cannot get the data"

	// flow/post-related error messages
	ERR_INVALID_REPLY = "no valid reply content entered"
	ERR_POST_UNAUTH_DELETE = "you only can delete your own posts!"
	ERR_POST_NOT_FOUND = "post not found"
	ERR_USER_NOT_FOUND = "user not found"
	ERR_PRIVATE_ACC = "this account is private"
)

package common

import (
	"net/http"
)

// Common helper function to decide the HTTP error according to the error contents.
var DecideStatusFromError = func(err error) int {
	// HTTP 200 condition.
	if err == nil {
		return http.StatusOK
	}

	// HTTP 400 conditions
	if err.Error() == ERR_USER_NOT_ACTIVATED ||
		err.Error() == ERR_PASSPHRASE_REQ_INCOMPLETE ||
		err.Error() == ERR_REQUEST_UUID_EXPIRED ||
		err.Error() == ERR_REQUEST_UUID_BLANK ||
		err.Error() == ERR_REQUEST_UUID_INVALID ||
		err.Error() == ERR_RESTRICTED_NICKNAME ||
		err.Error() == ERR_USER_NICKNAME_TAKEN ||
		err.Error() == ERR_NICKNAME_CHARSET_MISMATCH ||
		err.Error() == ERR_NICKNAME_TOO_LONG_SHORT ||
		err.Error() == ERR_WRONG_EMAIL_FORMAT ||
		err.Error() == ERR_INPUT_DATA_FAIL ||
		err.Error() == ERR_IMG_UNKNOWN_TYPE {
		return http.StatusBadRequest
	}

	// HTTP 403 conditions.
	if err.Error() == ERR_POLL_SELF_VOTE ||
		err.Error() == ERR_USER_DELETE_FOREIGN ||
		err.Error() == ERR_USER_PASSPHRASE_FOREIGN ||
		err.Error() == ERR_REGISTRATION_DISABLED ||
		err.Error() == ERR_POLL_EXISTING_VOTE ||
		err.Error() == ERR_POLL_INVALID_VOTE_COUNT {
		return http.StatusForbidden
	}

	// HTTP 404 condition.
	if err.Error() == ERR_POLL_NOT_FOUND ||
		err.Error() == ERR_USER_NOT_FOUND {
		return http.StatusNotFound
	}

	// HTTP 409 condition
	if err.Error() == ERR_EMAIL_ALREADY_USED ||
		err.Error() == ERR_PASSPHRASE_CURRENT_WRONG {
		return http.StatusConflict
	}

	// HTTP 500 as default.
	return http.StatusInternalServerError
}

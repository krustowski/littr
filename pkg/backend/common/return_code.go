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
	if err.Error() == ERR_USER_NOT_ACTIVATED {
		return http.StatusBadRequest
	}

	// HTTP 403 conditions.
	if err.Error() == ERR_POLL_SELF_VOTE ||
		err.Error() == ERR_POLL_EXISTING_VOTE ||
		err.Error() == ERR_POLL_INVALID_VOTE_COUNT {
		return http.StatusForbidden
	}

	// HTTP 404 condition.
	if err.Error() == ERR_POLL_NOT_FOUND {
		return http.StatusNotFound
	}

	// HTTP 500 as default.
	return http.StatusInternalServerError
}

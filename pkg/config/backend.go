package config

import (
	"os"
	"strconv"
)

const (
	DEFAULT_PORT = "8054"

	// Time interval after that a heartbeat event of type 'message' is to be sent to connected clients/subscribers.
	HEARTBEAT_SLEEP_TIME = 20
	SERVER_PORT          = "APP_PORT"
)

/*
 *  Registration
 */

var REGISTRATION_ENABLED bool = func() bool {
	if os.Getenv("REGISTRATION_ENABLED") != "" {
		boolVal, err := strconv.ParseBool(os.Getenv("REGISTRATION_ENABLED"))
		if err != nil {
			return false
		}
		return boolVal
	} else {
		return true
	}
}()

/*
 *  App environment
 */

var APP_ENVIRONMENT string = func() string {
	if os.Getenv("APP_ENVIRONMENT") != "" {
		return os.Getenv("APP_ENVIRONMENT")
	} else {
		return "dev"
	}
}()

var ServerPort = func() string {
	if os.Getenv(SERVER_PORT) != "" {
		return os.Getenv(SERVER_PORT)
	}

	return DEFAULT_PORT
}()

/*
 *  BE data migrations
 */

// Accounts to be ceased from the database inc. their posts.
// You would have to restart backend server for this to apply if you made changes there.
var UserDeletionList []string = []string{
	"admin",
	"administrator",
	"superuser",
	"moderator",
	"passphrase",
	"user",
	"nickname",
	"test",
	"tester",
	"littr",
	"voter",
}

// This array is used in a procedure's loop to manually unshade listed users.
// Thus listed accounts should have a zero (0) on stats page.
var UsersToUnshade []string = []string{}

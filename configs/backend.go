package configs

import (
	"os"
)

const (
	// Time interval after that a heartbeat event of type 'message' is to be sent to connected clients/subscribers.
	HEARTBEAT_SLEEP_TIME  = 20
	//REGISTERATION_ENABLED = false
)

/*
 *  Registeration
 */

var REGISTERATION_ENABLED bool = false

if os.Getenv("REGISTERATION_ENABLED") != "" {
	REGISTERATION_ENABLED = os.Getenv("REGISTERATION_ENABLED")
}

/*
 *  App environment
 */

var APP_ENVIRONMENT string = "dev"

if os.Getenv("APP_ENVIRONMENT") != "" {
	APP_ENVIRONMENT = os.Getenv("APP_ENVIRONMENT")
}

/*
 *  BE data migrations
 */

// Accounts to be ceased from the database inc. their posts.
// You would have to restart backend server for this to apply if you made changes there.
var UserDeletionList []string = []string{
	"admin",
}

// This array is used in a procedure's loop to manually unshade listed users.
// Thus listed accounts should have a zero (0) on stats page.
var UsersToUnshade []string = []string{}

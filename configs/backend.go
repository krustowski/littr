package configs

import (
// "os"
)

const (
	// Time interval after that a heartbeat event of type 'message' is to be sent to connected clients/subscribers.
	HEARTBEAT_SLEEP_TIME = 20
)

/*
 *  BE data migrations
 */

// Accounts to be ceased from the database inc. their posts.
// You would have to restart backend server for this to apply if you made changes there.
var UserDeletionList []string = []string{
	"fred",
	"fred2",
	"admin",
	"alternative",
	"Lmao",
	"lma0",
}

// This array is used in a procedure's loop to manually unshade listed users.
// Thus listed accounts should have a zero (0) on stats page.
var UsersToUnshade []string = []string{}

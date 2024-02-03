package backend

import (
	"time"

	"go.savla.dev/littr/models"
)

func noteUsersActivity(caller string) bool {
	// check if caller exists
	callerUser, found := getOne(UserCache, caller, models.User{})
	if !found {
		return false
	}

	// update user's activity timestamp
	callerUser.LastActiveTime = time.Now()

	return setOne(UserCache, caller, callerUser)
}

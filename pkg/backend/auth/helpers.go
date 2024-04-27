package auth

import (
	"time"

	"go.savla.dev/littr/pkg/backend/db"
	"go.savla.dev/littr/pkg/models"
)

func noteUsersActivity(caller string) bool {
	// check if caller exists
	callerUser, found := db.GetOne(db.UserCache, caller, models.User{})
	if !found {
		return false
	}

	// update user's activity timestamp
	callerUser.LastActiveTime = time.Now()

	return db.SetOne(db.UserCache, caller, callerUser)
}

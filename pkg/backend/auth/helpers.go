package auth

import (
	"time"

	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/models"
)

func noteUsersActivity(callerID string, cache db.Cacher) bool {
	// check if caller exists
	callerUser, found := cache.Load(callerID)
	if !found {
		return false
	}

	caller, ok := callerUser.(models.User)
	if !ok {
		return false
	}

	// update user's activity timestamp
	caller.LastActiveTime = time.Now()

	return cache.Store(callerID, caller)
}

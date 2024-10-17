package common

import (
	"go.vxn.dev/littr/pkg/models"
)

// helper function to flush sensitive user data in the export for response
func FlushUserData(users *map[string]models.User, callerID string) *map[string]models.User {
	if users == nil || callerID == "" {
		return nil
	}

	// flush unwanted properties
	for key, user := range *users {
		user.Passphrase = ""
		user.PassphraseHex = ""

		// these are kept for callerID
		if user.Nickname != callerID {
			user.Email = ""
			user.FlowList = nil
			user.ShadeList = nil

			// TODO map of user's options
			//user.Options = nil

			// return the caller's status in counterpart account's req. list only callerID's state if present
			if value, found := user.RequestList[callerID]; found {
				user.RequestList = make(map[string]bool)
				user.RequestList[callerID] = value
			} else {
				user.RequestList = nil
			}
		}

		(*users)[key] = user
	}
	return users
}

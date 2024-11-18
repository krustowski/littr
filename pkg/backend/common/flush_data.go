package common

import (
	"go.vxn.dev/littr/pkg/models"
)

// Helper function to flush sensitive user data in the export for response.
func FlushUserData(users *map[string]models.User, callerID string) *map[string]models.User {
	if users == nil || callerID == "" {
		return nil
	}

	// Flush unwanted properties.
	for key, user := range *users {
		user.Passphrase = ""
		user.PassphraseHex = ""

		// These are kept for callerID.
		if user.Nickname != callerID {
			user.Email = ""
			user.FlowList = nil
			user.ShadeList = nil

			// Flush user's options, keep the private state only.
			options := map[string]bool{}
			options["private"] = user.Options["private"]
			user.Options = options

			// Return the caller's status in counterpart account's req. list only callerID's state if present.
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

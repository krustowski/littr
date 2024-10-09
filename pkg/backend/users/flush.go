package users

import (
	"go.vxn.dev/littr/pkg/models"
)

func flushUserData(users *map[string]models.User, callerID string) *map[string]models.User {
	if users == nil || callerID == "" {
		return nil
	}

	// flush unwanted properties
	for key, user := range *users {
		user.Passphrase = ""
		user.PassphraseHex = ""

		if user.Nickname != callerID {
			user.Email = ""
			user.FlowList = nil
			user.ShadeList = nil

			// return the caller's status in counterpart account's req. list only
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

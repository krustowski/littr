package auth

import (
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/models"
)

func authUser(aUser models.User) (*models.User, bool) {
	// fetch one user from cache according to the login credential
	user, ok := db.GetOne(db.UserCache, aUser.Nickname, models.User{})
	if !ok {
		// not found
		return nil, false
	}

	// check the passhash
	if user.Passphrase == aUser.Passphrase || user.PassphraseHex == aUser.PassphraseHex {
		// update user's hexadecimal passphrase form, as the binary form is broken and cannot be used on BE
		if user.PassphraseHex == "" && aUser.PassphraseHex != "" {
			user.PassphraseHex = aUser.PassphraseHex
			_ = db.SetOne(db.UserCache, user.Nickname, user)
		}

		// auth granted
		return &user, true
	}

	// auth failed
	return nil, false
}

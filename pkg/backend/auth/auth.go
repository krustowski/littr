package auth

import (
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/models"
)

type AuthUser struct {
	// Nickname is the user's very username.
	Nickname string `json:"nickname"`

	// Passphrase is a legacy format converted to string from a raw byte stream
	// (do not use anymore as this will be removed in future versions).
	Passphrase string `json:"passphrase"`

	// PassphraseHex is a hexadecimal representation of a passphrase (a SHA-512 checksum).
	// Use 'echo $PASS | sha512sum' for example to get the hex format.
	PassphraseHex string `json:"passphrase_hex"`
}

func authUser(aUser *AuthUser) (*models.User, bool) {
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

			if saved := db.SetOne(db.UserCache, user.Nickname, user); !saved {
				// not very correct response --- user won't know this caused the failed auth
				return nil, false
			}
		}

		// auth granted
		return &user, true
	}

	// auth failed
	return nil, false
}

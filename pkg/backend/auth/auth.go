package auth

import (
// "go.vxn.dev/littr/pkg/backend/db"
// "go.vxn.dev/littr/pkg/models"
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

package models

import (
	"time"
)

// Token is a model structure which is to hold refresh token's properties.
type Token struct {
	// Unique hash = sha512 sum of refresh token's data.
	Hash string `json:"hash"`

	// User's name to easily fetch user's data from the database.
	Nickname string `json:"nickname"`

	// Timestamp of the refresh token's generation, should expire in 4 weeks after the initialization.
	CreatedAt time.Time `json:"created_at"`

	// Time to live, period of validity since the token creation.
	TTL time.Time `json:"ttl"`
}

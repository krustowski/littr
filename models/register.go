package models

type Registration struct {
	Nickname   string `json:"nickname"`
	Passphrase string `json:"passphrase"`
	Poll       int    `json:"poll_id"`
	Type       string `json:"type"`
}

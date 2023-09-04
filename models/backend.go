package backend

type Registration struct {
	ID         int    `json:""`
	Nickname   string `json:"nickname"`
	Passphrase string `json:"passphrase"`
	Poll       int    `json:"poll_id"`
}

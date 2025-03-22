package auth

type AuthUser struct {
	// Nickname is the user's very username.
	Nickname string `json:"nickname" example:"alice"`

	// PassphrasePlain is the plain-text format of the passphrase.
	PassphrasePlain string `json:"passphrase_plain"  example:"s3creTpauWussw0rt"`

	// PassphraseHex is a hexadecimal representation of a passphrase (a SHA-512 checksum).
	// Use 'echo $PASS | sha512sum' for example to get the hex format.
	PassphraseHex string `json:"passphrase_hex" example:"fb43b35a752b0e8045e2dd1b1e292983b9cbf4672a51e30caaa3f9b06c5a3b74d5096bc8092c9e90a2e047c1eab29eceb50c09d6c51e6995c1674beb3b06535e" swaggerignore:"true"`
}

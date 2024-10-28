// Common mailing (and templating) package.
package mail

type MessagePayload struct {
	Email      string
	Type       string
	UUID       string
	Passphrase string
	Nickname   string
}

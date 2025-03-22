// Common mailing (and templating) package.
package mail

import (
	"go.vxn.dev/littr/pkg/models"
)

type MessagePayload struct {
	Email      string
	Type       string
	UUID       string
	Passphrase string
	Nickname   string
}

type mailService struct{}

func NewMailService() models.MailServiceInterface {
	return &mailService{}
}

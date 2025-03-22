package mail

import (
	"errors"
	"os"
	"strings"

	gomail "github.com/wneessen/go-mail"
	"go.vxn.dev/littr/pkg/config"
)

var (
	ErrInvalidPayload    = errors.New("invalid input: not of type MessagePayload")
	ErrNoEmailAddress    = errors.New("invalid input: e-mail address field is empty")
	ErrActivationNoUUID  = errors.New("invalid input: UUID string is empty")
	ErrResetNoPassphrase = errors.New("invalid input: new passhrase to send is empty")
	ErrUnknownMailType   = errors.New("invalid input: unknown mail message type provided")

	ErrTemplateBakeFail = errors.New("message template baking failed, the composition is empty")
)

// ComposeMail is a function to prepare a go-mail-formatted message for sending.
func (s *mailService) ComposeMail(payloadI interface{}) (*gomail.Msg, error) {
	payload, ok := payloadI.(MessagePayload)
	if !ok {
		return nil, ErrInvalidPayload
	}

	m := gomail.NewMsg()

	// Compose the From field.
	if err := m.From(os.Getenv("VAPID_SUBSCRIBER")); err != nil {
		return nil, err
	}

	// Check if the e-mail address is given.
	if payload.Email == "" {
		return nil, ErrNoEmailAddress
	}

	// Compose the To field.
	if err := m.To(payload.Email); err != nil {
		return nil, err
	}

	// Ensure the nickname is never blank.
	nickname := func() string {
		if payload.Nickname != "" && strings.TrimSpace(payload.Nickname) != "" {
			return payload.Nickname
		}

		return "user"
	}()

	var tmplPayload *TemplatePayload

	// Swtich the message type(s).
	switch payload.Type {
	// User Activation procedure.
	case "user_activation":
		if payload.UUID == "" {
			return nil, ErrActivationNoUUID
		}

		tmplPayload = &TemplatePayload{
			Nickname:       nickname,
			MainURL:        config.ServerUrl,
			ActivationLink: "https://" + config.ServerUrl + "/activation/" + payload.UUID,
			TemplateSrc:    "/opt/templates/activation.tmpl",
		}

		m.Subject("User Activation Link")

	// Passphrase reset request.
	case "reset_request":
		if payload.UUID == "" {
			return nil, ErrActivationNoUUID
		}

		tmplPayload = &TemplatePayload{
			Nickname:    nickname,
			MainURL:     config.ServerUrl,
			ResetLink:   "https://" + config.ServerUrl + "/reset/" + payload.UUID,
			UUID:        payload.UUID,
			TemplateSrc: "/opt/templates/reset_request.tmpl",
		}

		m.Subject("Passphrase Reset Request")

	// New passphrase regeneration procedure.
	case "reset_passphrase":
		if payload.Passphrase == "" {
			return nil, ErrResetNoPassphrase
		}

		tmplPayload = &TemplatePayload{
			Nickname:    nickname,
			MainURL:     config.ServerUrl,
			Passphrase:  payload.Passphrase,
			TemplateSrc: "/opt/templates/new_passphrase.tmpl",
		}

		m.Subject("Your New Passphrase")

	default:
		return nil, ErrUnknownMailType
	}

	var composition string

	// Bake the given template (via TemplateSrc) to the <composition> pointer address.
	if err := bakeTemplate(tmplPayload, &composition); err != nil {
		return nil, err
	}

	// Check the composition content.
	if composition == "" {
		return nil, ErrTemplateBakeFail
	}

	// Set the mail's body.
	m.SetBodyString(gomail.TypeTextPlain, composition)

	return m, nil
}

package mail

import (
	"fmt"
	"os"
	"strings"

	gomail "github.com/wneessen/go-mail"
)

// ComposeMail is a function to prepare a go-mail-formatted message for sending.
func ComposeMail(payload MessagePayload) (*gomail.Msg, error) {
	m := gomail.NewMsg()

	// Compose the From field.
	if err := m.From(os.Getenv("VAPID_SUBSCRIBER")); err != nil {
		return nil, err
	}

	// Check if the e-mail address is given.
	if payload.Email == "" {
		return nil, fmt.Errorf("no new passhrase given for mail composition")
	}

	// Compose the To field.
	if err := m.To(payload.Email); err != nil {
		return nil, err
	}

	// Fetch the mail instance URL: used in the PS part of the message.
	mainURL := func() string {
		if os.Getenv("APP_URL_MAIN") != "" {
			return os.Getenv("APP_URL_MAIN")
		}

		return "www.littr.eu"
	}()

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
			return nil, fmt.Errorf("no UUID given for mail composition")
		}

		tmplPayload = &TemplatePayload{
			Nickname:       nickname,
			MainURL:        mainURL,
			ActivationLink: "https://" + mainURL + "/activate/" + payload.UUID,
			TemplateSrc:    "/opt/templates/activation.tmpl",
		}

		m.Subject("User Activation Link")

	// Passphrase reset request.
	case "reset_request":
		if payload.UUID == "" {
			return nil, fmt.Errorf("no UUID given for mail composition")
		}

		tmplPayload = &TemplatePayload{
			Nickname:    nickname,
			MainURL:     mainURL,
			ResetLink:   "https://" + mainURL + "/reset/" + payload.UUID,
			UUID:        payload.UUID,
			TemplateSrc: "/opt/templates/reset_request.tmpl",
		}

		m.Subject("Passphrase Reset Request")

	// New passphrase regeneration procedure.
	case "reset_passphrase":
		if payload.Passphrase == "" {
			return nil, fmt.Errorf("no new passhrase given for mail composition")
		}

		tmplPayload = &TemplatePayload{
			Nickname:    nickname,
			MainURL:     mainURL,
			Passphrase:  payload.Passphrase,
			TemplateSrc: "/opt/templates/new_passphrase.tmpl",
		}

		m.Subject("Your New Passphrase")

	default:
		return nil, fmt.Errorf("no mail Type specified")
	}

	var composition string

	// Bake the given template (via TemplateSrc) to the <composition> pointer address.
	if err := bakeTemplate(tmplPayload, &composition); err != nil {
		return nil, err
	}

	// Check the composition content.
	if composition == "" {
		return nil, fmt.Errorf("blank mail text composition")
	}

	// Set the mail's body.
	m.SetBodyString(gomail.TypeTextPlain, composition)

	return m, nil
}

package mail

import (
	"fmt"
	"os"

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

	// Swtich the message type(s).
	switch payload.Type {
	case "user_activation":
		if payload.Nickname == "" {
			return nil, fmt.Errorf("no user nickname given")
		}

		activationLink := "https://" + mainURL + "/activate/" + payload.UUID

		m.Subject("User Activation")
		m.SetBodyString(gomail.TypeTextPlain, "Dear "+payload.Nickname+",\n\nWe received a request to reset the passphrase for your account associated with this e-mail address: "+payload.Email+"\n\nTo reset your passphrase, please click the link below:\n\nReset Passphrase Link: "+activationLink+"\n\nYou can insert the generated UUID in the reset form too: "+payload.UUID+"\n\nIf you did not request a passphrase reset, please ignore this email. Your passphrase will remain unchanged.\n\nFor security reasons, this link will expire in 24 hours.\n\nThank you\nlittr\nhttps://"+mainURL)

	case "reset_request":
		if payload.UUID == "" {
			return nil, fmt.Errorf("no UUID given for mail composition")
		}

		resetLink := "https://" + mainURL + "/reset/" + payload.UUID

		m.Subject("Passphrase Reset Request")
		m.SetBodyString(gomail.TypeTextPlain, "Dear user,\n\nWe received a request to reset the passphrase for your account associated with this e-mail address: "+payload.Email+"\n\nTo reset your passphrase, please click the link below:\n\nReset Passphrase Link: "+resetLink+"\n\nYou can insert the generated UUID in the reset form too: "+payload.UUID+"\n\nIf you did not request a passphrase reset, please ignore this email. Your passphrase will remain unchanged.\n\nFor security reasons, this link will expire in 24 hours.\n\nThank you\nlittr\nhttps://"+mainURL)

	case "reset_passphrase":
		if payload.Passphrase == "" {
			return nil, fmt.Errorf("no new passhrase given for mail composition")
		}

		m.Subject("Your New Passphrase")
		m.SetBodyString(gomail.TypeTextPlain, "Dear user,\n\nThe requested passphrase regeneration process has been successful. Please use the generated string below to log-in again.\n\nNew passphrase: "+payload.Passphrase+"\n\nPlease do not forget to change the passphrase right after logging in in settings.\n\nThank you\nlittr\nhttps://"+mainURL)

	default:
		return nil, fmt.Errorf("no mail Type specified")
	}

	return m, nil
}

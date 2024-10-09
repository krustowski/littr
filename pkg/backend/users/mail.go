package users

import (
	"fmt"
	"os"
	"strconv"

	gomail "github.com/wneessen/go-mail"
)

func sendResetMail(msg *gomail.Msg) error {
	port, err := strconv.Atoi(os.Getenv("MAIL_PORT"))
	if err != nil {
		return err
	}

	c, err := gomail.NewClient(os.Getenv("MAIL_HOST"), gomail.WithPort(port), gomail.WithSMTPAuth(gomail.SMTPAuthPlain),
		gomail.WithUsername(os.Getenv("MAIL_SASL_USR")), gomail.WithPassword(os.Getenv("MAIL_SASL_PWD")), gomail.WithHELO(os.Getenv("MAIL_HELO")))
	if err != nil {
		return err
	}

	//c.SetTLSPolicy(mail.TLSOpportunistic)

	if err := c.DialAndSend(msg); err != nil {
		return err
	}

	return nil
}

func composeResetMail(payload msgPayload) (*gomail.Msg, error) {
	m := gomail.NewMsg()
	if err := m.From(os.Getenv("VAPID_SUBSCRIBER")); err != nil {
		return nil, err
	}

	if payload.Email == "" {
		return nil, fmt.Errorf("no new passhrase given for mail composition")
	}

	if err := m.To(payload.Email); err != nil {
		return nil, err
	}

	mainURL := os.Getenv("APP_URL_MAIN")
	if mainURL == "" {
		mainURL = "www.littr.eu"
	}

	switch payload.Type {
	case "request":
		if payload.UUID == "" {
			return nil, fmt.Errorf("no UUID given for mail composition")
		}

		resetLink := "https://" + mainURL + "/reset/" + payload.UUID

		m.Subject("Passphrase Reset Request")
		m.SetBodyString(gomail.TypeTextPlain, "Dear user,\n\nWe received a request to reset the passphrase for your account associated with this e-mail address: "+payload.Email+"\n\nTo reset your passphrase, please click the link below:\n\nReset Passphrase Link: "+resetLink+"\n\nYou can insert the generated UUID in the reset form too: "+payload.UUID+"\n\nIf you did not request a passphrase reset, please ignore this email. Your passphrase will remain unchanged.\n\nFor security reasons, this link will expire in 24 hours.\n\nThank you\nlittr\nhttps://"+mainURL)

	case "passphrase":
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

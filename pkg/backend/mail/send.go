package mail

import (
	"errors"
	"os"
	"strconv"

	gomail "github.com/wneessen/go-mail"
)

var (
	ErrIncompleteMailServerConfiguration = errors.New("mail server is not configured properly, check the settings")
)

var (
	mailHelo     = os.Getenv("MAIL_HELO")
	mailHost     = os.Getenv("MAIL_HOST")
	mailPort     = os.Getenv("MAIL_PORT")
	mailSaslUser = os.Getenv("MAIL_SASL_USR")
	mailSaslPass = os.Getenv("MAIL_SASL_PWD")
)

func (s *mailService) SendMail(msg *gomail.Msg) error {
	port, err := strconv.Atoi(mailPort)
	if err != nil {
		return err
	}

	if mailHost == "" || mailSaslUser == "" || mailSaslPass == "" || mailHelo == "" {
		return ErrIncompleteMailServerConfiguration
	}

	c, err := gomail.NewClient(mailHost, gomail.WithPort(port), gomail.WithSMTPAuth(gomail.SMTPAuthPlain),
		gomail.WithUsername(mailSaslUser), gomail.WithPassword(mailSaslPass), gomail.WithHELO(mailHelo))
	if err != nil {
		return err
	}

	//c.SetTLSPolicy(mail.TLSOpportunistic)

	if err := c.DialAndSend(msg); err != nil {
		return err
	}

	return nil
}

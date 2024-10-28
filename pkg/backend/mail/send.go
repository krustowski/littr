package mail

import (
	"fmt"
	"os"
	"strconv"

	gomail "github.com/wneessen/go-mail"
)

func SendMail(msg *gomail.Msg) error {
	port, err := strconv.Atoi(os.Getenv("MAIL_PORT"))
	if err != nil {
		return err
	}

	if os.Getenv("MAIL_HOST") == "" || os.Getenv("MAIL_SASL_USR") == "" || os.Getenv("MAIL_SASL_PWD") == "" || os.Getenv("MAIL_HELO") == "" {
		return fmt.Errorf("invalid mail server configuration, check the server settings")
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

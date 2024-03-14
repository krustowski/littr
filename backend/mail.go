package backend

import (
	"github.com/wneessen/go-mail"
	"log"
	"os"
)

func main() {
	m := mail.NewMsg()
	if err := m.From("littr@n0p.cz"); err != nil {
		log.Fatalf("failed to set From address: %s", err)
	}
	if err := m.To(""); err != nil {
		log.Fatalf("failed to set To address: %s", err)
	}
	m.Subject("This is my first mail with go-mail!")
	m.SetBodyString(mail.TypeTextPlain, "Do you like this mail? I certainly do!")

	c, err := mail.NewClient(os.Getenv("MAIL_HOST"), mail.WithPort(587), mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(os.Getenv("MAIL_SASL_USER")), mail.WithPassword(os.Getenv("MAIL_SASL_PWD")))
	if err != nil {
		log.Fatalf("failed to create mail client: %s", err)
	}

	//c.SetTLSPolicy(mail.TLSOpportunistic)

	if err := c.DialAndSend(m); err != nil {
		log.Fatalf("failed to send mail: %s", err)
	}
}

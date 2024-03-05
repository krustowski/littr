// package main
package backend

import (
	//"fmt"
	//"net/smtp"
	//
	"github.com/wneessen/go-mail"
	"log"
	//
	//sasl "github.com/emersion/go-sasl"
	//smtp "github.com/emersion/go-smtp"
	//"log"
	//"strings"
)

/*func main() {
	// https://www.loginradius.com/blog/engineering/sending-emails-with-golang/
	from := "littr@n0p.cz"
	//user := os.Getenv("MAIL_SASL_USR")
	//password := os.Getenv("MAIL_SASL_PWD")
	user := "littr_n0p_cz"
	password := "81875096ba5fd67f113fsd78787987a562221fab16060fds456454anm5f6f23"

	to := []string{
		"krusty@savla.dev",
	}

	smtpHost := "a90bffe1884b84d5e255f12ff0ecbd70f2edfc877b68d7c.n0p.cz"
	smtpPort := "587"
	message := []byte("test message")

	// https://pkg.go.dev/net/smtp#CRAMMD5Auth
	//auth := smtp.CRAMMD5Auth(user, password)
	auth := smtp.PlainAuth("", user, password, smtpHost)

	if err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message); err != nil {
		fmt.Println(err)
		return
	}
}*/

func main() {
	m := mail.NewMsg()
	if err := m.From("littr@n0p.cz"); err != nil {
		log.Fatalf("failed to set From address: %s", err)
	}
	if err := m.To("krusty@savla.dev"); err != nil {
		log.Fatalf("failed to set To address: %s", err)
	}
	m.Subject("This is my first mail with go-mail!")
	m.SetBodyString(mail.TypeTextPlain, "Do you like this mail? I certainly do!")

	c, err := mail.NewClient("a90bffe1884b84d5e255f12ff0ecbd70f2edfc877b68d7c.n0p.cz", mail.WithPort(587), mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername("littr_n0p_cz"), mail.WithPassword("81875096ba5fd67f113fsd78787987a562221fab16060fds456454anm5f6f23"))
	if err != nil {
		log.Fatalf("failed to create mail client: %s", err)
	}

	//c.SetTLSPolicy(mail.TLSOpportunistic)

	if err := c.DialAndSend(m); err != nil {
		log.Fatalf("failed to send mail: %s", err)
	}
}

/*func main() {
	// Set up authentication information.
	auth := sasl.NewPlainClient("", "littr_n0p_cz", "81875096ba5fd67f113fsd78787987a562221fab16060fds456454anm5f6f23")

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	to := []string{"krusty@savla.dev"}
	msg := strings.NewReader("To: krusty@savla.dev\r\n" +
		"Subject: discount Gophers!\r\n" +
		"\r\n" +
		"This is the email body.\r\n")
	err := smtp.SendMail("0bffe1884b84d5e255f12ff0ecbd70f2edfc877b68d7c.n0p.cz:587", auth, "littr@n0p.cz", to, msg)
	if err != nil {
		log.Fatal(err)
	}
}*/

package backend

//package main

import (
	"fmt"
	"net/smtp"
	"os"
)

func main() {
	// https://www.loginradius.com/blog/engineering/sending-emails-with-golang/
	from := "littr@n0p.cz"
	user := os.Getenv("MAIL_SASL_USR")
	password := os.Getenv("MAIL_SASL_PWD")

	to := []string{
		"krusty@savla.dev",
	}

	smtpHost := "frank.savla.net"
	smtpPort := "25"

	message := []byte("test message")

	// https://pkg.go.dev/net/smtp#CRAMMD5Auth
	// auth := smtp.CRAMMD5Auth(user, password)
	auth := smtp.PlainAuth("", user, password, smtpHost)

	if err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message); err != nil {
		fmt.Println(err)
		return
	}
}

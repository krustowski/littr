package backend

import (
	"crypto/sha512"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"

	mail "github.com/wneessen/go-mail"
)

func resetHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "reset")

	fetch := struct {
		Email string `json:"email"`
	}{}

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message+err.Error(), resp.Code)
		resp.Write(w)
		return
	}

	data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

	if err = json.Unmarshal(data, &fetch); err != nil {
		resp.Message = "backend error: cannot unmarshall request data"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message+err.Error(), resp.Code)
		resp.Write(w)
		return
	}

	rand.Seed(time.Now().UnixNano())
	random := randSeq(16)
	passHash := sha512.Sum512([]byte(random + os.Getenv("APP_PEPPER")))

	users, _ := getAll(UserCache, models.User{})

	// loop over users to find matching e-mail address
	var user models.User

	found := false
	for _, user = range users {
		if user.Email == fetch.Email {
			found = true
			break
		}
	}

	if !found {
		resp.Message = "backend error: matching user not found"
		resp.Code = http.StatusNotFound

		l.Println(resp.Message+err.Error(), resp.Code)
		resp.Write(w)
		return
	}

	user.Passphrase = string(passHash[:])
	if saved := setOne(UserCache, user.Nickname, user); !saved {
		resp.Message = "backend error: cannot update user in database"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message+err.Error(), resp.Code)
		resp.Write(w)
		return
	}

	email := user.Email

	// compose a mail
	m := mail.NewMsg()
	if err := m.From("littr@n0p.cz"); err != nil {
		resp.Message = "backend error: failed to set From address: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message+err.Error(), resp.Code)
		resp.Write(w)
		return
	}
	if err := m.To(email); err != nil {
		resp.Message = "backend error: failed to set To address: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message+err.Error(), resp.Code)
		resp.Write(w)
		return
	}
	m.Subject("Lost password recovery")
	m.SetBodyString(mail.TypeTextPlain, "Someone requested the password reset for the account linked to this e-mail.\n\n"+random)

	port, err := strconv.Atoi(os.Getenv("MAIL_PORT"))
	if err != nil {
		resp.Message = "backend error: cannot convert MAIL_PORT to int: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message+err.Error(), resp.Code)
		resp.Write(w)
		return
	}

	c, err := mail.NewClient(os.Getenv("MAIL_HOST"), mail.WithPort(port), mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(os.Getenv("MAIL_SASL_USR")), mail.WithPassword(os.Getenv("MAIL_SASL_PWD")))
	if err != nil {
		resp.Message = "backend error: failed to create mail client: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message+err.Error(), resp.Code)
		resp.Write(w)
		return
	}

	//c.SetTLSPolicy(mail.TLSOpportunistic)

	if err := c.DialAndSend(m); err != nil {
		resp.Message = "backend error: failed to sent e-mail: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message+err.Error(), resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "reset e-mail was rent"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
	return
}

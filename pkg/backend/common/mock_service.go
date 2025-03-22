package common

import (
	"go.vxn.dev/littr/pkg/models"

	gomail "github.com/wneessen/go-mail"
)

type MockMailService struct{}

func (m *MockMailService) ComposeMail(payload interface{}) (*gomail.Msg, error) {
	return &gomail.Msg{}, nil
}

func (m *MockMailService) SendMail(msg *gomail.Msg) error {
	return nil
}

// Implementation verification for compiler.
var _ models.MailServiceInterface = (*MockMailService)(nil)

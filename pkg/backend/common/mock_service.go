package common

import (
	"context"

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

type MockPagingService struct{}

func (m *MockPagingService) GetOne(ctx context.Context, opts interface{}, data ...interface{}) (interface{}, error) {
	return nil, nil
}

func (m *MockPagingService) GetMany(ctx context.Context, opts any) (any, error) {
	return nil, nil
}

// Implementation verification for compiler.
var _ models.PagingServiceInterface = (*MockPagingService)(nil)

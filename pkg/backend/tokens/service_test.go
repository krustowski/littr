package tokens

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.vxn.dev/littr/pkg/models"
)

const (
	nicknameToFind string = "mocker99"
)

type mockTokenRepository struct{}

type mockNickname string

func (m *mockTokenRepository) GetAll() (*map[string]models.Token, error) {
	return &map[string]models.Token{
		"ff00lerT5": {
			Hash:      "ff00lertT5",
			Nickname:  nicknameToFind,
			CreatedAt: time.Now().Add(-5 * 7 * 24 * time.Hour),
			TTL:       500,
		},
		"xx09wert": {
			Hash:      "xx00wert",
			Nickname:  "tweaker66",
			CreatedAt: time.Now(),
			TTL:       500,
		},
	}, nil
}

func (m *mockTokenRepository) GetByID(tokenID string) (*models.Token, error) {
	return &models.Token{
		Hash:      "ff00lertT5",
		Nickname:  nicknameToFind,
		CreatedAt: time.Now().Add(-5 * 7 * 24 * time.Hour),
		TTL:       500,
	}, nil
}

func (m *mockTokenRepository) Save(token *models.Token) error {
	return nil
}

func (m *mockTokenRepository) Delete(tokenID string) error {
	return nil
}

//
//  Tests
//

func newTestContext() context.Context {
	return context.WithValue(context.Background(), mockNickname("nickname"), "lawrents")
}

func newTestService(t *testing.T) models.TokenServiceInterface {
	service := NewTokenService(&mockTokenRepository{})
	if service == nil {
		t.Fatal("nil TokenService")
	}

	return service
}

func TestTokens_TokenServiceCreate(t *testing.T) {
	service := newTestService(t)
	ctx := newTestContext()

	tokens, err := service.Create(ctx, &models.User{Nickname: nicknameToFind})
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(tokens) != 2 {
		t.Fatal("too few tokens received")
	}
}

func TestTokens_TokenServiceFindByID(t *testing.T) {
	service := newTestService(t)
	ctx := newTestContext()

	token, err := service.FindByID(ctx, nicknameToFind)
	if err != nil {
		if errors.Is(err, errNotImplemented) {
			t.Skip("skipping unimplemented method")
		}

		t.Fatal(err.Error())
	}

	if token.Nickname != nicknameToFind {
		t.Errorf("wrong nickname returned: %s", token.Nickname)
	}
}

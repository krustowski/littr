package tokens

import (
	"context"
	"errors"
	"testing"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/models"
)

const (
	nicknameToFind string = "mocker99"
)

//
//  Tests
//

func newTestContext() context.Context {
	return context.WithValue(context.Background(), common.ContextUserKeyName, "lawrents")
}

func newTestService(t *testing.T) models.TokenServiceInterface {
	service := NewTokenService(&common.MockTokenRepository{})
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

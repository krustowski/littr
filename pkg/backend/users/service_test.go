package users

import (
	"context"
	"testing"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/models"
)

type mockNickname string

//
//  Tests
//

func newTestContext() context.Context {
	return context.WithValue(context.Background(), mockNickname("nickname"), "lawrents")
}

func newTestService(t *testing.T) models.UserServiceInterface {
	service := NewUserService(&common.MockMailService{}, &common.MockPollRepository{}, &common.MockPostRepository{}, &common.MockRequestRepository{}, &common.MockTokenRepository{}, &common.MockUserRepository{})
	if service == nil {
		t.Fatal("nil UserService")
	}

	return service
}

func TestUsers_UserServiceCreate(t *testing.T) {
	ctx := newTestContext()
	service := newTestService(t)

	req := &UserCreateRequest{
		Email:           "alice@example.com",
		Nickname:        "alice",
		PassphrasePlain: "bobdod",
	}

	if err := service.Create(ctx, req); err != nil {
		t.Error(err)
	}
}

func TestUsers_UserServiceActivate(t *testing.T) {
	ctx := newTestContext()
	service := newTestService(t)

	if err := service.Activate(ctx, common.MockUserNickname); err != nil {
		t.Error(err)
	}
}

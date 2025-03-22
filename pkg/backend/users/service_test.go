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
	service := NewUserService(&common.MockPollRepository{}, &common.MockPostRepository{}, &common.MockRequestRepository{}, &common.MockTokenRepository{}, &common.MockUserRepository{})
	if service == nil {
		t.Fatal("nil UserService")
	}

	return service
}

package common

import (
	"time"

	"go.vxn.dev/littr/pkg/models"
)

var MockUserNickname = "mocker99"

//
//  PollRepositoryInterface dummy implementation
//

type MockPollRepository struct{}

func (m *MockPollRepository) GetAll() (*map[string]models.Poll, error) {
	return &map[string]models.Poll{}, nil
}

func (m *MockPollRepository) GetPage(opts interface{}) (*map[string]models.Poll, error) {
	return &map[string]models.Poll{}, nil
}

func (m *MockPollRepository) GetByID(pollID string) (*models.Poll, error) {
	return &models.Poll{}, nil
}

func (m *MockPollRepository) Save(poll *models.Poll) error {
	return nil
}

func (m *MockPollRepository) Delete(pollID string) error {
	return nil
}

// Implementation verification for compiler.
var _ models.PollRepositoryInterface = (*MockPollRepository)(nil)

//
//  PostRepositoryInterface dummy implementation
//

type MockPostRepository struct{}

func (m *MockPostRepository) GetAll() (*map[string]models.Post, error) {
	return &map[string]models.Post{}, nil
}

func (m *MockPostRepository) GetPage(opts interface{}) (*map[string]models.Post, *map[string]models.User, error) {
	return &map[string]models.Post{}, &map[string]models.User{}, nil
}

func (m *MockPostRepository) GetByID(postID string) (*models.Post, error) {
	return &models.Post{}, nil
}

func (m *MockPostRepository) Save(post *models.Post) error {
	return nil
}

func (m *MockPostRepository) Delete(postID string) error {
	return nil
}

// Implementation verification for compiler.
var _ models.PostRepositoryInterface = (*MockPostRepository)(nil)

//
//  RequestRepositoryInterface dummy implementation
//

type MockRequestRepository struct{}

func (m *MockRequestRepository) GetByID(reqID string) (*models.Request, error) {
	return &models.Request{}, nil
}

func (m *MockRequestRepository) Save(req *models.Request) error {
	return nil
}

func (m *MockRequestRepository) Delete(reqID string) error {
	return nil
}

//
//  SubscriptionRepositoryInterface dummy implementation
//

type MockSubscriptionRepository struct{}

func (m *MockSubscriptionRepository) GetByUserID(userID string) (*[]models.Device, error) {
	return &[]models.Device{}, nil
}

func (m *MockSubscriptionRepository) Save(userID string, sub *[]models.Device) error {
	return nil
}

func (m *MockSubscriptionRepository) Delete(userID string) error {
	return nil
}

// Implementation verification for compiler.
var _ models.SubscriptionRepositoryInterface = (*MockSubscriptionRepository)(nil)

//
//  TokenRepositoryInterface dummy implementation
//

type MockTokenRepository struct{}

func (m *MockTokenRepository) GetAll() (*map[string]models.Token, error) {
	return &map[string]models.Token{
		"ff00lerT5": {
			Hash:      "ff00lertT5",
			Nickname:  MockUserNickname,
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

func (m *MockTokenRepository) GetByID(tokenID string) (*models.Token, error) {
	return &models.Token{
		Hash:      "ff00lertT5",
		Nickname:  MockUserNickname,
		CreatedAt: time.Now().Add(-5 * 7 * 24 * time.Hour),
		TTL:       500,
	}, nil
}

func (m *MockTokenRepository) Save(token *models.Token) error {
	return nil
}

func (m *MockTokenRepository) Delete(tokenID string) error {
	return nil
}

// Implementation verification for compiler.
var _ models.TokenRepositoryInterface = (*MockTokenRepository)(nil)

//
//  UserRepositoryInterface dummy implementation
//

type MockUserRepository struct{}

func (m *MockUserRepository) GetAll() (*map[string]models.User, error) {
	return &map[string]models.User{}, nil
}

func (m *MockUserRepository) GetPage(opts interface{}) (*map[string]models.User, error) {
	return &map[string]models.User{}, nil
}

func (m *MockUserRepository) GetByID(userID string) (*models.User, error) {
	return &models.User{}, nil
}

func (m *MockUserRepository) Save(user *models.User) error {
	return nil
}

func (m *MockUserRepository) Delete(userID string) error {
	return nil
}

// Implementation verification for compiler.
var _ models.UserRepositoryInterface = (*MockUserRepository)(nil)

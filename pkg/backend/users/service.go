package users

import (
	"context"
	"fmt"
	//"strconv"
	//"time"

	"go.vxn.dev/littr/pkg/backend/common"
	//"go.vxn.dev/littr/pkg/backend/live"
	"go.vxn.dev/littr/pkg/backend/pages"
	//"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"
)

type UserUpdateRequest struct {
	NewPassphraseHex     string `json:"new_passphrase_hex"`
	CurrentPassphraseHex string `json:"current_passphrase_hex"`
}

//
// models.UserServiceInterface implementation
//

type UserService struct {
	postRepository  models.PostRepositoryInterface
	userRepository  models.UserRepositoryInterface
	tokenRepository models.TokenRepositoryInterface
}

func NewUserService(
	postRepository models.PostRepositoryInterface,
	userRepository models.UserRepositoryInterface,
	tokenRepository models.TokenRepositoryInterface,
) models.UserServiceInterface {
	if postRepository == nil || userRepository == nil || tokenRepository == nil {
		return nil
	}

	return &UserService{
		postRepository:  postRepository,
		userRepository:  userRepository,
		tokenRepository: tokenRepository,
	}
}

func (s *UserService) Create(ctx context.Context, user *models.User) error {
	// Fetch the callerID/nickname type from the given context.
	_, ok := ctx.Value("nickname").(string)
	if !ok {
		return fmt.Errorf("could not decode the user request")
	}

	return fmt.Errorf("not yet implemented")
}

func (s *UserService) Update(ctx context.Context, userRequest interface{}) error {
	// Assert the type for the user update request.
	_, ok := userRequest.(*UserUpdateRequest)
	if !ok {
		return fmt.Errorf("could not decode the user request")
	}

	// Fetch the update type from the given context.
	reqType, ok := ctx.Value("updateType").(string)
	if !ok {
		return fmt.Errorf("could not decode the user request")
	}

	switch reqType {
	case "lists":
	case "options":
	case "passhrase":
	default:
		return fmt.Errorf("unknown request type")
	}

	return fmt.Errorf("not yet implemented")
}

func (s *UserService) Delete(ctx context.Context, userID string) error {
	return fmt.Errorf("not yet implemented")
}

func (s *UserService) FindAll(ctx context.Context) (*map[string]models.User, error) {
	// Fetch the caller's ID from the context.
	callerID, ok := ctx.Value("nickname").(string)
	if !ok {
		return nil, fmt.Errorf(common.ERR_CALLER_FAIL)
	}

	// Fetch the pageNo from the context.
	pageNo, ok := ctx.Value("pageNo").(int)
	if !ok {
		return nil, fmt.Errorf(common.ERR_PAGENO_INCORRECT)
	}

	// Request the caller from the user repository.
	caller, err := s.userRepository.GetByID(callerID)
	if err != nil {
		return nil, err
	}

	// Compose a pagination options object to paginate users.
	opts := &pages.PageOptions{
		CallerID: callerID,
		PageNo:   pageNo,
		FlowList: nil,

		Users: pages.UserOptions{
			Plain:       true,
			RequestList: &caller.RequestList,
		},
	}

	// Request the page of users from the user repository.
	users, err := s.userRepository.GetPage(opts)
	if err != nil {
		return nil, err
	}

	// Add the caller to users map.
	(*users)[callerID] = *caller

	// Patch the user's data for export.
	patchedUsers := common.FlushUserData(users, callerID)

	return patchedUsers, nil
}

func (s *UserService) FindByID(ctx context.Context, userID string) (*models.User, error) {
	// Fetch the user's ID from the context.
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return nil, fmt.Errorf(common.ERR_USERID_BLANK)
	}

	// Request the user's data from repository..
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Patch the user's data for export.
	patchedUser := (*common.FlushUserData(&map[string]models.User{userID: *user}, userID))[userID]

	return &patchedUser, nil
}

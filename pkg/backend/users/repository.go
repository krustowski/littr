package users

import (
	"fmt"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/backend/pages"
	"go.vxn.dev/littr/pkg/models"
)

// The implementation of pkg/models.UserRepositoryInterface.
type UserRepository struct {
	cache db.Cacher
}

func NewUserRepository(cache db.Cacher) models.UserRepositoryInterface {
	if cache == nil {
		return nil
	}

	return &UserRepository{
		cache: cache,
	}
}

func (r *UserRepository) GetAll() (*map[string]models.User, error) {
	rawUsers, count := r.cache.Range()
	if count == 0 {
		return nil, fmt.Errorf("no items found")
	}

	users := make(map[string]models.User)

	// Assert types to fetched interface map.
	for key, rawUser := range *rawUsers {
		user, ok := rawUser.(models.User)
		if !ok {
			return nil, fmt.Errorf("user's data corrupted")
		}

		users[key] = user
	}

	return &users, nil
}

func (r *UserRepository) GetPage(pageOpts interface{}) (*map[string]models.User, error) {
	// Assert type for pageOptions.
	opts, ok := pageOpts.(*pages.PageOptions)
	if !ok {
		return nil, fmt.Errorf("cannot read the page options at the repository level")
	}

	// Fetch page according to the calling user (in options).
	pagePtrs := pages.GetOnePage(*opts)
	if pagePtrs == (pages.PagePointers{}) || pagePtrs.Users == nil || (*pagePtrs.Users) == nil {
		return nil, fmt.Errorf(common.ERR_PAGE_EXPORT_NIL)
	}

	// If zero items were fetched, no need to continue asserting types.
	if len(*pagePtrs.Users) == 0 {
		return nil, fmt.Errorf("no users found in the database")
	}

	return pagePtrs.Users, nil

}

func (r *UserRepository) GetByID(userID string) (*models.User, error) {
	// Fetch the user from the cache.
	rawUser, found := r.cache.Load(userID)
	if !found {
		return nil, fmt.Errorf("requested user not found")
	}

	// Assert the type
	user, ok := rawUser.(models.User)
	if !ok {
		return nil, fmt.Errorf("user's data corrupted")
	}

	return &user, nil
}

func (r *UserRepository) Save(user *models.User) error {
	// Store the user using its key in the cache.
	saved := r.cache.Store(user.Nickname, *user)
	if !saved {
		return fmt.Errorf("an error occurred while saving a user")
	}

	return nil
}

func (r *UserRepository) Delete(userID string) error {
	// Simple user's deleting.
	deleted := r.cache.Delete(userID)
	if !deleted {
		return fmt.Errorf("user data could not be purged from the database")
	}

	return nil
}
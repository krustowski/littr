package tokens

import (
	"fmt"

	//"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	//"go.vxn.dev/littr/pkg/backend/pages"
	"go.vxn.dev/littr/pkg/models"
)

// The implementation of pkg/models.TokenRepositoryInterface.
type TokenRepository struct {
	cache db.Cacher
}

func NewTokenRepository(cache db.Cacher) models.TokenRepositoryInterface {
	if cache == nil {
		return nil
	}

	return &TokenRepository{
		cache: cache,
	}
}

/*func (r *TokenRepository) GetAll(pageOpts interface{}) (*map[string]models.Token, error) {
	// Assert type for pageOptions.
	opts, ok := pageOpts.(*pages.PageOptions)
	if !ok {
		return nil, fmt.Errorf("cannot read the page options at the repository level")
	}

	// Fetch page according to the calling user (in options).
	pagePtrs := pages.GetOnePage(*opts)
	if pagePtrs == (pages.PagePointers{}) || pagePtrs.Tokens == nil || (*pagePtrs.Tokens) == nil {
		return nil, fmt.Errorf(common.ERR_PAGE_EXPORT_NIL)
	}

	// If zero items were fetched, no need to continue asserting types.
	if len(*pagePtrs.Tokens) == 0 {
		return nil, fmt.Errorf("no tokens found in the database")
	}

	return pagePtrs.Tokens, nil

}*/

// GetTokenByID is a static function to export to other services.
func GetTokenByID(tokenID string, cache db.Cacher) (*models.Token, error) {
	if tokenID == "" || cache == nil {
		return nil, fmt.Errorf("tokenID is blank, or cache is nil")
	}

	// Fetch the token from the cache.
	tokRaw, found := cache.Load(tokenID)
	if !found {
		return nil, fmt.Errorf("could not find requested token")
	}

	token, ok := tokRaw.(models.Token)
	if !ok {
		return nil, fmt.Errorf("could not assert type *models.Token")
	}

	return &token, nil
}

func (r *TokenRepository) GetByID(tokenID string) (*models.Token, error) {
	// Use the static function to get such token.
	token, err := GetTokenByID(tokenID, r.cache)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (r *TokenRepository) Save(token *models.Token) error {
	// Store the token using its key in the cache.
	saved := r.cache.Store(token.Hash, *token)
	if !saved {
		return fmt.Errorf("an error occurred while saving a token")
	}

	return nil
}

func (r *TokenRepository) Delete(tokenID string) error {
	// Simple token's deletion.
	deleted := r.cache.Delete(tokenID)
	if !deleted {
		return fmt.Errorf("token data could not be purged from the database")
	}

	return nil
}

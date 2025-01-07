package push

import (
	"errors"
	"fmt"

	//"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	//"go.vxn.dev/littr/pkg/backend/pages"
	"go.vxn.dev/littr/pkg/models"
)

// The implementation of pkg/models.SubscriptionRepositoryInterface.
type SubscriptionRepository struct {
	cache db.Cacher
}

func NewSubscriptionRepository(cache db.Cacher) models.SubscriptionRepositoryInterface {
	if cache == nil {
		return nil
	}

	return &SubscriptionRepository{
		cache: cache,
	}
}

/*func (r *SubscriptionRepository) GetAll(pageOpts interface{}) (*map[string]models.Subscription, error) {
	// Assert type for pageOptions.
	opts, ok := pageOpts.(*pages.PageOptions)
	if !ok {
		return nil, fmt.Errorf("cannot read the page options at the repository level")
	}

	// Fetch page according to the calling user (in options).
	pagePtrs := pages.GetOnePage(*opts)
	if pagePtrs == (pages.PagePointers{}) || pagePtrs.Subscriptions == nil || (*pagePtrs.Subscriptions) == nil {
		return nil, fmt.Errorf(common.ERR_PAGE_EXPORT_NIL)
	}

	// If zero items were fetched, no need to continue asserting types.
	if len(*pagePtrs.Subscriptions) == 0 {
		return nil, fmt.Errorf("no subscriptions found in the database")
	}

	return pagePtrs.Subscriptions, nil

}*/

var (
	ErrSubscriptionNotFound error = errors.New("could not find requested Subscription")
)

// GetSubscriptionByID is a static function to export to other services.
func GetSubscriptionByID(userID string, cache db.Cacher) (*[]models.Device, error) {
	if userID == "" || cache == nil {
		return nil, fmt.Errorf("subscriptionID is blank, or cache is nil")
	}

	// Fetch the subscription from the cache.
	tokRaw, found := cache.Load(userID)
	if !found {
		return nil, ErrSubscriptionNotFound
	}

	subscription, ok := tokRaw.([]models.Device)
	if !ok {
		return nil, fmt.Errorf("could not assert type *models.Subscription")
	}

	return &subscription, nil
}

func (r *SubscriptionRepository) GetByUserID(userID string) (*[]models.Device, error) {
	// Use the static function to get such subscription.
	subscription, err := GetSubscriptionByID(userID, r.cache)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

func (r *SubscriptionRepository) Save(userID string, subscription *[]models.Device) error {
	// Store the subscription using its key in the cache.
	saved := r.cache.Store(userID, *subscription)
	if !saved {
		return fmt.Errorf("an error occurred while saving a subscription")
	}

	return nil
}

func (r *SubscriptionRepository) Delete(userID string) error {
	// Simple subscription's deletion.
	deleted := r.cache.Delete(userID)
	if !deleted {
		return fmt.Errorf("subscription data could not be purged from the database")
	}

	return nil
}

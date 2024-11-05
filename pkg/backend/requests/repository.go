package requests

import (
	"fmt"

	//"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	//"go.vxn.dev/littr/pkg/backend/pages"
	"go.vxn.dev/littr/pkg/models"
)

// The implementation of pkg/models.RequestRepositoryInterface.
type RequestRepository struct {
	cache db.Cacher
}

func NewRequestRepository(cache db.Cacher) models.RequestRepositoryInterface {
	if cache == nil {
		return nil
	}

	return &RequestRepository{
		cache: cache,
	}
}

/*func (r *RequestRepository) GetAll(pageOpts interface{}) (*map[string]models.Request, error) {
	// Assert type for pageOptions.
	opts, ok := pageOpts.(*pages.PageOptions)
	if !ok {
		return nil, fmt.Errorf("cannot read the page options at the repository level")
	}

	// Fetch page according to the calling user (in options).
	pagePtrs := pages.GetOnePage(*opts)
	if pagePtrs == (pages.PagePointers{}) || pagePtrs.Requests == nil || (*pagePtrs.Requests) == nil {
		return nil, fmt.Errorf(common.ERR_PAGE_EXPORT_NIL)
	}

	// If zero items were fetched, no need to continue asserting types.
	if len(*pagePtrs.Requests) == 0 {
		return nil, fmt.Errorf("no requests found in the database")
	}

	return pagePtrs.Requests, nil

}*/

// GetRequestByID is a static function to export to other services.
func GetRequestByID(requestID string, cache db.Cacher) (*models.Request, error) {
	if requestID == "" || cache == nil {
		return nil, fmt.Errorf("requestID is blank, or cache is nil")
	}

	// Fetch the request from the cache.
	reqRaw, found := cache.Load(requestID)
	if !found {
		return nil, fmt.Errorf("request not found")
	}

	request, ok := reqRaw.(*models.Request)
	if !ok {
		return nil, fmt.Errorf("could not assert type *models.Request")
	}

	return request, nil
}

func (r *RequestRepository) GetByID(requestID string) (*models.Request, error) {
	// Use the static function to get such request.
	request, err := GetRequestByID(requestID, r.cache)
	if err != nil {
		return nil, err
	}

	return request, nil
}

func (r *RequestRepository) Save(request *models.Request) error {
	// Store the request using its key in the cache.
	saved := r.cache.Store(request.ID, *request)
	if !saved {
		return fmt.Errorf("an error occurred while saving a request")
	}

	return nil
}

func (r *RequestRepository) Delete(requestID string) error {
	// Simple request's deletion.
	deleted := r.cache.Delete(requestID)
	if !deleted {
		return fmt.Errorf("request data could not be purged from the database")
	}

	return nil
}

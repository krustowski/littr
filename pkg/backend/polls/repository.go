package polls

import (
	"fmt"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/models"
)

// The implementation of pkg/models.PollRepositoryInterface.
type PollRepository struct {
	cache db.Cacher
}

func NewPollRepository(cache db.Cacher) models.PollRepositoryInterface {
	if cache == nil {
		return nil
	}

	return &PollRepository{
		cache: cache,
	}
}

func (r *PollRepository) GetAll() (*map[string]models.Poll, error) {
	rawPolls, count := r.cache.Range()
	if count == 0 {
		return nil, fmt.Errorf("no items found")
	}

	polls := make(map[string]models.Poll)

	// Assert types to fetched interface map.
	for key, rawPoll := range *rawPolls {
		poll, ok := rawPoll.(models.Poll)
		if !ok {
			return nil, fmt.Errorf("poll's data corrupted")
		}

		polls[key] = poll
	}

	return &polls, nil
}

func (r *PollRepository) GetByID(pollID string) (*models.Poll, error) {
	// Fetch the poll from the cache.
	rawPoll, found := r.cache.Load(pollID)
	if !found {
		return nil, fmt.Errorf(common.ERR_POLL_NOT_FOUND)
	}

	// Assert the type
	poll, ok := rawPoll.(models.Poll)
	if !ok {
		return nil, fmt.Errorf("poll's data corrupted")
	}

	return &poll, nil
}

func (r *PollRepository) Save(poll *models.Poll) error {
	// Store the poll using its key in the cache.
	saved := r.cache.Store(poll.ID, *poll)
	if !saved {
		return fmt.Errorf("an error occurred while saving a poll")
	}

	return nil
}

func (r *PollRepository) Delete(pollID string) error {
	// Simple poll's deleting.
	deleted := r.cache.Delete(pollID)
	if !deleted {
		return fmt.Errorf("poll data could not be purged from the database")
	}

	return nil
}

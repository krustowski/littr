package polls

import (
	"fmt"

	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/models"
)

// Repository implementation

type PollRepository struct {
	cache db.Cacher
}

func (r *PollRepository) GetAll() ([]*models.Poll, error) {
	rawPolls, count := r.cache.Range()

	if count == 0 {
		return nil, fmt.Errorf("no polls found in the database")
	}

	// Prepare the output array.
	polls := make([]*models.Poll, 0)

	// Assert the model type.
	for _, rawPoll := range *rawPolls {
		poll, ok := rawPoll.(*models.Poll)
		if !ok {
			continue
		}

		polls = append(polls, poll)
	}

	return polls, nil
}

func (r *PollRepository) GetByID(pollID string) (*models.Poll, error) {
	rawPoll, found := r.cache.Load(pollID)
	if !found {
		return nil, fmt.Errorf("requested poll not found")
	}

	// Assert the type
	poll, ok := rawPoll.(*models.Poll)
	if !ok {
		return nil, fmt.Errorf("poll's data corrupted")
	}

	return poll, nil
}

func (r *PollRepository) Save(poll *models.Poll) error {
	saved := r.cache.Store(poll.ID, poll)
	if !saved {
		return fmt.Errorf("an error occurred while saving a poll")
	}

	return nil
}

func (r *PollRepository) Update(poll *models.Poll) error {
	//data := ...

	updated := r.cache.Store(poll.ID, poll)
	if !updated {
		return fmt.Errorf("poll data could not be updated in the database")
	}

	return nil
}

func (r *PollRepository) Delete(pollID string) error {
	deleted := r.cache.Delete(pollID)
	if !deleted {
		return fmt.Errorf("poll data could not be purged from the database")
	}

	return nil
}

//
//
//

func NewPollRepository(cache db.Cacher) db.PollRepositoryInterface {
	return &PollRepository{
		cache: cache,
	}
}

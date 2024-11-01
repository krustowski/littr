package polls

import (
	"fmt"

	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/models"
)

type PollServiceInterface interface {
	Create(poll *models.Poll) error
	Update(poll *models.Poll) error
	Delete(pollID string) error
	FindAll() (polls []*models.Poll, err error)
	FindByID(pollID string) (poll *models.Poll, err error)
}

// PollServiceInterface implementation

type PollService struct {
	pollRepository db.PollRepositoryInterface
}

func (s *PollService) Create(poll *models.Poll) error {

	if err := s.pollRepository.Save(poll); err != nil {
		return fmt.Errorf("failed to save the poll: %s", err.Error())
	}

	return nil
}

func (s *PollService) Update(poll *models.Poll) error {
	err := s.pollRepository.Update(poll)

	return err
}

func (s *PollService) Delete(pollID string) error {
	err := s.pollRepository.Delete(pollID)

	return err
}

func (s *PollService) FindAll() ([]*models.Poll, error) {
	polls, err := s.pollRepository.GetAll()
	if err != nil {
		return nil, err
	}

	return polls, nil
}

func (s *PollService) FindByID(pollID string) (*models.Poll, error) {
	poll, err := s.pollRepository.GetByID(pollID)
	if err != nil {
		return nil, err
	}

	return poll, nil
}

//
//
//

func NewPollService(pollRepository db.PollRepositoryInterface) *PollService {
	return &PollService{
		pollRepository: pollRepository,
	}
}

package polls

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/backend/live"
	"go.vxn.dev/littr/pkg/backend/pages"
	"go.vxn.dev/littr/pkg/models"
)

type PollServiceInterface interface {
	Create(ctx context.Context, poll *models.Poll) error
	Update(ctx context.Context, poll *models.Poll) error
	Delete(ctx context.Context, pollID string) error
	FindAll(ctx context.Context) (*map[string]models.Poll, *models.User, error)
	FindByID(ctx context.Context, pollID string) (*models.Poll, *models.User, error)
}

//
// PollServiceInterface implementation
//

type PollService struct {
	pollRepository db.PollRepositoryInterface
	postRepository db.PostRepositoryInterface
	userRepository db.UserRepositoryInterface
}

func NewPollService(pollRepository db.PollRepositoryInterface, postRepository db.PostRepositoryInterface, userRepository db.UserRepositoryInterface) *PollService {
	return &PollService{
		pollRepository: pollRepository,
		postRepository: postRepository,
		userRepository: userRepository,
	}
}

func (s *PollService) Create(ctx context.Context, poll *models.Poll) error {
	// Fetch the caller's ID from the context.
	callerID, ok := ctx.Value("callerID").(string)
	if !ok {
		return fmt.Errorf(common.ERR_CALLER_FAIL)
	}

	// To patch loaded invalid user data from LocalStorage = caller's the Author now.
	if poll.Author == "" {
		poll.Author = callerID
	}

	// The caller must be the author of such poll to be added.
	if poll.Author != callerID {
		return fmt.Errorf(common.ERR_POLL_AUTHOR_MISMATCH)
	}

	// Patch the recurring option value (every option has to be unique).
	if (poll.OptionOne == poll.OptionTwo) || (poll.OptionOne == poll.OptionThree) || (poll.OptionTwo == poll.OptionThree) || (poll.Question == poll.OptionOne.Content) || (poll.Question == poll.OptionTwo.Content) || (poll.Question == poll.OptionThree.Content) {
		return fmt.Errorf(common.ERR_POLL_DUPLICIT_OPTIONS)
	}

	//
	//  Validation end --- dispatch the poll to repository.
	//

	if err := s.pollRepository.Save(poll); err != nil {
		return fmt.Errorf(common.ERR_POLL_SAVE_FAIL + ": " + err.Error())
	}

	// Prepare timestamps for a new system post to flow.
	postStamp := time.Now()
	postKey := strconv.FormatInt(postStamp.UnixNano(), 10)

	post := &models.Post{
		ID:        postKey,
		Type:      "poll",
		Nickname:  "system",
		Content:   "new poll has been added",
		PollID:    poll.ID,
		Timestamp: postStamp,
	}

	// Dispatch the new post aboat a new poll to postRepository.
	if err := s.postRepository.Save(post); err != nil {
		return fmt.Errorf(common.ERR_POLL_POST_FAIL + ": " + err.Error())
	}

	// Broadcast the new poll event.
	live.BroadcastMessage(live.EventPayload{Data: "poll," + poll.ID, Type: "message"})

	return nil
}

func (s *PollService) Update(ctx context.Context, poll *models.Poll) error {
	//err := s.pollRepository.Store(poll)

	return fmt.Errorf("not yet implemented")
}

func (s *PollService) Delete(ctx context.Context, pollID string) error {
	err := s.pollRepository.Delete(pollID)

	return err
}

func (s *PollService) FindAll(ctx context.Context) (*map[string]models.Poll, *models.User, error) {
	// Fetch the caller's ID from the context.
	callerID, ok := ctx.Value("nickname").(string)
	if !ok {
		return nil, nil, fmt.Errorf(common.ERR_CALLER_FAIL)
	}

	// Fetch the pageNo from the context.
	pageNo, ok := ctx.Value("pageNo").(int)
	if !ok {
		return nil, nil, fmt.Errorf(common.ERR_PAGENO_INCORRECT)
	}

	// Compose a pagination options object to paginate polls.
	opts := &pages.PageOptions{
		CallerID: callerID,
		PageNo:   pageNo,
		FlowList: nil,

		Polls: pages.PollOptions{
			Plain: true,
		},
	}

	// Request the page of polls from the poll repository.
	polls, err := s.pollRepository.GetAll(opts)
	if err != nil {
		return nil, nil, err
	}

	// Request the caller from the user repository.
	/*caller, err := s.userRepository.GetByID(callerID)
	if err != nil {
		return nil, nil, err
	}*/

	// Hide foreign poll's authors and voters.
	for key, poll := range *polls {
		var votedList []string

		// Loop over voters, anonymize them.
		for _, voter := range poll.Voted {
			if voter == callerID {
				votedList = append(votedList, callerID)
			} else {
				votedList = append(votedList, "voter")
			}
		}

		poll.Voted = votedList

		if poll.Author == callerID {
			continue
		}

		poll.Author = ""
		//(*pagePtrs.Polls)[key] = poll
		(*polls)[key] = poll
	}

	return polls, nil, nil
}

func (s *PollService) FindByID(ctx context.Context, pollID string) (*models.Poll, *models.User, error) {
	// Fetch the caller's ID from the context.
	/*callerID, ok := ctx.Value("nickname").(string)
	if !ok {
		return nil, nil, fmt.Errorf(common.ERR_CALLER_FAIL)
	}*/

	// Fetch the poll.
	poll, err := s.pollRepository.GetByID(pollID)
	if err != nil {
		return nil, nil, err
	}

	// Request the caller from the user repository.
	/*caller, err := s.userRepository.GetByID(callerID)
	if err != nil {
		return nil, nil, err
	}*/

	return poll, nil, nil
}

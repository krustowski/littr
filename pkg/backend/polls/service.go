package polls

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/live"
	"go.vxn.dev/littr/pkg/backend/pages"
	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"
)

//
// models.PollServiceInterface implementation
//

type PollService struct {
	pageService    models.PagingServiceInterface
	pollRepository models.PollRepositoryInterface
	postRepository models.PostRepositoryInterface
	userRepository models.UserRepositoryInterface
}

func NewPollService(
	pageService models.PagingServiceInterface,
	pollRepository models.PollRepositoryInterface,
	postRepository models.PostRepositoryInterface,
	userRepository models.UserRepositoryInterface,
) models.PollServiceInterface {
	if pageService == nil || pollRepository == nil || postRepository == nil || userRepository == nil {
		return nil
	}

	return &PollService{
		pageService:    pageService,
		pollRepository: pollRepository,
		postRepository: postRepository,
		userRepository: userRepository,
	}
}

func (s *PollService) Create(ctx context.Context, pollI interface{}) error {
	// Fetch the caller's ID from the context.
	callerID := common.GetCallerID(ctx)

	pollReq, ok := pollI.(*PollCreateRequest)
	if !ok {
		return fmt.Errorf("could not assert type for the poll creation request")
	}

	poll := new(models.Poll)

	poll.Author = callerID

	// The caller must be the author of such poll to be added.
	if poll.Author != callerID {
		return fmt.Errorf(common.ERR_POLL_AUTHOR_MISMATCH)
	}

	poll.Question = pollReq.Question
	poll.OptionOne.Content = pollReq.OptionOne
	poll.OptionTwo.Content = pollReq.OptionTwo
	poll.OptionThree.Content = pollReq.OptionThree

	// Patch the recurring option value (every option has to be unique).
	if (poll.OptionOne.Content == poll.OptionTwo.Content) ||
		(poll.OptionOne.Content == poll.OptionThree.Content) ||
		(poll.OptionTwo.Content == poll.OptionThree.Content) ||
		(poll.Question == poll.OptionOne.Content) ||
		(poll.Question == poll.OptionTwo.Content) ||
		(poll.Question == poll.OptionThree.Content) {

		return fmt.Errorf(common.ERR_POLL_DUPLICIT_OPTIONS)
	}

	//
	//  Validation end --- dispatch the poll to repository.
	//

	poll.Timestamp = time.Now()
	poll.ID = strconv.FormatInt(poll.Timestamp.UnixNano(), 10)

	if err := s.pollRepository.Save(poll); err != nil {
		return err
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
		return err
	}

	// Broadcast the new poll event.
	live.BroadcastMessage(live.EventPayload{Data: "poll," + poll.ID, Type: "message"})

	return nil
}

func (s *PollService) Update(ctx context.Context, pollI interface{}) error {
	// Fetch the caller's ID from the context.
	callerID := common.GetCallerID(ctx)

	// Assert type for the request data.
	pollReq, ok := pollI.(*PollUpdateRequest)
	if !ok {
		return fmt.Errorf("could not assert type for the poll update request")
	}

	// Fetch the actual poll to verify its content to be updated..
	dbPoll, err := s.pollRepository.GetByID(pollReq.ID)
	if err != nil {
		return err
	}

	// Check the poll's ownership. The author cannot vote on such poll.
	if dbPoll.Author == callerID {
		return fmt.Errorf(common.ERR_POLL_SELF_VOTE)
	}

	// Has the callerID already voted?
	if helpers.Contains(dbPoll.Voted, callerID) {
		return fmt.Errorf(common.ERR_POLL_EXISTING_VOTE)
	}

	// Verify that only one vote had been passed in; suppress vote count forgery.
	if (pollReq.OptionOneCount + pollReq.OptionTwoCount + pollReq.OptionThreeCount) != (dbPoll.OptionOne.Counter + dbPoll.OptionTwo.Counter + dbPoll.OptionThree.Counter + 1) {
		return fmt.Errorf(common.ERR_POLL_INVALID_VOTE_COUNT)
	}

	// Now, update the poll's data.
	dbPoll.Voted = append(dbPoll.Voted, callerID)
	dbPoll.OptionOne.Counter = pollReq.OptionOneCount
	dbPoll.OptionTwo.Counter = pollReq.OptionTwoCount
	dbPoll.OptionThree.Counter = pollReq.OptionThreeCount

	// Save the changes in repository.
	return s.pollRepository.Save(dbPoll)
}

func (s *PollService) Delete(ctx context.Context, pollID string) error {
	// Fetch the caller's ID from the context.
	callerID := common.GetCallerID(ctx)

	// Fetch the actual poll to verify it can be deleted at all.
	poll, err := s.pollRepository.GetByID(pollID)
	if err != nil {
		return err
	}

	// Check the poll's ownership.
	if poll.Author != callerID {
		return fmt.Errorf(common.ERR_POLL_DELETE_FOREIGN)
	}

	// Try to delete the poll.
	return s.pollRepository.Delete(pollID)
}

func (s *PollService) FindAll(ctx context.Context, pageOpts interface{}) (*map[string]models.Poll, *models.User, error) {
	// Fetch the caller's ID from the context.
	callerID := common.GetCallerID(ctx)

	req, ok := pageOpts.(*PollPagingRequest)
	if !ok {
		return nil, nil, fmt.Errorf(common.ERR_PAGENO_INCORRECT)
	}

	// Compose a pagination options object to paginate polls.
	opts := &pages.PageOptions{
		CallerID: callerID,
		PageNo:   req.PageNo,
		FlowList: nil,

		Polls: pages.PollOptions{
			Plain: true,
		},
	}

	allPolls, err := s.pollRepository.GetAll()
	if err != nil {
		return nil, nil, err
	}

	iface, err := s.pageService.GetOne(ctx, opts, allPolls)
	if err != nil {
		return nil, nil, err
	}

	ptrs, ok := iface.(pages.PagePointers)
	if !ok {
		return nil, nil, fmt.Errorf("cannot assert type map of polls")
	}

	// Request the caller from the user repository.
	caller, err := s.userRepository.GetByID(callerID)
	if err != nil {
		return nil, nil, err
	}

	// Patch the polls' data for export.
	polls := hidePollAuthorAndVoters(ptrs.Polls, callerID)

	// Patch the user's data for export.
	patchedCaller := (*common.FlushUserData(&map[string]models.User{callerID: *caller}, callerID))[callerID]

	return polls, &patchedCaller, nil
}

func (s *PollService) FindByID(ctx context.Context, pollID string) (*models.Poll, *models.User, error) {
	// Fetch the caller's ID from the context.
	callerID := common.GetCallerID(ctx)

	// Fetch the poll.
	poll, err := s.pollRepository.GetByID(pollID)
	if err != nil {
		return nil, nil, err
	}

	// Request the caller from the user repository.
	caller, err := s.userRepository.GetByID(callerID)
	if err != nil {
		return nil, nil, err
	}

	// Patch the polls' data for export.
	patchedPoll := (*hidePollAuthorAndVoters(&map[string]models.Poll{pollID: *poll}, callerID))[pollID]

	// Patch the user's data for export.
	patchedCaller := (*common.FlushUserData(&map[string]models.User{callerID: *caller}, callerID))[callerID]

	return &patchedPoll, &patchedCaller, nil
}

//
//  Helpers
//

func hidePollAuthorAndVoters(polls *map[string]models.Poll, callerID string) *map[string]models.Poll {
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

		// Return new voters list to such poll.
		poll.Voted = votedList

		// Hide poll's author.
		if poll.Author != callerID {
			poll.Author = ""
		}

		(*polls)[key] = poll
	}

	return polls
}

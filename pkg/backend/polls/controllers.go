package polls

import (
	"net/http"
	"strconv"
	"time"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/backend/live"
	"go.vxn.dev/littr/pkg/backend/pages"
	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"

	chi "github.com/go-chi/chi/v5"
)

// getPolls get a list of polls
//
// @Summary      Get a list of polls
// @Description  get a list of polls
// @Tags         polls
// @Accept       json
// @Produce      json
// @X-Page-No    {"pageNo": 0}
// @Param    	 X-Page-No header string true "page number"
// @Success      200  {object}   common.APIResponse{data=polls.getPolls.responseData}
// @Failure      400  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /polls [get]
func getPolls(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "polls")

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	type responseData struct {
		Polls map[string]models.Poll `json:"polls,omitempty"`
		User  models.User            `json:"user,omitempty"`
	}

	pl := responseData{}

	pageNoString := r.Header.Get("X-Page-No")
	pageNo, err := strconv.Atoi(pageNoString)
	if err != nil {
		l.Msg(common.ERR_PAGENO_INCORRECT).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	opts := pages.PageOptions{
		CallerID: callerID,
		PageNo:   pageNo,
		FlowList: nil,

		Polls: pages.PollOptions{
			Plain: true,
		},
	}

	// fetch page according to the logged user
	pagePtrs := pages.GetOnePage(opts)
	if pagePtrs == (pages.PagePointers{}) || pagePtrs.Polls == nil || (*pagePtrs.Polls) == nil {
		l.Msg(common.ERR_PAGE_EXPORT_NIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// hide foreign poll's authors and voters
	for key, poll := range *pagePtrs.Polls {
		var votedList []string

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
		(*pagePtrs.Polls)[key] = poll
	}

	// hack: include caller's models.User struct
	if pl.User, ok = db.GetOne(db.UserCache, callerID, models.User{}); !ok {
		l.Msg(common.ERR_CALLER_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	pl.Polls = *pagePtrs.Polls

	l.Msg("ok, dumping polls").Status(http.StatusOK).Log().Payload(&pl).Write(w)
	return
}

// addNewPoll ensures a new polls is created and saved
//
// @Summary      Add new poll
// @Description  add new poll
// @Tags         polls
// @Accept       json
// @Produce      json
// @Param    	 request body models.Poll true "query params"
// @Success      201  {object}  common.APIResponse "success"
// @Failure      400  {object}  common.APIResponse "bad/malformed input data, invalid cookies"
// @Failure      500  {object}  common.APIResponse "the poll saving process failed"
// @Router       /polls [post]
func addNewPoll(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "polls")

	// get the caller's nickname from context
	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var poll models.Poll

	// decode received data
	if err := common.UnmarshalRequestData(r, &poll); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// to patch wrongly loaded user data from LocalStorage
	if poll.Author == "" {
		poll.Author = callerID
	}

	key := poll.ID

	// caller must be the author of such poll to be added
	if poll.Author != callerID {
		l.Msg(common.ERR_POLL_AUTHOR_MISMATCH).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	if saved := db.SetOne(db.PollCache, key, poll); !saved {
		l.Msg(common.ERR_POLL_SAVE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// prepare timestamps for a new system post to flow
	postStamp := time.Now()
	postKey := strconv.FormatInt(postStamp.UnixNano(), 10)

	post := models.Post{
		ID:        postKey,
		Type:      "poll",
		Nickname:  "system",
		Content:   "new poll has been added",
		Timestamp: postStamp,
	}

	if saved := db.SetOne(db.FlowCache, postKey, post); !saved {
		l.Msg(common.ERR_POLL_POST_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// broadcast the new poll event
	live.BroadcastMessage("poll", "message")

	l.Msg("ok, adding new poll").Status(http.StatusCreated).Log().Payload(nil).Write(w)
	return
}

// getSinglePoll
//
// @Summary      Get single poll
// @Description  get single poll
// @Tags         polls
// @Produce      json
// @Param        pollID path string true "poll ID"
// @Success      200  {object}  common.APIResponse{data=polls.getSinglePoll.responseData}
// @Failure      400  {object}  common.APIResponse
// @Failure      404  {object}  common.APIResponse
// @Router       /polls/{pollID} [get]
func getSinglePoll(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "polls")

	// get the caller's nickname from context
	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var caller models.User

	// fetch caller's data object
	if caller, ok = db.GetOne(db.UserCache, callerID, models.User{}); !ok {
		l.Msg(common.ERR_CALLER_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	// take the param from path
	pollID := chi.URLParam(r, "pollID")
	if pollID == "" {
		l.Msg(common.ERR_POLLID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// prepare vars for the requested poll's fetch
	var poll models.Poll
	var found bool

	if poll, found = db.GetOne(db.PollCache, pollID, models.Poll{}); !found {
		l.Msg(common.ERR_POLL_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	type responseData struct {
		Poll models.Poll `json:"poll"`
		User models.User `json:"user"`
	}

	pl := responseData{Poll: poll, User: caller}

	l.Msg("ok, dumping single poll").Status(http.StatusOK).Log().Payload(&pl).Write(w)
	return
}

// updatePoll updates a given poll
//
// @Summary      Update a poll
// @Description  update a poll
// @Tags         polls
// @Accept       json
// @Produce      json
// @Param    	 updatedPoll body models.Poll true "update poll's body"
// @Param        pollID path string true "poll's ID"
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /polls/{pollID} [put]
func updatePoll(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "polls")

	// get the caller's nickname from context
	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var payload models.Poll

	// decode received data
	if err := common.UnmarshalRequestData(r, &payload); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	key := payload.ID

	var poll models.Poll
	var found bool

	// fetch the poll from database for comparison
	if poll, found = db.GetOne(db.PollCache, key, models.Poll{}); !found {
		l.Msg(common.ERR_POLL_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	// caller must not be the author of such poll to be voted on
	if poll.Author == callerID {
		l.Msg(common.ERR_POLL_SELF_VOTE).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	// has the caller already voted?
	if helpers.Contains(poll.Voted, callerID) {
		l.Msg(common.ERR_POLL_EXISTING_VOTE).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	// verify that only one vote had been passed in; supress vote count forgery
	if (payload.OptionOne.Counter + payload.OptionTwo.Counter + payload.OptionThree.Counter) != (poll.OptionOne.Counter + poll.OptionTwo.Counter + poll.OptionThree.Counter + 1) {
		l.Msg(common.ERR_POLL_INVALID_VOTE_COUNT).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	// now, update the poll's data
	poll.Voted = append(poll.Voted, l.CallerID)
	poll.OptionOne = payload.OptionOne
	poll.OptionTwo = payload.OptionTwo
	poll.OptionThree = payload.OptionThree

	// update the poll in database
	if saved := db.SetOne(db.PollCache, key, poll); !saved {
		l.Msg(common.ERR_POLL_SAVE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, poll updated").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// deletePoll removes a poll
//
// @Summary      Delete a poll by ID
// @Description  delete a poll by ID
// @Tags         polls
// @Produce      json
// @Param        pollID path string true "poll's ID to delete"
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      403  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /polls/{pollID} [delete]
func deletePoll(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "polls")

	// get caller's name
	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// take the param from path
	pollID := chi.URLParam(r, "pollID")
	if pollID == "" {
		l.Msg(common.ERR_POLLID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// fetch the poll from database for comparison
	poll, found := db.GetOne(db.PollCache, key, models.Poll{})
	if !found {
		l.Msg(common.ERR_POLL_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	// check for possible poll's deletion forgery
	if poll.Author != callerID {
		l.Msg(common.ERR_POLL_DELETE_FOREIGN).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	// delete requested poll
	if deleted := db.DeleteOne(db.PollCache, pollID); !deleted {
		l.Msg(common.ERR_POLL_DELETE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, poll removed").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

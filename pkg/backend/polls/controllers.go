package polls

import (
	"net/http"
	"strconv"
	"time"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/backend/pages"
	"go.vxn.dev/littr/pkg/backend/posts"
	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"

	sse "github.com/alexandrevicenzi/go-sse"
)

// getPolls get a list of polls
//
// @Summary      Get a list of polls
// @Description  get a list of polls
// @Tags         polls
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.APIResponse{Data=polls.getPolls.payload}
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

	//resp := common.Response{}
	type payload struct {
		Polls map[string]models.Poll `json:"polls,omitempty"`
		User  models.User            `json:"user,omitempty"`
	}

	pl := payload{}

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

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var poll models.Poll

	if err := common.UnmarshalRequestData(r, &poll); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// patch wrongly loaded user data from LocalStorage
	if poll.Author == "" {
		poll.Author = callerID
	}

	key := poll.ID

	if poll.Author != callerID {
		l.Msg(common.ERR_POLL_AUTHOR_MISMATCH).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	if saved := db.SetOne(db.PollCache, key, poll); !saved {
		l.Msg(common.ERR_POLL_SAVE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

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

	if posts.Streamer != nil {
		posts.Streamer.SendMessage("/api/v1/posts/live", sse.SimpleMessage("poll"))
	}

	l.Msg("ok, adding new poll").Status(http.StatusCreated).Log().Payload(nil).Write(w)
	return
}

// getSinglePoll
//
// @Summary      Get single poll
// @Description  get single poll
// @Tags         polls
// @Accept       json
// @Produce      json
// @Param        pollID path string true "poll ID"
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /polls/{pollID} [get]
func getSinglePoll(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "polls")

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var payload models.Poll

	if err := common.UnmarshalRequestData(r, &payload); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	key := payload.ID

	var poll models.Poll
	var found bool

	if poll, found = db.GetOne(db.PollCache, key, models.Poll{}); !found {
		l.Msg(common.ERR_POLL_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	if poll.Author == callerID {
		l.Msg(common.ERR_POLL_SELF_VOTE).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	if helpers.Contains(poll.Voted, callerID) {
		l.Msg(common.ERR_POLL_EXISTING_VOTE).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	poll.Voted = append(poll.Voted, l.CallerID)
	poll.OptionOne = payload.OptionOne
	poll.OptionTwo = payload.OptionTwo
	poll.OptionThree = payload.OptionThree

	if saved := db.SetOne(db.PollCache, key, poll); !saved {
		l.Msg(common.ERR_POLL_SAVE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, poll updated").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// updatePoll updates a given poll
//
// @Summary      Update a poll
// @Description  update a poll
// @Tags         polls
// @Accept       json
// @Produce      json
// @Param        pollID path string true "poll ID"
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /polls/{pollID} [put]
func updatePoll(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "polls")

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var payload models.Poll

	if err := common.UnmarshalRequestData(r, &payload); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	key := payload.ID

	var poll models.Poll
	var found bool

	if poll, found = db.GetOne(db.PollCache, key, models.Poll{}); !found {
		l.Msg(common.ERR_POLL_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	if poll.Author == callerID {
		l.Msg(common.ERR_POLL_SELF_VOTE).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	if helpers.Contains(poll.Voted, callerID) {
		l.Msg(common.ERR_POLL_EXISTING_VOTE).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	poll.Voted = append(poll.Voted, l.CallerID)
	poll.OptionOne = payload.OptionOne
	poll.OptionTwo = payload.OptionTwo
	poll.OptionThree = payload.OptionThree

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
// @Accept       json
// @Produce      json
// @Param        pollID path string true "poll ID"
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /polls/{pollID} [delete]
func deletePoll(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "polls")

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var poll models.Poll

	if err := common.UnmarshalRequestData(r, &poll); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	if poll.Author != callerID {
		l.Msg(common.ERR_POLL_DELETE_FOREIGN).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	key := poll.ID

	if deleted := db.DeleteOne(db.PollCache, key); !deleted {
		l.Msg(common.ERR_POLL_DELETE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, poll removed").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

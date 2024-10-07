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

type stub struct{}

// getPolls get a list of polls
//
// @Summary      Get a list of polls
// @Description  get a list of polls
// @Tags         polls
// @Accept       json
// @Produce      json
// @Success      200  {object}   polls.getPolls.response
// @Failure      400  {object}   polls.getPolls.response{polls=stub,user=stub}
// @Failure      500  {object}   polls.getPolls.response{polls=stub,user=stub}
// @Router       /polls [get]
func getPolls(w http.ResponseWriter, r *http.Request) {
	//resp := common.Response{}
	l := common.NewLogger(r, "polls")
	callerID, _ := r.Context().Value("nickname").(string)

	type response struct {
		Code    int                    `json:"code"`
		Message string                 `json:"message"`
		Polls   map[string]models.Poll `json:"polls,omitempty"`
		User    models.User            `json:"user,omitempty"`
	}
	resp := new(response)

	pageNoString := r.Header.Get("X-Page-No")
	pageNo, err := strconv.Atoi(pageNoString)
	if err != nil {
		resp.Message = "page No has to be specified as integer/number"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		common.WriteResponse(w, resp, resp.Code)
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
		resp.Message = "error while requesting more pages, one exported map is nil!"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		common.WriteResponse(w, resp, resp.Code)
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

	var ok bool

	// hack: include caller's models.User struct
	if resp.User, ok = db.GetOne(db.UserCache, callerID, models.User{}); !ok {
		resp.Message = "cannot fetch such callerID-named user"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		common.WriteResponse(w, resp, resp.Code)
		return
	}

	resp.Message = "ok, dumping polls"
	resp.Code = http.StatusOK
	resp.Polls = *pagePtrs.Polls

	l.Println(resp.Message, resp.Code)
	common.WriteResponse(w, resp, resp.Code)
}

// addNewPoll ensures a new polls is created and saved
//
// @Summary      Add new poll
// @Description  add new poll
// @Tags         polls
// @Accept       json
// @Produce      json
// @Success      201  {object}  common.Response
// @Failure      400  {object}  common.Response
// @Failure      500  {object}  common.Response
// @Router       /polls [post]
func addNewPoll(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "polls")
	callerID, _ := r.Context().Value("nickname").(string)

	var poll models.Poll

	if err := common.UnmarshalRequestData(r, &poll); err != nil {
		resp.Message = "input read error: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if poll.Author == "" {
		poll.Author = callerID
	}

	key := poll.ID

	if poll.Author != callerID {
		resp.Message = "backend error: such poll's author differs from callerID"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if saved := db.SetOne(db.PollCache, key, poll); !saved {
		resp.Message = "backend error: cannot save new poll (cache error)"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
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
		resp.Message = "backend error: cannot create a new post about new poll creation"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	posts.Streamer.SendMessage("/api/v1/posts/live", sse.SimpleMessage("poll"))

	resp.Message = "ok, adding new poll"
	resp.Code = http.StatusCreated

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

// getSinglePoll
//
// @Summary      Get single poll
// @Description  get single poll
// @Tags         polls
// @Accept       json
// @Produce      json
// @Param        pollID path string true "poll ID"
// @Success      200  {object}  common.Response
// @Failure      400  {object}  common.Response
// @Failure      500  {object}  common.Response
// @Router       /polls/{pollID} [get]
func getSinglePoll(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "polls")
	callerID, _ := r.Context().Value("nickname").(string)

	var payload models.Poll

	if err := common.UnmarshalRequestData(r, &payload); err != nil {
		resp.Message = "input read error: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	key := payload.ID

	var poll models.Poll
	var found bool

	if poll, found = db.GetOne(db.PollCache, key, models.Poll{}); !found {
		resp.Message = "unknown poll update requested"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if poll.Author == callerID {
		resp.Message = "you cannot vote on your own poll"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if helpers.Contains(poll.Voted, callerID) {
		resp.Message = "this user already voted on such poll"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	poll.Voted = append(poll.Voted, l.CallerID)
	poll.OptionOne = payload.OptionOne
	poll.OptionTwo = payload.OptionTwo
	poll.OptionThree = payload.OptionThree

	if saved := db.SetOne(db.PollCache, key, poll); !saved {
		resp.Message = "cannot update poll"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "ok, poll updated"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

// updatePoll updates a given poll
//
// @Summary      Update a poll
// @Description  update a poll
// @Tags         polls
// @Accept       json
// @Produce      json
// @Param        pollID path string true "poll ID"
// @Success      200  {object}  common.Response
// @Failure      400  {object}  common.Response
// @Failure      500  {object}  common.Response
// @Router       /polls/{pollID} [put]
func updatePoll(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "polls")
	callerID, _ := r.Context().Value("nickname").(string)

	var payload models.Poll

	if err := common.UnmarshalRequestData(r, &payload); err != nil {
		resp.Message = "input read error: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	key := payload.ID

	var poll models.Poll
	var found bool

	if poll, found = db.GetOne(db.PollCache, key, models.Poll{}); !found {
		resp.Message = "unknown poll update requested"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if poll.Author == callerID {
		resp.Message = "you cannot vote on your own poll"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if helpers.Contains(poll.Voted, callerID) {
		resp.Message = "this user already voted on such poll"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	poll.Voted = append(poll.Voted, l.CallerID)
	poll.OptionOne = payload.OptionOne
	poll.OptionTwo = payload.OptionTwo
	poll.OptionThree = payload.OptionThree

	if saved := db.SetOne(db.PollCache, key, poll); !saved {
		resp.Message = "cannot update poll"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "ok, poll updated"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

// deletePoll removes a poll
//
// @Summary      Delete a poll by ID
// @Description  delete a poll by ID
// @Tags         polls
// @Accept       json
// @Produce      json
// @Param        pollID path string true "poll ID"
// @Success      200  {object}  common.Response
// @Failure      400  {object}  common.Response
// @Failure      500  {object}  common.Response
// @Router       /polls/{pollID} [delete]
func deletePoll(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	callerID, _ := r.Context().Value("nickname").(string)
	l := common.NewLogger(r, "polls")

	var poll models.Poll

	if err := common.UnmarshalRequestData(r, &poll); err != nil {
		resp.Message = "input read error: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if poll.Author != callerID {
		resp.Message = "one can only delete their own poll"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	key := poll.ID

	// check for possible poll forgery
	//if caller != poll.Author {}

	if deleted := db.DeleteOne(db.PollCache, key); !deleted {
		resp.Message = "cannot delete the poll"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "ok, poll removed"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

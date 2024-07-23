package polls

import (
	"net/http"
	"strconv"
	"time"

	"go.savla.dev/littr/pkg/backend/common"
	"go.savla.dev/littr/pkg/backend/db"
	"go.savla.dev/littr/pkg/backend/posts"
	"go.savla.dev/littr/pkg/helpers"
	"go.savla.dev/littr/pkg/models"

	sse "github.com/alexandrevicenzi/go-sse"
)

// getPolls get a list of polls
//
// @Summary      Get a list of polls
// @Description  get a list of polls
// @Tags         polls
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.Response
// @Router       /polls/ [get]
func getPolls(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "polls")
	callerID, _ := r.Context().Value("nickname").(string)

	polls, _ := db.GetAll(db.PollCache, models.Poll{})

	// hide poll's author
	for key, poll := range polls {
		if poll.Author == callerID {
			continue
		}

		poll.Author = ""
		polls[key] = poll
	}

	uExport := make(map[string]models.User)

	// hack: include caller's models.User struct
	if caller, ok := db.GetOne(db.UserCache, callerID, models.User{}); !ok {
		resp.Message = "cannot fetch such callerID-named user"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	} else {
		uExport[callerID] = caller
	}

	// TODO: use DTO
	for key, user := range uExport {
		user.Passphrase = ""
		user.PassphraseHex = ""
		user.Email = ""

		if user.Nickname != callerID {
			user.FlowList = nil
			user.ShadeList = nil
			user.RequestList = nil
		}

		uExport[key] = user
	}

	resp.Message = "ok, dumping polls"
	resp.Code = http.StatusOK
	resp.Polls = polls

	resp.Users = uExport
	resp.Key = callerID

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
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
// @Router       /polls/ [post]
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

// updatePoll updates a given poll
//
// @Summary      Update a poll
// @Description  update a poll
// @Tags         polls
// @Accept       json
// @Produce      json
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
		resp.Message = "cannot update post"
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
// @Summary      Delete a poll
// @Description  delete a poll
// @Tags         polls
// @Accept       json
// @Produce      json
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

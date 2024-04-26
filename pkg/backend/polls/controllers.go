package polls

import (
	"net/http"

	"go.savla.dev/littr/pkg/backend/common"
	"go.savla.dev/littr/pkg/backend/db"
	"go.savla.dev/littr/pkg/models"
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

	polls, _ := db.GetAll(db.PollCache, models.Poll{})

	resp.Message = "ok, dumping polls"
	resp.Code = http.StatusOK
	resp.Polls = polls

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

	var poll models.Poll

	if err := common.UnmarshalRequestData(r, &poll); err != nil {
		resp.Message = "input read error: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	key := poll.ID

	if saved := db.SetOne(db.PollCache, key, poll); !saved {
		resp.Message = "backend error: cannot save new poll (cache error)"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

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
	l := common.NewLogger(r, "polls")

	var poll models.Poll

	if err := common.UnmarshalRequestData(r, &poll); err != nil {
		resp.Message = "input read error: " + err.Error()
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

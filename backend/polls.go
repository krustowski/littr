package backend

import (
	"encoding/json"
	"io"
	"net/http"

	"go.savla.dev/littr/models"
)

// getPolls get a list of polls
//
//	@Summary      Get a list of polls
//	@Description  get a list of polls
//	@Tags         polls
//	@Accept       json
//	@Produce      json
//	@Success      200  {object}   response
//	@Router       /polls/ [get]
func getPolls(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "polls")

	polls, _ := getAll(PollCache, models.Poll{})

	resp.Message = "ok, dumping polls"
	resp.Code = http.StatusOK
	resp.Polls = polls

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

// addNewPoll ensures a new polls is created and saved
//
//	@Summary      Add new poll
//	@Description  add new poll
//	@Tags         polls
//	@Accept       json
//	@Produce      json
//	@Success      201  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /polls/ [post]
func addNewPoll(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "polls")

	var poll models.Poll

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if err := json.Unmarshal(reqBody, &poll); err != nil {
		resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	key := poll.ID

	if saved := setOne(PollCache, key, poll); !saved {
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
//	@Summary      Update a poll
//	@Description  update a poll
//	@Tags         polls
//	@Accept       json
//	@Produce      json
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /polls/ [put]
func updatePoll(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "polls")

	var payload models.Poll

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if err := json.Unmarshal(reqBody, &payload); err != nil {
		resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	key := payload.ID

	var poll models.Poll
	var found bool

	if poll, found = getOne(PollCache, key, models.Poll{}); !found {
		resp.Message = "unknown poll update requested"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	poll.Voted = append(poll.Voted, l.CallerID)

	if saved := setOne(PollCache, key, poll); !saved {
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
//	@Summary      Delete a poll
//	@Description  delete a poll
//	@Tags         polls
//	@Accept       json
//	@Produce      json
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /polls/ [delete]
func deletePoll(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "polls")

	var poll models.Poll

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if err := json.Unmarshal(reqBody, &poll); err != nil {
		resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	key := poll.ID

	if deleted := deleteOne(PollCache, key); !deleted {
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

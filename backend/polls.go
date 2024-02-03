package backend

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"
)

func getPolls(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := Logger{
		CallerID:  r.Header.Get("X-API-Caller-ID"),
		IPAddress: r.Header.Get("X-Real-IP"),
		//IPAddress:  r.RemoteAddr,
		Method:     r.Method,
		Route:      r.URL.String(),
		WorkerName: "polls",
		Version:    r.Header.Get("X-App-Version"),
	}
	noteUsersActivity(r.Header.Get("X-API-Caller-ID"))

	polls, _ := getAll(PollCache, models.Poll{})

	resp.Message = "ok, dumping polls"
	resp.Code = http.StatusOK
	resp.Polls = polls

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

func addNewPoll(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := Logger{
		CallerID:  r.Header.Get("X-API-Caller-ID"),
		IPAddress: r.Header.Get("X-Real-IP"),
		//IPAddress:  r.RemoteAddr,
		Method:     r.Method,
		Route:      r.URL.String(),
		WorkerName: "polls",
		Version:    r.Header.Get("X-App-Version"),
	}
	noteUsersActivity(r.Header.Get("X-API-Caller-ID"))

	var poll models.Poll

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		break
	}

	data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

	if err := json.Unmarshal(data, &poll); err != nil {
		resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		break
	}

	key := poll.ID

	if saved := setOne(PollCache, key, poll); !saved {
		resp.Message = "backend error: cannot save new poll (cache error)"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		break
	}

	resp.Message = "ok, adding new poll"
	resp.Code = http.StatusCreated

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

func updatePoll(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := Logger{
		CallerID:  r.Header.Get("X-API-Caller-ID"),
		IPAddress: r.Header.Get("X-Real-IP"),
		//IPAddress:  r.RemoteAddr,
		Method:     r.Method,
		Route:      r.URL.String(),
		WorkerName: "polls",
		Version:    r.Header.Get("X-App-Version"),
	}
	noteUsersActivity(r.Header.Get("X-API-Caller-ID"))

	var poll models.Poll

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		break
	}

	data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

	if err := json.Unmarshal(data, &poll); err != nil {
		resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		break
	}

	key := poll.ID

	if _, found := getOne(PollCache, key, models.Poll{}); !found {
		resp.Message = "unknown poll update requested"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		break
	}

	if saved := setOne(PollCache, key, poll); !saved {
		resp.Message = "cannot update post"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		break
	}

	resp.Message = "ok, poll updated"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

func deletePoll(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := Logger{
		CallerID:  r.Header.Get("X-API-Caller-ID"),
		IPAddress: r.Header.Get("X-Real-IP"),
		//IPAddress:  r.RemoteAddr,
		Method:     r.Method,
		Route:      r.URL.String(),
		WorkerName: "polls",
		Version:    r.Header.Get("X-App-Version"),
	}
	noteUsersActivity(r.Header.Get("X-API-Caller-ID"))

	var poll models.Poll

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		break
	}

	data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

	if err := json.Unmarshal(data, &poll); err != nil {
		resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		break
	}

	key := poll.ID

	if deleted := deleteOne(PollCache, key); !deleted {
		resp.Message = "cannot delete the poll"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		break
	}

	resp.Message = "ok, poll removed"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

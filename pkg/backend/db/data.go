package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/models"
	"go.vxn.dev/swis/v5/pkg/core"
)

const (
	pollsFile         = "/opt/data/polls.json"
	postsFile         = "/opt/data/posts.json"
	requestsFile      = "/opt/data/requests.json"
	subscriptionsFile = "/opt/data/subscriptions.json"
	tokensFile        = "/opt/data/tokens.json"
	usersFile         = "/opt/data/users.json"
	void              = ""
)

func LoadAll() string {
	polls := makeLoadReport("polls", wrapLoadOutput(
		loadOne(PollCache, pollsFile, models.Poll{})))
	posts := makeLoadReport("posts", wrapLoadOutput(
		loadOne(FlowCache, postsFile, models.Post{})))
	reqs := makeLoadReport("requests", wrapLoadOutput(
		loadOne(RequestCache, requestsFile, models.Request{})))
	subs := makeLoadReport("subscriptions", wrapLoadOutput(
		loadOne(SubscriptionCache, subscriptionsFile, []models.Device{})))
	tokens := makeLoadReport("tokens", wrapLoadOutput(
		loadOne(TokenCache, tokensFile, void)))
	users := makeLoadReport("users", wrapLoadOutput(
		loadOne(UserCache, usersFile, models.User{})))

	return fmt.Sprintf("loaded: %s, %s, %s, %s, %s, %s", polls, posts, reqs, subs, tokens, users)
}

func DumpAll() {
	// TODO: catch errors!
	dumpOne(PollCache, pollsFile, models.Poll{})
	dumpOne(FlowCache, postsFile, models.Post{})
	dumpOne(RequestCache, requestsFile, models.Request{})
	dumpOne(SubscriptionCache, subscriptionsFile, []models.Device{})
	dumpOne(TokenCache, tokensFile, void)
	dumpOne(UserCache, usersFile, models.User{})
}

type load struct {
	count int
	total int
	err   error
}

func makeLoadReport(name string, ld load) string {
	var prct float64

	if ld.total == 0 {
		prct = 0
	} else {
		prct = float64(ld.count) / float64(ld.total) * 100
	}

	report := fmt.Sprintf("%d/%d %s (%.0f%%)", ld.count, ld.total, name, prct)

	if ld.err == nil {
		return report
	} else {
		return fmt.Sprintf("%s but err: %s", report, ld.err.Error())
	}
}

func wrapLoadOutput(count, total int, err error) load {
	return load{
		count: count,
		total: total,
		err:   err,
	}
}

func loadOne[T any](cache *core.Cache, filepath string, model T) (int, int, error) {
	l := common.NewLogger(nil, "data load")

	var count int
	var total int

	rawData, err := os.ReadFile(filepath)
	if err != nil {
		l.Println(err.Error(), http.StatusInternalServerError)
		return count, total, err
	}

	if string(rawData) == "" {
		l.Println("empty data on input", http.StatusBadRequest)
		return count, total, errors.New("empty data on input")
	}

	matrix := struct {
		Items map[string]T `json:"items"`
	}{}

	err = json.Unmarshal(rawData, &matrix)
	if err != nil {
		l.Println(err.Error(), http.StatusInternalServerError)
		return count, total, err
	}

	total = len(matrix.Items)

	for key, val := range matrix.Items {
		if key == "" || &val == nil {
			continue
		}

		if saved := SetOne(cache, key, val); !saved {
			msg := fmt.Sprintf("cannot load item from file '%s' (key: %s)", filepath, key)
			l.Println(msg, http.StatusInternalServerError)
			return count, total, fmt.Errorf(msg)
			//continue
		}

		count++
	}

	return count, total, nil
}

func dumpOne[T any](cache *core.Cache, filepath string, model T) error {
	l := common.NewLogger(nil, "data dump")

	if &model == nil {
		l.Println("nil pointer on input!", http.StatusBadRequest)
		return errors.New("nil pointer on input!")
	}

	matrix := struct {
		Items map[string]T `json:"items"`
	}{}

	matrix.Items, _ = GetAll(cache, model)

	jsonData, err := json.Marshal(matrix)
	if err != nil {
		return err
	}

	// system-critical short log hack
	msg := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{}

	err = os.WriteFile(filepath, jsonData, 0660)
	if err != nil {
		msg.Message = err.Error()
		err = nil
	} else {
		return nil
	}

	if err = os.WriteFile(filepath+".bak", jsonData, 0660); err != nil {
		msg.Message += "; cannot even dump the data to a backup file: " + err.Error()
		l.Println(msg.Message, http.StatusInternalServerError)
		err = nil
	}

	msg.Code = 500
	marsh, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(msg.Message)
		return err
	}

	l.Println(string(marsh), http.StatusInternalServerError)
	//fmt.Println(string(marsh))

	return fmt.Errorf(msg.Message)
}

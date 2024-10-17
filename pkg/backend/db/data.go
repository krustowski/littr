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
		loadOne(TokenCache, tokensFile, models.Token{})))

	users := makeLoadReport("users", wrapLoadOutput(
		loadOne(UserCache, usersFile, models.User{})))

	return fmt.Sprintf("loaded: %s, %s, %s, %s, %s, %s", polls, posts, reqs, subs, tokens, users)
}

func DumpAll() string {
	var report string

	report += prepareDumpReport("polls",
		dumpOne(PollCache, pollsFile, models.Poll{}))

	report += prepareDumpReport("posts",
		dumpOne(FlowCache, postsFile, models.Post{}))

	report += prepareDumpReport("requests",
		dumpOne(RequestCache, requestsFile, models.Request{}))

	report += prepareDumpReport("subscriptions",
		dumpOne(SubscriptionCache, subscriptionsFile, []models.Device{}))

	report += prepareDumpReport("tokens",
		dumpOne(TokenCache, tokensFile, models.Token{}))

	report += prepareDumpReport("users",
		dumpOne(UserCache, usersFile, models.User{}))

	return fmt.Sprintf("dump: %s", report)
}

/*
 *  helper functions --- loadOne stack
 */

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

/*
 *  helper functions --- dumpOne
 */

func prepareDumpReport(cacheName string, rep *dumpReport) string {
	if rep.Error == nil {
		return fmt.Sprintf("[%s] dumped: %d, ", cacheName, rep.Total)
	}
	return fmt.Sprintf("[%s] dump failed: %d (%s), ", cacheName, rep.Total, rep.Error.Error())
}

type dumpReport struct {
	Total int
	Error error
}

func dumpOne[T any](cache *core.Cache, filepath string, model T) *dumpReport {
	l := common.NewLogger(nil, "data dump")

	// check if the model is usable
	if &model == nil {
		l.Msg("nil pointer to model on input to!").Status(http.StatusBadRequest).Log()
		return &dumpReport{Total: 0, Error: fmt.Errorf("nil pointer to model on input!")}
	}

	// base struct to map the data to JSON
	matrix := struct {
		Items map[string]T `json:"items"`
	}{}

	var total int

	// dump the in-memoty running data
	matrix.Items, total = GetAll(cache, model)

	// prepare the JSON byte stream
	jsonData, err := json.Marshal(&matrix)
	if err != nil {
		return &dumpReport{Error: err}
	}

	// write dumped data to the file
	if err = os.WriteFile(filepath, jsonData, 0660); err == nil {
		// OK condition
		return &dumpReport{Total: total}
	}

	// log the first attempt fail, but continue
	l.Msg("write error: " + err.Error()).Status(http.StatusInternalServerError).Log()
	err = nil

	// try the backup file if previous write failed
	if err = os.WriteFile(filepath+".bak", jsonData, 0660); err != nil {
		l.Msg("backup write failed: " + err.Error()).Status(http.StatusInternalServerError).Log()
		return &dumpReport{Total: 0, Error: err}
	}

	// OK condition
	return &dumpReport{Total: total}
}

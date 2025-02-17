package db

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/metrics"
	"go.vxn.dev/littr/pkg/models"
	//"go.vxn.dev/swis/v5/pkg/core"
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
		loadOne(SubscriptionCache, subscriptionsFile, models.Devices{})))

	tokens := makeLoadReport("tokens", wrapLoadOutput(
		loadOne(TokenCache, tokensFile, models.Token{})))

	users := makeLoadReport("users", wrapLoadOutput(
		loadOne(UserCache, usersFile, models.User{})))

	runtime.GC()

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
		dumpOne(SubscriptionCache, subscriptionsFile, models.Devices{}))

	report += prepareDumpReport("tokens",
		dumpOne(TokenCache, tokensFile, models.Token{}))

	report += prepareDumpReport("users",
		dumpOne(UserCache, usersFile, models.User{}))

	runtime.GC()

	return fmt.Sprintf("dump: %s", report)
}

//
//  helper functions --- loadOne stack
//

type load struct {
	count int64
	total int64
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

func wrapLoadOutput(count, total int64, err error) load {
	return load{
		count: count,
		total: total,
		err:   err,
	}
}

type item interface {
}

func loadOne[T models.Item](cache Cacher, filepath string, model T) (int64, int64, error) {
	l := common.NewLogger(nil, "data load")

	var count int64
	var total int64

	/*rb, err := os.ReadFile(fmt.Sprintf("/opt/data/%s.bin", cache.GetName()))
	if err != nil {
		log.Fatal("read: ", err)
	}

	rbuf := bytes.NewReader(rb)
	dec := gob.NewDecoder(rbuf)

	var items []T
	if err := dec.Decode(&items); err != nil {
		log.Fatalf("decode: %s, err: %s", cache.GetName(), err)
	}

	for _, item := range items {
		count++
		if stored := cache.Store(item.GetID(), item); !stored {
			log.Fatal(cache.GetName())
		}
		total++
	}*/

	//
	//
	//

	rawData, err := os.ReadFile(filepath)
	if err != nil {
		l.Error(err).Status(http.StatusInternalServerError).Log()
		return count, total, err
	}

	if string(rawData) == "" {
		l.Msg("empty data on input").Status(http.StatusBadRequest).Log()
		return count, total, errors.New("empty data on input")
	}

	matrix := &struct {
		Items map[string]T `json:"items"`
	}{}

	err = json.Unmarshal(rawData, matrix)
	if err != nil {
		l.Error(err).Status(http.StatusInternalServerError).Log()
		return count, total, err
	}

	total = int64(len(matrix.Items))

	for key, val := range matrix.Items {
		if key == "" || &val == nil {
			continue
		}

		if saved := SetOne(cache, key, val); !saved {
			msg := fmt.Sprintf("cannot load item from file '%s' (key: %s)", filepath, key)
			l.Msg(msg).Status(http.StatusInternalServerError).Log()
			return count, total, fmt.Errorf(msg)
		}

		count++
	}

	matrix = &struct {
		Items map[string]T `json:"items"`
	}{}

	metrics.UpdateCountMetric(cache.GetName(), count, true)

	return count, total, nil
}

//
//  helper functions --- dumpOne
//

func prepareDumpReport(cacheName string, rep *dumpReport) string {
	if rep == nil || rep.Error == nil {
		var total int

		if rep != nil {
			total = int(rep.Total)
		}

		return fmt.Sprintf("[%s] dumped: %d, ", cacheName, total)
	}
	return fmt.Sprintf("[%s] dump failed: %d (%s), ", cacheName, rep.Total, rep.Error.Error())
}

type dumpReport struct {
	Total int64
	Error error
}

func dumpOne[T models.Item](cache Cacher, filepath string, model T) *dumpReport {
	l := common.NewLogger(nil, "data dump")

	// check if the model is usable
	/*if &model == nil {
		l.Msg("nil pointer to model on input to").Status(http.StatusBadRequest).Log()
		return &dumpReport{Total: 0, Error: fmt.Errorf("nil pointer to model on input")}
	}*/

	//
	//  Experimental feature (memory dump do binary)
	//

	var items []T

	rawItems, count := cache.Range()

	for _, rawItem := range *rawItems {
		item, ok := rawItem.(T)
		if ok {
			items = append(items, item)
		}
	}

	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(items); err != nil {
		l.Msg("write error: " + err.Error()).Status(http.StatusInternalServerError).Log()
		fmt.Printf("encode: %s", err.Error())
		return nil
	}

	os.WriteFile(fmt.Sprintf("/opt/data/%s.bin", cache.GetName()), buf.Bytes(), 0600)

	buf.Reset()

	return &dumpReport{Total: count}

	//
	//
	//

	// base struct to map the data to JSON
	/*matrix := struct {
		Items *map[string]T `json:"items"`
	}{}

	var (
		jsonData []byte
		err      error
	)

	defer func() {
		*matrix.Items = map[string]T{}

		matrix = struct {
			Items *map[string]T `json:"items"`
		}{}

		jsonData = []byte{}
	}()

	var total int64

	// dump the in-memoty running data
	matrix.Items, total = GetAll(cache, model)

	// prepare the JSON byte stream
	jsonData, err = json.Marshal(&matrix)
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
	return &dumpReport{Total: total}*/
}

package db

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/config"
	"go.vxn.dev/littr/pkg/models"
)

const (
	pollsFile    = "/opt/data/polls.json"
	postsFile    = "/opt/data/posts.json"
	requestsFile = "/opt/data/requests.json"
	tokensFile   = "/opt/data/tokens.json"
	usersFile    = "/opt/data/users.json"
)

func (d *defaultDatabaseKeeper) LoadAll() (string, error) {
	db := d.Database()

	polls := makeLoadReport("polls", wrapLoadOutput(
		loadOne(db["PollCache"], pollsFile, models.Poll{})))

	posts := makeLoadReport("posts", wrapLoadOutput(
		loadOne(db["FlowCache"], postsFile, models.Post{})))

	reqs := makeLoadReport("requests", wrapLoadOutput(
		loadOne(db["RequestCache"], requestsFile, models.Request{})))

	tokens := makeLoadReport("tokens", wrapLoadOutput(
		loadOne(db["TokenCache"], tokensFile, models.Token{})))

	users := makeLoadReport("users", wrapLoadOutput(
		loadOne(db["UserCache"], usersFile, models.User{})))

	defer runtime.GC()

	return fmt.Sprintf("loaded: %s, %s, %s, %s, %s", polls, posts, reqs, tokens, users), nil
}

func (d *defaultDatabaseKeeper) DumpAll() (string, error) {
	db := d.Database()

	caches := []Cacher{
		db["PollCache"],
		db["FlowCache"],
		db["RequestCache"],
		db["TokenCache"],
		db["UserCache"],
	}

	paths := []string{
		pollsFile,
		postsFile,
		requestsFile,
		tokensFile,
		usersFile,
	}

	report := runDumpEngine(caches, paths)

	defer runtime.GC()

	return fmt.Sprintf("dump: %s", report), nil
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

func loadOne[T models.Item](cache Cacher, filepath string, _ T) (int64, int64, error) {
	l := common.NewLogger(nil, "data load")

	var count int64
	var total int64

	switch config.DataLoadFormat {
	case "binary":
		rb, err := os.ReadFile(fmt.Sprintf("/opt/data/%s.bin", cache.GetName()))
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
		}

	default:
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
			if key == "" {
				continue
			}

			if saved := setOne(cache, key, val); !saved {
				msg := fmt.Sprintf("cannot load item from file '%s' (key: %s)", filepath, key)
				l.Msg(msg).Status(http.StatusInternalServerError).Log()
				return count, total, fmt.Errorf("%s", msg)
			}

			count++
		}

		matrix = &struct {
			Items map[string]T `json:"items"`
		}{}

		// metrics.UpdateCountMetric(cache.GetName(), count, true)

	}

	return count, total, nil
}

//
//  helper functions --- dumpOne
//

type dumpReport struct {
	CacheName string
	Total     int64
	Error     error
}

/*func prepareDumpReport(cacheName string, rep *dumpReport) string {
	if rep == nil || rep.Error == nil {
		var total int

		if rep != nil {
			total = int(rep.Total)
		}

		return fmt.Sprintf("[%s] dumped: %d, ", cacheName, total)
	}
	return fmt.Sprintf("[%s] dump failed: %d (%s), ", cacheName, rep.Total, rep.Error.Error())
}

func dumpOne[T models.Item](cache Cacher, filepath string, model T) *dumpReport {
	l := common.NewLogger(nil, "data dump")

	switch config.DataDumpFormat {
	case "binary":
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

		if err := os.WriteFile(fmt.Sprintf("/opt/data/%s.bin", cache.GetName()), buf.Bytes(), 0o600); err != nil {
			fmt.Print(err)
			return nil
		}

		buf.Reset()

		return &dumpReport{Total: count}

	default:
		// Base struct to map the data to JSON.
		matrix := struct {
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

		// Dump the in-memoty running data.
		matrix.Items, total = getAll(cache, model)

		// Prepare the JSON byte stream.
		jsonData, err = json.Marshal(&matrix)
		if err != nil {
			return &dumpReport{Error: err}
		}

		// Write dumped data to the file.
		if err = os.WriteFile(filepath, jsonData, 0o660); err == nil {
			// OK condition
			return &dumpReport{Total: total}
		}

		// Log the first attempt fail, but continue.
		l.Msg("write error: " + err.Error()).Status(http.StatusInternalServerError).Log()
		err = nil

		// Try the backup file if previous write failed.
		if err = os.WriteFile(filepath+".bak", jsonData, 0o660); err != nil {
			l.Msg("backup write failed: " + err.Error()).Status(http.StatusInternalServerError).Log()
			return &dumpReport{Total: 0, Error: err}
		}

		return &dumpReport{Total: total}
	}
}*/

func runDumpEngine(caches []Cacher, filePaths []string) string {
	if len(caches) != len(filePaths) || len(caches) == 0 {
		return "input error: check the length of input arrays"
	}

	chans := make([]chan interface{}, 0)

	var wg sync.WaitGroup

	for j, cache := range caches {
		ch := make(chan interface{})
		chans = append(chans, ch)

		wg.Add(1)
		go dumpOneRaw(cache, filePaths[j], ch, &wg)
	}

	reports := FanInChannels(nil, chans...)
	wg.Wait()

	var commonReport string

	for raw := range reports {
		report, ok := raw.(*dumpReport)
		if !ok {
			// Skip for now
			continue
		}

		if report.Error != nil {
			commonReport += fmt.Sprintf("[%s] dump failed: %d (%s), ", report.CacheName, report.Total, report.Error.Error())
			continue
		}

		commonReport += fmt.Sprintf("[%s] dumped: %d, ", report.CacheName, report.Total)
	}

	return commonReport
}

func dumpOneRaw(cache Cacher, filepath string, ch chan interface{}, wg *sync.WaitGroup) {
	mp := cache.Dump()
	total := len(*mp)

	report := &dumpReport{
		CacheName: cache.GetName(),
		Total:     int64(total),
	}

	defer func() {
		if ch != nil {
			ch <- report
			close(ch)
		}

		if wg != nil {
			wg.Done()
		}
	}()

	data := struct {
		Items *GenericMap `json:"items"`
	}{
		Items: mp,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		report.Error = err
		return
	}

	// Write dumped data to the file.
	err = os.WriteFile(filepath, jsonData, 0o660)
	if err != nil {
		report.Error = err
		return
	}
}

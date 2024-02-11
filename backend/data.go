package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"
	"go.savla.dev/swis/v5/pkg/core"

	"github.com/SherClockHolmes/webpush-go"
)

const (
	pollsFile         = "/opt/data/polls.json"
	postsFile         = "/opt/data/posts.json"
	subscriptionsFile = "/opt/data/subscriptions.json"
	tokensFile        = "/opt/data/tokens.json"
	usersFile         = "/opt/data/users.json"
	void              = ""
)

func LoadAll() {
	// TODO: catch errors!
	loadOne(PollCache, pollsFile, models.Poll{})
	loadOne(FlowCache, postsFile, models.Post{})
	loadOne(SubscriptionCache, subscriptionsFile, []webpush.Subscription{})
	loadOne(TokenCache, tokensFile, void)
	loadOne(UserCache, usersFile, models.User{})
}

func DumpAll() {
	// TODO: catch errors!
	dumpOne(PollCache, pollsFile, models.Poll{})
	dumpOne(FlowCache, postsFile, models.Post{})
	dumpOne(SubscriptionCache, subscriptionsFile, []webpush.Subscription{})
	dumpOne(TokenCache, tokensFile, void)
	dumpOne(UserCache, usersFile, models.User{})
}

func dumpHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{}

	// prepare the Logger instance
	l := Logger{
		CallerID: "system",
		//IPAddress:  r.RemoteAddr,
		IPAddress:  r.Header.Get("X-Real-IP"),
		Method:     r.Method,
		Route:      r.URL.String(),
		WorkerName: "dump",
		Version:    "system",
	}

	// check the incoming API token
	token := r.Header.Get("X-App-Token")

	if token == "" {
		resp.Message = "empty token"
		resp.Code = http.StatusUnauthorized

		l.Println(resp.Message, resp.Code)
		return
	}

	if token != os.Getenv("API_TOKEN") {
		resp.Message = "invalid token"
		resp.Code = http.StatusForbidden

		l.Println(resp.Message, resp.Code)
		return
	}

	DumpAll()

	resp.Code = http.StatusOK
	resp.Message = "data dumped successfully"

	l.Println(resp.Message, resp.Code)

	// dynamic encryption bypass hack --- we need unecrypted JSON for curl (healthcheck)
	if config.EncryptionEnabled {
		//log.Printf("[dump] disabling encryption (was %t)", config.EncryptionEnabled)
		config.EncryptionEnabled = !config.EncryptionEnabled

		resp.Write(w)

		//log.Printf("[dump] enabling encryption (was %t)", config.EncryptionEnabled)
		config.EncryptionEnabled = !config.EncryptionEnabled
	} else {
		resp.Write(w)
	}

	return
}

func loadOne[T any](cache *core.Cache, filepath string, model T) error {
	rawData, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	if string(rawData) == "" {
		return errors.New("empty data on input")
	}

	matrix := struct {
		Items map[string]T `json:"items"`
	}{}

	err = json.Unmarshal(rawData, &matrix)
	if err != nil {
		return err
	}

	for key, val := range matrix.Items {
		if key == "" || &val == nil {
			continue
		}

		if saved := setOne(cache, key, val); !saved {
			return fmt.Errorf("cannot load item from file '%s' (key: %s)", filepath, key)
			//continue
		}
	}

	return nil
}

func dumpOne[T any](cache *core.Cache, filepath string, model T) error {
	if &model == nil {
		return errors.New("nil pointer on input!")
	}

	matrix := struct {
		Items map[string]T `json:"items"`
	}{}

	matrix.Items, _ = getAll(cache, model)

	jsonData, err := json.Marshal(matrix)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath, jsonData, 0660)
	if err != nil {
		return err
	}

	return nil
}

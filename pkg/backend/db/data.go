package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"go.savla.dev/littr/pkg/backend/common"
	"go.savla.dev/littr/pkg/models"
	"go.savla.dev/swis/v5/pkg/core"
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
	loadOne(SubscriptionCache, subscriptionsFile, []models.Device{})
	loadOne(TokenCache, tokensFile, void)
	loadOne(UserCache, usersFile, models.User{})
}

func DumpAll() {
	// TODO: catch errors!
	dumpOne(PollCache, pollsFile, models.Poll{})
	dumpOne(FlowCache, postsFile, models.Post{})
	dumpOne(SubscriptionCache, subscriptionsFile, []models.Device{})
	dumpOne(TokenCache, tokensFile, void)
	dumpOne(UserCache, usersFile, models.User{})
}

func loadOne[T any](cache *core.Cache, filepath string, model T) error {
	l := common.NewLogger(nil, "data load")

	rawData, err := os.ReadFile(filepath)
	if err != nil {
		l.Println(err.Error(), http.StatusInternalServerError)
		return err
	}

	if string(rawData) == "" {
		l.Println("empty data on input", http.StatusBadRequest)
		return errors.New("empty data on input")
	}

	matrix := struct {
		Items map[string]T `json:"items"`
	}{}

	err = json.Unmarshal(rawData, &matrix)
	if err != nil {
		l.Println(err.Error(), http.StatusInternalServerError)
		return err
	}

	for key, val := range matrix.Items {
		if key == "" || &val == nil {
			continue
		}

		if saved := SetOne(cache, key, val); !saved {
			msg := fmt.Sprintf("cannot load item from file '%s' (key: %s)", filepath, key)
			l.Println(msg, http.StatusInternalServerError)
			return fmt.Errorf(msg)
			//continue
		}
	}

	return nil
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

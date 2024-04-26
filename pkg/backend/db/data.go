package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"go.savla.dev/littr/configs"
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

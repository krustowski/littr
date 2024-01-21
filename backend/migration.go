package backend

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.savla.dev/littr/models"
)

type migration func() bool

var migrations = map[string]migration{
	"migrateAvatarURL()":          migrateAvatarURL,
	"migrateUserDeletion()":       migrateUserDeletion,
	"migrateUserRegisteredTime()": migrateUserRegisteredTime,
	"migrateUserShadeList()":      migrateUserShadeList,
	"migrateUserUnshade()":        migrateUserUnshade,
}

const defaultAvatarImage = "/web/android-chrome-192x192.png"

// RunMigrations is a "wrapper" function for the migration registration and execution
func RunMigrations() bool {
	l := Logger{
		CallerID:   "system",
		WorkerName: "migration",
		Version:    "system",
	}

	for migName, migFunc := range migrations {
		code := http.StatusOK

		if ok := migFunc(); !ok {
			code = http.StatusInternalServerError
		}

		l.Println(migName, code)
	}

	return true
}

// migrateAvatarURL function take care of (re)assigning custom, or default avatars to all users having blank or default strings saved in their data chunk. Function returns bool based on the process result.
func migrateAvatarURL() bool {
	users, _ := getAll(UserCache, models.User{})

	for key, user := range users {
		if user.AvatarURL != "" && user.AvatarURL != defaultAvatarImage {
			continue
		}

		user.AvatarURL = GetGravatarURL(user.Email)
		if ok := setOne(UserCache, key, user); !ok {
			return false
		}
	}

	return true
}

// migrateUserDeletion function takes care of default users deletion from the database. Function returns bool based on the process result.
func migrateUserDeletion() bool {
	bank := []string{
		"fred",
		"fred2",
		"admin",
		"alternative",
		"Lmao",
		"lma0",
	}

	users, _ := getAll(UserCache, models.User{})

	for key, user := range users {
		if contains(bank, user.Nickname) {
			if deleted := deleteOne(UserCache, key); !deleted {
				//return false
				continue
			}
		}
	}

	posts, _ := getAll(FlowCache, models.Post{})

	for key, post := range posts {
		if contains(bank, post.Nickname) {
			if deleted := deleteOne(FlowCache, key); !deleted {
				//return false
				continue
			}
		}
	}

	return true
}

// migrateUserRegisteredTime function fixes the initial registration date if it defaults to the "null" time.Time string. Function returns bool based on the process result.
func migrateUserRegisteredTime() bool {
	users, _ := getAll(UserCache, models.User{})

	for key, user := range users {
		if user.RegisteredTime == time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC) {
			user.RegisteredTime = time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC)
			if ok := setOne(UserCache, key, user); !ok {
				return false
			}
		}
	}

	return true
}

// migrateUserShadeList function lists ShadeList items and ensures user shaded (no mutual following, no replying).
func migrateUserShadeList() bool {
	users, _ := getAll(UserCache, models.User{})

	for key, user := range users {
		shadeList := user.ShadeList
		flowList := user.FlowList

		if flowList == nil {
			flowList = make(map[string]bool)
		}

		// ShadeList map[string]bool `json:"shade_list"`
		for name, state := range shadeList {
			if state && name != user.Nickname {
				flowList[name] = false
				user.FlowList = flowList
				//setOne(UserCache, key, user)
			}

		}

		// ensure that users can see themselves
		flowList[key] = true
		user.FlowList = flowList
		if ok := setOne(UserCache, key, user); !ok {
			return false
		}
	}

	return true
}

// migrateUserUnshade function lists all users and unshades manually some explicitly list users
func migrateUserUnshade() bool {
	users, _ := getAll(UserCache, models.User{})

	usersToUnshade := []string{
		"amdulka",
		"nestolecek",
	}

	for key, user := range users {
		if !contains(usersToUnshade, key) {
			continue
		}

		shadeList := user.ShadeList

		for name := range shadeList {
			if contains(usersToUnshade, name) {
				shadeList[name] = false
			}
		}

		user.ShadeList = shadeList
		if ok := setOne(UserCache, key, user); !ok {
			return false
		}
	}

	return true
}

/*
 *  helpers
 */

// contains checks if a string is present in a slice.
// https://freshman.tech/snippets/go/check-if-slice-contains-element/
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// GetGravatarURL function returns the avatar image location/URL, or it defaults to a app logo.
func GetGravatarURL(emailInput string) string {
	// TODO: do not hardcode this
	baseURL := "https://littr.n0p.cz/"
	email := strings.ToLower(emailInput)
	size := 150

	// hash the emailInput
	byteEmail := []byte(email)
	hashEmail := md5.Sum(byteEmail)
	hashedStringEmail := hex.EncodeToString(hashEmail[:])

	url := "https://www.gravatar.com/avatar/" + hashedStringEmail + "?d=" + baseURL + "&s=" + strconv.Itoa(size)

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return defaultAvatarImage
	}

	return url
}

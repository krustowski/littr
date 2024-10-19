package db

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/config"
	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"
)

const defaultAvatarImage = "/web/android-chrome-192x192.png"

// migrationProp is a struct to hold the migration function reference, and array of interfaces (mainly pointers) of various length.
type migrationProp struct {
	// Migration name.
	N string

	// Migration function.
	F func(*common.Logger, []interface{}) bool

	// Migration resources.
	R []interface{}
}

// RunMigrations is a "wrapper" function for the migration registration and execution
func RunMigrations() bool {
	// init logger
	l := common.NewLogger(nil, "migrations")
	l.CallerID = "system"
	l.Version = "system"

	// fetch all the data
	users, _ := GetAll(UserCache, models.User{})
	//polls, _ := GetAll(PollCache, models.Poll{})
	posts, _ := GetAll(FlowCache, models.Post{})
	reqs, _ := GetAll(RequestCache, models.Request{})
	subs, _ := GetAll(SubscriptionCache, []models.Device{})
	tokens, _ := GetAll(TokenCache, models.Token{})

	// define the migrations map
	//var migrations = map[string]migrationProp{
	var migrations = []migrationProp{
		{
			N: "migrateExpired",
			F: migrateExpired,
			R: []interface{}{&reqs, &tokens},
		},
		{
			N: "migrateEmptyDeviceTags",
			F: migrateEmptyDeviceTags,
			R: []interface{}{&subs},
		},
		{
			N: "migrateAvatarURL",
			F: migrateAvatarURL,
			R: []interface{}{&users},
		},
		{
			N: "migrateFlowPurge",
			F: migrateFlowPurge,
			R: []interface{}{&posts, &users},
		},
		{
			N: "migrateUserDeletion",
			F: migrateUserDeletion,
			R: []interface{}{&posts, &users},
		},
		{
			N: "migrateUserRegisteredTime",
			F: migrateUserRegisteredTime,
			R: []interface{}{&users},
		},
		{
			N: "migrateUserShadeList",
			F: migrateUserShadeList,
			R: []interface{}{&users},
		},
		{
			N: "migrateUserUnshade",
			F: migrateUserUnshade,
			R: []interface{}{&users},
		},
		{
			N: "migrateBlankAboutText",
			F: migrateBlankAboutText,
			R: []interface{}{&users},
		},
		{
			N: "migrateSystemFlowOn",
			F: migrateSystemFlowOn,
			R: []interface{}{&users},
		},
	}

	// prepare the report var
	var report string

	// run migrations
	for _, mig := range migrations {
		report += fmt.Sprintf("[%s]: %t, ", mig.N, mig.F(l, mig.R))
	}

	l.Msg(report).Status(http.StatusOK).Log()

	return true
}

// migrateExpiredRequests procedure loops over requests and removes those expired already.
func migrateExpired(l *common.Logger, rawElems []interface{}) bool {
	var reqs *map[string]models.Request
	var tokens *map[string]models.Token

	// assert pointers from the interface array
	for _, raw := range rawElems {
		elem, ok := raw.(*map[string]models.Request)
		if ok {
			reqs = elem
			continue
		}

		elem2, ok := raw.(*map[string]models.Token)
		if ok {
			tokens = elem2
			continue
		}
	}

	if reqs == nil {
		l.Msg("migrateExpired: reqs are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// loop over and compare times with durations, if expired, yeet it
	for uuid, req := range *reqs {
		if time.Now().After(req.CreatedAt.Add(time.Hour * 24)) {
			if deleted := DeleteOne(RequestCache, uuid); !deleted {
				l.Msg("migrateExpired: could not delete request: " + uuid).Status(http.StatusInternalServerError).Log()
				continue
			}

			delete(*reqs, uuid)
		}
	}

	// migrateExpiredTokens procedure loop over tokens and removes those beyond the expiry.
	//func migrateExpiredTokens(l *common.Logger, tokens *map[string]models.Token) bool {
	if tokens == nil {
		l.Msg("migrateExpired: tokens are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// loop over and compare times with durations, if expired, yeet it
	for hash, token := range *tokens {
		if time.Now().After(token.CreatedAt.Add(common.TOKEN_TTL)) {
			if deleted := DeleteOne(TokenCache, hash); !deleted {
				l.Msg("migrateExpiredTokens: could not delete token: " + hash).Status(http.StatusInternalServerError).Log()
				continue
			}

			delete(*tokens, hash)
		}
	}

	return true
}

// migrateEmptyDeviceTags procedure takes care of filling empty device tags arrays.
func migrateEmptyDeviceTags(l *common.Logger, rawElems []interface{}) bool {
	var subs *map[string][]models.Device

	// assert pointers from the interface array
	for _, raw := range rawElems {
		elem, ok := raw.(*map[string][]models.Device)
		if ok {
			subs = elem
			continue
		}
	}

	if subs == nil {
		l.Msg("migrateEmptyDeviceTags: subs are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// look for empty tags of subscriptions/devices in nested loops
	for key, devs := range *subs {
		changed := false
		for idx, dev := range devs {
			if len(dev.Tags) == 0 {
				dev.Tags = []string{
					"reply",
					"mention",
				}
				changed = true
				devs[idx] = dev
			}
		}

		if changed {
			if saved := SetOne(SubscriptionCache, key, devs); !saved {
				l.Msg("migrateEmptyDeviceTags: cannot save changed dev: " + key).Status(http.StatusInternalServerError).Log()
				return false
			}
		}
	}

	return true
}

// migrateAvatarURL procedure takes care of (re)assigning custom, or default avatars to all users having blank or default strings saved in their data chunk. Function returns bool based on the process result.
func migrateAvatarURL(l *common.Logger, rawElems []interface{}) bool {
	var users *map[string]models.User
	var urlsChan chan string

	// assert pointers from the interface array
	for _, raw := range rawElems {
		elem, ok := raw.(*map[string]models.User)
		if ok {
			users = elem
			continue
		}
	}

	if users == nil {
		l.Msg("migrateAvatarURL: users are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// channel for gravatar routines' results
	urlsChan = make(chan string)

	var wg *sync.WaitGroup

	for key, user := range *users {
		if user.AvatarURL != "" && user.AvatarURL != defaultAvatarImage {
			continue
		}

		wg.Add(1)
		go GetGravatarURL(user.Email, urlsChan, wg)
		url := <-urlsChan

		if url != user.AvatarURL {
			user.AvatarURL = url

			if ok := SetOne(UserCache, key, user); !ok {
				l.Msg("migrateAvatarURL: cannot save an avatar: " + key).Status(http.StatusInternalServerError).Log()
				close(urlsChan)
				return false
			}
			(*users)[key] = user
		}
	}

	close(urlsChan)

	return true
}

// migrateFlowPurge procedure deletes all pseudoaccounts and their posts, those psaudeaccounts are not registered accounts, thus not real users.
func migrateFlowPurge(l *common.Logger, rawElems []interface{}) bool {
	var posts *map[string]models.Post
	var users *map[string]models.User

	// assert pointers from the interface array
	for _, raw := range rawElems {
		elem, ok := raw.(*map[string]models.User)
		if ok {
			users = elem
			continue
		}

		elem2, ok := raw.(*map[string]models.Post)
		if ok {
			posts = elem2
			continue
		}
	}

	// terminate on nil pointer(s)
	if users == nil || posts == nil {
		l.Msg("migrateFlowPurge: users or posts are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// loop and delete author-less posts
	for key, post := range *posts {
		if _, found := (*users)[post.Nickname]; !found {
			if deleted := DeleteOne(FlowCache, key); !deleted {
				l.Msg("migrateFlowPurge: cannot delete user: " + key).Status(http.StatusInternalServerError).Log()
				return false
			}
			delete(*posts, key)
		}
	}

	return true
}

// migrateUserDeletion procedure takes care of default users deletion from the database. Function returns bool based on the process result.
func migrateUserDeletion(l *common.Logger, rawElems []interface{}) bool {
	var posts *map[string]models.Post
	var users *map[string]models.User

	// assert pointers from the interface array
	for _, raw := range rawElems {
		elem, ok := raw.(*map[string]models.User)
		if ok {
			users = elem
			continue
		}

		elem2, ok := raw.(*map[string]models.Post)
		if ok {
			posts = elem2
			continue
		}
	}

	if users == nil || posts == nil {
		l.Msg("migrateUserDeletion: users or posts are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	bank := &config.UserDeletionList

	// delete all users matching the contents of restricted nickname list
	for key, user := range *users {
		if helpers.Contains(*bank, user.Nickname) {
			l.Msg("deleting " + user.Nickname).Status(http.StatusProcessing).Log()

			if deleted := DeleteOne(UserCache, key); !deleted {
				l.Msg("migrateUserDeletion: cannot delete an user: " + key).Status(http.StatusInternalServerError).Log()
				return false
			}

			delete(*users, key)
		}
	}

	// delete all user's posts
	for key, post := range *posts {
		if helpers.Contains(*bank, post.Nickname) {
			if deleted := DeleteOne(FlowCache, key); !deleted {
				l.Msg("migrateUserDeletion: cannot delete a post: " + key).Status(http.StatusInternalServerError).Log()
				return false
			}

			delete(*posts, key)
		}
	}

	return true
}

// migrateUserRegisteredTime procedure fixes the initial registration date if it defaults to the "null" time.Time string. Function returns bool based on the process result.
func migrateUserRegisteredTime(l *common.Logger, rawElems []interface{}) bool {
	var users *map[string]models.User

	// assert pointers from the interface array
	for _, raw := range rawElems {
		elem, ok := raw.(*map[string]models.User)
		if ok {
			users = elem
			continue
		}
	}

	if users == nil {
		l.Msg("migrateUserRegisteredTime: users are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// loop over users and fix their reg. datetime
	for key, user := range *users {
		if user.RegisteredTime == time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC) {
			user.RegisteredTime = time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC)
			if ok := SetOne(UserCache, key, user); !ok {
				l.Msg("migrateUserRegisteredTime: cannot save an user: " + key).Status(http.StatusInternalServerError).Log()
				return false
			}
			(*users)[key] = user
		}
	}

	return true
}

// migrateUserShadeList procedure lists ShadeList items and ensures user shaded (no mutual following, no replying).
func migrateUserShadeList(l *common.Logger, rawElems []interface{}) bool {
	var users *map[string]models.User

	// assert pointers from the interface array
	for _, raw := range rawElems {
		elem, ok := raw.(*map[string]models.User)
		if ok {
			users = elem
			continue
		}
	}

	if users == nil {
		l.Msg("migrateUserShadeList: users are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	for key, user := range *users {
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
		if saved := SetOne(UserCache, key, user); !saved {
			l.Msg("migrateUserShadeList: cannot save user: " + key).Status(http.StatusInternalServerError).Log()
			return false
		}

		(*users)[key] = user
	}

	return true
}

// migrateUserUnshade procedure lists all users and unshades manually some explicitly list users
func migrateUserUnshade(l *common.Logger, rawElems []interface{}) bool {
	var users *map[string]models.User

	// assert pointers from the interface array
	for _, raw := range rawElems {
		elem, ok := raw.(*map[string]models.User)
		if ok {
			users = elem
			continue
		}
	}

	if users == nil {
		l.Msg("migrateUserUnshade: users are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	usersToUnshade := &config.UsersToUnshade

	for key, user := range *users {
		if !helpers.Contains(*usersToUnshade, key) {
			continue
		}

		shadeList := user.ShadeList

		for name := range shadeList {
			if helpers.Contains(*usersToUnshade, name) {
				shadeList[name] = false
			}
		}

		user.ShadeList = shadeList
		if ok := SetOne(UserCache, key, user); !ok {
			l.Msg("migrateUserUnshade: cannot save an user: " + key).Status(http.StatusInternalServerError).Log()
			return false
		}

		(*users)[key] = user
	}

	return true
}

// migrateBlankAboutText procedure loops over user accounts and adds "newbie" where the about-text field is blank
func migrateBlankAboutText(l *common.Logger, rawElems []interface{}) bool {
	var users *map[string]models.User

	// assert pointers from the interface array
	for _, raw := range rawElems {
		elem, ok := raw.(*map[string]models.User)
		if ok {
			users = elem
			continue
		}
	}

	if users == nil {
		l.Msg("migrateBlankAboutText: users are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	for key, user := range *users {
		if len(user.About) == 0 {
			user.About = "newbie"
		}

		if saved := SetOne(UserCache, key, user); !saved {
			l.Msg("migrateBlankAboutText: cannot save an user: " + key).Status(http.StatusInternalServerError).Log()
			return false
		}

		(*users)[key] = user
	}

	return true
}

// migrateSystemFlowOn procedure ensures everyone has system account in the flow
func migrateSystemFlowOn(l *common.Logger, rawElems []interface{}) bool {
	var users *map[string]models.User

	// assert pointers from the interface array
	for _, raw := range rawElems {
		elem, ok := raw.(*map[string]models.User)
		if ok {
			users = elem
			continue
		}
	}

	if users == nil {
		l.Msg("migrateSystemFlowOn: users are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// loop over users and fix the system following
	for key, user := range *users {
		if user.FlowList == nil {
			user.FlowList = make(map[string]bool)
		}

		user.FlowList[user.Nickname] = true
		user.FlowList["system"] = true

		if saved := SetOne(UserCache, key, user); !saved {
			l.Msg("migrateSystemFlowOn: cannot save an user: " + key).Status(http.StatusInternalServerError).Log()
			return false
		}

		(*users)[key] = user
	}

	return true
}

/*
 *  helpers
 */

// GetGravatarURL function returns the avatar image location/URL, or it defaults to a app logo.
func GetGravatarURL(emailInput string, channel chan string, wg *sync.WaitGroup) string {
	if wg != nil {
		defer wg.Done()
	}

	email := strings.ToLower(emailInput)
	size := 150

	sha := sha256.New()
	sha.Write([]byte(email))

	hashedStringEmail := fmt.Sprintf("%x", sha.Sum(nil))

	// hash the emailInput
	//byteEmail := []byte(email)
	//hashEmail := md5.Sum(byteEmail)
	//hashedStringEmail := hex.EncodeToString(hashEmail[:])

	url := "https://www.gravatar.com/avatar/" + hashedStringEmail + "?s=" + strconv.Itoa(size)

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		//log.Println(resp.StatusCode)
		//log.Println(err.Error())
		url = defaultAvatarImage
	} else {
		resp.Body.Close()
	}

	// maybe we are running in a goroutine...
	if channel != nil {
		channel <- url
	}
	return url
}

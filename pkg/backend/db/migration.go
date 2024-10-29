package db

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
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
	// Migration's name.
	N string

	// Migration's function handle.
	F func(*common.Logger, []interface{}) bool

	// Migration's resources to process.
	R []interface{}
}

// RunMigrations is a wrapper function for the migration registration and execution.
func RunMigrations(l *common.Logger) string {
	// Fetch all the data for the migration procedures.
	users, _ := GetAll(UserCache, models.User{})
	//polls, _ := GetAll(PollCache, models.Poll{})
	posts, _ := GetAll(FlowCache, models.Post{})
	reqs, _ := GetAll(RequestCache, models.Request{})
	subs, _ := GetAll(SubscriptionCache, []models.Device{})
	tokens, _ := GetAll(TokenCache, models.Token{})

	// Define the migration procedures order.
	var migrationsOrderedList = []migrationProp{
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
		{
			N: "migrateUserActiveState",
			F: migrateUserActiveState,
			R: []interface{}{&users, &reqs},
		},
	}

	// Declare the migrations report variable.
	var report string

	// Execute the migration procedures.
	for _, mig := range migrationsOrderedList {
		report += fmt.Sprintf("[%s]: %t, ", mig.N, mig.F(l.SetPrefix(mig.N), mig.R))
	}

	l.RemovePrefix()
	return report
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
		l.Msg("reqs are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// loop over and compare times with durations, if expired, yeet it
	for uuid, req := range *reqs {
		if time.Now().After(req.CreatedAt.Add(time.Hour * 24)) {
			if deleted := DeleteOne(RequestCache, uuid); !deleted {
				l.Msg("could not delete request: " + uuid).Status(http.StatusInternalServerError).Log()
				continue
			}

			delete(*reqs, uuid)
		}
	}

	// migrateExpiredTokens subprocedure loop over tokens and removes those beyond the expiry.
	//func migrateExpiredTokens(l *common.Logger, tokens *map[string]models.Token) bool {
	if tokens == nil {
		l.Msg("tokens are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// loop over and compare times with durations, if expired, yeet it
	for hash, token := range *tokens {
		if time.Now().After(token.CreatedAt.Add(common.TOKEN_TTL)) {
			if deleted := DeleteOne(TokenCache, hash); !deleted {
				l.Msg("could not delete token: " + hash).Status(http.StatusInternalServerError).Log()
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
		l.Msg("subs are nil").Status(http.StatusInternalServerError).Log()
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
				l.Msg("cannot save changed dev: " + key).Status(http.StatusInternalServerError).Log()
				return false
			}
		}
	}

	return true
}

// migrateAvatarURL procedure takes care of (re)assigning custom, or default avatars to all users having blank or default strings saved in their data chunk. Function returns bool based on the process result.
func migrateAvatarURL(l *common.Logger, rawElems []interface{}) bool {
	var users *map[string]models.User

	// assert pointers from the interface array
	for _, raw := range rawElems {
		elem, ok := raw.(*map[string]models.User)
		if ok {
			users = elem
			continue
		}
	}

	// Exit when the users cannot be asserted, or are just nil in general.
	if users == nil {
		l.Msg("users are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	var wg sync.WaitGroup

	// Make multiple channels for the per user run goroutines.
	var channels = make([]chan avatarResult, len(*users))
	var i int

	// Loop over users to execute one goroutin per user.
	for _, user := range *users {
		// Disable this just for tests. Maybe add a feature flag to enable this via an env variable.
		//if user.AvatarURL != "" && user.AvatarURL != defaultAvatarImage {
		if !func() bool {
			val, err := strconv.ParseBool(os.Getenv("RUN_AVATAR_MIGRATION"))
			if err != nil {
				return false
			}
			return val
		}() {
			continue
		}

		wg.Add(1)
		channels[i] = make(chan avatarResult)

		// Run the gravatar goroutine.
		go GetGravatarURL(user, channels[i], &wg)
		i++
	}

	// Retrieve the results = merge the channels into one.
	results := fanInChannels(channels...)
	wg.Wait()

	// Collect the results.
	for result := range results {
		// Change the avatarURL only if the new URL differs from the previous one.
		if result.URL != "" && result.URL != result.User.AvatarURL {
			result.User.AvatarURL = result.URL

			// Update the user's avatar in the User database.
			if ok := SetOne(UserCache, result.User.Nickname, result.User); !ok {
				l.Msg("cannot save an avatar: " + result.User.Nickname).Status(http.StatusInternalServerError).Log()
				return false
			}

			// Save the user locally within the migrations.
			(*users)[result.User.Nickname] = result.User
		}
	}

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
		l.Msg("users or posts are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// loop and delete author-less posts
	for key, post := range *posts {
		if _, found := (*users)[post.Nickname]; !found {
			if deleted := DeleteOne(FlowCache, key); !deleted {
				l.Msg("cannot delete user: " + key).Status(http.StatusInternalServerError).Log()
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
		l.Msg("users or posts are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	bank := &config.UserDeletionList

	// delete all users matching the contents of restricted nickname list
	for key, user := range *users {
		if helpers.Contains(*bank, user.Nickname) {
			l.Msg("deleting " + user.Nickname).Status(http.StatusProcessing).Log()

			if deleted := DeleteOne(UserCache, key); !deleted {
				l.Msg("cannot delete an user: " + key).Status(http.StatusInternalServerError).Log()
				return false
			}

			delete(*users, key)
		}
	}

	// delete all user's posts
	for key, post := range *posts {
		if helpers.Contains(*bank, post.Nickname) {
			if deleted := DeleteOne(FlowCache, key); !deleted {
				l.Msg("cannot delete a post: " + key).Status(http.StatusInternalServerError).Log()
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
		l.Msg("users are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// loop over users and fix their reg. datetime
	for key, user := range *users {
		if user.RegisteredTime == time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC) {
			user.RegisteredTime = time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC)
			if ok := SetOne(UserCache, key, user); !ok {
				l.Msg("cannot save an user: " + key).Status(http.StatusInternalServerError).Log()
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
		l.Msg("users are nil").Status(http.StatusInternalServerError).Log()
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
			l.Msg("cannot save user: " + key).Status(http.StatusInternalServerError).Log()
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
		l.Msg("users are nil").Status(http.StatusInternalServerError).Log()
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
			l.Msg("cannot save an user: " + key).Status(http.StatusInternalServerError).Log()
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
		l.Msg("users are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	for key, user := range *users {
		if len(user.About) == 0 {
			user.About = "newbie"
		}

		if saved := SetOne(UserCache, key, user); !saved {
			l.Msg("cannot save an user: " + key).Status(http.StatusInternalServerError).Log()
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
		l.Msg("users are nil").Status(http.StatusInternalServerError).Log()
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
			l.Msg("cannot save an user: " + key).Status(http.StatusInternalServerError).Log()
			return false
		}

		(*users)[key] = user
	}

	return true
}

// migrateUserActiveState ensures all users registered before Oct 28, 2024 are activated; otherwise it alse tries to delete valid, but misdeleted activation requests from its database.
func migrateUserActiveState(l *common.Logger, rawElems []interface{}) bool {
	var users *map[string]models.User
	var reqs *map[string]models.Request

	// Assert pointers from the interface array.
	for _, raw := range rawElems {
		elem, ok := raw.(*map[string]models.User)
		if ok {
			users = elem
			continue
		}

		elem2, ok := raw.(*map[string]models.Request)
		if ok {
			reqs = elem2
			continue
		}
	}

	// Fail on nil pointer(s).
	if users == nil || reqs == nil {
		l.Msg("users and/or reqs are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// Iterate over requests to find misdeleted requests.
	for key, req := range *reqs {
		// Check the request validity = the activation request is still valid and the user has been already activated.
		if !time.Now().After(req.CreatedAt.Add(time.Hour*24)) && req.Type == "activation" && ((*users)[req.Nickname].Active || (*users)[req.Nickname].Options["active"]) {
			// Delete the misdeleted request.
			if deleted := DeleteOne(RequestCache, key); !deleted {
				l.Msg("cannot delete the request: " + key).Status(http.StatusInternalServerError).Log()
				return false
			}
		}
	}

	const timeLayout = "2006-Jan-02"

	// Iterate over users to patch the Active bool's state according to the user's registration date.
	for key, user := range *users {
		migrationCreationDate, err := time.Parse(timeLayout, "2024-Oct-24")
		if err != nil {
			l.Msg("cannot parse the migration createdAt date").Status(http.StatusInternalServerError).Log()
			return false
		}

		// The user is not activated and the registration time is (way) before the migration creating datetime = make active.
		if (!user.Active || !user.Options["active"]) && migrationCreationDate.After(user.RegisteredTime) {
			user.Active = true

			if user.Options == nil {
				user.Options = make(map[string]bool)
			}
			user.Options["active"] = true

			if saved := SetOne(UserCache, key, user); !saved {
				l.Msg("cannot save an user: " + key).Status(http.StatusInternalServerError).Log()
				return false
			}
		}
	}

	return true
}

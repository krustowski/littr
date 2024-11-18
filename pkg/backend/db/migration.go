package db

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/config"
	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"
)

// migrationFunc declares the unified type for any migration function.
type migrationFunc func(common.Logger, []interface{}) bool

// migrationProp is a struct to hold the migration function reference, and array of interfaces (mainly pointers) of various length.
type migrationProp struct {
	// Migration's name.
	N string

	// Migration's function handle.
	F migrationFunc

	// Migration's resources to process.
	R []interface{}
}

// RunMigrations is a wrapper function for the migration registration and execution.
func RunMigrations() string {
	l := common.NewLogger(nil, "migrations")

	// Fetch all the data for the migration procedures.
	//polls, pollCount := GetAll(PollCache, models.Poll{})
	posts, _ := GetAll(FlowCache, models.Post{})
	reqs, _ := GetAll(RequestCache, models.Request{})
	subs, _ := GetAll(SubscriptionCache, []models.Device{})
	tokens, _ := GetAll(TokenCache, models.Token{})
	users, _ := GetAll(UserCache, models.User{})

	// This is/was just to check whether there are any data entering the migration starter.
	//l.Msg(fmt.Sprintf("counts: posts, %d, reqs: %d, subs: %d, tokens: %d, users: %d", postCount, reqCount, subCount, tokenCount, userCount)).Status(http.StatusOK).Log()

	// Define the migration procedures order.
	var migrationsOrderedList = []migrationProp{
		{
			N: "migrateExpiredReuests",
			F: migrateExpiredRequests,
			R: []interface{}{&reqs},
		},
		{
			N: "migrateExpiredTokens",
			F: migrateExpiredTokens,
			R: []interface{}{&tokens},
		},
		{
			N: "migrateDeleteBlankDevices",
			F: migrateDeleteBlankDevices,
			R: []interface{}{&subs},
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
		{
			N: "migrateUserOptions",
			F: migrateUserOptions,
			R: []interface{}{&users},
		},
		/*{
			N: "migratePolls",
			F: migratePolls,
			R: []interface{}{&users, &polls},
		},*/
	}

	// Declare the migrations report variable.
	var report string

	// Execute the migration procedures.
	for _, mig := range migrationsOrderedList {
		report += fmt.Sprintf("[%s]: %t, ", mig.N, mig.F(l.SetPrefix(mig.N), mig.R))
	}

	//report += fmt.Sprintf("; aftermath: posts: %d, requests: %d, subscriptions: %d, tokens: %d, users: %d", len(posts), len(reqs), len(subs), len(tokens), len(users))

	// Run the GC to tidy up.
	runtime.GC()

	// Remove the prefix for the further Logger instance usage.
	l.RemovePrefix()
	return report
}

// migrateExpiredRequests procedure loops over requests and removes those expired already.
func migrateExpiredRequests(l common.Logger, rawElems []interface{}) bool {
	var reqs *map[string]models.Request

	// Assert pointers from the interface array.
	for _, raw := range rawElems {
		// Try the reqs pointer.
		elem, ok := raw.(*map[string]models.Request)
		if ok {
			reqs = elem
			continue
		}
	}

	// Exit if the reqs pointer is nil.
	if reqs == nil {
		l.Msg("reqs are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// Loop over reqs and compare their times with durations, if expired, yeet them.
	for uuid, req := range *reqs {
		/*t, err := time.Parse("2006-01-02T15:04:05.000000000-07:00", req.CreatedAt.String())
		if err != nil {
			l.Msg("could not parse request's time: " + uuid).Status(http.StatusInternalServerError).Log()
			return false
		}*/

		if time.Now().After(req.CreatedAt.Add(time.Hour * 24)) {
			// Expired request = delete it.
			if deleted := DeleteOne(RequestCache, uuid); !deleted {
				l.Msg("could not delete request: " + uuid).Status(http.StatusInternalServerError).Log()
				return false
			}

			// Delete from the reqs map locally within the migrations.
			delete(*reqs, uuid)
		}
	}

	return true
}

// migrateExpiredTokens procedure loop over tokens and removes those beyond the expiry.
func migrateExpiredTokens(l common.Logger, rawElems []interface{}) bool {
	var tokens *map[string]models.Token

	// Assert pointers from the interface array.
	for _, raw := range rawElems {
		// Try the tokens pointer.
		elem, ok := raw.(*map[string]models.Token)
		if ok {
			tokens = elem
			continue
		}
	}

	// Exit if the tokens pointer is nil.
	if tokens == nil {
		l.Msg("tokens are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// Loop over tokens and compare their times with durations, if expired, yeet them.
	for hash, token := range *tokens {
		if time.Now().After(token.CreatedAt.Add(common.TOKEN_TTL)) {
			// Expired token = delete it.
			if deleted := DeleteOne(TokenCache, hash); !deleted {
				l.Msg("could not delete token: " + hash).Status(http.StatusInternalServerError).Log()
				return false
			}

			// Delete from the tokens map locally within the migrations.
			delete(*tokens, hash)
		}
	}

	return true
}

// migrateDeleteBlankDevices procedure ensures blank devices are omitted from the user's device list in SubscriptionCache.
func migrateDeleteBlankDevices(l common.Logger, rawElems []interface{}) bool {
	var subs *map[string][]models.Device

	// Assert pointers from the interface array.
	for _, raw := range rawElems {
		elem, ok := raw.(*map[string][]models.Device)
		if ok {
			subs = elem
			continue
		}
	}

	// Exit if the subs pointer is nil.
	if subs == nil {
		l.Msg("subs are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	for userID, devs := range *subs {
		var newDevs []models.Device

		for _, dev := range devs {
			if reflect.DeepEqual(dev, (models.Device{})) {
				continue
			}

			tags := dev.Tags
			dev.Tags = nil

			if reflect.DeepEqual(dev, (models.Device{})) {
				continue
			}

			dev.Tags = tags

			newDevs = append(newDevs, dev)
		}

		if len(newDevs) == 0 {
			if deleted := DeleteOne(SubscriptionCache, userID); !deleted {
				l.Msg("could not delete devices: " + userID).Status(http.StatusInternalServerError).Log()
				return false
			}

			delete(*subs, userID)
			return true
		}

		if saved := SetOne(SubscriptionCache, userID, newDevs); !saved {
			l.Msg("could not save new devices: " + userID).Status(http.StatusInternalServerError).Log()
			return false
		}

		(*subs)[userID] = newDevs
	}

	return true
}

// migrateEmptyDeviceTags procedure takes care of filling empty device tags arrays.
func migrateEmptyDeviceTags(l common.Logger, rawElems []interface{}) bool {
	var subs *map[string][]models.Device

	// Assert pointers from the interface array.
	for _, raw := range rawElems {
		elem, ok := raw.(*map[string][]models.Device)
		if ok {
			subs = elem
			continue
		}
	}

	// Exit if the subs pointer is nil.
	if subs == nil {
		l.Msg("subs are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// Look for empty tags of subscriptions/devices in nested loops.
	for key, devs := range *subs {
		changed := false

		// Iterate over devices.
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
			// Save the changes found and made.
			if saved := SetOne(SubscriptionCache, key, devs); !saved {
				l.Msg("cannot save changed dev: " + key).Status(http.StatusInternalServerError).Log()
				return false
			}

			// Update the subs map locally within the migrations.
			(*subs)[key] = devs
		}
	}

	return true
}

// migrateAvatarURL procedure takes care of (re)assigning custom, or default avatars to all users having blank or default strings saved in their data chunk. Function returns bool based on the process result.
func migrateAvatarURL(l common.Logger, rawElems []interface{}) bool {
	var users *map[string]models.User

	// Assert pointers from the interface array.
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

	// Make zero-size channel array for the per user run goroutines. The array is incremented/appended dynamically to ensure proper channel closures.
	var channels = make([]chan interface{}, 0)
	//var channels = make([]chan avatarResult, len(*users))
	var wg sync.WaitGroup
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

		// Skip the custom avatar uploaded via the settings view/page.
		if strings.Contains(user.AvatarURL, "/web/pix/thumb_") {
			continue
		}

		// Add one new worker to the WaitGroup, and create a channel for it.
		wg.Add(1)
		channels = append(channels, make(chan interface{}, 1))

		// Run the gravatar goroutine, increment the chan/goroutines count.
		go GetGravatarURL(user, channels[i], &wg)
		i++
	}

	// Retrieve the results = merge the channels into one. See pkg/backend/db/gravatar.go for more.
	results := FanInChannels(l, channels...)
	wg.Wait()

	// Collect the results.
	for rawR := range results {
		result, ok := rawR.(*avatarResult)
		if !ok {
			l.Msg("corrupted output from the avatarResult chan...").Status(http.StatusInternalServerError).Log()
		}

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
func migrateFlowPurge(l common.Logger, rawElems []interface{}) bool {
	var posts *map[string]models.Post
	var users *map[string]models.User

	// Assert pointers from the interface array.
	for _, raw := range rawElems {
		// Try the users pointer.
		elem, ok := raw.(*map[string]models.User)
		if ok {
			users = elem
			continue
		}

		// Try the posts pointer.
		elem2, ok := raw.(*map[string]models.Post)
		if ok {
			posts = elem2
			continue
		}
	}

	// Exit on nil pointer(s).
	if users == nil || posts == nil {
		l.Msg("users or posts are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// Loop over posts and delete author-less posts.
	for key, post := range *posts {
		// If the post exists, but its author not, delete the post first.
		if _, found := (*users)[post.Nickname]; !found {
			if deleted := DeleteOne(FlowCache, key); !deleted {
				l.Msg("cannot delete post: " + key).Status(http.StatusInternalServerError).Log()
				return false
			}

			// Delete from the posts map locally within the migrations.
			delete(*posts, key)
		}
	}

	return true
}

// migrateUserDeletion procedure takes care of default users deletion from the database. Function returns bool based on the process result.
func migrateUserDeletion(l common.Logger, rawElems []interface{}) bool {
	var posts *map[string]models.Post
	var users *map[string]models.User

	// Assert pointers from the interface array.
	for _, raw := range rawElems {
		// Try the users pointer.
		elem, ok := raw.(*map[string]models.User)
		if ok {
			users = elem
			continue
		}

		// Try the posts pointer.
		elem2, ok := raw.(*map[string]models.Post)
		if ok {
			posts = elem2
			continue
		}
	}

	// Exit if any of the pointer is nil.
	if users == nil || posts == nil {
		l.Msg("users and/or posts are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// Fetch the bank of user deletion list.
	bank := &config.UserDeletionList

	// Delete all users matching the contents of restricted nickname list.
	for key, user := range *users {
		if helpers.Contains(*bank, user.Nickname) {
			l.Msg("deleting " + user.Nickname).Status(http.StatusProcessing).Log()

			// Delete the user from the User database.
			if deleted := DeleteOne(UserCache, key); !deleted {
				l.Msg("cannot delete an user: " + key).Status(http.StatusInternalServerError).Log()
				return false
			}

			// Delete the user locally within the migrations.
			delete(*users, key)
		}
	}

	// Delete all user's posts.
	for key, post := range *posts {
		if helpers.Contains(*bank, post.Nickname) {
			// Delete the post from the Flow database.
			if deleted := DeleteOne(FlowCache, key); !deleted {
				l.Msg("cannot delete a post: " + key).Status(http.StatusInternalServerError).Log()
				return false
			}

			// Delete the post locally within the migrations.
			delete(*posts, key)
		}
	}

	return true
}

// migrateUserRegisteredTime procedure fixes the initial registration date if it defaults to the "null" time.Time string. Function returns bool based on the process result.
func migrateUserRegisteredTime(l common.Logger, rawElems []interface{}) bool {
	var users *map[string]models.User

	// Assert pointers from the interface array.
	for _, raw := range rawElems {
		elem, ok := raw.(*map[string]models.User)
		if ok {
			users = elem
			continue
		}
	}

	// Exit if the users pointer is nil.
	if users == nil {
		l.Msg("users are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// Loop over users and fix their registered datetime.
	for key, user := range *users {
		// TDA way of defining a first year AD.
		if user.RegisteredTime == time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC) {
			// Set the registration time to Sep 1, 2023 by default.
			user.RegisteredTime = time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC)

			// Update the user in the User database.
			if ok := SetOne(UserCache, key, user); !ok {
				l.Msg("cannot save an user: " + key).Status(http.StatusInternalServerError).Log()
				return false
			}

			// Update the users map locally within the migrations.
			(*users)[key] = user
		}
	}

	return true
}

// migrateUserShadeList procedure lists ShadeList items and ensures user shaded (no mutual following, no replying).
func migrateUserShadeList(l common.Logger, rawElems []interface{}) bool {
	var users *map[string]models.User

	// Assert pointers from the interface array.
	for _, raw := range rawElems {
		elem, ok := raw.(*map[string]models.User)
		if ok {
			users = elem
			continue
		}
	}

	// Exit if the users pointer is nil.
	if users == nil {
		l.Msg("users are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// Loop over users, create the lists if needed.
	for key, user := range *users {
		flowList := user.FlowList
		shadeList := user.ShadeList

		// Patch the flowList/shadeList nil map.
		if flowList == nil {
			flowList = make(map[string]bool)
		}
		if shadeList == nil {
			shadeList = make(map[string]bool)
		}

		// Loop over the shadeList, making sure that the user themselves is not shaded, nor the system user.
		for name, shaded := range shadeList {
			// If shaded and the shaded nickname differs from the user's nickname.
			if shaded && name != user.Nickname {
				flowList[name] = false
				user.FlowList = flowList

				// Update the user in the User database. It is enough to save the user later below.
				/*if saved := setOne(UserCache, key, user); !saved {
					l.Msg("cannot save user: " + key).Status(http.StatusInternalServerError).Log()
					return false
				}*/
			}

		}

		// Ensure that users can see themselves.
		flowList[key] = true
		user.FlowList = flowList

		// Save the user again in the User dataabse.
		if saved := SetOne(UserCache, key, user); !saved {
			l.Msg("cannot save user: " + key).Status(http.StatusInternalServerError).Log()
			return false
		}

		// Update the users map locally within the migrations.
		(*users)[key] = user
	}

	return true
}

// migrateUserUnshade procedure lists all users and unshades manually some explicitly list users.
func migrateUserUnshade(l common.Logger, rawElems []interface{}) bool {
	var users *map[string]models.User

	// Assert pointers from the interface array.
	for _, raw := range rawElems {
		elem, ok := raw.(*map[string]models.User)
		if ok {
			users = elem
			continue
		}
	}

	// Exit if the users pointer is nil.
	if users == nil {
		l.Msg("users are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// Fetch the bank of users to unshade.
	usersToUnshade := &config.UsersToUnshade

	// Loop over users, find those needing an acute unshading.
	for key, user := range *users {
		if !helpers.Contains(*usersToUnshade, key) {
			continue
		}

		// Patch the nil shadeList map.
		shadeList := user.ShadeList
		if shadeList == nil {
			shadeList = make(map[string]bool)
		}

		// Loop over the shadeList, find those needing an acure unshading.
		for name := range shadeList {
			if helpers.Contains(*usersToUnshade, name) {
				shadeList[name] = false
			}
		}

		// Update the user's shadeList.
		user.ShadeList = shadeList

		// Update the user's shadeList in the User database.
		if ok := SetOne(UserCache, key, user); !ok {
			l.Msg("cannot save an user: " + key).Status(http.StatusInternalServerError).Log()
			return false
		}

		// Update the users map locally within the migrations.
		(*users)[key] = user
	}

	return true
}

// migrateBlankAboutText procedure loops over user accounts and adds "newbie" where the about-text field is blank.
func migrateBlankAboutText(l common.Logger, rawElems []interface{}) bool {
	var users *map[string]models.User

	// Assert pointers from the interface array.
	for _, raw := range rawElems {
		elem, ok := raw.(*map[string]models.User)
		if ok {
			users = elem
			continue
		}
	}

	// Exit if the users pointer is nil.
	if users == nil {
		l.Msg("users are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// Loop over users to find if anyone has the empty About text string.
	for key, user := range *users {
		if len(user.About) == 0 {
			// Set the default about text (bio).
			user.About = "newbie"
		}

		// Update the user in the User database.
		if saved := SetOne(UserCache, key, user); !saved {
			l.Msg("cannot save an user: " + key).Status(http.StatusInternalServerError).Log()
			return false
		}

		// Update the users map locally within the migrations.
		(*users)[key] = user
	}

	return true
}

// migrateSystemFlowOn procedure ensures everyone has system account in the flow.
func migrateSystemFlowOn(l common.Logger, rawElems []interface{}) bool {
	var users *map[string]models.User

	// Assert pointers from the interface array.
	for _, raw := range rawElems {
		elem, ok := raw.(*map[string]models.User)
		if ok {
			users = elem
			continue
		}
	}

	// Exit if the users pointer is nil.
	if users == nil {
		l.Msg("users are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// Loop over users and fix the system followment.
	for key, user := range *users {
		// Patch the flowList nil map.
		if user.FlowList == nil {
			user.FlowList = make(map[string]bool)
		}

		// Set the flowList defaults.
		user.FlowList[user.Nickname] = true
		user.FlowList["system"] = true

		// Update the user in the User database.
		if saved := SetOne(UserCache, key, user); !saved {
			l.Msg("cannot save an user: " + key).Status(http.StatusInternalServerError).Log()
			return false
		}

		// Update the users map locally within the migrations.
		(*users)[key] = user
	}

	return true
}

// migrateUserActiveState ensures all users registered before Oct 28, 2024 are activated; otherwise it also tries to delete valid, but misdeleted activation requests from its database.
func migrateUserActiveState(l common.Logger, rawElems []interface{}) bool {
	var users *map[string]models.User
	var reqs *map[string]models.Request

	// Assert pointers from the interface array.
	for _, raw := range rawElems {
		// Try the users pointer.
		elem, ok := raw.(*map[string]models.User)
		if ok {
			users = elem
			continue
		}

		// Try the reqs pointer.
		elem2, ok := raw.(*map[string]models.Request)
		if ok {
			reqs = elem2
			continue
		}
	}

	// Exit on the nil pointer(s).
	if users == nil || reqs == nil {
		l.Msg("users and/or reqs are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// Iterate over requests to find misdeleted requests.
	for key, req := range *reqs {
		// Check the request validity = the activation request could still be valid, but the user has been already activated.
		if !time.Now().After(req.CreatedAt.Add(time.Hour*24)) &&
			req.Type == "activation" &&
			((*users)[req.Nickname].Active ||
				(*users)[req.Nickname].Options["active"]) {

			// Delete the misdeleted request.
			if deleted := DeleteOne(RequestCache, key); !deleted {
				l.Msg("cannot delete the request: " + key).Status(http.StatusInternalServerError).Log()
				return false
			}

			// Delete from the reqs map locally within the migrations.
			delete(*reqs, key)
		}
	}

	// Just a datetime layout for the upcoming hardcoded datetime parsing.
	const timeLayout = "2006-Jan-02"

	// Iterate over users to patch the Active bool's state according to the user's registration date.
	for key, user := range *users {
		// The date when this migration subprocedure was created. For comparison with the later user registrations.
		migrationCreationDate, err := time.Parse(timeLayout, "2024-Oct-28")
		if err != nil {
			l.Msg("cannot parse the migration createdAt date").Status(http.StatusInternalServerError).Log()
			return false
		}

		// The user is not activated and the registration time is (way) before the migration creating datetime = make active automatically.
		if (!user.Active || !user.Options["active"]) && migrationCreationDate.After(user.RegisteredTime) {
			user.Active = true

			// Patch the nil options map.
			if user.Options == nil {
				user.Options = make(map[string]bool)
			}
			user.Options["active"] = true

			// Update the user in the User database.
			if saved := SetOne(UserCache, key, user); !saved {
				l.Msg("cannot save an user: " + key).Status(http.StatusInternalServerError).Log()
				return false
			}

			// Update the users map locally within the migrations.
			(*users)[key] = user
		}
	}

	return true
}

// migrateUserOptions procedure ensures that every user has a proper set of all options according to the legacy models.User fields.
func migrateUserOptions(l common.Logger, rawElems []interface{}) bool {
	var users *map[string]models.User

	// Assert pointers from the interface array.
	for _, raw := range rawElems {
		// Try the users pointer.
		elem, ok := raw.(*map[string]models.User)
		if ok {
			users = elem
			continue
		}
	}

	// Exit on the nil pointer(s).
	if users == nil {
		l.Msg("users are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// Loop over all user accounts to patch their (not only nil) options maps.
	for key, user := range *users {
		// Patch the possible nil map.
		if user.Options == nil {
			user.Options = models.DefaultUserOptionsMap
		}

		options := user.Options

		// active | ativation | activated: false

		if _, found := options["active"]; !found {
			options["active"] = user.Active
		}

		// GDPR | gdpr: true

		if value, found := options["gdpr"]; !value || !found {
			// Every single user accepts the GDPR notice at the registration time, thus this option should always be true.
			options["gdpr"] = true
		}

		// private: false

		if _, found := options["private"]; !found {
			options["private"] = user.Private
		}

		// uiDarkMode: true

		if _, found := options["uiDarkMode"]; !found {
			options["uiDarkMode"] = user.UIDarkMode
		}

		// liveMode: true

		if value, found := options["liveMode"]; !value || !found {
			// For now, just sitck with the fact that this option always has got its switch disabled in settings,
			// thus no one could possibly change it to anything else so far.
			options["liveMode"] = true
		}

		// localTimeMode: true

		if _, found := options["localTimeMode"]; !found {
			options["localTimeMode"] = user.LocalTimeMode
		}

		// Assign the options map back to its owner, and update the owner in the User database.
		user.Options = options

		if saved := SetOne(UserCache, key, user); !saved {
			l.Msg("cannot save an user: " + key).Status(http.StatusInternalServerError).Log()
			return false
		}

		// Update the users map locally within the migration procedures.
		(*users)[key] = user
	}

	return true
}

// migratePolls procedure loops over all polls making sure that any poll with the non-existing Author will be omitted/deleted.
func migratePolls(l common.Logger, rawElems []interface{}) bool {
	var polls *map[string]models.Poll
	var users *map[string]models.User

	// Assert pointers from the interface array.
	for _, raw := range rawElems {
		// Try the polls pointer.
		elem, ok := raw.(*map[string]models.Poll)
		if ok {
			polls = elem
			continue
		}

		// Try the users pointer.
		elem2, ok := raw.(*map[string]models.User)
		if ok {
			users = elem2
			continue
		}
	}

	// Exit on the nil pointer(s).
	if polls == nil || users == nil {
		l.Msg("polls and/or users are nil").Status(http.StatusInternalServerError).Log()
		return false
	}

	// Loop over all polls to check if their Author is still there.
	for key, poll := range *polls {
		// Check the user's existence in the local users export.
		if _, found := (*users)[poll.Author]; !found {
			// Delete the poll once for all.
			if deleted := DeleteOne(PollCache, key); !deleted {
				l.Msg("cannot delete a poll: " + key).Status(http.StatusInternalServerError).Log()
				return false
			}

			// Delete the poll locally too within the migration procedures.
			delete(*polls, key)
		}
	}

	return true
}

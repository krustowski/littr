package stats

import (
	"net/http"
	"time"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/models"
)

// GetStats acts like a handler for stats page
//
// @Summary      Get stats
// @Description  get stats
// @Tags         stats
// @Accept       json
// @Produce      json
// @Success      200  {array}   common.APIResponse
// @Router       /stats/ [get]
func getStats(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "stats")

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// fetch the data
	polls, _ := db.GetAll(db.PollCache, models.Poll{})
	posts, postCount := db.GetAll(db.FlowCache, models.Post{})
	users, _ := db.GetAll(db.UserCache, models.User{})

	// flush email addresses
	for key, user := range users {
		user.Email = ""
		users[key] = user
	}

	// prepare the maps for export
	flowStats := make(map[string]int)
	userStats := make(map[string]models.UserStat)

	flowers := make(map[string]int)
	shades := make(map[string]int)

	// init the flowStats
	flowStats["posts"] = postCount
	//flowStats["users"] = userCount
	flowStats["users"] = -1
	flowStats["stars"] = 0

	// iterate over all posts, compose stats results
	for _, val := range posts {
		// increment user's stats
		stat, ok := userStats[val.Nickname]
		if !ok {
			// create a blank stat
			stat = models.UserStat{}
			stat.Searched = true
		}

		stat.PostCount++
		stat.ReactionCount += val.ReactionCount
		userStats[val.Nickname] = stat
		flowStats["stars"] += val.ReactionCount
	}

	// iterate over all users, compose global flower and shade count
	for _, user := range users {
		for key, enabled := range user.FlowList {
			if enabled && key != user.Nickname {
				flowers[key]++
			}
		}

		for key, shaded := range user.ShadeList {
			if shaded && key != user.Nickname {
				shades[key]++
			}
		}

		// check the online status
		diff := time.Since(user.LastActiveTime)
		if diff < 15*time.Minute {
			flowStats["online"]++
		}

		flowStats["users"]++
	}

	// iterate over composed flowers, assign the count to an user
	for key, count := range flowers {
		stat := userStats[key]

		// FlowList also contains the user itself
		stat.FlowerCount = count
		userStats[key] = stat
	}

	// iterate over compose shades, assign the count to an user
	for key, count := range shades {
		stat := userStats[key]

		// FlowList also contains the user itself
		stat.ShadeCount = count
		userStats[key] = stat
	}

	// iterate over all polls, count them good
	for _, poll := range polls {
		flowStats["polls"]++

		flowStats["votes"] += poll.OptionOne.Counter
		flowStats["votes"] += poll.OptionTwo.Counter
		flowStats["votes"] += poll.OptionThree.Counter
	}

	pl := struct {
		FlowStats map[string]int             `json:"flow_stats"`
		UserStats map[string]models.UserStat `json:"user_stats"`
		Users     map[string]models.User     `json:"users"`
	}{
		FlowStats: flowStats,
		UserStats: userStats,
		Users:     *common.FlushUserData(&users, callerID),
	}

	l.Msg("ok, dumping user and system stats").Status(http.StatusOK).Log().Payload(&pl).Write(w)
	return
}

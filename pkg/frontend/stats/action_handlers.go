package stats

import (
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func (c *Content) handleSearch(ctx app.Context, a app.Action) {
	matchedList := []string{}

	val, ok := a.Value.(string)
	if !ok {
		return
	}

	ctx.Async(func() {
		users := c.userStats

		// iterate over calculated stats' "rows" and find matchings
		for key, user := range users {
			//user := users[key]
			user.Searched = false

			// use lowecase to search across UNICODE letters
			lval := strings.ToLower(val)
			lkey := strings.ToLower(key)

			if strings.Contains(lkey, lval) {
				user.Searched = true

				//matchedList = append(matchedList, key)
			}

			users[key] = user
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.userStats = users
			c.nicknames = matchedList

			c.loaderShow = false
		})
		return
	})

}

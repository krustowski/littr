package frontend

import (
	"encoding/json"
	"log"
	"strconv"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type StatsPage struct {
	app.Compo
}

type statsContent struct {
	app.Compo

	postCount int

	posts map[string]models.Post
	stats map[string]int

	loaderShow bool
}

func (p *StatsPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("stats / littr")
}

func (p *StatsPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		&footer{},
		&statsContent{},
	)
}

func (c *statsContent) OnMount(ctx app.Context) {
	c.loaderShow = true
}


func (c *statsContent) OnNav(ctx app.Context) {
	ctx.Async(func() {
		postsRaw := struct {
			Posts map[string]models.Post `json:"posts"`
			Count int `json:"count"`
		}{}

		if byteData, _ := litterAPI("GET", "/api/flow", nil); byteData != nil {
			err := json.Unmarshal(*byteData, &postsRaw)
			if err != nil {
				log.Println(err.Error())
				return
			}
		} else {
			log.Println("cannot fetch post flow list")
			return
		}

		// Storing HTTP response in component field:
		ctx.Dispatch(func(ctx app.Context) {
			c.posts = postsRaw.Posts
			c.postCount = postsRaw.Count
			//c.sortedPosts = posts

			c.loaderShow = false
			log.Println("dispatch ends")
		})
		return
	})
}

func (c *statsContent) calculateStats() map[string]int {
	stats := make(map[string]int)

	stats["flow_posts_total_count"] = c.postCount

	// iterate over all posts, compose stats results
	for _, val := range c.posts {
		// increment user's stats
		stats["users_" + val.Nickname + "_posts_count"]++
		stats["users_" + val.Nickname + "_total_reactions"] += val.ReactionCount
	}

	return stats
}

func (c *statsContent) Render() app.UI {
	stats := c.calculateStats()

	loaderActiveClass := ""
	if c.loaderShow {
		loaderActiveClass = " active"
	}

	return app.Main().Class("responsive").Body(
		app.H5().Text("littr stats").Style("padding-top", config.HeaderTopPadding),
		app.P().Text("wanna know your flow stats? how many you got in the flow and vice versa? yo"),
		app.Div().Class("space"),

		app.If(c.loaderShow,
			app.Div().Class("small-space"),
			app.Div().Class("loader center large deep-orange"+loaderActiveClass),
		),

		app.Range(stats).Map(func(key string) app.UI {
			return app.P().Body(
				app.Text(key + ": " + strconv.FormatInt(int64(stats[key]), 10)),
				app.Br(),
			)
		}),
	)
}

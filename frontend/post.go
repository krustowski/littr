package frontend

import (
	"log"
	"net/url"
	"strconv"
	"time"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type PostPage struct {
	app.Compo
}

type postContent struct {
	app.Compo

	newPost    string
	newFigLink string

	pollQuestion  string
	pollOptionI   string
	pollOptionII  string
	pollOptionIII string

	toastShow bool
	toastText string
}

func (p *PostPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("post / littr")
}

func (p *PostPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		&footer{},
		&postContent{},
	)
}

func (c *postContent) onClick(ctx app.Context, e app.Event) {
	// post, fig, poll
	postType := ctx.JSSrc().Get("id").String()
	content := ""
	poll := models.Poll{}

	var payload interface{}

	switch postType {
	case "fig":
		if c.newFigLink == "" {
			c.toastShow = true
			c.toastText = "fig link must be filled"
			return
		}

		if _, err := url.ParseRequestURI(c.newFigLink); err != nil {
			c.toastShow = true
			c.toastText = "fig link prolly not a valid URL"
			return
		}
		content = c.newFigLink

		break

	case "poll":
		if c.pollOptionI == "" || c.pollOptionII == "" || c.pollQuestion == "" {
			c.toastShow = true
			c.toastText = "poll question and at least two option have to be filled"
			return
		}
		now := time.Now()
		content = strconv.FormatInt(now.UnixNano(), 10)

		poll.ID = content
		poll.Question = c.pollQuestion
		poll.OptionOne.Content = c.pollOptionI
		poll.OptionTwo.Content = c.pollOptionII
		poll.OptionThree.Content = c.pollOptionIII
		poll.Timestamp = now

		break

	case "post":
		if c.newPost == "" {
			c.toastShow = true
			c.toastText = "post textarea must be filled"
			return
		}
		content = c.newPost

		break

	default:
		return
		break
	}

	ctx.Async(func() {
		var enUser string
		var user models.User

		ctx.LocalStorage().Get("user", &enUser)

		// decode, decrypt and unmarshal the local storage string
		if err := prepare(enUser, &user); err != nil {
			c.toastText = "frontend decoding/decryption failed: " + err.Error()
			c.toastShow = true
			return
		}

		author := user.Nickname
		path := "/api/flow"

		if postType == "post" || postType == "fig" {
			payload = models.Post{
				Nickname:  author,
				Type:      postType,
				Content:   content,
				Timestamp: time.Now(),
				PollID:    poll.ID,
			}
		} else if postType == "poll" {
			path = "/api/polls"
			payload = poll
		}

		// add new post/poll to backend struct
		if _, ok := litterAPI("POST", path, payload); !ok {
			c.toastShow = true
			c.toastText = "backend error: cannot add new content"
			log.Println("cannot post new content to API!")
			return
		}

		c.toastShow = false
		ctx.Navigate("/flow")
	})
}

func (c *postContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastShow = false
}

func (c *postContent) Render() app.UI {
	toastActiveClass := ""
	if c.toastShow {
		toastActiveClass = " active"
	}

	return app.Main().Class("responsive").Body(
		app.H5().Text("add flow post").Style("padding-top", config.HeaderTopPadding),
		app.P().Text("drop it, drop it"),

		app.A().OnClick(c.dismissToast).Body(
			app.Div().Class("toast red10 white-text top"+toastActiveClass).Body(
				app.I().Text("error"),
				app.Span().Text(c.toastText),
			),
		),

		app.Div().Class("field textarea label border invalid extra deep-orange-text").Body(
			app.Textarea().Class("active").Name("newPost").OnChange(c.ValueTo(&c.newPost)).AutoFocus(true),
			app.Label().Text("text contents").Class("active"),
		),
		app.Button().ID("post").Class("responsive deep-orange7 white-text bold").Text("post text").OnClick(c.onClick),

		app.Div().Class("large-divider"),

		app.H5().Text("add flow fig").Style("padding-top", config.HeaderTopPadding),
		app.P().Text("provide me with the image URL, papi"),
		app.Div().Class("space"),

		app.Div().Class("field label border invalid extra deep-orange-text").Body(
			app.Input().Class("active").Type("text").OnChange(c.ValueTo(&c.newFigLink)),
			//app.Input().Class("active").Type("file"),
			app.Label().Text("fig link").Class("active"),
			app.I().Text("attach_file"),
		),
		app.Button().ID("fig").Class("responsive deep-orange7 white-text bold").Text("post fig").OnClick(c.onClick),

		app.Div().Class("large-divider"),

		app.H5().Text("add flow poll").Style("padding-top", config.HeaderTopPadding),
		app.P().Text("lmao gotem"),
		app.Div().Class("space"),

		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("text").OnChange(c.ValueTo(&c.pollQuestion)).Required(true).Class("active").MaxLength(50),
			app.Label().Text("question").Class("active"),
		),
		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("text").OnChange(c.ValueTo(&c.pollOptionI)).Required(true).Class("active").MaxLength(50),
			app.Label().Text("option one").Class("active"),
		),
		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("text").OnChange(c.ValueTo(&c.pollOptionII)).Required(true).Class("active").MaxLength(50),
			app.Label().Text("option two").Class("active"),
		),
		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("text").OnChange(c.ValueTo(&c.pollOptionIII)).Required(false).Class("active").MaxLength(60),
			app.Label().Text("option three (optional)").Class("active"),
		),
		app.Button().ID("poll").Class("responsive deep-orange7 white-text bold").Text("post poll").OnClick(c.onClick).Disabled(false),
	)
}

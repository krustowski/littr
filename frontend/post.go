package frontend

import (
	"log"
	//"net/url"
	"strconv"
	"strings"
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
	newFigFile string
	newFigData []byte

	pollQuestion  string
	pollOptionI   string
	pollOptionII  string
	pollOptionIII string

	toastShow bool
	toastText string
	toastType string

	postButtonsDisabled bool
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

// https://github.com/maxence-charriere/go-app/issues/882
func (c *postContent) handleFigUpload(ctx app.Context, e app.Event) {
	var toastText string

	file := e.Get("target").Get("files").Index(0)

	//log.Println("name", file.Get("name").String())
	//log.Println("size", file.Get("size").Int())
	//log.Println("type", file.Get("type").String())

	c.postButtonsDisabled = true

	ctx.Async(func() {
		if data, err := readFile(file); err != nil {
			toastText = err.Error()

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return

		} else {
			/*payload := models.Post{
				Nickname:  author,
				Type:      "fig",
				Content:   file.Get("name").String(),
				Timestamp: time.Now(),
				Data:      data,
			}*/

			// add new post/poll to backend struct
			/*if _, ok := litterAPI("POST", path, payload, user.Nickname, 0); !ok {
				toastText = "backend error: cannot add new content"
				log.Println("cannot post new content to API!")
			} else {
				ctx.Navigate("/flow")
			}*/

			toastText = "image uploaded"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastType = "success"
				c.toastText = toastText
				c.toastShow = (toastText != "")

				c.newFigFile = file.Get("name").String()
				c.newFigData = data
			})
			return

		}
	})
}

func (c *postContent) onClick(ctx app.Context, e app.Event) {
	// post, fig, poll
	postType := ctx.JSSrc().Get("id").String()
	content := ""
	poll := models.Poll{}
	toastText := ""

	var payload interface{}

	c.postButtonsDisabled = true

	ctx.Async(func() {
		switch postType {
		case "poll":
			// trim the padding spaces on the extremities
			// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
			pollQuestion := strings.TrimSpace(c.pollQuestion)
			pollOptionI := strings.TrimSpace(c.pollOptionI)
			pollOptionII := strings.TrimSpace(c.pollOptionII)
			pollOptionIII := strings.TrimSpace(c.pollOptionIII)

			if pollOptionI == "" || pollOptionII == "" || pollQuestion == "" {
				toastText = "poll question and at least two options have to be filled"
				break
			}

			now := time.Now()
			content = strconv.FormatInt(now.UnixNano(), 10)

			poll.ID = content
			poll.Question = pollQuestion
			poll.OptionOne.Content = pollOptionI
			poll.OptionTwo.Content = pollOptionII
			poll.OptionThree.Content = pollOptionIII
			poll.Timestamp = now

			break

		case "post":
			// trim the padding spaces on the extremities
			// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
			newPost := strings.TrimSpace(c.newPost)

			if newPost == "" {
				toastText = "post textarea must be filled"
				break
			}
			content = newPost

			break

		default:
			break
		}

		if toastText != "" {
			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		var enUser string
		var user models.User

		ctx.LocalStorage().Get("user", &enUser)

		// decode, decrypt and unmarshal the local storage user data
		if err := prepare(enUser, &user); err != nil {
			toastText = "frontend decoding/decryption failed: " + err.Error()
		}

		author := user.Nickname
		path := "/api/flow"

		if postType == "post" {
			payload = models.Post{
				Nickname: author,
				Type:     postType,
				Content:  content,
				PollID:   poll.ID,
				Figure:   c.newFigFile,
				Data:     c.newFigData,
				//Timestamp: time.Now(),
			}
		} else if postType == "poll" {
			path = "/api/polls"
			poll.Author = user.Nickname
			payload = poll
		}

		// add new post/poll to backend struct
		if _, ok := litterAPI("POST", path, payload, user.Nickname, 0); !ok {
			toastText = "backend error: cannot add new content"
			log.Println("cannot post new content to API!")
		} else {
			ctx.Navigate("/flow")
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = toastText
			c.toastShow = (toastText != "")
		})
		return
	})
}

func (c *postContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastText = ""
	c.toastShow = (c.toastText != "")
	c.toastType = ""
	c.postButtonsDisabled = false
}

func (c *postContent) Render() app.UI {
	toastColor := ""

	switch c.toastType {
	case "success":
		toastColor = "green10"
		break

	case "info":
		toastColor = "blue10"
		break

	default:
		toastColor = "red10"
	}

	return app.Main().Class("responsive").Body(
		app.H5().Text("add flow post").Style("padding-top", config.HeaderTopPadding),
		app.P().Text("drop it, drop it"),

		// snackbar
		app.A().OnClick(c.dismissToast).Body(
			app.If(c.toastText != "",
				app.Div().Class("snackbar white-text top active "+toastColor).Body(
					app.I().Text("error"),
					app.Span().Text(c.toastText),
				),
			),
		),

		// new post textarea
		app.Div().Class("field textarea label border extra deep-orange-text").Body(
			app.Textarea().Class("active").Name("newPost").OnChange(c.ValueTo(&c.newPost)).AutoFocus(true),
			app.Label().Text("post content").Class("active deep-orange-text"),
		),
		/*app.Button().ID("post").Class("responsive deep-orange7 white-text bold").OnClick(c.onClick).Disabled(c.postButtonsDisabled).Body(
			app.If(c.postButtonsDisabled,
				app.Progress().Class("circle white-border small"),
			),
			app.Text("post text"),
		),*/

		// new fig input
		app.Div().Class("field border label extra deep-orange-text").Body(
			app.Input().ID("fig-upload").Class("active").Type("file").OnChange(c.ValueTo(&c.newFigLink)).OnInput(c.handleFigUpload),
			app.Input().Class("active").Type("text").Value(c.newFigFile).Disabled(true),
			app.Label().Text("image").Class("active deep-orange-text"),
			app.I().Text("image"),
		),
		app.Div().Class("row").Body(
			app.Button().ID("post").Class("max deep-orange7 white-text bold").Style("border-radius", "8px").OnClick(c.onClick).Disabled(c.postButtonsDisabled).Body(
				app.If(c.postButtonsDisabled,
					app.Progress().Class("circle white-border small"),
				),
				app.Text("send new post"),
			),
		),

		app.Div().Class("space"),

		// new poll header text
		app.H5().Text("add flow poll").Style("padding-top", config.HeaderTopPadding),
		app.P().Text("lmao gotem"),
		app.Div().Class("space"),

		// newx poll input area
		app.Div().Class("field border label deep-orange-text").Body(
			app.Input().Type("text").OnChange(c.ValueTo(&c.pollQuestion)).Required(true).Class("active").MaxLength(50),
			app.Label().Text("question").Class("active deep-orange-text"),
		),
		app.Div().Class("field border label deep-orange-text").Body(
			app.Input().Type("text").OnChange(c.ValueTo(&c.pollOptionI)).Required(true).Class("active").MaxLength(50),
			app.Label().Text("option one").Class("active deep-orange-text"),
		),
		app.Div().Class("field border label deep-orange-text").Body(
			app.Input().Type("text").OnChange(c.ValueTo(&c.pollOptionII)).Required(true).Class("active").MaxLength(50),
			app.Label().Text("option two").Class("active deep-orange-text"),
		),
		app.Div().Class("field border label deep-orange-text").Body(
			app.Input().Type("text").OnChange(c.ValueTo(&c.pollOptionIII)).Required(false).Class("active").MaxLength(60),
			app.Label().Text("option three (optional)").Class("active deep-orange-text"),
		),
		app.Div().Class("row").Body(
			app.Button().ID("poll").Class("max deep-orange7 white-text bold").Style("border-radius", "8px").OnClick(c.onClick).Disabled(c.postButtonsDisabled).Body(
				app.If(c.postButtonsDisabled,
					app.Progress().Class("circle white-border small"),
				),
				app.Text("send new poll"),
			),
		),
		app.Div().Class("space"),
	)
}

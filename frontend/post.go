package frontend

import (
	"fmt"
	"log"
	"net/url"
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

	pollQuestion  string
	pollOptionI   string
	pollOptionII  string
	pollOptionIII string

	toastShow bool
	toastText string

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
			//fmt.Println(string(data))

			var enUser string
			var user models.User

			ctx.LocalStorage().Get("user", &enUser)

			// decode, decrypt and unmarshal the local storage user data
			if err := prepare(enUser, &user); err != nil {
				toastText = "frontend decoding/decryption failed: " + err.Error()
			}

			author := user.Nickname
			path := "/api/flow"

			payload := models.Post{
				Nickname:  author,
				Type:      "fig",
				Content:   file.Get("name").String(),
				Timestamp: time.Now(),
				Data:      data,
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

		}
	})
}

func readFile(file app.Value) (data []byte, err error) {
	done := make(chan bool)

	// https://developer.mozilla.org/en-US/docs/Web/API/FileReader
	reader := app.Window().Get("FileReader").New()
	reader.Set("onloadend", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		done <- true
		return nil
	}))
	reader.Call("readAsArrayBuffer", file)
	<-done

	readerError := reader.Get("error")
	if !readerError.IsNull() {
		err = fmt.Errorf("file reader error : %s", readerError.Get("message").String())
	} else {
		uint8Array := app.Window().Get("Uint8Array").New(reader.Get("result"))
		data = make([]byte, uint8Array.Length())
		app.CopyBytesToGo(data, uint8Array)
	}
	return data, err
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
		case "fig":
			// trim the padding spaces on the extremities
			// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
			newFigLink := strings.TrimSpace(c.newFigLink)

			if newFigLink == "" {
				toastText = "fig link must be filled"
				break
			}

			// check the URL/URI format
			if _, err := url.ParseRequestURI(newFigLink); err != nil {
				toastText = "fig link prolly not a valid URL"
				break
			}
			content = newFigLink

			break

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
	c.postButtonsDisabled = false
}

func (c *postContent) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.H5().Text("add flow post").Style("padding-top", config.HeaderTopPadding),
		app.P().Text("drop it, drop it"),

		// snackbar
		app.A().OnClick(c.dismissToast).Body(
			app.If(c.toastText != "",
				app.Div().Class("snackbar red10 white-text top active").Body(
					app.I().Text("error"),
					app.Span().Text(c.toastText),
				),
			),
		),

		// new post textarea
		app.Div().Class("field textarea label border invalid extra deep-orange-text").Body(
			app.Textarea().Class("active").Name("newPost").OnChange(c.ValueTo(&c.newPost)).AutoFocus(true),
			app.Label().Text("post content").Class("active"),
		),
		app.Button().ID("post").Class("responsive deep-orange7 white-text bold").OnClick(c.onClick).Disabled(c.postButtonsDisabled).Body(
			app.If(c.postButtonsDisabled,
				app.Progress().Class("circle white-border small"),
			),
			app.Text("post text"),
		),

		app.Div().Class("large-divider"),

		// new fig header text
		app.H5().Text("add flow fig").Style("padding-top", config.HeaderTopPadding),
		app.P().Text("provide me with the image URL, papi"),
		app.Div().Class("space"),

		// new fig input
		app.Div().Class("field label border invalid extra deep-orange-text").Body(
			app.Input().ID("fig-upload").Class("active").Type("file").OnChange(c.ValueTo(&c.newFigLink)).OnInput(c.handleFigUpload),
			app.Input().Class("active").Type("text"),
			app.Label().Text("fig link").Class("active"),
			app.I().Text("attach_file"),
		),
		app.Button().ID("fig").Class("responsive deep-orange7 white-text bold").OnClick(c.onClick).Disabled(c.postButtonsDisabled).Body(
			app.If(c.postButtonsDisabled,
				app.Progress().Class("circle white-border small"),
			),
			app.Text("post fig"),
		),

		app.Div().Class("large-divider"),

		// new poll header text
		app.H5().Text("add flow poll").Style("padding-top", config.HeaderTopPadding),
		app.P().Text("lmao gotem"),
		app.Div().Class("space"),

		// newx poll input area
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
		app.Button().ID("poll").Class("responsive deep-orange7 white-text bold").OnClick(c.onClick).Disabled(c.postButtonsDisabled).Body(
			app.If(c.postButtonsDisabled,
				app.Progress().Class("circle white-border small"),
			),
			app.Text("post poll"),
		),
		app.Div().Class("space"),
	)
}

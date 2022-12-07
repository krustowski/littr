package pages

import (
	"bytes"
	"encoding/json"
	"litter-go/backend"
	"log"
	"net/http"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type PostPage struct {
	app.Compo
}

type postContent struct {
	app.Compo
	newPost string

	toastShow bool
	toastText string
}

func postAPI(post backend.Post) bool {
	jsonData, err := json.Marshal(post)
	if err != nil {
		log.Println("cannot marshal post to register API")
		log.Println(err.Error())
		return false
	}

	bodyReader := bytes.NewReader([]byte(jsonData))

	req, err := http.NewRequest("POST", "/api/flow", bodyReader)
	if err != nil {
		log.Println(err.Error())
	}
	req.Header.Set("Content-Type", "application/byte")

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	log.Println("new post pushed to API")
	defer res.Body.Close()

	return true
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
	ctx.Async(func() {
		c.toastShow = true
		if c.newPost == "" {
			c.toastText = "post textarea must be filled"
			return
		}

		// add new post to backend struct
		if ok := postAPI(backend.Post{
			Content: c.newPost,
		}); !ok {
			log.Println("cannot post new post to API!")
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
		app.H5().Text("new flow post"),
		app.P().Text("pleasure to be lit"),

		app.A().OnClick(c.dismissToast).Body(
			app.Div().Class("toast red10 white-text top"+toastActiveClass).Body(
				app.I().Text("error"),
				app.Span().Text(c.toastText),
			),
		),

		app.Div().Class("field textarea label border invalid extra deep-orange-text").Body(
			app.Textarea().Name("newPost").OnChange(c.ValueTo(&c.newPost)).AutoFocus(true),
			app.Label().Text("contents"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").Text("post").OnClick(c.onClick),
	)
}

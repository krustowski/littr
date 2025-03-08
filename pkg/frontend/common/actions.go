package common

import (
	"strconv"

	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/models"
)

func HandleImageUpload(ctx app.Context, a app.Action, user *models.User, callback func()) {
	file, ok := a.Value.(app.Value)
	if !ok {
		callback()
		return
	}

	toast := Toast{AppContext: &ctx}

	ctx.Async(func() {
		defer callback()

		var (
			err      error
			imgBytes []byte
		)

		// Get the image bytes.
		imgBytes, err = ReadFile(file)
		if err != nil {
			toast.Text(err.Error()).Type(TTYPE_ERR).Dispatch()
			return
		}

		path := "/api/v1/users/" + user.Nickname + "/avatar"

		payload := models.Post{
			Nickname: user.Nickname,
			Figure:   file.Get("name").String(),
			Data:     imgBytes,
		}

		input := &CallInput{
			Method:      "POST",
			Url:         path,
			Data:        payload,
			CallerID:    user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		type dataModel struct {
			Key string
		}

		output := &Response{Data: &dataModel{}}

		if ok := FetchData(input, output); !ok {
			toast.Text(ERR_CANNOT_REACH_BE).Type(TTYPE_ERR).Dispatch()
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(TTYPE_ERR).Dispatch()
			return
		}

		data, ok := output.Data.(*dataModel)
		if !ok {
			toast.Text(ERR_CANNOT_GET_DATA).Type(TTYPE_ERR).Dispatch()
			return
		}

		user.AvatarURL = "/web/pix/thumb_" + data.Key

		// Update the LocalStorage.
		SaveUser(user, &ctx)

		toast.Text(MSG_AVATAR_CHANGE_SUCCESS).Type(TTYPE_SUCCESS).Dispatch()
	})
}

func HandleLink(ctx app.Context, a app.Action, path, pathAlt string) {
	id, ok := a.Value.(string)
	if !ok {
		return
	}

	url := ctx.Page().URL()
	scheme := url.Scheme
	host := url.Host

	if _, err := strconv.ParseFloat(id, 64); err != nil {
		path = pathAlt
	}

	// Write the link to browsers's clipboard.
	navigator := app.Window().Get("navigator")
	if !navigator.IsNull() {
		clipboard := navigator.Get("clipboard")
		if !clipboard.IsNull() && !clipboard.IsUndefined() {
			clipboard.Call("writeText", scheme+"://"+host+path+id)
		}
	}

	ctx.Navigate(path + id)
}

func HandleMouseEnter(ctx app.Context, a app.Action) {
	id, ok := a.Value.(string)
	if !ok {
		return
	}

	if elem := app.Window().GetElementByID(id); !elem.IsNull() {
		elem.Get("style").Call("setProperty", "font-size", "1.2rem")
	}
}

func HandleMouseLeave(ctx app.Context, a app.Action) {
	id, ok := a.Value.(string)
	if !ok {
		return
	}

	if elem := app.Window().GetElementByID(id); !elem.IsNull() {
		elem.Get("style").Call("setProperty", "font-size", "1rem")
	}
}

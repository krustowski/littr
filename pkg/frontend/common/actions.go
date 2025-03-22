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
		if err := SaveUser(user, &ctx); err != nil {
			toast.Text(ErrLocalStorageUserSave).Type(TTYPE_ERR).Dispatch()
			return
		}

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

// HandleUserShade is an action handler function that enables one to shade other accounts.
func HandleUserShade(ctx app.Context, a app.Action, userToShade models.User, callback func()) {
	// Fetch the requested ID (nickname) and assert it type string.
	key, ok := a.Value.(string)
	if !ok {
		return
	}

	var loggedUser models.User
	ctx.GetState(StateNameUser, &loggedUser)

	// One cannot shade themselves.
	if loggedUser.Nickname == key {
		return
	}

	// Patch the to-be-(un)shaded counterpart user's flowList.
	if userToShade.FlowList == nil {
		userToShade.FlowList = make(map[string]bool)
	}

	if userToShade.RequestList != nil {
		reqList := userToShade.RequestList
		reqList[loggedUser.Nickname] = false
		userToShade.RequestList = reqList
	}

	// Fetch and negate the current shade status.
	shadeListItem := loggedUser.ShadeList[key]

	// Patch the controlling user's flowList/shadeList nil map.
	if loggedUser.FlowList == nil {
		loggedUser.FlowList = make(map[string]bool)
	}
	if loggedUser.ShadeList == nil {
		loggedUser.ShadeList = make(map[string]bool)
	}
	if loggedUser.RequestList != nil {
		reqList := loggedUser.RequestList
		reqList[userToShade.Nickname] = false
		loggedUser.RequestList = reqList
	}

	// Only (un)shade user accounts different from the controlling user's one.
	if key != loggedUser.Nickname {
		loggedUser.ShadeList[key] = !shadeListItem
	}

	// Disable the following of the controlling user in the counterpart user's flowList. And vice versa.
	if loggedUser.ShadeList[key] {
		userToShade.FlowList[loggedUser.Nickname] = false
		loggedUser.FlowList[key] = false
	}

	// Instantiate the toast.
	toast := Toast{AppContext: &ctx}

	ctx.Async(func() {
		defer ctx.Dispatch(func(ctx app.Context) {
			callback()
		})

		// Prepare the request body data structure.
		payload := struct {
			FlowList    map[string]bool `json:"flow_list"`
			RequestList map[string]bool `json:"request_list"`
		}{
			FlowList:    userToShade.FlowList,
			RequestList: userToShade.RequestList,
		}

		// Compose the API input payload.
		input := &CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + userToShade.Nickname + "/lists",
			Data:        payload,
			CallerID:    loggedUser.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		// Prepare the blank API call response object.
		output := &Response{}

		// Patch the counterpart's lists.
		if ok := FetchData(input, output); !ok {
			toast.Text(ERR_CANNOT_REACH_BE).Type(TTYPE_ERR).Dispatch()
			return
		}

		// Check for the HTTP 200/201 response code(s), otherwise print the API response message in the toast.
		if output.Code != 200 && output.Code != 201 {
			toast.Text(output.Message).Type(TTYPE_ERR).Dispatch()
			return
		}

		// Prepare the second list update payload.
		payload2 := struct {
			FlowList    map[string]bool `json:"flow_list"`
			ShadeList   map[string]bool `json:"shade_list"`
			RequestList map[string]bool `json:"request_list"`
		}{
			FlowList:    loggedUser.FlowList,
			ShadeList:   loggedUser.ShadeList,
			RequestList: loggedUser.RequestList,
		}

		// Compsoe the second API call input.
		input2 := &CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + loggedUser.Nickname + "/lists",
			Data:        payload2,
			CallerID:    loggedUser.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		// Prepare the blank API response object.
		output2 := &Response{}

		// Patch the controlling user's lists.
		if ok := FetchData(input2, output2); !ok {
			toast.Text(ERR_CANNOT_REACH_BE).Type(TTYPE_ERR).Dispatch()
			return
		}

		// Check for the HTTP 200 response code, otherwise print the API response message in the toast.
		if output2.Code != 200 {
			toast.Text(output2.Message).Type(TTYPE_ERR).Dispatch()
			return
		}

		// Update the current user's state in the LocalStorage.
		ctx.SetState(StateNameUser, loggedUser).Persist()

		toast.Text(MSG_SHADE_SUCCESSFUL).Type(TTYPE_SUCCESS).Dispatch()
	})
}

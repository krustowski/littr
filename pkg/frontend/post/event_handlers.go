package post

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"strconv"
	"strings"
	"time"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/dsoprea/go-exif/v3"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// onClick is a callback function to post generic input (post, or poll).
func (c *Content) onClick(ctx app.Context, e app.Event) {
	// Prevent the double-posting.
	if c.postButtonsDisabled {
		return
	}

	// Fetch the type: post, fig, or poll.
	postType := ctx.JSSrc().Get("id").String()

	// Null the content (field of the payload, read on).
	content := ""
	poll := models.Poll{}

	// Instantiate the toast.
	toast := common.Toast{AppContext: &ctx}

	var payload interface{}

	// Nasty way on how to disable the buttons. (Use action handler and Dispatch function instead.)
	c.postButtonsDisabled = true

	ctx.Async(func() {
		defer ctx.Dispatch(func(ctx app.Context) {
			c.postButtonsDisabled = false
		})

		var leave bool

		// Determine the post Type.
		switch postType {
		case "poll":
			// Trim the padding spaces on the extremities.
			// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
			pollQuestion := strings.TrimSpace(c.pollQuestion)
			pollOptionI := strings.TrimSpace(c.pollOptionI)
			pollOptionII := strings.TrimSpace(c.pollOptionII)
			pollOptionIII := strings.TrimSpace(c.pollOptionIII)

			if pollOptionI == "" || pollOptionII == "" || pollQuestion == "" {
				toast.Text(common.ERR_POLL_FIELDS_REQUIRED).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
				leave = true
				break
			}

			if pollOptionIII == "" {
				pollOptionIII = strings.TrimSpace(app.Window().GetElementByID("poll-option-iii").Get("value").String())
			}

			// Compose a timestamp and the derived key (content).
			now := time.Now()
			content = strconv.FormatInt(now.UnixNano(), 10)

			// Assign various poll's field inputs to the generic poll.
			poll.ID = content
			poll.Question = pollQuestion
			poll.OptionOne.Content = pollOptionI
			poll.OptionTwo.Content = pollOptionII
			poll.OptionThree.Content = pollOptionIII
			poll.Timestamp = now

		case "post":
			// This is to hotfix the fact that the input can be CTRL-Entered in.
			textarea := app.Window().GetElementByID("post-textarea").Get("value").String()

			// Trim the padding spaces on the extremities.
			// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
			newPost := strings.TrimSpace(textarea)

			// Allow a just picture posting.
			if newPost == "" && c.newFigFile == "" {
				toast.Text(common.ERR_POST_TEXTAREA_EMPTY).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
				leave = true
				break
			}

			// Content is the new post itself.
			content = newPost

		default:
			toast.Text(common.ERR_POST_UNKNOWN_TYPE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			leave = true
			return
		}

		// Fixing the bug: cannot use return and break together in switch according to Sonar.
		if leave {
			return
		}

		var user models.User

		// Decode and unmarshal the local storage user data.
		if err := common.LoadUser(&user, &ctx); err != nil {
			toast.Text(common.ERR_LOCAL_STORAGE_LOAD_FAIL).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		path := "/api/v1/posts"

		// Compose a post payload.
		if postType == "post" {
			payload = models.Post{
				Nickname: user.Nickname,
				Type:     postType,
				Content:  content,
				PollID:   poll.ID,
				Figure:   c.newFigFile,
				Data:     c.newFigData,
			}

			// Compose a poll payload.
		} else if postType == "poll" {
			path = "/api/v1/polls"
			poll.Author = user.Nickname

			payload = struct {
				Question    string `json:"question"`
				OptionOne   string `json:"option_one"`
				OptionTwo   string `json:"option_two"`
				OptionThree string `json:"option_three"`
			}{
				Question:    poll.Question,
				OptionOne:   poll.OptionOne.Content,
				OptionTwo:   poll.OptionTwo.Content,
				OptionThree: poll.OptionThree.Content,
			}
		}

		// Compose the API input payload.
		input := &common.CallInput{
			Method:      "POST",
			Url:         path,
			Data:        payload,
			CallerID:    user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		// Prepare the blank API response object.
		output := &common.Response{}

		// Send new post/poll to backend struct.
		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		// Check for the HTTP 201 response code, otherwise print the API response message in the toast.
		if output.Code != 201 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		// Delete the draft(s) from LocalStorage.
		ctx.LocalStorage().Set("newPostDraft", nil)
		ctx.LocalStorage().Set("newPostFigFile", nil)
		ctx.LocalStorage().Set("newPostFigData", nil)

		// Redirection according to the post type.
		if postType == "poll" {
			ctx.Navigate("/polls")
		} else {
			ctx.Navigate("/flow")
		}

		return
	})
}

// onDismissToast is a callback function that call a new action to handle that request for you.
func (c *Content) onDismissToast(ctx app.Context, e app.Event) {
	// Cast a new valueless action.
	ctx.NewAction("dismiss")
}

// onKeyDown is a callback function that takes care of the keyboard key UI controlling.
func (c *Content) onKeyDown(ctx app.Context, e app.Event) {
	// IS the key an Escape/Esc? Then the dismiss action call.
	if e.Get("key").String() == "Escape" || e.Get("key").String() == "Esc" {
		ctx.NewAction("dismiss")
		return
	}

	// Fetch the post textarea.
	textarea := app.Window().GetElementByID("post-textarea")

	// If null, we null.
	if textarea.IsNull() {
		return
	}

	// Otherwise utilize the CTRL-Enter combination and send the post in.
	if e.Get("ctrlKey").Bool() && e.Get("key").String() == "Enter" && len(textarea.Get("value").String()) != 0 {
		app.Window().GetElementByID("post").Call("click")
	}
}

func (c *Content) onTextareaBlur(ctx app.Context, e app.Event) {
	// Save a new post draft, if the focus on textarea is lost.
	ctx.LocalStorage().Set("newPostDraft", ctx.JSSrc().Get("value").String())
}

// handleFixUpload is a callbackfunction that takes care of the figure/image uploading.
// https://github.com/maxence-charriere/go-app/issues/882
func (c *Content) handleFigUpload(ctx app.Context, e app.Event) {
	// Fetch the first file in the row (index 0).
	file := e.Get("target").Get("files").Index(0)

	//log.Println("name", file.Get("name").String())
	//log.Println("size", file.Get("size").Int())
	//log.Println("type", file.Get("type").String())

	// Nasty way on how to disable buttons.
	c.postButtonsDisabled = true

	defer ctx.Dispatch(func(ctx app.Context) {
		c.postButtonsDisabled = false
	})

	// Instantiate the toast.
	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		var (
			data []byte
			err  error
		)

		// Read the figure/image data.
		data, err = common.ReadFile(file)
		if err != nil {
			toast.Text(err.Error()).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		// Inflate the image to memory.
		img, format, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			toast.Text(err.Error()).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		imgP := &img

		// Fix the image orientation for decoded image.
		imgP, err = FixOrientation(imgP, &data)
		if err != nil {
			toast.Text(err.Error()).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		var buf bytes.Buffer

		// Encode depending on the format.
		switch format {
		case "jpeg":
			err := jpeg.Encode(&buf, *imgP, nil)
			if err != nil {
				toast.Text(err.Error()).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
				return
			}
		case "png":
			err := png.Encode(&buf, *imgP)
			if err != nil {
				toast.Text(err.Error()).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
				return
			}
		}

		// Cast the image ready message.
		toast.Text(common.MSG_IMAGE_READY).Type(common.TTYPE_INFO).Dispatch(c, dispatch)

		// Load the image data to the Content structure.
		ctx.Dispatch(func(ctx app.Context) {
			c.newFigFile = file.Get("name").String()
			c.newFigData = buf.Bytes()

			// Save the figure data in LS as a backup.
			ctx.LocalStorage().Set("newPostFigFile", file.Get("name").String())
			ctx.LocalStorage().Set("newPostFigData", buf.Bytes())
		})
	})
}

// FixOrientation checks the EXIF orientation tag and corrects the image's orientation if necessary
func FixOrientation(img *image.Image, imgBytes *[]byte) (*image.Image, error) {
	rawExif, err := exif.SearchAndExtractExif(*imgBytes)
	if err != nil {
		if err == exif.ErrNoExif {
			return img, nil // If there's no EXIF data, return the original image
		}
		return nil, err
	}

	// Parse the EXIF data
	entries, _, err := exif.GetFlatExifData(rawExif, nil)
	if err != nil {
		return nil, err
	}

	// Find the Orientation tag
	for _, entry := range entries {
		if entry.TagName == "Orientation" {
			//fmt.Printf("orientation: entry.Value: %v\n", entry.Value)
			orientationRaw := entry.Value.([]uint16) // Orientation should be a uint16 value
			orientation := orientationRaw[0]

			//orientation := entry.Formatted
			//fmt.Println("Orientation tag found:", orientation)

			switch orientation {
			case 3: // 180 degrees
				*img = rotate180(img)
			case 6: // 90 degrees clockwise
				*img = rotate90(img)
			case 8: // 90 degrees counterclockwise
				*img = rotate270(img)
			}
		}
	}

	return img, nil
}

// Rotate image 90 degrees clockwise
func rotate90(img *image.Image) image.Image {
	bounds := (*img).Bounds()
	rotated := image.NewRGBA(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))

	for x := 0; x < bounds.Dx(); x++ {
		for y := 0; y < bounds.Dy(); y++ {
			rotated.Set(bounds.Dy()-y-1, x, (*img).At(x, y))
		}
	}

	return rotated
}

// Rotate image 180 degrees
func rotate180(img *image.Image) image.Image {
	bounds := (*img).Bounds()
	rotated := image.NewRGBA(bounds)

	for x := 0; x < bounds.Dx(); x++ {
		for y := 0; y < bounds.Dy(); y++ {
			rotated.Set(bounds.Dx()-x-1, bounds.Dy()-y-1, (*img).At(x, y))
		}
	}

	return rotated
}

// Rotate image 270 degrees (90 degrees counter-clockwise)
func rotate270(img *image.Image) image.Image {
	bounds := (*img).Bounds()
	rotated := image.NewRGBA(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))

	for x := 0; x < bounds.Dx(); x++ {
		for y := 0; y < bounds.Dy(); y++ {
			rotated.Set(y, bounds.Dx()-x-1, (*img).At(x, y))
		}
	}

	return rotated
}

package molecules

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"go.vxn.dev/littr/pkg/frontend/common"
)

type ImageInput struct {
	app.Compo

	ImageData []byte
	ImageFile string
	ImageLink string

	ButtonsDisabled bool
}

// https://github.com/maxence-charriere/go-app/issues/882
func (i *ImageInput) onImageInput(ctx app.Context, e app.Event) {
	file := e.Get("target").Get("files").Index(0)

	//log.Println("name", file.Get("name").String())
	//log.Println("size", file.Get("size").Int())
	//log.Println("type", file.Get("type").String())

	i.ButtonsDisabled = true

	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		defer ctx.Dispatch(func(ctx app.Context) {
			i.ButtonsDisabled = false
		})

		var (
			data         []byte
			err          error
			processedImg *[]byte
		)

		// Read the figure/image data.
		data, err = common.ReadFile(file)
		if err != nil {
			toast.Text(err.Error()).Type(common.TTYPE_ERR).Dispatch(i)
			return
		}

		processedImg, err = common.ProcessImage(&data)
		if err != nil {
			toast.Text(err.Error()).Type(common.TTYPE_ERR).Dispatch(i)
			return
		}

		// Load the image data to the Content structure.
		ctx.Dispatch(func(ctx app.Context) {
			i.ImageFile = file.Get("name").String()
			i.ImageData = *processedImg

			// Save the figure data in LS as a backup.
			ctx.LocalStorage().Set("newPostFigFile", file.Get("name").String())
			ctx.LocalStorage().Set("newPostFigData", *processedImg)
		})

		// Cast the image ready message.
		toast.Text(common.MSG_IMAGE_READY).Type(common.TTYPE_INFO).Dispatch(i)
	})
}

/*func (i *ImageInput) onImageInput(ctx app.Context, e app.Event) {
	ctx.NewActionWithValue(i.OnImageUploadActionName, e.Get("id").String())
}*/

func (i *ImageInput) Render() app.UI {
	return app.Div().Class("field label border extra deep-orange-text thicc").Body(
		app.Input().ID("fig-upload").Class("active").Type("file").OnChange(i.ValueTo(&i.ImageLink)).OnInput(i.onImageInput).Accept("image/*"),
		app.Input().Class("active").Type("text").Value(i.ImageFile).Disabled(true),
		app.Label().Text("Image").Class("active deep-orange-text"),
		app.I().Text("image"),
	)

}

package atoms

import (
	"fmt"
	"strconv"

	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/models"
)

type PollResult struct {
	app.Compo

	OptionShare int64
	Option      models.PollOption

	OrangelLevel int
}

func (p *PollResult) composeClass() string {
	if p.OrangelLevel == 0 {
		p.OrangelLevel = 3
	}

	return fmt.Sprintf("bold progress left blue%d small-padding thicc", p.OrangelLevel)
}

func (p *PollResult) Render() app.UI {
	return app.Div().Class("thicc").Body(
		app.Div().Class(p.composeClass()).Style("clip-path", "polygon(0% 0%, 0% 98%, "+strconv.FormatInt(p.OptionShare, 10)+"% 98%, "+strconv.FormatInt(p.OptionShare, 10)+"% 0%);"),

		//app.Progress().Value(strconv.Itoa(optionOneShare)).Max(100).Class("deep-orange-text padding medium bold left"),
		//app.Div().Class("progress left light-green"),

		app.Div().Class("bold").Body(
			app.Span().Text(p.Option.Content+" ("+strconv.FormatInt(p.OptionShare, 10)+"%)"),
		),
	)
}

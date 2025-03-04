package organisms

import (
	"slices"
	"time"

	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/molecules"
	"go.vxn.dev/littr/pkg/models"
)

type PollFeed struct {
	app.Compo

	LoggedUser models.User

	SortedPolls []models.Poll

	Polls map[string]models.Poll

	Pagination int
	PageNo     int

	ButtonsDisabled bool
	LoaderShowImage bool

	OnClickOptionOneActionName   string
	OnClickOptionTwoActionName   string
	OnClickOptionThreeActionName string

	OnClickDeleteModalShowActionName string
	OnClickLinkActionName            string
	OnMouseEnterActionName           string
	OnMouseLeaveActionName           string

	pollTimestamp    string
	userVoted        bool
	optionOneShare   int64
	optionTwoShare   int64
	optionThreeShare int64
}

func (p *PollFeed) clearProps() {
	p.pollTimestamp = ""
	p.userVoted = false
	p.optionOneShare = 0
	p.optionTwoShare = 0
	p.optionThreeShare = 0
}

func (p *PollFeed) processPoll(poll models.Poll) bool {
	p.clearProps()

	p.userVoted = slices.Contains(poll.Voted, p.LoggedUser.Nickname)

	var pollCounterSum int64

	pollCounterSum = poll.OptionOne.Counter + poll.OptionTwo.Counter
	if poll.OptionThree.Content != "" {
		pollCounterSum += poll.OptionThree.Counter
	}

	// At least one vote has to be already recorded to show the progresses.
	if pollCounterSum > 0 {
		p.optionOneShare = poll.OptionOne.Counter * 100 / pollCounterSum
		p.optionTwoShare = poll.OptionTwo.Counter * 100 / pollCounterSum
		p.optionThreeShare = poll.OptionThree.Counter * 100 / pollCounterSum
	}

	// Use JS toLocaleString() function to reformat the timestamp
	if !p.LoggedUser.LocalTimeMode {
		pollLocale := app.Window().
			Get("Date").
			New(poll.Timestamp.Format(time.RFC3339))

		p.pollTimestamp = pollLocale.Call("toLocaleString", "en-GB").String()
	} else {
		p.pollTimestamp = poll.Timestamp.Format("Jan 02, 2006 / 15:04:05")
	}

	return true
}

func (p *PollFeed) Render() app.UI {
	return app.Div().Class("post-feed").Body(
		app.Range(p.SortedPolls).Slice(func(idx int) app.UI {
			poll := p.SortedPolls[idx]

			if !p.processPoll(poll) {
				return nil
			}

			return app.Div().Class("post").Body(
				&molecules.PollHeader{
					Poll:                  poll,
					ButtonsDisabled:       p.ButtonsDisabled,
					OnClickLinkActionName: p.OnClickLinkActionName,
				},

				&molecules.PollBody{
					Poll:       poll,
					LoggedUser: p.LoggedUser,
					RenderProps: struct {
						PollTimestamp    string
						UserVoted        bool
						OptionOneShare   int64
						OptionTwoShare   int64
						OptionThreeShare int64
					}{
						PollTimestamp:    p.pollTimestamp,
						UserVoted:        p.userVoted,
						OptionOneShare:   p.optionOneShare,
						OptionTwoShare:   p.optionTwoShare,
						OptionThreeShare: p.optionThreeShare,
					},

					OnClickOptionOneActionName:   p.OnClickOptionOneActionName,
					OnClickOptionTwoActionName:   p.OnClickOptionTwoActionName,
					OnClickOptionThreeActionName: p.OnClickOptionThreeActionName,

					ButtonDisabled:  p.ButtonsDisabled,
					LoaderShowImage: p.LoaderShowImage,
				},

				&molecules.PollFooter{
					Poll:                    poll,
					LoggedUserNickname:      p.LoggedUser.Nickname,
					PollTimestamp:           p.pollTimestamp,
					ButtonsDisabled:         p.ButtonsDisabled,
					OnClickDeleteActionName: p.OnClickDeleteModalShowActionName,
				},
			)
		}),
	)
}

package polls

import (
	"sort"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/frontend/atomic/organisms"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func (c *Content) sortPolls() []models.Poll {
	var sortedPolls []models.Poll

	for _, sortedPoll := range c.polls {
		sortedPolls = append(sortedPolls, sortedPoll)
	}

	// order polls by timestamp DESC
	sort.SliceStable(sortedPolls, func(i, j int) bool {
		return sortedPolls[i].Timestamp.After(sortedPolls[j].Timestamp)
	})

	// prepare polls according to the actual pagination and pageNo
	pagedPolls := []models.Poll{}

	end := len(sortedPolls)
	start := 0

	stop := func(c *Content) int {
		var pos int

		if c.pagination > 0 {
			// (c.pageNo - 1) * c.pagination + c.pagination
			pos = c.pageNo * c.pagination
		}

		if pos > end {
			// kill the eventListener (observers scrolling)
			//c.scrollEventListener()
			c.paginationEnd = true

			return (end)
		}

		if pos < 0 {
			return 0
		}

		return pos
	}(c)

	if end > 0 && stop > 0 {
		pagedPolls = sortedPolls[start:stop]
	}

	return pagedPolls
}

func (c *Content) Render() app.UI {
	return app.Main().Class("responsive").Body(
		&atoms.PageHeading{
			Title: "polls",
		},

		// Poll deletion modal.
		&organisms.ModalPollDelete{
			PollID:                   c.polls[c.interactedPollKey].ID,
			ModalShow:                c.deletePollModalShow,
			ModalButtonsDisabled:     c.deleteModalButtonsDisabled,
			OnClickDismissActionName: "dismiss",
			OnClickDeleteActionName:  "delete",
		},

		// The very polls feed.
		&organisms.PollFeed{
			LoggedUser: c.user,

			SortedPolls: c.sortPolls(),

			Pagination: c.pagination,
			PageNo:     c.pageNo,

			ButtonsDisabled: c.pollsButtonDisabled,
			LoaderShowImage: c.loaderShow,

			OnClickOptionOneActionName:   "option-one-click",
			OnClickOptionTwoActionName:   "option-two-click",
			OnClickOptionThreeActionName: "option-three-click",

			OnClickDeleteModalShowActionName: "delete-click",
			OnClickLinkActionName:            "link",
			OnMouseEnterActionName:           "mouse-enter",
			OnMouseLeaveActionName:           "mouse-leave",
		},

		&atoms.Loader{
			ID:         "page-end-anchor",
			ShowLoader: c.loaderShow,
		},
	)
}

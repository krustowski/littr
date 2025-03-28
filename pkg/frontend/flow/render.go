package flow

import (
	"sort"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/frontend/atomic/molecules"
	"go.vxn.dev/littr/pkg/frontend/atomic/organisms"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func (c *Content) sortPosts() []models.Post {
	var sortedPosts []models.Post

	posts := c.posts
	if posts == nil {
		posts = make(map[string]models.Post)
	}

	flowList := c.user.FlowList
	if len(flowList) == 0 {
		return sortedPosts
	}

	// fetch posts and put them in an array
	for _, sortedPost := range posts {
		// do not append a post that is not meant to be shown
		if !c.user.FlowList[sortedPost.Nickname] && sortedPost.Nickname != "system" && sortedPost.Nickname != c.userFlowNick {
			continue
		}

		sortedPosts = append(sortedPosts, sortedPost)
	}

	// order posts by timestamp DESC
	sort.SliceStable(sortedPosts, func(i, j int) bool {
		if c.singlePostID != "" {
			return sortedPosts[i].Timestamp.Before(sortedPosts[j].Timestamp)
		}

		return sortedPosts[i].Timestamp.After(sortedPosts[j].Timestamp)
	})

	return sortedPosts
}

func (c *Content) Render() app.UI {
	return app.Main().Class("responsive").Body(
		// Page heading
		&molecules.FlowHeader{
			SingleUser:      c.users[c.userFlowNick],
			SinglePostID:    c.singlePostID,
			Hashtag:         c.hashtag,
			ButtonsDisabled: c.buttonDisabled,
			RefreshClicked:  c.refreshClicked,
		},

		// SingleUser view (profile mode)
		&organisms.SingleUserProfile{
			LoggedUser:      c.user,
			SingleUser:      c.users[c.userFlowNick],
			ButtonsDisabled: c.buttonDisabled,
			//
			OnClickAskActionName:      "ask",
			OnClickShadeActionName:    "shade",
			OnClickCancelActionName:   "cancel",
			OnClickFollowActionName:   "follow",
			OnClickUnfollowActionName: "unfollow",
		},
		app.Div().Class("space"),

		// Post deletion modal.
		&organisms.ModalPostDelete{
			PostID:                   c.posts[c.interactedPostKey].ID,
			ModalShow:                c.deletePostModalShow,
			ModalButtonsDisabled:     c.deleteModalButtonsDisabled,
			OnClickDismissActionName: "dismiss",
			OnClickDeleteActionName:  "delete",
		},

		// Post reply modal.
		&organisms.ModalPostReply{
			ReplyPostContent:         &c.replyPostContent,
			ImageData:                &c.newFigData,
			ImageLink:                &c.newFigLink,
			ImageFile:                &c.newFigFile,
			PostOriginal:             c.posts[c.interactedPostKey],
			ModalShow:                c.modalReplyActive,
			ModalButtonsDisabled:     &c.postButtonsDisabled,
			OnClickDismissActionName: "dismiss",
			OnClickReplyActionName:   "reply",
			OnFigureUploadActionName: "image-upload",
			OnBlurActionName:         "blur-post",
		},

		// The very post feed.
		&organisms.PostFeed{
			Pagination:      c.pagination,
			PageNo:          c.pageNo,
			HideReplies:     c.hideReplies,
			SinglePostID:    c.singlePostID,
			SingleUserID:    c.userFlowNick,
			Posts:           c.posts,
			Users:           c.users,
			LoaderShowImage: c.loaderShowImage,
			ButtonsDisabled: c.buttonDisabled,
			LoggedUser:      c.user,
			SortedPosts:     c.sortPosts(),
			//
			OnClickImageActionName:   "image-click",
			OnClickStarActionName:    "star",
			OnClickReplyActionName:   "modal-post-reply",
			OnClickLinkActionName:    "link",
			OnClickHistoryActionName: "history",
			OnClickDeleteActionName:  "modal-post-delete",
			OnClickUserActionName:    "user",
			OnMouseEnterActionName:   "mouse-enter",
			OnMouseLeaveActionName:   "mouse-leave",
		},

		&atoms.Loader{
			ID:         "page-end-anchor",
			ShowLoader: c.loaderShow,
		},
	)
}

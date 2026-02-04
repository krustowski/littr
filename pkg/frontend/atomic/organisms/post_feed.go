package organisms

import (
	"net/url"
	"strings"
	"time"

	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/config"
	"go.vxn.dev/littr/pkg/frontend/atomic/molecules"
	"go.vxn.dev/littr/pkg/models"
)

type PostFeed struct {
	app.Compo

	LoggedUser models.User

	SortedPosts []models.Post

	Posts map[string]models.Post
	Users map[string]models.User

	Pagination int
	PageNo     int

	HideReplies bool

	ButtonsDisabled bool
	LoaderShowImage bool

	SinglePostID string
	SingleUserID string

	OnClickDeleteActionName  string
	OnClickHistoryActionName string
	OnClickImageActionName   string
	OnClickLinkActionName    string
	OnClickReplyActionName   string
	OnClickStarActionName    string
	OnClickUserActionName    string
	OnMouseEnterActionName   string
	OnMouseLeaveActionName   string

	imageSource     string
	postSummary     string
	originalContent string
	originalSummary string
	postTimestamp   string
	systemLink      string
	postClass       string
}

func (p *PostFeed) clearProps() {
	p.imageSource = ""
	p.postSummary = ""
	p.originalContent = ""
	p.originalSummary = ""
	p.postTimestamp = ""
	p.systemLink = ""
	p.postClass = "post"
}

func (p *PostFeed) processPost(post models.Post) bool {
	p.clearProps()

	var counter int

	counter++
	if counter > p.Pagination*p.PageNo {
		return false
	}

	// Original post that is replied to.
	if post.ReplyToID != "" && !p.HideReplies {
		if originalPost, found := p.Posts[post.ReplyToID]; found {
			if flowListValue, foundUser := p.LoggedUser.FlowList[originalPost.Nickname]; (!flowListValue || !foundUser) && (p.Users[originalPost.Nickname].Private || p.Users[originalPost.Nickname].Options["private"]) {
				p.originalContent = "this content is private"
			} else if value, found := p.LoggedUser.ShadeList[originalPost.Nickname]; found && value {
				p.originalContent = "the content is shaded"
			} else if value, found := p.Users[originalPost.Nickname].ShadeList[p.LoggedUser.Nickname]; found && value {
				p.originalContent = "the content is shaded"
			} else {
				p.originalContent = originalPost.Nickname + " posted: " + originalPost.Content
			}
		} else {
			p.originalContent = "the post was deleted bye"
		}
	}

	// Filter out non single-post items.
	if p.SinglePostID != "" {
		if post.ID != p.SinglePostID && p.SinglePostID != post.ReplyToID {
			return false
		}

		if post.ID == p.SinglePostID {
			p.postClass = "post original"
		}
	}

	// Remove the border at the last post to render
	// Note: this one works only for len <= PageSize
	if post.ID == p.SortedPosts[len(p.SortedPosts)-1].ID {
		p.postClass += " last"
	}

	// Filter out non single-user items.
	if p.SingleUserID != "" {
		if post.Nickname != p.SingleUserID {
			return false
		}
	}

	// Show posts of users in one's flowList only.
	/*if !p.LoggedUser.FlowList[post.Nickname] && post.Nickname != "system" {
		return false
	}*/

	// Check the post's length, on threshold use <details> tag.
	if len(post.Content) > config.MaxPostLength {
		p.postSummary = post.Content[:config.MaxPostLength/10] + "- [...]"
	}

	// The same as above with the previous post's length for reply render.
	if len(p.originalContent) > config.MaxPostLength {
		p.originalSummary = p.originalContent[:config.MaxPostLength/10] + "- [...]"
	}

	// Compose the image source string.
	switch post.Type {
	case "fig":
		// Check the URL/URI format.
		if _, err := url.ParseRequestURI(post.Content); err == nil {
			p.imageSource = post.Content
		}
	case "post":
		if _, err := url.ParseRequestURI(post.Figure); err == nil {
			p.imageSource = post.Figure
		}
	}

	if p.imageSource == "" {
		fileExplode := strings.Split(post.Figure, ".")
		extension := fileExplode[len(fileExplode)-1]

		p.imageSource = "/web/pix/thumb_" + post.Figure

		if extension == "gif" {
			p.imageSource = "/web/click-to-see.gif"
		}
	}

	// Use JS toLocaleString() function to reformat the timestamp, use negated LocalTimeMode boolean as true!
	if !p.LoggedUser.LocalTimeMode || !p.LoggedUser.Options["localTimeMode"] {
		postLocale := app.Window().
			Get("Date").
			New(post.Timestamp.Format(time.RFC3339))

		p.postTimestamp = postLocale.Call("toLocaleString", "en-GB").String()
	} else {
		p.postTimestamp = post.Timestamp.Format("Jan 02, 2006 / 15:04:05")
	}

	// Omit older system messages for new users.
	if post.Nickname == "system" && post.Timestamp.Before(p.LoggedUser.RegisteredTime) {
		return false
	}

	// Link on system messages.
	p.systemLink = func() string {
		// A system post about a new poll.
		if post.PollID != "" {
			return "/polls/" + post.PollID
		}

		// A system post about a new user.
		if post.Nickname == "system" && post.Type == "user" {
			return "/flow/users/" + post.Figure
		}

		// A system post about a new poll (legacy).
		return "/polls"
	}()

	return true
}

func (p *PostFeed) Render() app.UI {
	return app.Div().Class("post-feed").Body(
		app.Range(p.SortedPosts).Slice(func(idx int) app.UI {
			post := p.SortedPosts[idx]

			if !p.processPost(post) {
				return nil
			}

			return app.Div().Class(p.postClass).Body(
				&molecules.PostHeader{
					PostAuthor:      post.Nickname,
					PostAvatarURL:   p.Users[post.Nickname].AvatarURL,
					PostID:          post.ID,
					ButtonsDisabled: p.ButtonsDisabled,
					//
					OnClickLinkActionName:  p.OnClickLinkActionName,
					OnClickUserActionName:  p.OnClickUserActionName,
					OnMouseEnterActionName: p.OnMouseEnterActionName,
					OnMouseLeaveActionName: p.OnMouseLeaveActionName,
				},

				&molecules.PostBody{
					Post: post,
					RenderProps: struct {
						ImageSource     string
						PostSummary     string
						OriginalContent string
						OriginalSummary string
						PostTimestamp   string
						SystemLink      string
					}{
						ImageSource:     p.imageSource,
						PostSummary:     p.postSummary,
						OriginalContent: p.originalContent,
						OriginalSummary: p.originalSummary,
						PostTimestamp:   p.postTimestamp,
						SystemLink:      p.systemLink,
					},
					ButtonDisabled:  p.ButtonsDisabled,
					LoaderShowImage: p.LoaderShowImage,
					//
					OnClickImageActionName:   p.OnClickImageActionName,
					OnClickHistoryActionName: p.OnClickHistoryActionName,
				},

				&molecules.PostFooter{
					Post:               post,
					PostTimestamp:      p.postTimestamp,
					ButtonsDisabled:    p.ButtonsDisabled,
					LoggedUserNickname: p.LoggedUser.Nickname,
					//
					OnClickDeleteActionName: p.OnClickDeleteActionName,
					OnClickStarActionName:   p.OnClickStarActionName,
					OnClickReplyActionName:  p.OnClickReplyActionName,
				},
			)
		}),
	)
}

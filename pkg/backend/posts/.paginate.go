package posts

import (
	"sort"
	"strings"

	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"
)

/*
 *
 *
 *  LEGACY CODE
 *
 * 
 */

const PageSize int = 25

type PageOptions struct {
	CallerID string          `json:"caller_id"`
	PageNo   int             `json:"page_no"`
	FlowList map[string]bool `json:"folow_list"`

	// flow subroutes booleans and vars
	SinglePost   bool   `json:"single_post" default:false`
	UserFlow     bool   `json:"user_flow" default:false`
	SinglePostID string `json:"single_post_id"`
	UserFlowNick string `json:"user_Flow_nick"`

	Hashtag     string `json:"hashtag" default:""`
	HideReplies bool   `json:"hide_replies" default:"false"`
}

// for now, let us use it for posts/flow exclusively only
func GetOnePage(opts PageOptions) (map[string]models.Post, map[string]models.User) {
	user, ok := db.GetOne(db.UserCache, opts.CallerID, models.User{})
	if !ok {
		return nil, nil
	}

	// fetch the flow + users and combine them into one response
	// those variables are both of type map[string]T
	allPosts, _ := db.GetAll(db.FlowCache, models.Post{})
	allUsers, _ := db.GetAll(db.UserCache, models.User{})

	// pagination draft
	// + only select N latest posts for such user according to their FlowList
	// + include previous posts to a reply
	// + only include users mentioned

	posts := []models.Post{}
	num := 0

	// extract requested post
	/*if opts.SinglePost && opts.SinglePostID != "" {
		if single, found := posts[opts.SinglePostID]; found {
			posts = append(posts, single)
		}
	}*/

	// overload flowList
	flowList := user.FlowList
	if opts.FlowList != nil {
		flowList = opts.FlowList
	}

	// assign reply count to each post
	for _, post := range allPosts {
		if post.ReplyToID == "" {
			continue
		}

		// calculate the reply count for each post
		origo, found := allPosts[post.ReplyToID]
		if found {
			origo.ReplyCount++
			allPosts[origo.ID] = origo
		}
	}

	// filter out all posts for such callerID
	for _, post := range allPosts {
		// check the caller's flow list, skip on unfollowed, or unknown user
		if value, found := flowList[post.Nickname]; !found || !value {
			continue
		}

		if opts.Hashtag != "" {
			if strings.Contains(post.Content, "#"+opts.Hashtag) {
				posts = append(posts, post)
			}
			continue
		}

		// filter replies out
		if opts.HideReplies && post.ReplyToID != "" {
			continue
		}

		// exctract replies to the single post
		if opts.SinglePost && opts.SinglePostID != "" {
			if post.ReplyToID == opts.SinglePostID || post.ID == opts.SinglePostID {
				posts = append(posts, post)
			}
			continue
		}

		if opts.UserFlow && opts.UserFlowNick != "" {
			if value, found := user.FlowList[opts.UserFlowNick]; (!value || !found) && allUsers[opts.UserFlowNick].Private {
				continue
			}

			if post.Nickname == opts.UserFlowNick {
				posts = append(posts, post)
			}
			continue
		}

		posts = append(posts, post)
	}

	// order posts by timestamp DESC
	sort.SliceStable(posts, func(i, j int) bool {
		return posts[i].Timestamp.After(posts[j].Timestamp)
	})

	// cut the <pageSize>*2 number of posts only
	var part []models.Post

	pageNo := opts.PageNo
	start := (PageSize * 2) * pageNo
	end := (PageSize * 2) * (pageNo + 1)

	if len(posts) > start {
		// only valid for the very first page
		//part = posts[0:(pageSize * 2)]

		if len(posts) <= end {
			// the very last page
			part = posts[start:]
		} else {
			// the middle page
			part = posts[start:end]
		}
	} else {
		// the very single page
		part = posts
	}

	// loop through the array and manually include other posts too
	// watch for users as well
	pExport := make(map[string]models.Post)
	uExport := make(map[string]models.User)

	num = 0
	for _, post := range part {
		// increase the count of posts
		num++

		// export one (1) post
		pExport[post.ID] = post
		uExport[post.Nickname] = allUsers[post.Nickname]

		// we can have multiple keys from a single post -> its interractions
		repKey := post.ReplyToID
		if repKey != "" {
			if prePost, found := allPosts[repKey]; found {
				num++

				// export previous user too
				nick := prePost.Nickname
				uExport[nick] = allUsers[nick]

				// mange private content
				if value, found := user.FlowList[nick]; (!value || !found) && allUsers[nick].Private {
					prePost.Content = ""
				}

				// increase the reply count
				prePost.ReplyCount++
				pExport[repKey] = prePost
			}
		}

		// this makes sure only N posts are returned, but it cuts off the tail posts
		/*if num > pageSize {
			break
		}*/
	}

	// ensure the UserFlowNick is always included too
	if _, found := uExport[opts.UserFlowNick]; !found {
		if _, found = allUsers[opts.UserFlowNick]; found {
			uExport[opts.UserFlowNick] = allUsers[opts.UserFlowNick]
		}
	}

	return pExport, uExport
}

func convertMapToArray[T any](m map[string]T, reverseOutput bool) ([]string, []T) {
	var keys = []string{}
	var vals = []T{}

	for key, val := range m {
		keys = append(keys, key)
		vals = append(vals, val)
	}

	if reverseOutput {
		helpers.Reverse(keys)
		//reverse(vals)
	}

	return keys, vals
}

// universal cache pagination tooling
package pages

import (
	"sort"
	"strings"

	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"
)

const PAGE_SIZE int = 25

type PageOptions struct {
	CallerID string          `json:"caller_id"`
	PageNo   int             `json:"page_no"`
	FlowList map[string]bool `json:"folow_list"`

	Flow  flowOptions `json:"flow_args"`
	Polls pollOptions `json:"poll_args"`
	Users userOptions `json:"user_args"`
}

type flowOptions struct {
	// flow subroutes' booleans and vars
	SinglePost   bool   `json:"single_post"`
	UserFlow     bool   `json:"user_flow"`
	SinglePostID string `json:"single_post_id"`
	UserFlowNick string `json:"user_Flow_nick"`
	Hashtag      string `json:"hashtag" default:""`
	HideReplies  bool   `json:"hide_replies"`
}

type pollOptions struct {
	// polls
	SinglePoll   bool   `json:"single_poll"`
	SinglePollID string `json:"single_poll_id"`
}

type userOptions struct {
	// users
	SingleUser   bool   `json:"single_user"`
	SingleUserID string `json:"single_user_id"`
}

// DTO for fillMaps output aggregation
type maps struct {
	Polls map[string]models.Poll
	Posts map[string]models.Post
	Users map[string]models.User
}

// DTO for GetOnePage pointer output aggregation
type PagePointers struct {
	PtrPolls *map[string]models.Poll
	PtrPosts *map[string]models.Post
	PtrUsers *map[string]models.User
}

// fillMaps is a function, that prepares raw maps of all (related) items for further processing according to input options
func fillMaps(opts PageOptions) *maps {
	if opts.Flow != nil {
		posts, _ := db.GetAll(db.FlowCache, models.Post{})
		users, _ := db.GetAll(db.UserCache, models.User{})

		return &maps{Posts: posts, Users: users}
	}

	if opts.Polls != nil {
		polls, _ := db.GetAll(db.PollCache, models.Poll{})
		// users are not needed necessarily there for now...
		//users, _ := db.GetAll(db.UserCache, models.User{})

		return &maps{Polls: polls}
	}

	if opts.Polls != nil {
		users, _ := db.GetAll(db.UserCache, models.User{})

		return &maps{Users: users}
	}

	return nil
}

func GetOnePage(opts PageOptions) PagePointers { 
	// validate the callerID is a legitimate user's ID
	caller, ok := db.GetOne(db.UserCache, opts.CallerID, models.User{})
	if !ok {
		// unregistred caller
		return nil
	}

	// pointer to maps of all items (based on and related to the opts input)
	ptrMaps := fillMaps(opts)
	if ptrMaps == nil {
		// invalid input options = resulted in empty maps only
		return ptrMaps
	}

	if opts.Flow != nil {
		return onePageFlow(opts, ptrMaps)
	}

	if opts.Polls != nil {
		return onePagePolls(opts, ptrMaps)
	}

	if opts.Users != nil {
		return onePageUsers(opts, ptrMaps)
	}

	return nil
}

	// BTW those variables are both of type map[string]T
	var allPolls map[string]models.Poll
	var allPosts map[string]models.Post
	var allUsers map[string]models.User


	// pagination draft
	// + only select N latest posts for such user according to their FlowList
	// + include previous posts to a reply
	// + only include users mentioned

	posts := []models.Post{}
	num := 0

	// overload flowList
	flowList := caller.FlowList
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
	start := (PAGE_SIZE * 2) * pageNo
	end := (PAGE_SIZE * 2) * (pageNo + 1)

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

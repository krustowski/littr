package pages

import (
	"sort"
	"strings"

	"go.vxn.dev/littr/pkg/models"
)

func onePagePosts(opts *PageOptions, data []interface{}) *PagePointers {
	// BTW those variables are both of type map[string]T
	var (
		allPosts *map[string]models.Post
		allUsers *map[string]models.User
		posts    []models.Post
	)

	defer func() {
		allPosts = nil
		allUsers = nil
		posts = []models.Post{}
	}()

	for _, iface := range data {
		p, ok := iface.(*map[string]models.Post)
		if ok {
			allPosts = p
			continue
		}

		u, ok := iface.(*map[string]models.User)
		if ok {
			allUsers = u
			continue
		}
	}

	// pagination draft
	// + only select N latest posts for such user according to their FlowList
	// + include previous posts to a reply
	// + only include users mentioned

	num := 0

	// overload flowList
	flowList := opts.Caller.FlowList
	if opts.FlowList != nil {
		flowList = opts.FlowList
	}

	// assign reply count to each post
	for _, post := range *allPosts {
		if post.ReplyToID == "" {
			continue
		}

		// calculate the reply count for each post
		origo, found := (*allPosts)[post.ReplyToID]
		if found {
			origo.ReplyCount++
			(*allPosts)[origo.ID] = origo
		}
	}

	// filter out all posts for such callerID
	for _, post := range *allPosts {
		// check the caller's flow list, skip on unfollowed, or unknown user
		if value, found := flowList[post.Nickname]; !found || !value {
			continue
		}

		if opts.Flow.Hashtag != "" {
			if strings.Contains(post.Content, "#"+opts.Flow.Hashtag) {
				posts = append(posts, post)
			}
			continue
		}

		// filter replies out
		if opts.Flow.HideReplies && post.ReplyToID != "" {
			continue
		}

		// exctract replies to the single post
		if opts.Flow.SinglePost && opts.Flow.SinglePostID != "" {
			if post.ReplyToID == opts.Flow.SinglePostID || post.ID == opts.Flow.SinglePostID {
				posts = append(posts, post)
			}
			continue
		}

		if opts.Flow.UserFlow && opts.Flow.UserFlowNick != "" {
			if value, found := opts.Caller.FlowList[opts.Flow.UserFlowNick]; (!value || !found) && (*allUsers)[opts.Flow.UserFlowNick].Private {
				continue
			}

			if post.Nickname == opts.Flow.UserFlowNick {
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

	// cut the PAGE_SIZE*2 number of posts only
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
		uExport[post.Nickname] = (*allUsers)[post.Nickname]

		// we can have multiple keys from a single post -> its interractions
		repKey := post.ReplyToID
		if repKey != "" {
			if prePost, found := (*allPosts)[repKey]; found {
				num++

				// export previous user too
				nick := prePost.Nickname
				uExport[nick] = (*allUsers)[nick]

				// mange private content
				if value, found := opts.Caller.FlowList[nick]; (!value || !found) && (*allUsers)[nick].Private {
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
	if _, found := uExport[opts.Flow.UserFlowNick]; !found {
		if _, found = (*allUsers)[opts.Flow.UserFlowNick]; found {
			uExport[opts.Flow.UserFlowNick] = (*allUsers)[opts.Flow.UserFlowNick]
		}
	}

	return &PagePointers{Posts: &pExport, Users: &uExport}
}

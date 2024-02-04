package backend

import (
	"sort"

	"go.savla.dev/littr/models"
)

const pageSize int = 25

// for now, let us use it for posts/flow exclusively only
func getOnePage(pageNo int, callerID string) (map[string]models.Post, map[string]models.User) {
	user, ok := getOne(UserCache, callerID, models.User{})
	if !ok {
		return nil, nil
	}

	// fetch the flow + users and combine them into one response
	// those variables are both of type map[string]T
	allPosts, _ := getAll(FlowCache, models.Post{})
	allUsers, _ := getAll(UserCache, models.User{})

	// pagination draft
	// + only select N latest posts for such user according to their FlowList
	// + include previous posts to a reply
	// + only include users mentioned

	posts := []models.Post{}
	num := 0

	// filter out all posts for such callerID
	for _, post := range allPosts {
		// check the caller's flow list, skip on unfollowed, or unknown user
		if value, found := user.FlowList[post.Nickname]; !found || !value {
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

	start := (pageSize * 2) * pageNo
	end := (pageSize * 2) * (pageNo + 1)

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
			num++
			pExport[repKey] = allPosts[repKey]

			// export previous user too
			nick := allPosts[repKey].Nickname
			uExport[nick] = allUsers[nick]
		}

		// this makes sure only N posts are returned, but it cuts off the tail posts
		/*if num > pageSize {
			break
		}*/
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
		reverse(keys)
		//reverse(vals)
	}

	return keys, vals
}

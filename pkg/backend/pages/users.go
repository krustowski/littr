package pages

import (
	"sort"

	"go.vxn.dev/littr/pkg/models"
)

func onePageUsers(opts PageOptions, ptrMaps *PagePointers) PagePointers {
	var allUsers *map[string]models.User = ptrMaps.Users

	users := []models.User{}
	caller := models.User{}

	for key, user := range *allUsers {
		// check and correct the corresponding item's key
		if key != user.Nickname {
			user.Nickname = key
		}

		if key == opts.CallerID {
			caller = user
		}

		users = append(users, user)
	}

	// sort by name
	sort.Slice(users, func(i, j int) bool {
		return users[i].Nickname < users[j].Nickname
	})

	// cut the PAGE_SIZE number of posts only
	var part []models.User

	pageNo := opts.PageNo
	start := (PAGE_SIZE) * pageNo
	end := (PAGE_SIZE) * (pageNo + 1)

	if len(users) > start {
		// only valid for the very first page
		//part = posts[0:(pageSize * 2)]

		if len(users) <= end {
			// the very last page
			part = users[start:]
		} else {
			// the middle page
			part = users[start:end]
		}
	} else {
		// the very single page
		//part = users[len(users)-PAGE_SIZE-1 : len(users)-1]
		if len(users) > PAGE_SIZE*(pageNo-1) {
			part = users[PAGE_SIZE*(pageNo-1):]
		}
	}

	if opts.Users.RequestList == nil {
		return PagePointers{}
	}

	// Add all (for now) users from the requestList to render properly at the top of the users page.
	for nick, requested := range *opts.Users.RequestList {
		if requested {
			part = append(part, (*allUsers)[nick])
		}
	}

	uExport := make(map[string]models.User)

	for _, user := range part {
		uExport[user.Nickname] = user
	}

	uExport[opts.CallerID] = caller

	return PagePointers{Users: &uExport}
}

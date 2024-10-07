package pages

import (
	"sort"

	"go.vxn.dev/littr/pkg/models"
)

func onePageUsers(opts PageOptions, ptrMaps *rawMaps) PagePointers {
	var allUsers *map[string]models.User = ptrMaps.Users

	users := []models.User{}

	for key, user := range *allUsers {
		// check and correct the corresponding item's key
		if key != user.Nickname {
			user.Nickname = key
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
		part = users
	}

	uExport := make(map[string]models.User)

	for _, user := range part {
		uExport[user.Nickname] = user
	}

	return PagePointers{Users: &uExport}
}

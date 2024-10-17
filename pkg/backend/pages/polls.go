package pages

import (
	"sort"

	"go.vxn.dev/littr/pkg/models"
)

func onePagePolls(opts PageOptions, ptrMaps *PagePointers) PagePointers {
	var allPolls *map[string]models.Poll = ptrMaps.Polls
	//var allUsers *map[string]models.User = ptrMaps.Users

	polls := []models.Poll{}

	// filter out all posts for such callerID
	for key, poll := range *allPolls {
		// check and correct the corresponding item's key
		if key != poll.ID {
			poll.ID = key
		}

		// check the caller's flow list, skip on unfollowed, or unknown user
		/*if value, found := flowList[poll.Author]; !found || !value {
			continue
		}*/

		if poll.Private || poll.Hidden {
			continue
		}

		polls = append(polls, poll)
	}

	// order polls by timestamp DESC
	sort.SliceStable(polls, func(i, j int) bool {
		return polls[i].Timestamp.After(polls[j].Timestamp)
	})

	// cut the PAGE_SIZE*2 number of posts only
	var part []models.Poll

	pageNo := opts.PageNo
	//start := (PAGE_SIZE * 2) * pageNo
	start := (PAGE_SIZE) * pageNo
	//end := (PAGE_SIZE * 2) * (pageNo + 1)
	end := (PAGE_SIZE) * (pageNo + 1)

	if len(polls) > start {
		// only valid for the very first page
		//part = posts[0:(pageSize * 2)]

		if len(polls) <= end {
			// the very last page
			part = polls[start:]
		} else {
			// the middle page
			part = polls[start:end]
		}
	} else {
		// the very single page
		part = polls
	}

	pExport := make(map[string]models.Poll)
	//uExport := make(map[string]models.User)

	for _, poll := range part {
		pExport[poll.ID] = poll
	}

	return PagePointers{Polls: &pExport}
}

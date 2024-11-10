package common

import (
	"bufio"
	"fmt"
	"strings"

	"go.vxn.dev/littr/pkg/models"
	//"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type Event struct {
	LastEventID string
	Type        string
	Data        string
}

func (e *Event) Dump() string {
	return fmt.Sprintf("Type:\t%s\nData:\t%s", e.Type, e.Data)
}

func NewSSEEvent(input string) *Event {
	// Read again and associate fields.
	r := strings.NewReader(input)
	b := bufio.NewReader(r)

	event := &Event{}

	if len(strings.Split(input, "\n")) >= 3 {
		for i := 0; i < 3; i++ {
			lineB, err := b.ReadSlice(byte('\n'))
			if err != nil {
				fmt.Println(err.Error())
				continue
			}

			// Split the event string by ':', trim spaces at the extremites, and join the second field back together.
			parts := strings.Split(string(lineB), ":")
			parts[1] = strings.TrimSpace(parts[1])
			line := strings.Join(parts[1:], " ")

			// Associate the line to event's fields.
			switch i {
			case 0:
				event.LastEventID = line
			case 1:
				event.Type = line
			case 2:
				event.Data = line
			}
		}
	}
	return event
}

func (e *Event) ParseEventData(user *models.User) (string, string) {
	//
	//  Parse the event data
	//

	var text string
	var link string

	// Explode the data CSV string.
	slice := strings.Split(e.Data, ",")

	switch slice[0] {
	// Server is stopping, being stopped, restarting etc.
	case "server-stop":
		text = "server is restarting..."

	// Server is booting up (just started).
	case "server-start":
		text = "server has just started"

	// New post added.
	case "post":
		author := slice[1]
		if author == user.Nickname {
			return "", ""
		}

		// Exit when the author is not followed, nor is found in the user's flowList.
		if user != nil {
			if flowed, found := user.FlowList[author]; !flowed || !found {
				return "", ""
			}
		}

		// Notify the user via toast.
		text = "new post added by " + author

	// New poll added.
	case "poll":
		pollID := slice[1]
		if pollID == "" {
			text = "new poll has been added"
			link = "/polls"
		} else {

			text = "new poll has been added"
			link = "/polls/" + pollID
		}
	}

	return text, link
}

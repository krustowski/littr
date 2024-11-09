package common

import (
	"bufio"
	"fmt"
	"strings"
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

			//
			line := strings.Join(
				strings.Split(
					string(lineB), ":")[1:], " ")

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

package common

import (
	"fmt"
)

type Event struct {
	LastEventID string
	Type        string
	Data        string
}

func (e *Event) Format() string {
	return fmt.Sprintf("Type:\t%s\nData:\t%s", e.Type, e.Data)
}

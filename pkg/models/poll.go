package models

import (
	"time"
)

type Poll struct {
	// ID is an unique poll's identifier.
	ID string `json:"id"`

	// Question is to describe the main purpose of such poll.
	Question string `json:"question"`

	// OptionOne is the answer numero uno.
	OptionOne PollOption `json:"option_one"`

	// OptionTwo is the answer numero dos.
	OptionTwo PollOption `json:"option_two"`

	// OptionThree is the answer numero tres.
	OptionThree PollOption `json:"option_three"`

	// VodeList is the list of user nicknames voted on such poll already.
	Voted []string `json:"voted_list"`

	// Timestamp is an UNIX timestamp indication the poll's creation time; should be identical to the upstream post's Timestamp.
	Timestamp time.Time `json:"timestamp"`

	// Author is the back key to the user originally posting that poll.
	Author string `json:"author"`

	// ReactionCount counts the number of item's reactions.
	ReactionCount int64 `json:"reaction_count"`

	// Experimental fields.
	Hidden  bool     `json:"hidden"`
	Private bool     `json:"private"`
	Tags    []string `json:"tags"`
}

type PollOption struct {
	// Content describes the very content of such poll's option/answer.
	Content string `json:"content"`

	// Counter hold a number of votes being committed to such option.
	Counter int64 `json:"counter"`
}

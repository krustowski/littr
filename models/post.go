package models

import (
	"time"
)

type Post struct {
	// ID is an unique post's identificator.
	ID int `json:"id"`

	// Type describes the post's type --- post, poll, reply, img.
	Type string `json:"type"`

	// Nickname is a name of the post's author's name.
	Nickname string `json:"nickname"`

	// Content contains the very post's data to be shown as a text typed in by the author when created.
	Content string `json:"content"`

	// Timestamp is an UNIX timestamp, indicates the creation time.
	Timestamp time.Time `json:"timestamp"`

	// Poll is an integrated Poll structure/object.
	Poll Poll `json:"poll"`

	// ReplyTo is a referrence to another post, that is being replied to.
	ReplyTo int `json:"reply_to"`

	// ReactionCount counts the number of item's reactions.
	ReactionCount int `json:"reaction_count"`
}

type Poll struct {
	// ID is an unique poll's identifier.
	ID int `json:"id"`

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
}

type PollOption struct {
	// Content describes the very content of such poll's option/answer.
	Content string `json:"content"`

	// Counter hold a number of votes being committed to such option.
	Counter int `json:"counter"`
}

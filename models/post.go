package models

import "time"

type Post struct {
	// Unique post ID.
	ID int `json:"id"`

	// Post type --- post, poll, reply, img
	Type string `json:"type"`

	// Author's account name.
	Nickname string `json:"nickname"`

	// Base64 encoded content.
	Content string `json:"content"`

	// UNIX timestamp of the post publication.
	Timestamp time.Time `json:"timestamp"`

	// Poll content.
	Poll Poll `json:"poll"`

	// Post ID being replied to.
	ReplyTo int `json:"reply_to"`

	// ReactionCount counts the number of item's reactions.
	ReactionCount int `json:"reaction_count"`
}

type Poll struct {
	ID          int        `json:"id"`
	Question    string     `json:"question"`
	OptionOne   PollOption `json:"option_one"`
	OptionTwo   PollOption `json:"option_two"`
	OptionThree PollOption `json:"option_three"`
	Voted       []string   `json:"voted_list"`
	Timestamp   time.Time  `json:"timestamp"`
}

type PollOption struct {
	Content string `json:"content"`
	Counter int    `json:"counter"`
}

package models

import (
	"time"
)

type Post struct {
	// ID is an unique post's identificator.
	ID string `json:"id"`

	// Type describes the post's type --- post, poll, reply, img.
	Type string `json:"type"`

	// Nickname is a name of the post's author's name.
	Nickname string `json:"nickname"`

	// Content contains the very post's data to be shown as a text typed in by the author when created.
	Content string `json:"content"`

	// Figure hold the filename of the uploaded figure to post with some provided text.
	Figure string `json:"figure"`

	// Timestamp is an UNIX timestamp, indicates the creation time.
	Timestamp time.Time `json:"timestamp"`

	// PollID is an identification of the Poll structure/object.
	PollID string `json:"poll_id"`

	// ReplyTo is a reference key to another post, that is being replied to.
	ReplyTo   int    `json:"reply_to"`
	ReplyToID string `json:"reply_to_id"`

	// ReactionCount counts the number of item's reactions.
	ReactionCount int64 `json:"reaction_count"`

	// ReplyCount hold the count of replies for such post.
	ReplyCount int64 `json:"reply_count"`

	// Data is a helper field for the actual figure upload.
	Data []byte `json:"data"`
}

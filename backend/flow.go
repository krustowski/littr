package backend

import (
	"sort"
	"time"
)

var (
	tm1, _        = time.Parse("2006-Jan-02", "1669997122")
	posts  []Post = []Post{
		{Nickname: "system", Content: "welcome onboard bruh, lit ngl", Timestamp: tm1},
	}
)

func GetPosts() *[]Post {
	// order posts by timestamp DESC
	sort.SliceStable(posts, func(i, j int) bool {
		return posts[i].Timestamp.After(posts[j].Timestamp)
	})

	return &posts
}

func AddPost(content string) bool {
	var post Post = Post{
		Nickname:  "random",
		Content:   content,
		Timestamp: time.Now(),
	}

	posts = append(posts, post)

	return true
}

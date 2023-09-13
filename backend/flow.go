package backend

import (
	"encoding/json"
	"log"
	"os"
	//"sort"
	//"strconv"
	//"time"

	"go.savla.dev/littr/models"
)

type Posts struct {
	Posts map[string]models.Post `json:"posts"`
}

func getPosts() *map[string]models.Post {
	var posts Posts

	dat, err := os.ReadFile("/opt/data/posts.json")
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	err = json.Unmarshal(dat, &posts)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	// order posts by timestamp DESC
	/*sort.SliceStable(posts.Posts, func(i, j int) bool {
		return posts.Posts[i].Timestamp.After(posts.Posts[j].Timestamp)
	})*/

	return &posts.Posts
}

func addPost(post models.Post) bool {
	var posts *map[string]models.Post = getPosts()

	/*var newPost models.Post = models.Post{
		Nickname:  post.Nickname,
		Content:   post.Content,
		Timestamp: time.Now(),
	}*/

	//timestamp := strconv.FormatInt(newPost.Timestamp.Unix(), 10)
	//*posts[timestamp] = newPost

	postsToWrite := &Posts{
		Posts: *posts,
	}

	jsonData, err := json.Marshal(postsToWrite)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	err = os.WriteFile("/opt/data/posts.json", jsonData, 0644)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	return true
}

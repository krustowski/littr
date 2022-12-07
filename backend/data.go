package backend

import (
	"encoding/json"
	"log"
	"os"
)

type Data struct {
	users []User `json:"users"`
	posts []Post `json:"posts"`
}

func NewData() *Data {
	var d *Data = &Data{}
	return d
}

func (d *Data) Read(field string) *Data {
	if field == "" {
		return nil
	}

	//f, err := os.Open("/opt/" + field)
	stream, err := os.ReadFile("/opt/data/" + field)
	if err != nil {
		log.Println(err)
		return nil
	}

	switch field {
	case "posts":
		//path = "/opt/flow.json"

		err = json.Unmarshal(stream, &d.posts)
		if err != nil {
			log.Println(err)

			return nil
		}
		break
	case "users":
		//path = "/opt/users.json"

		err = json.Unmarshal(stream, &d.users)
		if err != nil {
			log.Println(err)

			return nil
		}
		break
	}

	return d
}

func (d *Data) SetData(field string, newData interface{}) *Data {
	if field == "" {
		return nil
	}

	switch field {
	case "posts":
		d.posts = append(d.posts, newData.(Post))
		break
	case "users":
		d.users = append(d.users, newData.(User))
		break
	}

	return d
}

func (d *Data) Write(field string) *Data {
	if field == "" {
		return nil
	}

	var (
		err      error
		jsonData []byte
	)

	switch field {
	case "posts":
		jsonData, err = json.Marshal(d.posts)
		break
	case "users":
		jsonData, err = json.Marshal(d.users)
		break
	}

	err = os.WriteFile("/opt/data/"+field, jsonData, 0644)
	if err != nil {
		log.Println(err)
		return nil
	}

	return d
}

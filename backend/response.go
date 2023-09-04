package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"litter-go/models"
)

type response struct {
	AuthGranted bool                   `json:"auth_granted" default:false`
	Code        int                    `json:"code"`
	FlowList    []string               `json:"flow_records"`
	Key         string                 `json:"key"`
	Message     string                 `json:"message"`
	Posts       map[string]models.Post `json:"posts"`
	Users       map[string]models.User `json:"users"`
}

func (r *response) Write(w http.ResponseWriter) error {
	jsonData, err := json.Marshal(r)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	io.WriteString(w, fmt.Sprintf("%s", jsonData))
	return nil
}

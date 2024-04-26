package common

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"go.savla.dev/littr/configs"
	"go.savla.dev/littr/pkg/models"
)

type Response struct {
	AuthGranted bool `json:"auth_granted" default:false`
	Code        int  `json:"code"`

	PublicKey string `json:"public_key,omitempty"`
	Key       string `json:"key,omitempty"`
	Message   string `json:"message"`
	Count     int    `json:"count,omitempty"`

	Subscribed bool          `json:"subscribed"`
	Devices    []models.Device `json:"devices,omitempty"`

	Polls    map[string]models.Poll `json:"polls,omitempty"`
	Posts    map[string]models.Post `json:"posts,omitempty"`
	Users    map[string]models.User `json:"users,omitempty"`
	FlowList []string              `json:"flow_records,omitempty"`

	Data []byte `json:"data,omitempty"`

	// very stats properties
	FlowStats map[string]int            `json:"flow_stats,omitempty"`
	UserStats map[string]models.UserStat `json:"user_stats,omitempty"`

	// auth tokens (JWT)
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

func (r *Response) Write(w http.ResponseWriter) error {
	jsonData, err := json.Marshal(r)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(r.Code)

	io.WriteString(w, fmt.Sprintf("%s", jsonData))

	return nil
}

func (r *Response) WritePix(w http.ResponseWriter) error {
	w.Write(r.Data)

	return nil
}

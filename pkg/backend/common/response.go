package system

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"go.savla.dev/littr/configs"
	"go.savla.dev/littr/pkg/backend/polls"
	"go.savla.dev/littr/pkg/backend/posts"
	"go.savla.dev/littr/pkg/backend/push"
	"go.savla.dev/littr/pkg/backend/stats"
	"go.savla.dev/littr/pkg/backend/users"
)

type Response struct {
	AuthGranted bool `json:"auth_granted" default:false`
	Code        int  `json:"code"`

	PublicKey string `json:"public_key,omitempty"`
	Key       string `json:"key,omitempty"`
	Message   string `json:"message"`
	Count     int    `json:"count,omitempty"`

	Subscribed bool            `json:"subscribed"`
	Devices    []push.Device `json:"devices,omitempty"`

	Polls    map[string]polls.Poll `json:"polls,omitempty"`
	Posts    map[string]posts.Post `json:"posts,omitempty"`
	Users    map[string]users.User `json:"users,omitempty"`
	FlowList []string               `json:"flow_records,omitempty"`

	Data []byte `json:"data,omitempty"`

	// very stats properties
	FlowStats map[string]int            `json:"flow_stats,omitempty"`
	UserStats map[string]stats.UserStat `json:"user_stats,omitempty"`

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

	if configs.EncryptionEnabled {
		w.Header().Add("Content-Type", "application/octet-stream")
	} else {
		w.Header().Add("Content-Type", "application/json")
	}
	w.WriteHeader(r.Code)

	enData := configs.Encrypt([]byte(os.Getenv("APP_PEPPER")), jsonData)
	io.WriteString(w, fmt.Sprintf("%s", enData))

	return nil
}

func (r *Response) WritePix(w http.ResponseWriter) error {
	w.Write(r.Data)

	return nil
}

package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"
	"go.savla.dev/littr/backend/stats"
)

type Response struct {
	AuthGranted bool `json:"auth_granted" default:false`
	Code        int  `json:"code"`

	PublicKey string `json:"public_key,omitempty"`
	Key       string `json:"key,omitempty"`
	Message   string `json:"message"`
	Count     int    `json:"count,omitempty"`

	Subscribed bool            `json:"subscribed"`
	Devices    []models.Device `json:"devices,omitempty"`

	Polls    map[string]models.Poll `json:"polls,omitempty"`
	Posts    map[string]models.Post `json:"posts,omitempty"`
	Users    map[string]models.User `json:"users,omitempty"`
	FlowList []string               `json:"flow_records,omitempty"`

	Data []byte `json:"data,omitempty"`

	// very stats properties
	FlowStats map[string]int      `json:"flow_stats,omitempty"`
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

	if config.EncryptionEnabled {
		w.Header().Add("Content-Type", "application/octet-stream")
	} else {
		w.Header().Add("Content-Type", "application/json")
	}
	w.WriteHeader(r.Code)

	enData := config.Encrypt([]byte(os.Getenv("APP_PEPPER")), jsonData)
	io.WriteString(w, fmt.Sprintf("%s", enData))

	return nil
}

func (r *Response) WritePix(w http.ResponseWriter) error {
	w.Write(r.Data)

	return nil
}

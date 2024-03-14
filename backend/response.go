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
)

type response struct {
	AuthGranted bool `json:"auth_granted" default:false`
	Code        int  `json:"code"`

	PublicKey string `json:"public_key"`
	Key       string `json:"key"`
	Message   string `json:"message"`
	Count     int    `json:"count"`

	Subscribed bool `json:"subscribed"`

	Polls    map[string]models.Poll `json:"polls"`
	Posts    map[string]models.Post `json:"posts"`
	Users    map[string]models.User `json:"users"`
	FlowList []string               `json:"flow_records"`

	Data []byte `json:"data"`

	// very stats properties
	FlowStats map[string]int      `json:"flow_stats"`
	UserStats map[string]userStat `json:"user_stats"`

	// auth tokens (JWT)
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (r *response) Write(w http.ResponseWriter) error {
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

func (r *response) WritePix(w http.ResponseWriter) error {
	enData := config.Encrypt([]byte(os.Getenv("APP_PEPPER")), r.Data)
	w.Write(enData)

	return nil
}

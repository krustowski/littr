package common

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	//"go.vxn.dev/littr/pkg/models"
)

/*func (r *Response) Write(w http.ResponseWriter) error {
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
	// input type: []byte
	w.Write(r.Data)

	return nil
}*/

// new generic API response schema idea
type APIResponse struct {
	// Common fields for all responses
	Message   string `json:"message" example:"a generic success info, or a processing problem/error description"`
	Timestamp int64  `json:"timestamp" example:"1734778064068087800"`

	// Data field for any payload
	Data interface{} `json:"data"`
}

func WriteResponse(w http.ResponseWriter, resp interface{}, code int) error {
	jsonData, err := json.Marshal(resp)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)

	io.WriteString(w, fmt.Sprintf("%s", jsonData))

	return nil
}

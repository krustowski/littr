package common

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	//"go.vxn.dev/littr/pkg/models"
)

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

	if _, err := io.Writer.Write(w, jsonData); err != nil {
		return err
	}

	return nil
}

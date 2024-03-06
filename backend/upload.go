package backend

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	"go.savla.dev/littr/config"
)

func scsAPICall(postID string, userID string, data []byte) error {
	return nil
}

func gscAPICall(filename string, data []byte) error {
	gscToken := os.Getenv("GSC_TOKEN")
	gscURL := os.Getenv("GSC_URL")

	url := gscURL + "uploadImage?api_key=" + gscToken
	method := "POST"

	var bodyReader *bytes.Reader
	var req *http.Request
	var err error

	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}

		payload := config.Encrypt([]byte(os.Getenv("APP_PEPPER")), jsonData)

		bodyReader = bytes.NewReader([]byte(payload))

		req, err = http.NewRequest(method, url, bodyReader)
		if err != nil {
			return err
		}
	} else {
		return errors.New("no data provided!")
	}

	//req.Header.Set("X-API-Caller-ID", caller)

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	_, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return nil
}

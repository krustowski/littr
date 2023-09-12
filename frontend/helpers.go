package frontend

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"litter-go/config"
)

func litterAPI(method, url string, data interface{}) (*[]byte, bool) {
	var bodyReader *bytes.Reader = nil
	var req *http.Request
	var err error

	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Println("cannot marshal data")
			log.Println(err.Error())
			return nil, false
		}

		bodyReader = bytes.NewReader([]byte(jsonData))
		req, err = http.NewRequest(method, url, bodyReader)
		if err != nil {
			log.Println(err.Error())
			return nil, false
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			log.Println(err.Error())
			return nil, false
		}
		req.Header.Set("Content-Type", "application/json")
	}

	//req.Header.Set("X-API-Token", config.BackendToken)

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
		return nil, false
	}

	defer res.Body.Close()

	respData, err := io.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
		return nil, false
	}

	// decrypt the data
	decrData := config.Decrypt([]byte(config.Pepper), respData)

	return &decrData, true
}

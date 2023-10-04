package frontend

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"
)

func litterAPI(method, url string, data interface{}, caller string) (*[]byte, bool) {
	var bodyReader *bytes.Reader
	var req *http.Request
	var err error

	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Println("cannot marshal data")
			log.Println(err.Error())
			return nil, false
		}

		payload := config.Encrypt(config.Pepper, string(jsonData))

		bodyReader = bytes.NewReader([]byte(payload))

		req, err = http.NewRequest(method, url, bodyReader)
		if err != nil {
			log.Println(err.Error())
			return nil, false
		}
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			log.Println(err.Error())
			return nil, false
		}
	}

	if config.EncryptionEnabled {
		req.Header.Set("Content-Type", "application/octet-stream")
	} else {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("X-API-Caller-ID", caller)

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

func prepare[T any](localStorageInput string, model *T) error {
	if localStorageInput == "" {
		return nil
	}

	// beware base64 being used by the framework/browser
	decodedString, err := base64.StdEncoding.DecodeString(string(localStorageInput))
	if err != nil {
		return err
	}

	// decrypt the decoded string if the encryption is enabled (config.EncryptionEnabled)
	// returns the decodedString if the encryption is disabled
	decryptedString := config.Decrypt(
		[]byte(config.Pepper),
		[]byte(decodedString),
	)

	// finally, unmarshal the byte stream into a model
	if err := json.Unmarshal(decryptedString, model); err != nil {
		return err
	}

	return nil
}

func reload[T any](model T, stream *[]byte) error {
	if &model == nil {
		return errors.New("reload: input model is blank")
	}

	preStream, err := json.Marshal(model)
	if err != nil {
		return errors.New("marshal error: model marshal failed" + err.Error())
	}

	encryptedStream := config.Encrypt(
		config.Pepper,
		string(preStream),
	)

	*stream = encryptedStream

	return nil
}

func verifyUser(encodedUser string) bool {
	var user models.User

	if encodedUser == "" {
		return false
	}

	// decode, decrypt and unmarshal the local storage string
	if err := prepare(encodedUser, &user); err != nil {
		log.Println("verification err:" + err.Error())
		return false
	}

	// those fields should not be empty when the whole user struct is encoded/encrypted after login
	if user.Nickname != "" && user.Passphrase != "" {
		return true
	}

	return false
}

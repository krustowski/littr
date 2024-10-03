package common

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"go.vxn.dev/littr/configs"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

var (
	// those vars are used during the build --- linker (ld) bakes the values in
	appVersion     string
	appPepper      string
	vapidPublicKey string
)

type CallInput struct {
	Method      string
	Url         string
	Data        interface{}
	CallerID    string
	PageNo      int
	HideReplies bool
}

func CallAPI[T any](input CallInput, output *T) bool {
	var req *http.Request
	var err error

	if input.Data != nil {
		jsonData, err := json.Marshal(input.Data)
		if err != nil {
			log.Println(err.Error())
			return false
		}

		payload := bytes.NewReader(jsonData)

		req, err = http.NewRequest(input.Method, input.Url, payload)
		if err != nil {
			log.Println(err.Error())
			return false
		}
	} else {
		req, err = http.NewRequest(input.Method, input.Url, nil)
		if err != nil {
			log.Println(err.Error())
			return false
		}
	}

	pageNoString := strconv.FormatInt(int64(input.PageNo), 10)

	version := app.Getenv("APP_VERSION")
	if version == "" {
		version = appVersion
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Caller-ID", input.CallerID)
	req.Header.Set("X-App-Version", version)
	req.Header.Set("X-Page-No", pageNoString)
	req.Header.Set("X-Hide-Replies", fmt.Sprintf("%t", input.HideReplies))

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	defer res.Body.Close()

	respData, err := io.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
		return false
	}

	err = json.Unmarshal(respData, output)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	return true
}

//
//  prolly to be deleted soon
//

func prepare[T any](localStorageInput string, model *T) error {
	if localStorageInput == "" {
		return nil
	}

	// beware base64 being used by the framework/browser
	decodedString, err := base64.StdEncoding.DecodeString(string(localStorageInput))
	if err != nil {
		return err
	}

	// decrypt the decoded string if the encryption is enabled (configs.EncryptionEnabled)
	// returns the decodedString if the encryption is disabled
	decryptedString := configs.Decrypt(
		[]byte(app.Getenv("APP_PEPPER")),
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

	encryptedStream := configs.Encrypt(
		[]byte(app.Getenv("APP_PEPPER")),
		preStream,
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

func readFile(file app.Value) (data []byte, err error) {
	done := make(chan bool)

	// https://developer.mozilla.org/en-US/docs/Web/API/FileReader
	reader := app.Window().Get("FileReader").New()
	reader.Set("onloadend", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		done <- true
		return nil
	}))
	reader.Call("readAsArrayBuffer", file)
	<-done

	readerError := reader.Get("error")
	if !readerError.IsNull() {
		err = fmt.Errorf("file reader error : %s", readerError.Get("message").String())
	} else {
		uint8Array := app.Window().Get("Uint8Array").New(reader.Get("result"))
		data = make([]byte, uint8Array.Length())
		app.CopyBytesToGo(data, uint8Array)
	}
	return data, err
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// https://stackoverflow.com/a/31832326
// https://stackoverflow.com/a/22892986
func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

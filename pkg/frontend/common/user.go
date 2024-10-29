package common

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"go.vxn.dev/littr/pkg/models"
)

func LoadUser(baseString string) *models.User {
	var user models.User

	// Decode the user.
	str, err := base64.StdEncoding.DecodeString(baseString)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	// Unmarshal the result to get an User struct.
	err = json.Unmarshal(str, &user)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	return &user
}

func LoadUser2(encoded string, user *models.User) error {
	if encoded == "" {
		return fmt.Errorf("string input is empty")
	}

	if user == nil {
		return fmt.Errorf("user pointer input is nil")
	}

	// beware base64 being used by the framework/browser
	decodedString, err := base64.StdEncoding.DecodeString(string(encoded))
	if err != nil {
		return err
	}

	// finally, unmarshal the byte stream into a model
	if err := json.Unmarshal(decodedString, user); err != nil {
		return err
	}

	return nil
}

func SaveUser(plain *string, user *models.User) error {
	if plain == nil {
		return fmt.Errorf("string pointer input is empty")
	}

	if user == nil {
		return fmt.Errorf("user pointer input is nil")
	}

	return nil
}

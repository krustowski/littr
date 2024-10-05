package common

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"go.vxn.dev/littr/pkg/models"
)

func LoadUser(encoded string, user *models.User) error {
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

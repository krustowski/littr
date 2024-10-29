package common

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// LoadUser uses the app.Context pointer to load the encoded user string from the LocalStorage to decode it back to the models.User struct.
func LoadUser(user *models.User, ctx *app.Context) error {
	var baseString string

	(*ctx).LocalStorage().Get("user", &baseString)

	// Decode the user.
	str, err := base64.StdEncoding.DecodeString(baseString)
	if err != nil {
		return err
	}

	// Unmarshal the result to get an User struct.
	err = json.Unmarshal(str, user)
	if err != nil {
		return err
	}

	return nil
}

// SaveUser uses the app.Context pointer to save the given pointer to models.User and encode it into a JSON string.
func SaveUser(user *models.User, ctx *app.Context) error {
	// Encode (marshal) the user model into JSON byte stream.
	userJSON, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("%v", ERR_LOCAL_STORAGE_USER_FAIL)
	}

	// Save the encoded user data to the LocalStorage.
	(*ctx).LocalStorage().Set("user", userJSON)
	(*ctx).LocalStorage().Set("authGranted", true)

	return nil
}

//
//  Other attempts.
//

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

func SaveUser2(plain *string, user *models.User) error {
	if plain == nil {
		return fmt.Errorf("string pointer input is empty")
	}

	if user == nil {
		return fmt.Errorf("user pointer input is nil")
	}

	return nil
}

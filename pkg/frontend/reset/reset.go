package reset

import (
	"fmt"

	"go.vxn.dev/littr/pkg/frontend/common"
)

func (c *Content) handleResetRequest(email, uuid string) error {
	if email == "" && uuid == "" {
		return fmt.Errorf("invalid payload data")
	}

	path := "request"

	if uuid != "" {
		path = "reset"
	}

	payload := struct {
		Email string `json:"email"`
		UUID  string `json:"uuid"`
	}{
		Email: email,
		UUID:  uuid,
	}

	input := common.CallInput{
		Method:      "POST",
		Url:         "/api/v1/users/passphrase/" + path,
		Data:        payload,
		CallerID:    "",
		PageNo:      0,
		HideReplies: false,
	}

	response := struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}{}

	if ok := common.CallAPI(input, &response); !ok {
		return fmt.Errorf("communication with backend failed")
	}

	if response.Code != 200 {
		return fmt.Errorf("%s", response.Message)
	}

	return nil
}

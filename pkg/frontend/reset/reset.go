package reset

import (
	"fmt"

	"go.vxn.dev/littr/pkg/frontend/common"
)

func (c *Content) handleResetRequest(email, uuid string) error {
	if email == "" && uuid == "" {
		return fmt.Errorf(common.ERR_RESET_INVALID_INPUT_DATA)
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

	input := &common.CallInput{
		Method:      "POST",
		Url:         "/api/v1/users/passphrase/" + path,
		Data:        payload,
		CallerID:    "",
		PageNo:      0,
		HideReplies: false,
	}

	output := &common.Response{}

	if ok := common.FetchData(input, output); !ok {
		return fmt.Errorf(common.ERR_CANNOT_REACH_BE)
	}

	if output.Code != 200 || output.Code != 201 {
		return fmt.Errorf("%s", output.Message)
	}

	return nil
}

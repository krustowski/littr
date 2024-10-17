package common

import (
	"encoding/json"
	//"errors"
	"io"
	"net/http"
	//"go.vxn.dev/littr/pkg/backend/db"
	//"go.vxn.dev/swis/v5/pkg/core"
)

func UnmarshalRequestData[T any](r *http.Request, model *T) error {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(reqBody, model)
	if err != nil {
		return err
	}

	return nil
}

package db

import (
	"net/http"
	"os"

	"go.vxn.dev/littr/pkg/backend/common"
)

// dumpHandler is the dv package controller function to process system data dump request.
//
//	@Summary		Perform system data dump
//	@Description		This function call is used primarily by the healthcheck function inside the Docker compose stack to periodically dump running data into the JSON files.
//	@Tags			dump
//	@Produce		json
//	@Param			X-Dump-Token	header		string	true	"A special app's dump token."
//	@Success		200				{object}	common.APIResponse{data=models.Stub}	"The dumping process was successful."
//	@Failure		400				{object}	common.APIResponse{data=models.Stub}	"Invalid input data (e.g. a blank token)."
//	@Failure		403				{object}	common.APIResponse{data=models.Stub}	"User unauthorized (e.g. invalid token)."
//	@Failure		429				{object}	common.APIResponse{data=models.Stub}	"Too many requests, try again later."
//	@Router			/dump [get]
func dumpHandler(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "dump")

	// check the incoming API token
	token := r.Header.Get(common.HDR_DUMP_TOKEN)
	if token == "" {
		l.Msg(common.ERR_API_TOKEN_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// validate the incoming token
	if token != os.Getenv("API_TOKEN") {
		l.Msg(common.ERR_API_TOKEN_INVALID).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	//go DumpAll()
	l.Msg(DumpAll()).Status(http.StatusOK).Log().Payload(nil).Write(w)
}

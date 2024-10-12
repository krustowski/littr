package db

import (
	"net/http"
	"os"

	"go.vxn.dev/littr/pkg/backend/common"
)

// dumpHandler is the dv package controller function to process system data dump request.
//
// @Summary      Perform system data dump
// @Description  perform system data dump
// @SecurityDefinitions.apikey ApiKeyAuth
// @in 		 header
// @name 	 X-Dump-Token
// @Tags         dump
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.APIResponse
// @Failure 	 400  {object}   common.APIResponse
// @Failure 	 403  {object}   common.APIResponse
// @Router       /dump [get]
func dumpHandler(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "dump")
	l.CallerID = "system"
	l.Version = "system"

	// check the incoming API token
	token := r.Header.Get("X-Dump-Token")

	if token == "" {
		l.Msg(common.ERR_API_TOKEN_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	if token != os.Getenv("API_TOKEN") {
		l.Msg(common.ERR_API_TOKEN_INVALID).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	//go DumpAll()
	l.Msg(DumpAll()).Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

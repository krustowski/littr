package db

import (
	"net/http"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/config"
)

type dumpController struct {
	db DatabaseKeeper
}

func NewDumpController(db DatabaseKeeper) *dumpController {
	return &dumpController{
		db: db,
	}
}

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
func (c *dumpController) DumpAll(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "dumpController")

	// check the incoming API token
	token := r.Header.Get(common.HDR_DUMP_TOKEN)
	if token == "" {
		l.Msg(common.ERR_API_TOKEN_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// validate the incoming token
	if token != config.DataDumpToken {
		l.Msg(common.ERR_API_TOKEN_INVALID).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	//go DumpAll()
	report, err := c.db.DumpAll()
	if err != nil {
		l.Error(err).Status(http.StatusInternalServerError).Log().Write(w)
	} else {
		l.Msg(report).Status(http.StatusOK).Log().Write(w)
	}
}

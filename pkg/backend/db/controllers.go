package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"go.savla.dev/littr/pkg/backend/common"
)

// dumpHandler is the dv package controller function to process system data dump request.
//
// @Summary      Perform system data dump
// @Description  perform system data dump
// @Tags         dump
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.Response
// @Failure 	 401  {object}   common.Response
// @Failure 	 403  {object}   common.Response
// @Router       /dump [get]
func dumpHandler(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}

	l := common.NewLogger(r, "dump")
	l.CallerID: "system"
	l.Version:    "system"

	// check the incoming API token
	token := r.Header.Get("X-Dump-Token")

	if token == "" {
		resp.Message = "empty token"
		resp.Code = http.StatusUnauthorized

		l.Println(resp.Message, resp.Code)
		return
	}

	if token != os.Getenv("API_TOKEN") {
		resp.Message = "invalid token"
		resp.Code = http.StatusForbidden

		l.Println(resp.Message, resp.Code)
		return
	}

	go DumpAll()

	resp.Code = http.StatusOK
	resp.Message = "data dumped successfully"

	l.Println(resp.Message, resp.Code)
	resp.Write(w)

	return

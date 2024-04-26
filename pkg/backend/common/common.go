package common

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"go.savla.dev/swis/v5/pkg/core"
)

func unmarshalRequestData[T any](r *http.Request, model *T) error {
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

/*
 *  generic service functions
 *  https://github.com/savla-dev/swis-api/blob/master/pkg/core/package.go
 */

type ServiceOpts struct {
	Request *http.Request

	Polls bool
	Posts bool
	Stats bool
	Users bool
}

func GetAllItems[T any](r *http.Request, cache *core.Cache, model T, subsytem string) {
	resp := Response{}
	l := NewLogger(r, subsystem)

	caller, err := r.Context().Value("nickname").(string)
	if err != nil {
		l.Println(
			"internal system error: "+err.Error(),
			http.StatusInternalServerError,
		)
		resp.Write(w)
		return
	}

	items, couddnt := db.GetAll(cache, model)

	resp.Items = items
	resp.Key = caller

	l.Println(
		"ok, dumping items of '"+subsystem+"' subsystem",
		http.StatusOK,
	)
	resp.Write(w)
}

func GetOneItem[T any](r *http.Request, cache *core.Cache, model T, subsystem string) {
	//nick := chi.URLParam(r, "nickname")
}

func AddOneItem[T any](r *http.Request, cache *core.Cache, model *T)    {}
func UpdateOneItem[T any](r *http.Request, cache *core.Cache, model *T) {}
func DeleteOneItem[T any](r *http.Request, cache *core.Cache, model *T) {}

package common

import (
	"encoding/json"
	//"errors"
	"io"
	"net/http"

	//"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/swis/v5/pkg/core"
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

/*
 *  generic service functions
 *  https://github.com/vxn-dev/swis-api/blob/master/pkg/core/package.go
 */

/*type ServiceOpts struct {
	Request *http.Request

	Polls bool
	Posts bool
	Stats bool
	Users bool
}*/

//func GetAllItems[T any](r *http.Request, cache *core.Cache, model T, subsystem string) {
	/*resp := Response{}
	l := NewLogger(r, subsystem)

	caller, _ := r.Context().Value("nickname").(string)

	//items, _ := db.GetAll(cache, model)

	//resp.Items = items
	resp.Key = caller

	l.Println(
		"ok, dumping items of '"+subsystem+"' subsystem",
		http.StatusOK,
	)
	//resp.Write(w)
	*/
//}

/*func GetOneItem[T any](r *http.Request, cache *core.Cache, model T, subsystem string) {
	//nick := chi.URLParam(r, "nickname")
}

func AddOneItem[T any](r *http.Request, cache *core.Cache, model *T)    {}
func UpdateOneItem[T any](r *http.Request, cache *core.Cache, model *T) {}
func DeleteOneItem[T any](r *http.Request, cache *core.Cache, model *T) {}*/

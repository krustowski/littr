package navbars

import (
	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// custom implementation of common.Toast.Dispatch method
func dispatch(t *common.Toast, ic interface{}) {
	c, ok := ic.(*Header)
	if !ok || t.AppContext == nil {
		return
	}

	(*t.AppContext).Dispatch(func(ctx app.Context) {
		c.toast = *t
	})
}

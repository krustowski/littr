package flow

import (
	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

// Custom implementation of common.Toast.Dispatch method.
func dispatch(t *common.Toast, ic interface{}) {
	c, ok := ic.(*Content)
	if !ok || t.AppContext == nil {
		return
	}

	(*t.AppContext).Dispatch(func(ctx app.Context) {
		c.toast = *t
	})
}

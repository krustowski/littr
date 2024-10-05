package polls

import (
	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func dispatch(t *common.Toast, ic interface{}) {
	c, ok := ic.(*Content)
	if !ok {
		return
	}

	if t.AppContext == nil {
		return
	}

	(*t.AppContext).Dispatch(func(ctx app.Context) {
		c.toast = *t
	})
}

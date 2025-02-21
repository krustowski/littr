// The very ToS (terms of service) view logic package.
package tos

import (
	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Content struct {
	app.Compo

	toast     common.Toast
	toastText string
	toastShow bool
}

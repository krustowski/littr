package navbars

import (
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/models"
)

func (h *Header) ensureUIColors() {
	body := app.Window().Get("document").Call("querySelector", "body")
	currentClass := body.Get("className").String()

	var newClass string

	// Check dark/light mode
	switch h.user.UIMode {
	case false:
		newClass = "dark"

	case true:
		newClass = "light"
	}

	// Check UI theme
	switch h.user.UITheme {
	case models.ThemeOrang:
		newClass += "-orang"

	default:
		newClass += "-blu"
	}

	if strings.Contains(currentClass, newClass) {
		return
	}

	body.Set("className", newClass)
}

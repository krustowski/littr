package mail

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
)

type TemplatePayload struct {
	// All templates.
	MainURL     string
	Nickname    string
	TemplateSrc string

	// Reset template.
	Email     string
	ResetLink string
	UUID      string

	// Activation template.
	ActivationLink string

	// Passphrase template.
	Passphrase string
}

var fileAsString = func(templateName string) string {
	// Parse the custom Service Worker template string for the app handler.
	//tpl, err := os.ReadFile("pkg/backend/mail/templates/activation.tmpl")
	tpl, err := os.ReadFile(templateName)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s", tpl)
}

// bakeTemplate parses the given template with the associated payload and writes the output to the <output> pointer address.
func bakeTemplate(payload *TemplatePayload, output *string) error {
	// Ensure new template loaded.
	tmpl := template.Must(template.New(payload.TemplateSrc).Parse(fileAsString(payload.TemplateSrc)))

	var buf bytes.Buffer

	// Bake the template into buffer.
	err := tmpl.Execute(&buf, *payload)
	if err != nil {
		return err
	}

	// Save the output string to the output's address.
	*output = buf.String()

	return nil
}

package softimdb

import (
	"fmt"
	"html"
	"os"

	"github.com/hultan/dialog"
)

func cleanString(text string) string {
	text = html.EscapeString(text)
	return text
}

func reportError(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	dialog.Title(applicationTitle).Text(err.Error()).
		ErrorIcon().OkButton().Show()
}

package softimdb

import (
	"fmt"
	"html"
	"os"

	"github.com/hultan/softteam/framework"
)

func cleanString(text string) string {
	text = html.EscapeString(text)
	return text
}

func reportError(err error) {
	fw := framework.NewFramework()
	fmt.Fprintln(os.Stderr, err)
	fw.Gtk.Title(applicationTitle).Text(err.Error()).
		ErrorIcon().OkButton().Show()
}

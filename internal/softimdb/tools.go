package softimdb

import (
	"fmt"
	"html"
	"os"

	"github.com/hultan/softteam/framework"
)

// ErrorCheckWithPanic : panics on error
func ErrorCheckWithPanic(err error, message string) {
	if err != nil {
		panic(err.Error() + " : " + message)
	}
}

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

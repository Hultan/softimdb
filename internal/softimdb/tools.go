package softimdb

import (
	"html"
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

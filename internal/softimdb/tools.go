package softimdb

import (
	"strings"
)

// ErrorCheckWithPanic : panics on error
func ErrorCheckWithPanic(err error, message string) {
	if err != nil {
		panic(err.Error() + " : " + message)
	}
}

func cleanString(text string) string {
	return strings.Replace(text,"&", "&amp;",-1)
}

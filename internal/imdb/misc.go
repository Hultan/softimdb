package imdb

import (
	"strings"
)

type parameter struct {
	key, value string
}

// replaceParameters : Returns the url with all parameters replaced
func replaceParameters(url string, parameters []parameter) string {
	for _, p := range parameters {
		url = strings.Replace(url, p.key, p.value, 1)
	}

	return url
}

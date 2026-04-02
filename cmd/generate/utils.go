package main

import (
	"fmt"
	"strings"

	"github.com/invopop/gobl/cbc"
)

// backtick returns a literal backtick character for use in templates
// where the surrounding Go raw string literal prevents using backticks directly.
func backtick() string {
	return "`"
}

func joinKeys(keys []cbc.Key) string {
	var s []string
	for _, k := range keys {
		s = append(s, fmt.Sprintf("<code>%s</code>", k.String()))
	}
	return strings.Join(s, ", ")
}

func codeMap(m cbc.CodeMap) string {
	var s []string
	for k, v := range m {
		s = append(s, fmt.Sprintf("<code>%s:%s</code>", k, v))
	}
	return strings.Join(s, ", ")
}

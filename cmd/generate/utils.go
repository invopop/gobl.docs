package main

import (
	"fmt"
	"strings"

	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/tax"
)

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

func extMap(m tax.Extensions) string {
	var s []string
	for k, v := range m {
		s = append(s, fmt.Sprintf("<code>%s:%s</code>", k, v))
	}
	return strings.Join(s, ", ")
}

// scenarioTitle builds a concise title for a scenario accordion.
func scenarioTitle(sc *tax.Scenario) string {
	if sc.Name != nil {
		if n := sc.Name.String(); n != "" {
			return n
		}
	}
	var parts []string
	for _, t := range sc.Types {
		parts = append(parts, t.String())
	}
	for _, t := range sc.Tags {
		parts = append(parts, "#"+t.String())
	}
	if sc.ExtKey != "" {
		label := sc.ExtKey.String()
		if sc.ExtCode != "" {
			label += "=" + sc.ExtCode.String()
		}
		parts = append(parts, label)
	}
	if len(parts) == 0 {
		return "Scenario"
	}
	return strings.Join(parts, ", ")
}

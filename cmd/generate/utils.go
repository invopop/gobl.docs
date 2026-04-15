package main

import (
	"fmt"
	"html"
	"strings"

	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/tax"
)

func joinKeys(keys []cbc.Key) string {
	var s []string
	for _, k := range keys {
		s = append(s, fmt.Sprintf("`%s`", k.String()))
	}
	return strings.Join(s, ", ")
}

func codeMap(m cbc.CodeMap) string {
	var s []string
	for k, v := range m {
		s = append(s, fmt.Sprintf("`%s:%s`", k, v))
	}
	return strings.Join(s, ", ")
}

func extMap(m tax.Extensions) string {
	var s []string
	for k, v := range m {
		s = append(s, fmt.Sprintf("`%s:%s`", k, v))
	}
	return strings.Join(s, ", ")
}

func sentenceCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// rootTypeLabel returns the short type name for display when a rule has no field path
// (e.g. "bill.Invoice" → "Invoice", "tax.Combo" → "Combo").
func rootTypeLabel(qualifiedName string) string {
	if qualifiedName == "" {
		return ""
	}
	if i := strings.LastIndex(qualifiedName, "."); i >= 0 {
		return qualifiedName[i+1:]
	}
	return qualifiedName
}

func fieldCell(field string, calculated bool, qualifiedTypeName string) string {
	calcLabel := "<small class=\"gobl-field-calculated\">Calculated</small>"
	if field == "" {
		label := rootTypeLabel(qualifiedTypeName)
		if label == "" {
			label = "document"
		}
		esc := html.EscapeString(label)
		if calculated {
			return "<small>" + esc + "</small><br />" + calcLabel
		}
		return "<small>" + esc + "</small>"
	}
	if calculated {
		return fmt.Sprintf("`%s`<br />%s", field, calcLabel)
	}
	return fmt.Sprintf("`%s`", field)
}

func codeMessage(code, desc string) string {
	return fmt.Sprintf("`%s`<br />%s", code, sentenceCase(desc))
}

func testList(parts []string, calculated bool) string {
	var buf strings.Builder
	buf.WriteString("<ul class=\"gobl-test\">")
	for _, p := range parts {
		if p == "Present" && !calculated {
			buf.WriteString("<li class=\"gobl-test-present\">")
		} else {
			buf.WriteString("<li>")
		}
		buf.WriteString(p)
		buf.WriteString("</li>")
	}
	buf.WriteString("</ul>")
	return buf.String()
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

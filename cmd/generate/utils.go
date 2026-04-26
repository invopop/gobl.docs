package main

import (
	"encoding/json"
	"fmt"
	"html"
	"reflect"
	"runtime"
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

// customFilterLabel produces the marker shown in the Tags column for a scenario
// that uses a custom Filter function. It prefers the scenario's own Desc/Name
// (i18n strings the GOBL data model exposes for documentation) and falls back
// to a "(custom)" link pointing at the function's source on GitHub so an
// author who forgot to fill in Desc still gives readers somewhere to look.
func customFilterLabel(sc *tax.Scenario) string {
	if desc := sc.Desc.String(); desc != "" {
		return fmt.Sprintf("*%s*", desc)
	}
	if name := sc.Name.String(); name != "" {
		return fmt.Sprintf("*%s*", name)
	}
	if src := customFilterSource(sc.Filter); src != "" {
		return fmt.Sprintf("[custom](%s)", src)
	}
	return "custom"
}

// customFilterSource returns a GitHub permalink (pinned at the gobl version
// this generator is built against) to the source location of the supplied
// custom filter function, or "" if the source can't be resolved.
func customFilterSource(filter func(doc any) bool) string {
	if filter == nil {
		return ""
	}
	pc := reflect.ValueOf(filter).Pointer()
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return ""
	}
	file, line := fn.FileLine(pc)
	const marker = "github.com/invopop/gobl@"
	idx := strings.Index(file, marker)
	if idx < 0 {
		return ""
	}
	rest := file[idx+len(marker):]
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) != 2 {
		return ""
	}
	return fmt.Sprintf("https://github.com/invopop/gobl/blob/%s/%s#L%d", parts[0], parts[1], line)
}

// scenarioTagsCell renders the Tags column of a scenario row, including any
// ExtKey/ExtCode filter or custom-filter marker so no filter information is
// silently dropped.
func scenarioTagsCell(sc *tax.Scenario) string {
	var parts []string
	for _, t := range sc.Tags {
		parts = append(parts, fmt.Sprintf("`%s`", t.String()))
	}
	if sc.ExtKey != "" {
		label := sc.ExtKey.String()
		if sc.ExtCode != "" {
			label += "=" + sc.ExtCode.String()
		}
		parts = append(parts, fmt.Sprintf("`%s`", label))
	}
	if sc.Filter != nil {
		parts = append(parts, customFilterLabel(sc))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "<br />")
}

func scenarioTypesCell(sc *tax.Scenario) string {
	if len(sc.Types) == 0 {
		return "-"
	}
	var s []string
	for _, t := range sc.Types {
		s = append(s, fmt.Sprintf("`%s`", t.String()))
	}
	return strings.Join(s, "<br />")
}

func scenarioCategoriesCell(sc *tax.Scenario) string {
	if len(sc.Categories) == 0 {
		return "-"
	}
	var s []string
	for _, c := range sc.Categories {
		s = append(s, fmt.Sprintf("`%s`", c.String()))
	}
	return strings.Join(s, "<br />")
}

// scenarioOutputJSON is used to marshal scenario outputs in a stable field
// order (note → codes → ext) regardless of which fields are present.
type scenarioOutputJSON struct {
	Note  *tax.Note      `json:"note,omitempty"`
	Codes cbc.CodeMap    `json:"codes,omitempty"`
	Ext   tax.Extensions `json:"ext,omitempty"`
}

// scenarioOutputCell renders the Output column as an indented JSON snippet
// showing the scenario outputs in the shape GOBL emits when applying them to a
// document. Newlines are rendered as <br/> and leading spaces as &nbsp; so the
// content fits on a single source line in a markdown table cell while
// preserving indentation. Braces and pipes are escaped so MDX and the table
// parser don't choke.
func scenarioOutputCell(sc *tax.Scenario) string {
	if sc.Note == nil && len(sc.Codes) == 0 && len(sc.Ext) == 0 {
		return "-"
	}
	out := scenarioOutputJSON{Note: sc.Note, Codes: sc.Codes, Ext: sc.Ext}
	raw, _ := json.MarshalIndent(out, "", "  ")
	lines := strings.Split(string(raw), "\n")
	for i, line := range lines {
		n := 0
		for n < len(line) && line[n] == ' ' {
			n++
		}
		lines[i] = strings.Repeat("&nbsp;", n) + line[n:]
	}
	body := strings.Join(lines, "<br/>")
	body = strings.ReplaceAll(body, "{", "&#123;")
	body = strings.ReplaceAll(body, "}", "&#125;")
	body = strings.ReplaceAll(body, "|", "\\|")
	return `<code class="code-block">` + body + "</code>"
}

// scenarioTable renders a complete markdown table for a scenario set, omitting
// any filter columns whose values are empty across all rows.
func scenarioTable(set *tax.ScenarioSet) string {
	if set == nil || len(set.List) == 0 {
		return ""
	}
	var showTags, showTypes, showCategories bool
	for _, sc := range set.List {
		if len(sc.Tags) > 0 || sc.ExtKey != "" || sc.Filter != nil {
			showTags = true
		}
		if len(sc.Types) > 0 {
			showTypes = true
		}
		if len(sc.Categories) > 0 {
			showCategories = true
		}
	}
	var headers []string
	if showTags {
		headers = append(headers, "Tags")
	}
	if showTypes {
		headers = append(headers, "Type")
	}
	if showCategories {
		headers = append(headers, "Categories")
	}
	headers = append(headers, "Output")

	var b strings.Builder
	b.WriteString("| " + strings.Join(headers, " | ") + " |\n")
	b.WriteString("|" + strings.Repeat(" --- |", len(headers)) + "\n")
	for _, sc := range set.List {
		var cells []string
		if showTags {
			cells = append(cells, scenarioTagsCell(sc))
		}
		if showTypes {
			cells = append(cells, scenarioTypesCell(sc))
		}
		if showCategories {
			cells = append(cells, scenarioCategoriesCell(sc))
		}
		cells = append(cells, scenarioOutputCell(sc))
		b.WriteString("| " + strings.Join(cells, " | ") + " |\n")
	}
	return b.String()
}

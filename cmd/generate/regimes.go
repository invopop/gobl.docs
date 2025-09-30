package main

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"

	_ "github.com/invopop/gobl"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/i18n"
	"github.com/invopop/gobl/pkg/here"
	"github.com/invopop/gobl/tax"
)

type regimeGenerator struct {
	generator
	regime *tax.RegimeDef
}

// RateRow represents a single row in the tax rates table
type RateRow struct {
	Rate        cbc.Key
	Keys        []cbc.Key
	Name        i18n.String
	Description i18n.String
	Percent     string
	Extensions  map[cbc.Key]string // extension key -> extension value
}

func newRegimeGenerator(r *tax.RegimeDef) *regimeGenerator {
	g := &regimeGenerator{
		generator: generator{
			buf: new(bytes.Buffer),
		},
		regime: r,
	}
	g.tmpl = template.New("base")
	g.tmpl.Funcs(template.FuncMap{
		"t": func(s i18n.String) string {
			return s.String()
		},
		"rate":          g.taxRateValue,
		"extension":     g.taxRateExtension,
		"extensionKeys": g.getExtensionKeys,
		"rateRows":      g.getRateRows,
		"joinKeys":      joinKeys,
		"codeMap":       codeMap,
	})
	return g
}

func (g *regimeGenerator) generate() error {
	if err := g.base(); err != nil {
		return err
	}
	if err := g.taxCategories(); err != nil {
		return err
	}
	if err := g.preceding(g.regime.Corrections); err != nil {
		return err
	}
	/*
		if err := g.scenarios(); err != nil {
			return err
		}
	*/
	if err := g.extensions(g.regime.Extensions); err != nil {
		return err
	}
	return nil
}

func (g *regimeGenerator) base() error {
	return g.process(here.Doc(`
		---
		title: {{t .Name}}
		---

		{{- if .Description}}
		{{t .Description}}
		{{- end}}

		## Base Details

		| Key | Value |
		| --- | ----- |
		| Tax Country Code | <code>{{.Country}}</code> |
		| Currency | <code>{{.Currency}}</code> |
		| Base Time Zone | <code>{{.TimeZone}}</code> |
	`))
}

func (g *regimeGenerator) taxCategories() error {
	err := g.process(here.Doc(`


		## Tax Categories

		| Code | Name | Title |
		| ---- | ---- | ----- |
		{{- range .Categories }}
		| <code>{{ .Code }}</code> | {{t .Name }} | {{t .Title }} |
		{{- end}}


	`))
	if err != nil {
		return err
	}
	for _, c := range g.regime.Categories {
		if err := g.taxCategory(c); err != nil {
			return err
		}
	}
	return nil
}

func (g *regimeGenerator) taxCategory(tc *tax.CategoryDef) error {
	if tc.Description == nil && len(tc.Rates) == 0 {
		return nil
	}
	return g.processWith(here.Doc(`


		### {{t .Name}} Rates
		{{- if .Description }}	

		{{t .Description }}
		{{- end }}

		{{- if .Rates }}
		{{- $extKeys := extensionKeys . }}
		{{- $rows := rateRows . }}

		| Rate | Keys | Name |{{- range $extKeys }} {{.}} |{{- end }} Percents | Description |
		| ---- | ---- | ---- |{{- range $extKeys }} --------- |{{- end }} -------- | ----------- |
		{{- range $row := $rows }}
		| <code>{{ $row.Rate }}</code> | {{ joinKeys $row.Keys }} | {{t $row.Name }} |{{- range $extKeys }} {{ index $row.Extensions . }} |{{- end }} {{ $row.Percent }} | {{t $row.Description }} |
		{{- end }}
		{{- else }}
		No rates defined.
		{{- end }}
	`), tc)
}

func (g *regimeGenerator) taxRateValue(tr *tax.RateDef) string {
	if len(tr.Values) == 0 {
		return ""
	}

	// First, try to find a value without extensions (default case)
	for _, v := range tr.Values {
		if len(v.Ext) == 0 {
			item := v.Percent.String()
			if v.Surcharge != nil {
				item = fmt.Sprintf("%s (+%s)", item, v.Surcharge.String())
			}
			return item
		}
	}

	// If no value without extensions found, use the first one
	v := tr.Values[0]
	item := v.Percent.String()
	if v.Surcharge != nil {
		item = fmt.Sprintf("%s (+%s)", item, v.Surcharge.String())
	}
	return item
}

func (g *regimeGenerator) taxRateExtension(tr *tax.RateDef) string {
	if len(tr.Values) == 0 {
		return ""
	}

	// Check if there are any values with extensions
	hasExtensions := false
	for _, v := range tr.Values {
		if len(v.Ext) > 0 {
			hasExtensions = true
			break
		}
	}

	if !hasExtensions {
		return ""
	}

	// If there are extensions, show them
	extensions := make([]string, 0)
	for _, v := range tr.Values {
		if len(v.Ext) > 0 {
			for key, value := range v.Ext {
				extInfo := fmt.Sprintf("%s: %s (%s)", key, value, v.Percent.String())
				extensions = append(extensions, extInfo)
			}
		}
	}

	if len(extensions) == 0 {
		return ""
	}

	// Join all extensions with line breaks for better readability
	result := ""
	for i, ext := range extensions {
		if i > 0 {
			result += "<br/>"
		}
		result += ext
	}
	return result
}

// getExtensionKeys returns all unique extension keys used in a tax category
func (g *regimeGenerator) getExtensionKeys(tc *tax.CategoryDef) []cbc.Key {
	keySet := make(map[cbc.Key]bool)

	for _, rate := range tc.Rates {
		for _, value := range rate.Values {
			for key := range value.Ext {
				keySet[key] = true
			}
		}
	}

	keys := make([]cbc.Key, 0, len(keySet))
	for key := range keySet {
		keys = append(keys, key)
	}

	// Sort keys for consistent output
	sort.Slice(keys, func(i, j int) bool {
		return string(keys[i]) < string(keys[j])
	})

	return keys
}

// getRateRows converts tax rates into individual rows for the table
func (g *regimeGenerator) getRateRows(tc *tax.CategoryDef) []*RateRow {
	var rows []*RateRow

	for _, rate := range tc.Rates {
		if len(rate.Values) == 0 {
			rows = append(rows, &RateRow{
				Rate: rate.Rate, Keys: rate.Keys, Name: rate.Name, Description: rate.Description,
				Percent: "", Extensions: make(map[cbc.Key]string),
			})
			continue
		}

		// Group values by extension signature, keeping only the most recent for each
		groups := make(map[string]*tax.RateValueDef)
		for _, value := range rate.Values {
			sig := g.extSignature(value.Ext)
			if _, exists := groups[sig]; !exists {
				groups[sig] = value
			}
		}

		// Separate default and extension values
		var defaultVal *tax.RateValueDef
		var extVals []*tax.RateValueDef
		for sig, val := range groups {
			if sig == "default" {
				defaultVal = val
			} else {
				extVals = append(extVals, val)
			}
		}

		// Sort extensions by key for consistency
		sort.Slice(extVals, func(i, j int) bool {
			for k1 := range extVals[i].Ext {
				for k2 := range extVals[j].Ext {
					return string(k1) < string(k2)
				}
			}
			return false
		})

		// Build combined row
		var percents []string
		combinedExt := make(map[cbc.Key]string)

		// Add default first
		if defaultVal != nil {
			percents = append(percents, g.formatPercent(defaultVal))
		}

		// Add extensions
		extKeyVals := make(map[cbc.Key][]string)
		for _, val := range extVals {
			percents = append(percents, g.formatPercent(val))
			for key, extVal := range val.Ext {
				extKeyVals[key] = append(extKeyVals[key], extVal.String())
			}
		}

		// Format extension columns
		for key, vals := range extKeyVals {
			prefix := ""
			if defaultVal != nil {
				prefix = "<br/>"
			}
			combinedExt[key] = prefix + strings.Join(vals, "<br/>")
		}

		rows = append(rows, &RateRow{
			Rate: rate.Rate, Keys: rate.Keys, Name: rate.Name, Description: rate.Description,
			Percent: strings.Join(percents, " <br/>"), Extensions: combinedExt,
		})
	}

	return rows
}

// extSignature creates a unique signature for extension combinations
func (g *regimeGenerator) extSignature(ext tax.Extensions) string {
	if len(ext) == 0 {
		return "default"
	}
	var keys []string
	for k, v := range ext {
		keys = append(keys, fmt.Sprintf("%s:%s", k, v.String()))
	}
	sort.Strings(keys)
	return strings.Join(keys, "|")
}

// formatPercent formats a percentage with optional surcharge
func (g *regimeGenerator) formatPercent(val *tax.RateValueDef) string {
	percent := val.Percent.String()
	if val.Surcharge != nil {
		percent = fmt.Sprintf("%s (+%s)", percent, val.Surcharge.String())
	}
	return percent
}

func (g *regimeGenerator) scenarios() error {
	if len(g.regime.Scenarios) == 0 {
		return nil
	}
	return g.process(here.Doc(`


		## Scenarios

		{{- range .Scenarios }}
		For <code>{{ .Schema }}</code>:

		| Types | Tags | Name | Note Applied |
		| ----- | ---- | ---- | ------------ |
		{{- range .List }}
		| {{ joinKeys .Types }} | {{ joinKeys .Tags }} | {{t .Name }} | {{ if .Note }}{{ .Note.Text }}{{ end }} |
		{{- end}}
		{{- end}}

	`))
}

func (g *regimeGenerator) process(doc string) error {
	return g.generator.process(doc, g.regime)
}

func (g *regimeGenerator) processWith(doc string, data interface{}) error {
	return g.generator.process(doc, data)
}

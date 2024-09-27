package main

import (
	"bytes"
	"fmt"
	"text/template"

	_ "github.com/invopop/gobl"
	"github.com/invopop/gobl/i18n"
	"github.com/invopop/gobl/pkg/here"
	"github.com/invopop/gobl/tax"
)

type regimeGenerator struct {
	generator
	regime *tax.RegimeDef
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
		"rate":     g.taxRateValue,
		"joinKeys": joinKeys,
		"codeMap":  codeMap,
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

		| Key | Name | Percents | Description |
		| --- | ---- | -------- | ----------- |
		{{- range .Rates }}
		| <code>{{ .Key }}</code> | {{t .Name }} | {{ rate . }} | {{t .Description }} |
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
	v := tr.Values[0]
	item := v.Percent.String()
	if v.Surcharge != nil {
		item = fmt.Sprintf("%s (+%s)", item, v.Surcharge.String())
	}
	return item
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

package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	_ "github.com/invopop/gobl"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/i18n"
	"github.com/invopop/gobl/pkg/here"
	"github.com/invopop/gobl/tax"
)

func main() {
	// Build regime definitions
	regimes := tax.Regimes()

	for _, r := range regimes.All() {
		g := newGenerator(r)
		if err := g.generate(); err != nil {
			panic(err)
		}
		name := fmt.Sprintf("./regimes/%s.mdx", strings.ToLower(r.Country.String()))
		if err := os.WriteFile(name, g.bytes(), 0664); err != nil {
			panic(err)
		}
		fmt.Printf("Wrote %s\n", name)
	}

}

type generator struct {
	regime *tax.Regime
	buf    *bytes.Buffer
	tmpl   *template.Template
}

func newGenerator(r *tax.Regime) *generator {
	g := &generator{
		regime: r,
		buf:    new(bytes.Buffer),
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

func (g *generator) bytes() []byte {
	return g.buf.Bytes()
}

func (g *generator) generate() error {
	if err := g.base(); err != nil {
		return err
	}
	if err := g.taxCategories(); err != nil {
		return err
	}
	if err := g.preceding(); err != nil {
		return err
	}
	/*
		if err := g.scenarios(); err != nil {
			return err
		}
	*/
	if err := g.extensions(); err != nil {
		return err
	}
	return nil
}

func (g *generator) base() error {
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

func (g *generator) taxCategories() error {
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

func (g *generator) taxCategory(tc *tax.Category) error {
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

func (g *generator) taxRateValue(tr *tax.Rate) string {
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

func (g *generator) preceding() error {
	defs := g.regime.Corrections
	if defs == nil {
		return nil
	}

	var cd *tax.CorrectionDefinition
	for _, d := range defs {
		if d.Schema == bill.ShortSchemaInvoice {
			cd = d
		}
	}
	if cd == nil {
		return nil
	}

	return g.processWith(here.Doc(`


		## Correction Definitions

		This tax regime supports auto-generation of corrective invoices
		or credit and debit notes.

		{{- if .ReasonRequired }}

		This tax regime requires a reason to be provided in the <code>reason</code> field
		when submitting the correction options.

		{{- end }}
		{{- if .Types }}

		### Invoice Types

		The types of invoices that can be created with a preceding definition:
		{{- range .Types }}
		- <code>{{ . }}</code>
		{{- end }}
		{{- end }}

		{{- if .Stamps }}

		### Stamp Keys
		
		Stamp keys from the previous invoice that need to be referenced:

		{{- range .Stamps }}
		- <code>{{ . }}</code>
		{{- end }}
		{{- end}}

		{{- if .Extensions }}

		### Extension Keys

		One or all of the following extensions may be required as part of the correction
		options. See the [Extensions](#extensions) section for possible values.

		{{- range .Extensions }}
		- <code>{{ . }}</code>
		{{- end }}
		{{- end}}

	`), cd)
}

func (g *generator) extensions() error {
	if len(g.regime.Extensions) == 0 {
		return nil
	}

	if err := g.process(here.Doc(`
		

		## Extensions

		The following extensions are supported by this tax regime.
	`)); err != nil {
		return err
	}

	for _, kd := range g.regime.Extensions {
		if err := g.extension(kd); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) extension(kd *cbc.KeyDefinition) error {

	return g.processWith(here.Doc(`


		### {{t .Name}}
		
		Key: <code>{{ .Key }}</code>
		
		{{- if .Desc }}	

		{{t .Desc }}

		{{- end }}

 		{{- if .Keys }}

		| Key | Name |
		| ---- | ---- |
		{{- range .Keys }}
		| <code>{{ .Key }}</code> | {{t .Name }} |
		{{- end }}
		{{- end }}

		{{- if .Codes }}

		| Code | Name |
		| ---- | ---- |
		{{- range .Codes }}
		| <code>{{ .Code }}</code> | {{t .Name }} |
		{{- end }}
		{{- end }}
	`), kd)
}

func codeMap(m cbc.CodeMap) string {
	var s []string
	for k, v := range m {
		s = append(s, fmt.Sprintf("<code>%s:%s</code>", k, v))
	}
	return strings.Join(s, ", ")
}

func (g *generator) scenarios() error {
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

func joinKeys(keys []cbc.Key) string {
	var s []string
	for _, k := range keys {
		s = append(s, fmt.Sprintf("<code>%s</code>", k.String()))
	}
	return strings.Join(s, ", ")
}

func (g *generator) process(doc string) error {
	tmpl, err := g.tmpl.Parse(doc)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(g.buf, g.regime); err != nil {
		return err
	}
	return nil
}

func (g *generator) processWith(doc string, data interface{}) error {
	tmpl, err := g.tmpl.Parse(doc)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(g.buf, data); err != nil {
		return err
	}
	return nil
}

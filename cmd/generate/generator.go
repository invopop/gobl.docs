package main

import (
	"bytes"
	"text/template"

	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/pkg/here"
	"github.com/invopop/gobl/tax"
)

// generator defines the base structure for generator things,
// including some common templates.
type generator struct {
	buf  *bytes.Buffer
	tmpl *template.Template
}

func (g *generator) bytes() []byte {
	return g.buf.Bytes()
}

func (g *generator) process(doc string, data any) error {
	tmpl, err := g.tmpl.Parse(doc)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(g.buf, data); err != nil {
		return err
	}
	return nil
}

func (g *generator) preceding(defs tax.CorrectionSet) error {
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

	return g.process(here.Doc(`


		## Correction Definitions

		Auto-generation of corrective invoices or credit and debit notes is
		supported.

		{{- if .ReasonRequired }}

		A reason is required in the <code>reason</code> field
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

func (g *generator) extensions(exts []*cbc.Definition) error {
	if len(exts) == 0 {
		return nil
	}

	if err := g.process(here.Doc(`
		

		## Extensions

		The following extensions are supported.
	`), nil); err != nil {
		return err
	}

	for _, kd := range exts {
		if err := g.extension(kd); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) extension(kd *cbc.Definition) error {

	return g.process(here.Doc(`


		### {{t .Name}}
		
		{{- if .Key }}
		Key: <code>{{ .Key }}</code>
		{{- else }}
		Code: <code>{{ .Code }}</code>
		{{- end }}

		{{- if .Desc }}	

		{{t .Desc }}

		{{- end }}

 		{{- if .Values }}

		| Code | Name |
		| ---- | ---- |
		{{- range .Values }}
		| <code>{{ .Code }}</code> | {{t .Name }} |
		{{- end }}
		{{- end }}

	`), kd)
}

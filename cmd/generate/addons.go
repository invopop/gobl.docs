package main

import (
	"bytes"
	"text/template"

	"github.com/invopop/gobl/i18n"
	"github.com/invopop/gobl/pkg/here"
	"github.com/invopop/gobl/tax"
)

type addonGenerator struct {
	generator
	addon *tax.AddonDef
}

func newAddonGenerator(a *tax.AddonDef) *addonGenerator {
	g := &addonGenerator{
		generator: generator{
			buf: new(bytes.Buffer),
		},
		addon: a,
	}
	g.tmpl = template.New("base")
	g.tmpl.Funcs(template.FuncMap{
		"t": func(s i18n.String) string {
			return s.String()
		},
		"joinKeys": joinKeys,
		"codeMap":  codeMap,
	})
	return g
}

func (g *addonGenerator) generate() error {
	if err := g.base(); err != nil {
		return err
	}
	if err := g.preceding(g.addon.Corrections); err != nil {
		return err
	}
	if err := g.extensions(g.addon.Extensions); err != nil {
		return err
	}
	return nil
}

func (g *addonGenerator) base() error {
	return g.process(here.Doc(`
		---
		title: {{t .Name}}
		---

		Key: <code>{{ .Key }}</code>

		{{- if .Description}}

		{{t .Description}}
		{{- end}}

		{{- if .Sources}}

		## Sources

		{{- range .Sources}}
		- [{{t .Title}}]({{ .URL }})
		{{- end}}
		{{- end}}
	`))
}

func (g *addonGenerator) process(doc string) error {
	return g.generator.process(doc, g.addon)
}

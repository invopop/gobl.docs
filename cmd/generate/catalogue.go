package main

import (
	"bytes"
	"text/template"

	"github.com/invopop/gobl/i18n"
	"github.com/invopop/gobl/pkg/here"
	"github.com/invopop/gobl/tax"
)

type catalogueGenerator struct {
	generator
	catalogue *tax.CatalogueDef
}

func newCatalogueGenerator(d *tax.CatalogueDef) *catalogueGenerator {
	g := &catalogueGenerator{
		generator: generator{
			buf: new(bytes.Buffer),
		},
		catalogue: d,
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

func (g *catalogueGenerator) generate() error {
	if err := g.base(); err != nil {
		return err
	}
	if err := g.extensions(g.catalogue.Extensions); err != nil {
		return err
	}
	return nil
}

func (g *catalogueGenerator) base() error {
	return g.process(here.Doc(`
		---
		title: {{t .Name}}
		---

		{{- if .Description}}
		{{t .Description}}
		{{- end}}

	`))
}

func (g *catalogueGenerator) process(doc string) error {
	return g.generator.process(doc, g.catalogue)
}

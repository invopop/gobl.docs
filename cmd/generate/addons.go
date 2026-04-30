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
		"joinKeys":      joinKeys,
		"codeMap":       codeMap,
		"extMap":        extMap,
		"scenarioTitle": scenarioTitle,
		"codeMessage":   codeMessage,
		"testList":      testList,
		"fieldCell":     fieldCell,
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
	if err := g.scenarios(); err != nil {
		return err
	}
	if err := g.extensions(g.addon.Extensions); err != nil {
		return err
	}
	if err := g.validationRules(g.getAddonRuleSections()); err != nil {
		return err
	}
	return nil
}

func (g *addonGenerator) getAddonRuleSections() []RuleSection {
	topSet := findAddonSet(g.addon.Key)
	if topSet == nil {
		return nil
	}
	return ruleSections(topSet)
}

func (g *addonGenerator) base() error {
	return g.process(here.Doc(`
		---
		title: {{t .Name}}
		---

		Key: ~{{ .Key }}~

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

func (g *addonGenerator) scenarios() error {
	if len(g.addon.Scenarios) == 0 {
		return nil
	}
	return g.process(here.Doc(`


		## Scenarios

		{{- range .Scenarios }}

		### {{ .Schema }}

		{{- range .List }}

		<Accordion title="{{ scenarioTitle . }}">

		**Filters:**

		{{- if .Types }}
		- **Types:** {{ joinKeys .Types }}
		{{- end }}
		{{- if .Tags }}
		- **Tags:** {{ joinKeys .Tags }}
		{{- end }}
		{{- if .ExtKey }}
		- **Extension Key:** ~{{ .ExtKey }}~
		{{- end }}
		{{- if .ExtCode }}
		- **Extension Code:** ~{{ .ExtCode }}~
		{{- end }}
		{{- if .Filter }}
		- **Filter:** *(custom)*
		{{- end }}
		{{- if not .Types }}{{ if not .Tags }}{{ if not .ExtKey }}{{ if not .Filter }}
		- *(none)*
		{{- end }}{{ end }}{{ end }}{{ end }}

		**Output:**

		{{- if .Note }}
		- **Note:** {{ .Note.Text }}{{ if .Note.Key }} ({{ .Note.Key }}){{ end }}
		{{- end }}
		{{- if .Codes }}
		- **Codes:** {{ codeMap .Codes }}
		{{- end }}
		{{- if .Ext }}
		- **Extensions:** {{ extMap .Ext }}
		{{- end }}
		{{- if not .Note }}{{ if not .Codes }}{{ if not .Ext }}
		- *(none)*
		{{- end }}{{ end }}{{ end }}
		</Accordion>
		{{- end }}
		{{- end }}

	`))
}

func (g *addonGenerator) process(doc string) error {
	return g.generator.process(doc, g.addon)
}

package main

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"

	ublterms "github.com/invopop/gobl.ubl/terms"
	"github.com/invopop/gobl/pkg/here"
)

// goblToBTFile is the GOBL → EN16931 Business Term mapping (the semantic leg),
// authored and maintained in this repo. It shares the schema of the converter
// mapping files, so the same parser is reused.
const goblToBTFile = "./terms/en16931.yaml"

// converterMapping describes one converter doc page built by joining the
// GOBL → BT leg with a converter's BT → target leg on the Business Term ID.
type converterMapping struct {
	// Title and Key identify the generated page.
	Title string
	Key   string
	// Repo is the converter's module/repo, surfaced as provenance.
	Repo string
	// TargetName labels the converter's column (e.g. "UBL 2.1").
	TargetName string
	// gobl and target are the two legs to join.
	gobl   *ublterms.Mapping
	target *ublterms.Mapping
}

// mappingRow is one joined Business Term: its GOBL source paths and the
// converter target paths.
type mappingRow struct {
	ID         string
	Name       string
	GOBLPaths  []string
	TargetPath []string
	Notes      string
}

func generateConverterPages() error {
	goblData, err := os.ReadFile(goblToBTFile)
	if err != nil {
		return fmt.Errorf("reading GOBL→BT mapping: %w", err)
	}
	goblBT, err := ublterms.Parse(goblData)
	if err != nil {
		return err
	}

	ublBT, err := ublterms.EN16931UBL()
	if err != nil {
		return err
	}

	cm := &converterMapping{
		Title:      "GOBL to UBL",
		Key:        "ubl",
		Repo:       "github.com/invopop/gobl.ubl",
		TargetName: "UBL 2.1",
		gobl:       goblBT,
		target:     ublBT,
	}

	g := newConverterGenerator(cm)
	if err := g.generate(); err != nil {
		return err
	}
	name := fmt.Sprintf("./converters/%s.mdx", cm.Key)
	if err := os.WriteFile(name, g.bytes(), 0664); err != nil {
		return err
	}
	fmt.Printf("Wrote %s\n", name)
	return nil
}

// rows joins the two legs on Business Term ID, keeping only terms that both
// legs map (an unmapped BT on either side has nothing to document as a
// GOBL→target path). Rows are ordered by BT number.
func (cm *converterMapping) rows() []mappingRow {
	gobl := cm.gobl.Flatten()
	target := cm.target.Flatten()

	rows := make([]mappingRow, 0, len(gobl))
	for id, gt := range gobl {
		tt, ok := target[id]
		if !ok || len(gt.Paths) == 0 || len(tt.Paths) == 0 {
			continue
		}
		rows = append(rows, mappingRow{
			ID:         id,
			Name:       gt.Name,
			GOBLPaths:  gt.Paths,
			TargetPath: tt.Paths,
			Notes:      gt.Notes,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		return btLess(rows[i].ID, rows[j].ID)
	})
	return rows
}

// btLess orders "BT-1" < "BT-2" < "BT-10" numerically, falling back to string
// comparison for anything that doesn't parse.
func btLess(a, b string) bool {
	na, oka := btNum(a)
	nb, okb := btNum(b)
	if oka && okb {
		return na < nb
	}
	return a < b
}

func btNum(id string) (int, bool) {
	_, num, ok := strings.Cut(id, "-")
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(num)
	return n, err == nil
}

type converterGenerator struct {
	generator
	cm *converterMapping
}

func newConverterGenerator(cm *converterMapping) *converterGenerator {
	g := &converterGenerator{
		generator: generator{buf: new(bytes.Buffer)},
		cm:        cm,
	}
	g.tmpl = template.New("base")
	g.tmpl.Funcs(template.FuncMap{
		"pathCell": pathCell,
	})
	return g
}

// pathCell renders one or more paths into a single Markdown table cell,
// each as inline code on its own line.
func pathCell(paths []string) string {
	parts := make([]string, len(paths))
	for i, p := range paths {
		parts[i] = "`" + p + "`"
	}
	return strings.Join(parts, "<br />")
}

func (g *converterGenerator) generate() error {
	return g.process(here.Doc(`
		---
		title: {{.Title}}
		---

		This table maps GOBL invoice fields to their {{.TargetName}} equivalents,
		joined on the [EN16931](/addons/eu-en16931-v2017) Business Term they share.
		The GOBL→Business Term leg is maintained in this repository; the
		Business Term→{{.TargetName}} leg is imported from
		[~{{.Repo}}~](https://{{.Repo}}).

		| Business Term | GOBL | {{.TargetName}} |
		| ------------- | ---- | --------------- |
		{{- range .Rows}}
		| {{.ID}} {{.Name}} | {{pathCell .GOBLPaths}} | {{pathCell .TargetPath}} |
		{{- end}}

	`), struct {
		*converterMapping
		Rows []mappingRow
	}{g.cm, g.cm.rows()})
}

package main

import (
	"fmt"
	"os"
	"strings"

	_ "github.com/invopop/gobl"
	"github.com/invopop/gobl/tax"
)

func main() {
	// Phase 1: Generate draft-0 schema pages
	store, err := newSchemaStore("../gobl/data/schemas")
	if err != nil {
		panic(fmt.Errorf("loading schemas: %w", err))
	}
	buildCalculatedLookup(store)
	if err := generateSchemaPages(store, "./draft-0"); err != nil {
		panic(fmt.Errorf("generating schema pages: %w", err))
	}

	// Phase 2: Generate regime pages
	for _, r := range tax.AllRegimeDefs() {
		g := newRegimeGenerator(r)
		if err := g.generate(); err != nil {
			panic(err)
		}
		name := fmt.Sprintf("./regimes/%s.mdx", strings.ToLower(r.Country.String()))
		if err := os.WriteFile(name, g.bytes(), 0664); err != nil {
			panic(err)
		}
		fmt.Printf("Wrote %s\n", name)
	}

	// Phase 3: Generate addon pages
	for _, ad := range tax.AllAddonDefs() {
		g := newAddonGenerator(ad)
		if err := g.generate(); err != nil {
			panic(err)
		}
		name := fmt.Sprintf("./addons/%s.mdx", ad.Key)
		if err := os.WriteFile(name, g.bytes(), 0664); err != nil {
			panic(err)
		}
		fmt.Printf("Wrote %s\n", name)
	}

	// Phase 4: Generate catalogue pages
	for _, d := range tax.AllCatalogueDefs() {
		g := newCatalogueGenerator(d)
		if err := g.generate(); err != nil {
			panic(err)
		}
		name := fmt.Sprintf("./catalogues/%s.mdx", d.Key)
		if err := os.WriteFile(name, g.bytes(), 0664); err != nil {
			panic(err)
		}
		fmt.Printf("Wrote %s\n", name)
	}

	// Phase 5: Update docs.json navigation
	if err := updateDocsNavigation("./docs.json", "./draft-0"); err != nil {
		panic(fmt.Errorf("updating navigation: %w", err))
	}
	fmt.Println("Updated docs.json navigation")
}

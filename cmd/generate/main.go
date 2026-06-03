package main

import (
	"fmt"
	"os"
	"strings"

	_ "github.com/invopop/gobl"
	"github.com/invopop/gobl/tax"

	// External addon modules. Every addon on gobl's approved external-addon
	// list (tax.ApprovedAddons) must be imported here so its definitions are
	// registered and documented; the guard in main enforces this.
	_ "github.com/invopop/gobl.fr.ctc/addon"
)

func main() {
	// Phase 0: Ensure every approved external addon is loaded. The approved
	// list in gobl core is the manifest of external addons these docs must
	// cover; an entry missing from the runtime registry means a module
	// import is missing above.
	for _, ea := range tax.ApprovedAddons() {
		if tax.AddonForKey(ea.Key) == nil {
			panic(fmt.Errorf("approved external addon %q is not loaded: import its module (%s) in cmd/generate", ea.Key, ea.Module))
		}
	}

	// Phase 1: Generate draft-0 schema pages. Schemas are loaded from the
	// data.Content embedded FS of the pinned gobl module, so no sibling
	// checkout of github.com/invopop/gobl is required.
	//
	// ./draft-0 is wiped first so schemas removed upstream don't linger as
	// orphaned pages. The entire directory is treated as generator-owned;
	// don't hand-edit files under it.
	if err := os.RemoveAll("./draft-0"); err != nil {
		panic(fmt.Errorf("clearing draft-0: %w", err))
	}
	store, err := newSchemaStore()
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

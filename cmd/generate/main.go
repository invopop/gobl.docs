package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/invopop/gobl"
	"github.com/invopop/gobl/tax"

	// Addons are no longer all bundled in GOBL core: approved external addons
	// (e.g. fr-ctc, sa-zatca) live in their own modules and register themselves
	// when imported. The gobl.dev bundle package is the curated list of those
	// modules, so importing it here keeps these docs aligned with the addon set
	// that gobl.dev ships.
	_ "github.com/invopop/gobl.dev/bundle"
)

func main() {
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
	regimePages := map[string]bool{"overview.mdx": true}
	regimeKeys := []string{}
	for _, r := range tax.AllRegimeDefs() {
		g := newRegimeGenerator(r)
		if err := g.generate(); err != nil {
			panic(err)
		}
		key := strings.ToLower(r.Country.String())
		regimeKeys = append(regimeKeys, key)
		regimePages[key+".mdx"] = true
		name := fmt.Sprintf("./regimes/%s.mdx", key)
		if err := os.WriteFile(name, g.bytes(), 0664); err != nil {
			panic(err)
		}
		fmt.Printf("Wrote %s\n", name)
	}
	if err := removeStalePages("./regimes", regimePages); err != nil {
		panic(fmt.Errorf("cleaning regime pages: %w", err))
	}

	// Phase 3: Generate addon pages. Registered addons come from GOBL core
	// plus the external modules imported via the gobl.dev bundle (see imports
	// above). Pages for addons that no longer exist are removed so they don't
	// linger after an addon moves out of core or is retired.
	addonPages := map[string]bool{"overview.mdx": true}
	addonKeys := []string{}
	for _, ad := range tax.AllAddonDefs() {
		addonKeys = append(addonKeys, ad.Key.String())
		g := newAddonGenerator(ad)
		if err := g.generate(); err != nil {
			panic(err)
		}
		base := fmt.Sprintf("%s.mdx", ad.Key)
		addonPages[base] = true
		name := filepath.Join("./addons", base)
		if err := os.WriteFile(name, g.bytes(), 0664); err != nil {
			panic(err)
		}
		fmt.Printf("Wrote %s\n", name)
	}
	if err := removeStalePages("./addons", addonPages); err != nil {
		panic(fmt.Errorf("cleaning addon pages: %w", err))
	}

	// Warn about approved external addons whose module is not imported here;
	// they're valid `$addons` keys but we have no definition to document.
	registered := make(map[string]bool)
	for _, ad := range tax.AllAddonDefs() {
		registered[ad.Key.String()] = true
	}
	for _, ea := range tax.ApprovedAddons() {
		if !registered[ea.Key.String()] {
			fmt.Printf("WARNING: approved addon %q (%s) is not registered; no page generated\n", ea.Key, ea.Module)
		}
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
	if err := updateDocsNavigation("./docs.json", "./draft-0", addonKeys, regimeKeys); err != nil {
		panic(fmt.Errorf("updating navigation: %w", err))
	}
	fmt.Println("Updated docs.json navigation")
}

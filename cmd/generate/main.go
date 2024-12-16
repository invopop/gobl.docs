package main

import (
	"fmt"
	"os"
	"strings"

	_ "github.com/invopop/gobl"
	"github.com/invopop/gobl/tax"
)

func main() {
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
}

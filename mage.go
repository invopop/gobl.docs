//go:build mage
// +build mage

package main

import (
	"github.com/invopop/gobl.docs/internal"
)

// Schema generates the JSON Schema from the base models
func Schema() error {
	return internal.Generate("../gobl/build/schema")
}

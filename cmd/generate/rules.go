package main

import (
	"strings"

	"github.com/invopop/gobl/rules"
	"github.com/invopop/gobl/tax"
)

// calculatedLookup maps "pkg.Type.field" to true for calculated properties.
// Populated once from the schema store in main().
var calculatedLookup map[string]bool

// RuleRow is a single flattened assertion for display in a documentation table.
type RuleRow struct {
	Code       string   // e.g. "GOBL-FR-BILL-INVOICE-01"
	Field      string   // dot-path, e.g. "supplier.tax_id.code", empty = object-level
	Tests      []string // individual guard condition + assertion tests
	Desc       string   // human-readable assertion description
	Calculated bool     // true when the field is auto-populated during calculate
}

// RuleSection groups rows by struct name for a regime/addon page.
type RuleSection struct {
	Name string // e.g. "bill.Invoice"
	Rows []RuleRow
}

// findSetByName scans the global rules registry for a top-level Set with
// the given Name (e.g. "fr" for France, "es-sii-v1" for an addon).
func findSetByName(name string) *rules.Set {
	for _, s := range rules.Registry() {
		if s.Name == name {
			return s
		}
	}
	return nil
}

// ruleSections iterates the direct Subsets of a top-level registered Set
// and produces one RuleSection per struct-level subset that contains rules.
func ruleSections(topSet *rules.Set) []RuleSection {
	var sections []RuleSection
	for _, sub := range topSet.Subsets {
		rows := flattenAssertions(sub, "", "")
		if len(rows) == 0 {
			continue
		}
		sections = append(sections, RuleSection{
			Name: sub.Name,
			Rows: rows,
		})
	}
	annotateCalculated(sections)
	return sections
}

// coreRuleSectionsForStruct returns rule sections for core (unguarded) rules
// that apply to the given struct name (e.g. "bill.Invoice").
// It excludes regime and addon rule sets (which are also unguarded but belong
// to specific regimes/addons).
func coreRuleSectionsForStruct(structName string) []RuleSection {
	exclude := regimeAndAddonNames()
	var sections []RuleSection
	for _, topSet := range rules.Registry() {
		// Core sets have no guard
		if topSet.Guard != nil {
			continue
		}
		// Skip regime and addon sets
		if exclude[topSet.Name] {
			continue
		}
		for _, sub := range topSet.Subsets {
			if sub.Name == structName {
				rows := flattenAssertions(sub, "", "")
				if len(rows) > 0 {
					sections = append(sections, RuleSection{
						Name: sub.Name,
						Rows: rows,
					})
				}
			}
		}
	}
	annotateCalculated(sections)
	return sections
}

// regimeAndAddonNames returns a set of top-level rule set names that
// correspond to regimes or addons (not core rules).
func regimeAndAddonNames() map[string]bool {
	names := make(map[string]bool)
	for _, r := range tax.AllRegimeDefs() {
		names[strings.ToLower(r.Country.String())] = true
	}
	for _, a := range tax.AllAddonDefs() {
		names[a.Key.String()] = true
	}
	return names
}

// flattenAssertions recursively descends a Set tree, accumulating RuleRows.
// field is the dot-path built so far; cond is the most-recently-seen guard
// condition (non-present).
func flattenAssertions(s *rules.Set, field, cond string) []RuleRow {
	// Append field name component.
	if s.FieldName != "" {
		if field == "" {
			field = s.FieldName
		} else {
			field = field + "." + s.FieldName
		}
	}
	// Append [*] for each-element subsets.
	if s.Each {
		field = field + "[*]"
	}
	// Update condition when there is a meaningful guard (not the presentGuard).
	if s.Guard != nil {
		gs := s.Guard.String()
		if gs != "present" {
			cond = gs
		}
	}

	var rows []RuleRow

	// Emit a row for each assertion on this set.
	for _, a := range s.Assert {
		rows = append(rows, RuleRow{
			Code:  string(a.ID),
			Field: field,
			Tests: buildTestParts(cond, a),
			Desc:  a.Desc,
		})
	}

	// Recurse into subsets.
	for _, sub := range s.Subsets {
		rows = append(rows, flattenAssertions(sub, field, cond)...)
	}

	return rows
}

// buildCalculatedLookup iterates all schemas and populates calculatedLookup
// with entries like "bill.Invoice.totals" → true for every property that has
// "calculated": true in the JSON Schema.
func buildCalculatedLookup(store *schemaStore) {
	calculatedLookup = make(map[string]bool)
	for _, schema := range store.schemas {
		if schema.Defs == nil {
			continue
		}
		pkg := schemaIDPackage(schema.ID)
		for rawDefName, def := range schema.Defs {
			if def.Properties == nil {
				continue
			}
			// Strip package prefix (e.g. "bill.Invoice" → "Invoice")
			defName := rawDefName
			if idx := strings.LastIndex(defName, "."); idx >= 0 {
				defName = defName[idx+1:]
			}
			structName := pkg + "." + defName
			for pair := def.Properties.Oldest(); pair != nil; pair = pair.Next() {
				if pair.Value.Calculated {
					calculatedLookup[structName+"."+pair.Key] = true
				}
			}
		}
	}
}

// isCalculatedField checks whether the top-level field in a given struct is calculated.
func isCalculatedField(structName, fieldPath string) bool {
	if fieldPath == "" {
		return false
	}
	topField := strings.SplitN(fieldPath, ".", 2)[0]
	topField = strings.TrimSuffix(topField, "[*]")
	return calculatedLookup[structName+"."+topField]
}

// annotateCalculated fills in the Calculated flag on each row
// using the schema-derived lookup.
func annotateCalculated(sections []RuleSection) {
	for i, sec := range sections {
		for j, row := range sec.Rows {
			sections[i].Rows[j].Calculated = isCalculatedField(sec.Name, row.Field)
		}
	}
}

// buildTestParts returns individual test strings (sentence-cased, pipe-escaped).
func buildTestParts(cond string, a *rules.Assertion) []string {
	var parts []string
	if cond != "" {
		parts = append(parts, cond)
	}
	for _, t := range a.Tests {
		ts := t.String()
		if ts != "" {
			parts = append(parts, ts)
		}
	}
	for i, p := range parts {
		parts[i] = sentenceCase(strings.ReplaceAll(p, "|", "\\|"))
	}
	return parts
}

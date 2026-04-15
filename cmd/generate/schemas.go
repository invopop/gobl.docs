package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// JSONSchema represents a minimal JSON Schema subset covering what GOBL uses.
type JSONSchema struct {
	Schema            string                                      `json:"$schema,omitempty"`
	ID                string                                      `json:"$id,omitempty"`
	Ref               string                                      `json:"$ref,omitempty"`
	Defs              map[string]*JSONSchema                      `json:"$defs,omitempty"`
	Type              string                                      `json:"type,omitempty"`
	Title             string                                      `json:"title,omitempty"`
	Description       string                                      `json:"description,omitempty"`
	Properties        *orderedmap.OrderedMap[string, *JSONSchema]  `json:"properties,omitempty"`
	Required          []string                                    `json:"required,omitempty"`
	Items             *JSONSchema                                 `json:"items,omitempty"`
	OneOf             []*JSONSchema                               `json:"oneOf,omitempty"`
	AnyOf             []*JSONSchema                               `json:"anyOf,omitempty"`
	Const             any                                         `json:"const,omitempty"`
	Pattern           string                                      `json:"pattern,omitempty"`
	Format            string                                      `json:"format,omitempty"`
	PatternProperties map[string]*JSONSchema                      `json:"patternProperties,omitempty"`
	Recommended       []string                                    `json:"recommended,omitempty"`
	Calculated        bool                                        `json:"calculated,omitempty"`
	If                *JSONSchema                                 `json:"if,omitempty"`
	Then              *JSONSchema                                 `json:"then,omitempty"`
}

// schemaStore holds all loaded JSON schemas keyed by $id URL.
type schemaStore struct {
	schemas map[string]*JSONSchema // keyed by $id
	baseDir string
}

func newSchemaStore(baseDir string) (*schemaStore, error) {
	store := &schemaStore{
		schemas: make(map[string]*JSONSchema),
		baseDir: baseDir,
	}
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var s JSONSchema
		if err := json.Unmarshal(data, &s); err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}
		if s.ID != "" {
			store.schemas[s.ID] = &s
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return store, nil
}

// schemaGenerator produces a single draft-0 MDX page for one $defs entry.
type schemaGenerator struct {
	store      *schemaStore
	schemaID   string      // full $id of the parent file
	defName    string      // name inside $defs, e.g. "Invoice"
	def        *JSONSchema // the resolved definition
	parentFile *JSONSchema // root schema file
	isPrimary  bool        // true when file's $ref points to this def
}

// generateSchemaPages iterates all schemas and writes draft-0 MDX pages.
func generateSchemaPages(store *schemaStore, outDir string) error {
	for _, schema := range store.schemas {
		if schema.Defs == nil {
			continue
		}
		// Determine which def is the primary (referenced by $ref)
		primaryDef := refDefName(schema.Ref)

		for rawDefName, def := range schema.Defs {
			// Strip package prefix (e.g. "bill.Invoice" → "Invoice")
			defName := rawDefName
			if idx := strings.LastIndex(defName, "."); idx >= 0 {
				defName = defName[idx+1:]
			}
			sg := &schemaGenerator{
				store:      store,
				schemaID:   schema.ID,
				defName:    defName,
				def:        def,
				parentFile: schema,
				isPrimary:  defName == primaryDef,
			}
			content := sg.generate()

			// Determine output path
			relPath := schemaIDToFilePath(schema.ID, defName, defName == primaryDef)
			outPath := filepath.Join(outDir, relPath+".mdx")

			if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
				return err
			}
			if err := os.WriteFile(outPath, []byte(content), 0664); err != nil {
				return err
			}
			fmt.Printf("Wrote %s\n", outPath)
		}
	}
	return nil
}

// generate produces the MDX content for a single type.
func (sg *schemaGenerator) generate() string {
	var buf bytes.Buffer

	// Frontmatter
	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("title: %s\n", sg.defName))
	buf.WriteString("comment: \n")
	buf.WriteString("---\n")

	// Description
	desc := sg.def.Description
	if desc != "" {
		buf.WriteString("\n")
		// Replace newlines with spaces for inline description
		buf.WriteString(cleanDescription(desc))
		buf.WriteString("\n")
	}

	// Determine type category
	isObject := sg.def.Type == "object" && sg.def.Properties != nil && sg.def.Properties.Len() > 0
	isArray := sg.def.Type == "array" && sg.def.Items != nil
	isMap := sg.def.Type == "object" && sg.def.Properties == nil && len(sg.def.PatternProperties) > 0
	isEnum := len(sg.def.OneOf) > 0 && hasConstEntries(sg.def.OneOf)

	// Array - show "An array of [Type](...)"
	if isArray {
		buf.WriteString("\n")
		buf.WriteString(fmt.Sprintf("An array of %s\n", sg.typeString(sg.def.Items)))
	}

	// Schema ID (skip for map types)
	if !isMap {
		buf.WriteString("\n## Schema ID\n\n")
		buf.WriteString(fmt.Sprintf("`%s`\n", sg.fullSchemaID()))
		buf.WriteString("\n")
	}

	// Properties table
	if isObject {
		sg.writePropertiesTable(&buf)
	}

	// Enum values for top-level type (like org.Unit)
	if isEnum {
		sg.writeTopLevelEnumTable(&buf, sg.def.OneOf)
	}

	// Property enum tables
	if isObject {
		sg.writePropertyEnumTables(&buf)
	}

	// Core validation rules
	coreSections := coreRuleSectionsForStruct(sg.structDotName())
	if len(coreSections) > 0 {
		sg.writeValidationRules(&buf, coreSections)
	}

	return buf.String()
}

// fullSchemaID returns the complete schema ID for this definition.
func (sg *schemaGenerator) fullSchemaID() string {
	if sg.isPrimary {
		return sg.schemaID
	}
	return sg.schemaID + "#/$defs/" + sg.defName
}

// structDotName returns the pkg.TypeName notation, e.g. "bill.Invoice".
func (sg *schemaGenerator) structDotName() string {
	pkg := schemaIDPackage(sg.schemaID)
	return pkg + "." + sg.defName
}

// writePropertiesTable writes the properties markdown table.
func (sg *schemaGenerator) writePropertiesTable(buf *bytes.Buffer) {
	buf.WriteString("## Properties\n\n")
	buf.WriteString("| Title | Property | Type | Description |\n")
	buf.WriteString("|-------|----------|------|-------------|\n")

	for pair := sg.def.Properties.Oldest(); pair != nil; pair = pair.Next() {
		propName := pair.Key
		prop := pair.Value

		title := prop.Title
		typeStr := sg.typeString(prop)
		desc := cleanDescription(prop.Description)

		buf.WriteString(fmt.Sprintf("| %s | `%s` | %s | %s |\n",
			title, propName, typeStr, desc))
	}
	buf.WriteString("\n")
}

// writePropertyEnumTables writes enum value tables for properties that have oneOf/anyOf with const entries.
func (sg *schemaGenerator) writePropertyEnumTables(buf *bytes.Buffer) {
	for pair := sg.def.Properties.Oldest(); pair != nil; pair = pair.Next() {
		prop := pair.Value
		entries := constEntries(prop.OneOf)
		if len(entries) == 0 {
			entries = constEntries(prop.AnyOf)
		}
		if len(entries) == 0 {
			continue
		}

		title := prop.Title
		if title == "" {
			title = pair.Key
		}
		buf.WriteString(fmt.Sprintf("## %s Values\n\n", title))
		buf.WriteString("| Value | Description |\n")
		buf.WriteString("|-------|-------------|\n")
		for _, e := range entries {
			val := fmt.Sprintf("%v", e.Const)
			desc := e.Description
			if desc == "" {
				desc = e.Title
			}
			buf.WriteString(fmt.Sprintf("| `%s` | %s |\n", val, cleanDescription(desc)))
		}
		buf.WriteString("\n")
	}
}

// writeTopLevelEnumTable writes the values table for a type that is itself an enum (like org.Unit).
func (sg *schemaGenerator) writeTopLevelEnumTable(buf *bytes.Buffer, entries []*JSONSchema) {
	filtered := constEntries(entries)
	if len(filtered) == 0 {
		return
	}

	// Check if any entries have descriptions (separate from title)
	hasDescs := false
	for _, e := range filtered {
		if e.Description != "" {
			hasDescs = true
			break
		}
	}

	buf.WriteString("## Values\n\n")
	if hasDescs {
		buf.WriteString("| Value | Title | Description |\n")
		buf.WriteString("|-------|-------|-------------|\n")
		for _, e := range filtered {
			val := fmt.Sprintf("%v", e.Const)
			buf.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n",
				val, e.Title, cleanDescription(e.Description)))
		}
	} else {
		buf.WriteString("| Value | Title |\n")
		buf.WriteString("|-------|-------|\n")
		for _, e := range filtered {
			val := fmt.Sprintf("%v", e.Const)
			buf.WriteString(fmt.Sprintf("| `%s` | %s |\n", val, e.Title))
		}
	}
	buf.WriteString("\n")
}

// writeValidationRules writes the validation rules section.
// Since this is the struct's own page, the table is rendered directly
// without an accordion wrapper.
func (sg *schemaGenerator) writeValidationRules(buf *bytes.Buffer, sections []RuleSection) {
	tmpl := template.Must(template.New("rules").Funcs(template.FuncMap{
		"codeMessage": codeMessage,
		"testList":    testList,
		"fieldCell":   fieldCell,
	}).Parse(`
## Validation Rules

| Field | Test | Validation Code / Message |
| ----- | ---- | ------------------------- |
{{- range .}}
{{- $sec := .}}
{{- range .Rows}}
| {{fieldCell .Field .Calculated $sec.Name}} | {{testList .Tests .Calculated}} | {{codeMessage .Code .Desc}} |
{{- end}}
{{- end}}
`))
	tmpl.Execute(buf, sections)
}

// typeString converts a JSON Schema property to its display type string.
func (sg *schemaGenerator) typeString(prop *JSONSchema) string {
	if prop == nil {
		return ""
	}

	// $ref to another type
	if prop.Ref != "" {
		return sg.refTypeString(prop.Ref)
	}

	// Array type
	if prop.Type == "array" && prop.Items != nil {
		itemType := sg.typeString(prop.Items)
		return "array of " + itemType
	}

	// Basic types
	switch prop.Type {
	case "string":
		return "string"
	case "integer":
		return "integer"
	case "number":
		return "number"
	case "boolean":
		return "boolean"
	case "object":
		return "object"
	}

	return ""
}

// refTypeString converts a $ref URL to a linked type string.
func (sg *schemaGenerator) refTypeString(ref string) string {
	// Local reference: #/$defs/Name or #/$defs/pkg.Name
	if strings.HasPrefix(ref, "#/$defs/") {
		defName := ref[len("#/$defs/"):]
		if idx := strings.LastIndex(defName, "."); idx >= 0 {
			defName = defName[idx+1:]
		}
		pkg := schemaIDPackage(sg.schemaID)
		path := schemaIDToDocPath(sg.schemaID, defName, false)
		return fmt.Sprintf("[%s.%s](%s)", pkg, defName, path)
	}

	// External GOBL reference
	if strings.HasPrefix(ref, "https://gobl.org/") {
		// May contain an anchor: https://gobl.org/draft-0/tax/set#Combo
		baseRef := ref
		defName := ""
		if idx := strings.Index(ref, "#"); idx >= 0 {
			baseRef = ref[:idx]
			defName = ref[idx+1:]
		}

		// Look up the schema to find the primary def name
		if defName == "" {
			if s, ok := sg.store.schemas[baseRef]; ok {
				defName = refDefName(s.Ref)
			}
		}

		displayName := refDisplayName(baseRef, defName)
		path := refToDocPath(baseRef, defName)
		return fmt.Sprintf("[%s](%s)", displayName, path)
	}

	return ref
}

// refDefName extracts the type name from "#/$defs/Invoice" or "#/$defs/bill.Invoice",
// stripping any package prefix.
func refDefName(ref string) string {
	if strings.HasPrefix(ref, "#/$defs/") {
		name := ref[len("#/$defs/"):]
		if idx := strings.LastIndex(name, "."); idx >= 0 {
			name = name[idx+1:]
		}
		return name
	}
	return ""
}

// schemaIDPackage extracts the package name from a schema ID.
// e.g. "https://gobl.org/draft-0/bill/invoice" → "bill"
// e.g. "https://gobl.org/draft-0/regimes/mx/food-vouchers" → "regimes.mx"
func schemaIDPackage(id string) string {
	path := strings.TrimPrefix(id, "https://gobl.org/draft-0/")
	parts := strings.Split(path, "/")
	if len(parts) <= 1 {
		return ""
	}
	// Package is everything except the last segment
	return strings.Join(parts[:len(parts)-1], ".")
}

// schemaIDToFilePath converts a schema ID and def name to a file path relative to draft-0/.
// e.g. ("https://gobl.org/draft-0/bill/invoice", "Invoice", true) → "bill/invoice"
// e.g. ("https://gobl.org/draft-0/bill/line", "LineCharge", false) → "bill/line_charge"
func schemaIDToFilePath(id string, defName string, isPrimary bool) string {
	path := strings.TrimPrefix(id, "https://gobl.org/draft-0/")
	if isPrimary {
		// Primary def uses the file path directly, converting hyphens to underscores
		return strings.ReplaceAll(path, "-", "_")
	}
	// Non-primary def: use the package path + snake_case def name
	parts := strings.Split(path, "/")
	pkg := strings.Join(parts[:len(parts)-1], "/")
	return pkg + "/" + toSnakeCase(defName)
}

// schemaIDToDocPath converts a schema ID + def name to a doc path for linking.
func schemaIDToDocPath(id string, defName string, isPrimary bool) string {
	return "/draft-0/" + schemaIDToFilePath(id, defName, isPrimary)
}

// refToDocPath converts an external $ref URL to a doc path.
func refToDocPath(baseRef string, defName string) string {
	path := strings.TrimPrefix(baseRef, "https://gobl.org/draft-0/")
	path = strings.ReplaceAll(path, "-", "_")

	if defName != "" {
		// Non-primary: use snake_case of def name
		parts := strings.Split(path, "/")
		pkg := strings.Join(parts[:len(parts)-1], "/")
		if pkg != "" {
			return "/draft-0/" + pkg + "/" + toSnakeCase(defName)
		}
		return "/draft-0/" + toSnakeCase(defName)
	}

	return "/draft-0/" + path
}

// refDisplayName creates a display name for a $ref link.
// e.g. "https://gobl.org/draft-0/org/party" + "" → "org.Party"
// e.g. "https://gobl.org/draft-0/bill/line" + "LineCharge" → "bill.LineCharge"
func refDisplayName(baseRef string, defName string) string {
	path := strings.TrimPrefix(baseRef, "https://gobl.org/draft-0/")
	parts := strings.Split(path, "/")

	pkg := ""
	if len(parts) > 1 {
		pkg = strings.Join(parts[:len(parts)-1], ".")
	}

	typeName := defName
	if typeName == "" {
		// Use the last path segment converted to PascalCase
		typeName = hyphenToPascalCase(parts[len(parts)-1])
	}

	if pkg != "" {
		return pkg + "." + typeName
	}
	return typeName
}

// constEntries filters oneOf/anyOf entries to only those with a const value (no pattern-only entries).
func constEntries(entries []*JSONSchema) []*JSONSchema {
	var result []*JSONSchema
	for _, e := range entries {
		if e.Const != nil {
			result = append(result, e)
		}
	}
	return result
}

// hasConstEntries checks if any entries have const values.
func hasConstEntries(entries []*JSONSchema) bool {
	for _, e := range entries {
		if e.Const != nil {
			return true
		}
	}
	return false
}

// cleanDescription normalizes a description string for single-line table display.
func cleanDescription(desc string) string {
	// Replace newlines with spaces
	desc = strings.ReplaceAll(desc, "\n", " ")
	// Collapse multiple spaces
	for strings.Contains(desc, "  ") {
		desc = strings.ReplaceAll(desc, "  ", " ")
	}
	return strings.TrimSpace(desc)
}

// toSnakeCase converts PascalCase to snake_case.
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(r + 32) // toLower
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// hyphenToPascalCase converts "payment-details" to "PaymentDetails".
func hyphenToPascalCase(s string) string {
	parts := strings.Split(s, "-")
	var result strings.Builder
	for _, p := range parts {
		if len(p) > 0 {
			result.WriteString(strings.ToUpper(p[:1]) + p[1:])
		}
	}
	return result.String()
}

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// groupNames maps directory names to display names for navigation.
var groupNames = map[string]string{
	"bill":     "Bill",
	"cal":      "Cal",
	"cbc":      "CBC",
	"currency": "Currency",
	"dsig":     "DSig",
	"head":     "Head",
	"i18n":     "I18n",
	"l10n":     "L10n",
	"note":     "Note",
	"num":      "Num",
	"org":      "Org",
	"pay":      "Pay",
	"regimes":  "Addons",
	"schema":   "Schema",
	"tax":      "Tax",
}

// updateDocsNavigation reads docs.json, rebuilds the Schemas (Draft 0)
// navigation group from the generated draft-0/ pages, and writes it back.
func updateDocsNavigation(docsJSONPath, draft0Dir string) error {
	data, err := os.ReadFile(docsJSONPath)
	if err != nil {
		return err
	}

	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return err
	}

	// Build the new draft-0 navigation
	nav, err := buildDraft0Nav(draft0Dir)
	if err != nil {
		return err
	}

	// Find and replace the "Schemas (Draft 0)" group in the navigation
	if err := replaceSchemasGroup(doc, nav); err != nil {
		return err
	}

	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	// Ensure trailing newline
	out = append(out, '\n')
	return os.WriteFile(docsJSONPath, out, 0664)
}

// buildDraft0Nav walks draft-0/ and builds the navigation structure.
func buildDraft0Nav(draft0Dir string) (any, error) {
	// Collect all MDX files grouped by directory
	topLevel := []string{}
	groups := make(map[string][]string)     // dir → sorted page paths
	subGroups := make(map[string]map[string][]string) // dir → subdir → sorted page paths

	err := filepath.Walk(draft0Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".mdx" {
			return nil
		}

		rel, _ := filepath.Rel(filepath.Dir(draft0Dir), path)
		// Remove .mdx extension for the page path
		pagePath := strings.TrimSuffix(rel, ".mdx")

		// Determine depth within draft-0/
		inner := strings.TrimPrefix(pagePath, "draft-0/")
		parts := strings.Split(inner, "/")

		switch len(parts) {
		case 1:
			// Top-level pages (e.g. draft-0/envelope)
			// Skip the index page (draft-0 itself is handled separately)
			topLevel = append(topLevel, pagePath)
		case 2:
			// Package pages (e.g. draft-0/bill/invoice)
			dir := parts[0]
			groups[dir] = append(groups[dir], pagePath)
		case 3:
			// Sub-package pages (e.g. draft-0/regimes/mx/food_vouchers)
			dir := parts[0]
			subDir := parts[1]
			if subGroups[dir] == nil {
				subGroups[dir] = make(map[string][]string)
			}
			subGroups[dir][subDir] = append(subGroups[dir][subDir], pagePath)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Sort everything
	sort.Strings(topLevel)
	for k := range groups {
		sort.Strings(groups[k])
	}
	for k := range subGroups {
		for sk := range subGroups[k] {
			sort.Strings(subGroups[k][sk])
		}
	}

	// Build the pages array
	pages := []any{"draft-0"}
	for _, p := range topLevel {
		pages = append(pages, p)
	}

	// Sort group keys for consistent output
	dirKeys := make([]string, 0, len(groups))
	for k := range groups {
		dirKeys = append(dirKeys, k)
	}
	// Also include dirs that only appear in subGroups
	for k := range subGroups {
		if _, exists := groups[k]; !exists {
			dirKeys = append(dirKeys, k)
		}
	}
	sort.Strings(dirKeys)

	for _, dir := range dirKeys {
		displayName := groupNames[dir]
		if displayName == "" {
			displayName = strings.ToUpper(dir[:1]) + dir[1:]
		}

		groupPages := make([]any, 0)
		for _, p := range groups[dir] {
			groupPages = append(groupPages, p)
		}

		// Add sub-groups
		if subs, ok := subGroups[dir]; ok {
			subKeys := make([]string, 0, len(subs))
			for k := range subs {
				subKeys = append(subKeys, k)
			}
			sort.Strings(subKeys)
			for _, subDir := range subKeys {
				subDisplayName := strings.ToUpper(subDir)
				subPages := make([]any, 0)
				for _, p := range subs[subDir] {
					subPages = append(subPages, p)
				}
				groupPages = append(groupPages, map[string]any{
					"group": subDisplayName,
					"pages": subPages,
				})
			}
		}

		pages = append(pages, map[string]any{
			"group": displayName,
			"pages": groupPages,
		})
	}

	return map[string]any{
		"group": "Schemas (Draft 0)",
		"pages": pages,
	}, nil
}

// replaceSchemasGroup finds and replaces the "Schemas (Draft 0)" group in docs.json.
func replaceSchemasGroup(doc map[string]any, newGroup any) error {
	nav, ok := doc["navigation"].(map[string]any)
	if !ok {
		return fmt.Errorf("navigation not found in docs.json")
	}

	anchors, ok := nav["anchors"].([]any)
	if !ok {
		return fmt.Errorf("anchors not found in navigation")
	}

	for _, anchor := range anchors {
		a, ok := anchor.(map[string]any)
		if !ok {
			continue
		}
		groups, ok := a["groups"].([]any)
		if !ok {
			continue
		}
		for i, group := range groups {
			g, ok := group.(map[string]any)
			if !ok {
				continue
			}
			if g["group"] == "Schemas (Draft 0)" {
				groups[i] = newGroup
				return nil
			}
		}
	}
	return fmt.Errorf("Schemas (Draft 0) group not found in docs.json")
}

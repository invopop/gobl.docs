---
name: gobl
description: Use when writing, validating, correcting, signing or converting GOBL (Go Business Language) documents such as invoices, credit notes, parties and orders, or when you need a country's tax categories, rate keys, tags, extension codes or correction types. Covers the build-first workflow, the rules agents most often get wrong (string amounts and the percent sign, regime inference, lowercase keys versus uppercase codes, addons and extensions, tags, dates, corrections, sealed envelopes) and where the authoritative Markdown pages are for every schema, regime and addon.
license: Apache-2.0
metadata:
  version: "2.0"
  source: https://github.com/invopop/gobl.docs
  canonical: https://docs.gobl.org/AGENTS.md
---

# GOBL

GOBL is an open-source JSON format and Go library for business documents with the tax rules of 27 countries built in. You write only the facts of a document; `build` normalises them, calculates every derived value for the document's regime (country) and addons (local formats), and validates the result with coded rules. Built documents are wrapped in a signed envelope and converted to local formats by companion libraries. Invopop uses GOBL as its document format.

## When to use

- Turning application data into a `bill/invoice`, `org/party` or other GOBL document.
- Finding what a country requires: tax categories, rate keys, tags, extension codes, correction types.
- Diagnosing a build or validation error by its fault code.
- Producing a credit note or corrective invoice from an existing document.
- Building, signing or verifying envelopes from the CLI, the REST API, the MCP server or Go.

## Read first

1. https://docs.gobl.org/AGENTS.md - the condensed rules. Keep it in context.
2. https://docs.gobl.org/llms.md - the same rules with worked examples and prompts.
3. https://docs.gobl.org/llms.txt - the index of every page. Fetch pages as Markdown by appending `.md`.
4. For the country at hand: `https://docs.gobl.org/regimes/<code>.md` and the relevant `https://docs.gobl.org/addons/<key>.md`.

## Working method

1. Identify the supplier's tax country and any local format required. Read the regime page for rate keys, tags and correction types, and the addon page for extension codes. Or call the MCP tools `regime` and `addon`.
2. Write the minimal document: `$schema`, `$regime`, `$addons` if needed, `series`, `code` if you own numbering, `issue_date` as `YYYY-MM-DD`, `supplier` and `customer` with `tax_id`, `lines` with `quantity`, `item.name`, `item.price` as strings and `taxes` as an uppercase `cat` plus a lowercase `rate` key. No sums, totals, currency, uuid or percentages.
3. Build it: `POST https://gobl.dev/v0/build` with `{"data": <document>}` (no auth), or `gobl build -i doc.json` (CLI from `go install github.com/invopop/gobl.dev/cmd/gobl@latest`), or the MCP `build` tool, or `gobl.NewEnvelope().Insert(doc)` in Go.
4. Read every fault by `code` and `paths`; look the code up on the schema, regime or addon page named by its prefix; fix; rebuild.
5. For corrections use `gobl correct --credit`, `POST /v0/correct` or the MCP `correct` tool. Never write `preceding` by hand.
6. To seal: `gobl build -e` for an envelope, `gobl keygen` then `gobl sign`, `gobl verify` to check. Never edit a signed envelope; correct it.

## Rules that prevent the common failures

- Let build calculate. Hand-written `totals` are silently replaced.
- Amounts and percentages are strings. A percentage without `%` is read as a factor: `"21"` becomes `2100%` with no error.
- Set `$regime` explicitly. Without it and without a supplier `tax_id` the error is `currency: missing or invalid`. Greece is `EL`.
- `{"cat": "VAT", "rate": "standard"}` resolves per regime and `issue_date` to `key: standard, rate: general, percent: ...`. Unknown rate keys fail. Categories are uppercase and regime specific (`VAT`, `ST`, `IVA`, `GST`, `IRPF`).
- Keys are lowercase hyphenated (`credit-note`, `reverse-charge`); codes stay as given. `"cat": "vat"` fails.
- Never invent extension codes; take them from the addon page.
- Only tags the regime or addon defines are allowed (`simplified`, `reverse-charge`, `self-billed`, ...). Omitting the customer under an addon usually needs `$tags: ["simplified"]`.
- Dates are `YYYY-MM-DD`.
- Tax IDs are normalised and checksum-validated; use real identities when testing.
- `code` is optional to build but required to sign.
- `validate` is not `build`: validating a partial document fails.
- Editing an enveloped or signed document without rebuilding fails with `GOBL-ENVELOPE-11`.

## Reference

- Invoice schema: https://docs.gobl.org/draft-0/bill/invoice.md
- Party schema: https://docs.gobl.org/draft-0/org/party.md
- Tax combo: https://docs.gobl.org/draft-0/tax/combo.md
- Validation rules and fault codes: https://docs.gobl.org/overview/validation.md
- Numbers and rounding: https://docs.gobl.org/overview/numbers.md, https://docs.gobl.org/overview/rounding.md
- API: https://docs.gobl.org/api/introduction.md, MCP: https://docs.gobl.org/api/mcp.md
- Sending documents to authorities and networks with Invopop: https://docs.invopop.com/llms.md

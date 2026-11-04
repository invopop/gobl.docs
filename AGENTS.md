# GOBL for AI agents

This file is published at https://docs.gobl.org/AGENTS.md. It is the condensed, copy-into-your-context version of https://docs.gobl.org/llms.md for agents that write, validate or convert GOBL documents. If you are editing the docs.gobl.org repository itself, see its README.

## What GOBL is

GOBL (Go Business Language, https://gobl.org) is an open-source format and library for business documents: invoices, credit notes, orders, deliveries, payments, parties and items. A document is JSON (YAML is accepted on input). You write the raw facts; **build** normalises the data, calculates totals and taxes from the country's **regime** and any **addons** (local formats such as VERI*FACTU, CFDI, FatturaPA, EN 16931), and validates the result with coded rules. A built document can be wrapped in an **envelope** (header with UUID and digest, plus signatures) and converted to local XML or PDF by companion libraries. Invopop (https://docs.invopop.com/llms.md) uses GOBL as its document format.

- Docs: any page as Markdown by appending `.md`; index at https://docs.gobl.org/llms.txt.
- Public API, no auth: `https://gobl.dev/v0` (`POST /build`, `/validate`, `/correct`, `/replicate`, `/sign`, `/verify`; `GET /regimes`, `/regimes/{code}`, `/addons`, `/addons/{key}`, `/schemas`).
- MCP: hosted `https://gobl.dev/v0/mcp` or local `gobl mcp`; tools `build`, `validate`, `correct`, `replicate`, `schema`, `regime`, `regime_list`, `addon`, `addon_list`.
- CLI: `go install github.com/invopop/gobl.dev/cmd/gobl@latest` (the CLI lives in gobl.dev, not the core library; binaries at https://github.com/invopop/gobl.dev/releases).
- Go: `go get github.com/invopop/gobl`; `env := gobl.NewEnvelope(); err := env.Insert(inv)` calculates the document; call `env.Validate()` afterwards.
- Skill: `npx skills add https://docs.gobl.org` installs https://docs.gobl.org/skill.md.

## Mental model

1. **Document**: JSON whose `$schema` names its type, `https://gobl.org/draft-0/bill/invoice`, `.../org/party`, `.../bill/order`, `.../bill/delivery`, `.../bill/payment`, `.../bill/status`, `.../org/item`, `.../note/message`.
2. **Build**: normalise, calculate, validate. Fills `$regime`, `type`, `currency`, `uuid`, line `sum`/`total`, tax `key`/`rate`/`percent`, `totals`, and addon extension defaults. Idempotent.
3. **Envelope**: `{"$schema": "https://gobl.org/draft-0/envelope", "head": {uuid, dig, stamps, tags, meta}, "doc": <document>, "sigs": [...]}`. The digest seals the document; signatures cover the header.
4. **Regime** (`$regime`, a country code): tax categories, rate keys with percentages by date, tags, correction types, validation rules. Page: `https://docs.gobl.org/regimes/<code>.md`.
5. **Addon** (`$addons`, a list of keys): rules and **extensions** (`ext` code maps) for one local format. Page: `https://docs.gobl.org/addons/<key>.md`.
6. **Tag** (`$tags`): a scenario switch defined by the regime or addon (`simplified`, `reverse-charge`, `self-billed`, ...).
7. **Keys vs codes**: keys are GOBL's own lowercase hyphenated identifiers (`standard`, `credit-note`, `credit-transfer`); codes come from the outside world and are kept as given (`VAT`, `B98602642`, `SAMPLE`).

## Rules

1. **Send the minimum, let build calculate.** Never write `sum`, `total`, `totals`, `currency`, `uuid` or tax `percent` (unless the regime has no rates). Hand-written `totals` are silently replaced, never checked.
2. **Numbers are strings** with meaningful precision: `"100.00"`, `"10%"`, `"21.0%"`. JSON numbers are accepted and returned as strings.
3. **Percentages need the `%` sign.** `"percent": "21"` is read as a factor and becomes `2100%` without any error. Write `"21%"`.
4. **`$schema` is mandatory**; a document without it fails with `unknown-schema`. The envelope has its own `$schema` and the document's goes inside `doc`.
5. **Set `$regime` explicitly** to the supplier's tax country. It is inferred from `supplier.tax_id.country`, but with neither present the error is `currency: missing or invalid`. `$regime` wins over the supplier's country. Regimes: `AE AR AT BE BR CA CH CO DE DK EL ES FI FR GB IE IN IT MX NL NO PL PT SA SE SG US` (`EL` is Greece; a `GR` tax identity is normalised to `EL`). Countries without a regime need explicit `currency` and tax `percent` values.
6. **Taxes are `cat` + rate key.** `{"cat": "VAT", "rate": "standard"}` → `key: standard, rate: general, percent: 21.0%` for the regime and `issue_date` (`standard` and `general` are aliases). Other rates: `reduced`, `super-reduced`, `intermediate`, `zero`. Keys without a percent: `exempt`, `reverse-charge`. Unknown keys fail (`'bogus' rate not defined for key 'standard'`). Category codes differ by regime (`VAT`, `ST`, `IVA`, `GST`, plus `IGIC`, `IRPF` in Spain); read them from the regime page. Retained taxes (for example `IRPF`) lower `totals.payable` below `totals.total_with_tax`. `tax.prices_include: "VAT"` treats prices as tax-inclusive.
7. **Keys are lowercase, codes are uppercase as given.** `"cat": "vat"` → `invalid-category: 'vat' not defined in regime`; `"rate": "Standard"` fails too.
8. **`$addons` selects a local format**; build fills the extension defaults it can (VERI*FACTU alone yields `es-verifactu-doc-type: F1` and per-tax `es-verifactu-op-class`/`es-verifactu-regime`), and reports the rest as faults naming the extension key. Never invent an extension value; copy it from the addon page, the MCP `addon` tool or `GET /v0/addons/<key>`. Addons: `ar-arca-v4 br-nfe-v4 br-nfse-v1 co-dian-v2 de-xrechnung-v3 de-zugferd-v2 dk-oioubl-v2 es-facturae-v3 es-sii-v1 es-tbai-v1 es-verifactu-v1 eu-en16931-v2017 fi-finvoice-v3 fr-choruspro-v1 fr-ctc-flow2-v1 fr-ctc-flow6-v1 fr-ctc-flow10-v1 fr-facturx-v1 gr-mydata-v1 it-sdi-v1 it-ticket-v1 mx-cfdi-v4 pl-favat-v3 pt-saft-v1 sa-zatca-v1`.
9. **Only defined tags are allowed**: `$tags: ["bogus"]` fails with `$tags: 'bogus' undefined`. `simplified` lets you omit the customer (under VERI*FACTU the customer is otherwise required, `GOBL-ES-VERIFACTU-BILL-INVOICE-06`, and `simplified` sets doc type `F2`); `reverse-charge` adds the regime's legal note under `tax.notes`. Legacy `tax.tags` is moved to `$tags`.
10. **Dates are `YYYY-MM-DD`.** Anything else is an input error. `issue_date` defaults to today and selects the tax rates (a 2012 Spanish invoice gets 18%).
11. **`series` + `code` is the document number.** Optional at build, `code` required to sign. Tax IDs are normalised (`ESB23103039` → `B23103039`, `es-b-986.026.42` → `B98602642`) and checksum-validated (`GOBL-GB-TAX-IDENTITY-03`). Non-tax identifiers go in `identities`.
12. **Never edit an issued invoice; correct it.** `type` `credit-note` (extends), `debit-note` (adds charges) or `corrective` (replaces), with the original in `preceding`; amounts stay positive. Allowed types are on the regime page (Spain: all three; UK: `credit-note`). Use `gobl correct -i --credit built.json`, `POST /v0/correct` or the MCP `correct` tool; never build `preceding` by hand. `gobl correct --options` shows the allowed options (`type`, `series`, `issue_date`, `reason`, `copy_tax`, `stamps`, addon extensions).
13. **Envelopes are sealed.** `gobl build -e` adds `head.uuid` and `head.dig`; `gobl sign` needs a key from `gobl keygen` and a `code` on the document; `gobl verify` checks. Editing a signed or enveloped document without rebuilding fails with `GOBL-ENVELOPE-11 envelope digest does not match document contents`.
14. **`validate` is not `build`.** Validating a partial document fails; build first.
15. **Empty `lines` fail** (`GOBL-BILL-INVOICE-10`) unless the document carries discounts or charges. Line and document `discounts`/`charges` take `percent` or `amount`.

## Minimal examples

Invoice input (Spain), builds to `payable: 1089.00`:

```json
{
  "$schema": "https://gobl.org/draft-0/bill/invoice",
  "$regime": "ES",
  "series": "SAMPLE",
  "code": "001",
  "issue_date": "2026-09-03",
  "supplier": {
    "name": "Provider One S.L.",
    "tax_id": { "country": "ES", "code": "B98602642" }
  },
  "customer": {
    "name": "Sample Consumer S.L.",
    "tax_id": { "country": "ES", "code": "B63272603" }
  },
  "lines": [
    {
      "quantity": "10",
      "item": { "name": "Development services", "price": "100.00" },
      "discounts": [{ "percent": "10%" }],
      "taxes": [{ "cat": "VAT", "rate": "standard" }]
    }
  ]
}
```

Same invoice for VERI*FACTU (build adds `tax.ext` and per-tax `ext` codes automatically):

```json
{
  "$schema": "https://gobl.org/draft-0/bill/invoice",
  "$regime": "ES",
  "$addons": ["es-verifactu-v1"],
  "series": "SAMPLE",
  "code": "002",
  "issue_date": "2026-09-03",
  "supplier": {
    "name": "Provider One S.L.",
    "tax_id": { "country": "ES", "code": "B98602642" }
  },
  "customer": {
    "name": "Sample Consumer S.L.",
    "tax_id": { "country": "ES", "code": "B63272603" }
  },
  "lines": [
    {
      "quantity": "10",
      "item": { "name": "Development services", "price": "100.00" },
      "taxes": [{ "cat": "VAT", "rate": "general" }]
    }
  ]
}
```

Simplified B2C invoice (no customer) under VERI*FACTU:

```json
{
  "$schema": "https://gobl.org/draft-0/bill/invoice",
  "$regime": "ES",
  "$addons": ["es-verifactu-v1"],
  "$tags": ["simplified"],
  "series": "SIMP",
  "code": "001",
  "issue_date": "2026-09-03",
  "supplier": {
    "name": "Provider One S.L.",
    "tax_id": { "country": "ES", "code": "B98602642" }
  },
  "lines": [
    {
      "quantity": "1",
      "item": { "name": "Coffee", "price": "2.50" },
      "taxes": [{ "cat": "VAT", "rate": "reduced" }]
    }
  ]
}
```

Party:

```json
{
  "$schema": "https://gobl.org/draft-0/org/party",
  "name": "Provider One S.L.",
  "tax_id": { "country": "ES", "code": "B98602642" },
  "addresses": [
    { "num": "42", "street": "Calle Pradillo", "locality": "Madrid", "region": "Madrid", "code": "28002", "country": "ES" }
  ],
  "emails": [{ "addr": "billing@example.com" }]
}
```

Public API request:

```json
{ "data": { "$schema": "https://gobl.org/draft-0/bill/invoice", "...": "..." }, "envelop": false }
```

CLI loop:

```bash
gobl build -i invoice.json                 # calculate and validate
gobl build invoice.json > built.json
gobl correct -i --credit built.json        # credit note with preceding
gobl keygen && gobl sign -i built.json > signed.json
gobl verify signed.json
```

## Decoding errors

- Shape: `{"key": "validation", "faults": [{"code": "GOBL-...", "paths": ["$.lines[0].taxes[0]"], "message": "..."}]}` or `{"key": "calculation", "message": "..."}` / `{"key": "input", "message": "..."}` for errors before validation.
- Fault codes are `GOBL-<PACKAGE>-<TYPE>-<NN>` for core rules (`GOBL-BILL-INVOICE-10`), `GOBL-<REGIME>-...` for regime rules (`GOBL-GB-TAX-IDENTITY-03`), `GOBL-<ADDON-KEY>-...` for addon rules (`GOBL-ES-VERIFACTU-BILL-INVOICE-06`). The rule text is in the Validation Rules table of the schema, regime or addon page.
- `currency: missing or invalid` → no regime could be determined: set `$regime` or give the supplier a `tax_id`, or for a country without a regime set `currency` and explicit `percent`s.
- `GOBL-TAX-COMBO-04 tax combo percent required` → you gave a `key` with no rate and no percent, or ran `validate` on an unbuilt document.
- `invalid-category: 'vat' not defined in regime` → category codes are uppercase and regime specific.
- `'x' rate not defined for key 'standard'` → unknown rate key; read the regime page.
- `$tags: 'x' undefined` → the regime or addon does not define that tag.
- `unknown-schema` → missing or wrong `$schema`.
- `parsing time "15/01/2026" as "2006-01-02"` → dates must be `YYYY-MM-DD`.
- `GOBL-ENVELOPE-11 envelope digest does not match document contents` → an enveloped or signed document was edited; rebuild, or correct instead.
- `envelope doc is not ready to be signed, check code or other key fields` → signing requires `code`.

## Where to look

| Question | URL pattern |
| --- | --- |
| Fields, calculated markers and validation rules of a type | `https://docs.gobl.org/draft-0/<package>/<type>.md` (`bill/invoice`, `bill/line`, `org/party`, `tax/identity`, `tax/combo`, `pay/terms`, `pay/instructions`, `org/document_ref`) |
| Tax categories, rate keys, percentages, tags, correction types of a country | `https://docs.gobl.org/regimes/<code>.md` |
| Extension codes and scenarios of a local format | `https://docs.gobl.org/addons/<key>.md` |
| Shared code lists (UNTDID, ISO, CEF) | `https://docs.gobl.org/catalogues/<key>.md` |
| Validation layering and fault codes | `https://docs.gobl.org/overview/validation.md` |
| Numbers, rounding, canonical JSON | `https://docs.gobl.org/overview/numbers.md`, `.../rounding.md`, `.../canonicalization.md` |
| CLI and signing walkthrough | `https://docs.gobl.org/quick-start/cli.md`, `.../quick-start/invoices.md` |
| REST API and MCP | `https://docs.gobl.org/api/introduction.md`, `.../api/mcp.md` |
| Delivering documents to tax authorities and networks | `https://docs.invopop.com/llms.md` |

## Vocabulary

| You might say | GOBL says |
| --- | --- |
| Invoice number | `series` + `code` |
| VAT number, NIF, RFC, tax ID | `tax_id` with `country` and `code` |
| Registration number, Peppol ID, GLN | an `identities` entry |
| 21% VAT | `{"cat": "VAT", "rate": "standard"}` |
| Prices include tax | `tax.prices_include: "VAT"` |
| Withholding | a second tax combo (`IRPF`), see `totals.retained_tax` |
| Refund | `type: credit-note` + `preceding`, via `correct --credit` |
| Replace a wrong invoice | `type: corrective` where the regime allows it |
| Receipt without customer | `$tags: ["simplified"]` |
| Reverse charge | `$tags: ["reverse-charge"]`, tax `key: reverse-charge` |
| Bank transfer, card, direct debit | `payment.instructions.key`: `credit-transfer`, `card`, `direct-debit` |
| Due in 30 days | `payment.terms.key: due-date` with `due_dates` |
| Country format, e-invoicing standard | `$addons` + `ext` codes |
| Hash, signature, authority seal | `head.dig`, `sigs`, `head.stamps` |
| Custom fields | `meta` |

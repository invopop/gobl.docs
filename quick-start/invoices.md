# Invoices

For this guide, we assume you already have already successfully followed in the steps in the CLI Quick Start.

Creating an invoice is one of the most useful features of GOBL as it demonstrates the full power of the library. For the sake of the examples on this page, we're using Spain as the base region, but everything here should be supported with minor modifications in any other supported country.

## Prepare Source

To get started we're going to need some base data. Take the following yaml, modify as you see fit, and output to a file like `invoice.yaml`:

```yaml
region: "ES"
currency: "EUR"
issue_date: "2022-02-01"
code: "SAMPLE-001"

supplier:
  tax_id:
    country: "ES"
    code: "B98602642" # random
  name: "Provider One S.L."
  emails:
    - addr: "billing@example.com"
  addresses:
    - num: "42"
      street: "Calle Pradillo"
      locality: "Madrid"
      region: "Madrid"
      code: "28002"
      country: "ES"

customer:
  tax_id:
    country: "ES"
    code: "54387763P"
  name: "Sample Consumer"

lines:
  - quantity: 10
    item:
      name: "Item being purchased"
      price: "100.00"
    discounts:
      - rate: "10%"
    taxes:
      - cat: "VAT"
        code: "STD"
```

Take note of some of the key details there:

- The `region` is set to `ES`, the 2-letter ISO code for spain.
- `supplier` identifies the issuer of the invoice.
- `customer` includes the minimum amount of data possible for a customer.
- `lines` defines a simple list of items, including the tax codes, in this case "standard VAT".
- A single discount of "10%" has been defined for the line.
- There are no totals or other calculations, just the basic raw data.

## Building

Take the invoice details and send them to the `gobl envelop` command:

```bash
gobl envelop -t bill.Invoice -i invoice.yaml
```

You should get back something like:

```json
{
  "$schema": "https://gobl.org/draft-0/envelope",
  "head": {
    "uuid": "80c678a5-a481-11ec-aa03-3e7e00ce5635",
    "dig": {
      "alg": "sha256",
      "val": "8783ae676923eaa6add4ef53c7e0a6d0dc5a249ba0b8597f4b4bc25197f67b8f"
    }
  },
  "doc": {
    "$schema": "https://gobl.org/draft-0/bill/invoice",
    "region": "ES",
    "code": "SAMPLE-001",
    "currency": "EUR",
    "issue_date": "2022-02-01",
    "supplier": {
      "tax_id": {
        "country": "ES",
        "code": "B98602642"
      },
      "name": "Provider One S.L.",
      "addresses": [
        {
          "num": "42",
          "street": "Calle Pradillo",
          "locality": "Madrid",
          "region": "Madrid",
          "code": "28002",
          "country": "ES"
        }
      ],
      "emails": [
        {
          "addr": "billing@example.com"
        }
      ]
    },
    "customer": {
      "tax_id": {
        "country": "ES",
        "code": "54387763P"
      },
      "name": "Sample Consumer"
    },
    "lines": [
      {
        "i": 1,
        "quantity": "10",
        "item": {
          "name": "Item being purchased",
          "price": "100.00"
        },
        "sum": "1000.00",
        "discounts": [
          {
            "rate": "10%",
            "amount": "100.00"
          }
        ],
        "taxes": [
          {
            "cat": "VAT",
            "code": "STD"
          }
        ],
        "total": "900.00"
      }
    ],
    "totals": {
      "sum": "900.00",
      "total": "900.00",
      "taxes": {
        "categories": [
          {
            "code": "VAT",
            "rates": [
              {
                "code": "STD",
                "base": "900.00",
                "percent": "21.0%",
                "amount": "189.00"
              }
            ],
            "base": "900.00",
            "amount": "189.00"
          }
        ],
        "sum": "189.00"
      },
      "total_with_tax": "1089.00",
      "payable": "1089.00"
    }
  },
  "sigs": [
    "eyJhbGciOiJFUzI1NiIsImtpZCI6IjBhMjg2MDAwLTM2MGEtNGU2Ni04MWFhLTU2ZDQ0YmI4ZjEwNyJ9.eyJ1dWlkIjoiODBjNjc4YTUtYTQ4MS0xMWVjLWFhMDMtM2U3ZTAwY2U1NjM1IiwiZGlnIjp7ImFsZyI6InNoYTI1NiIsInZhbCI6Ijg3ODNhZTY3NjkyM2VhYTZhZGQ0ZWY1M2M3ZTBhNmQwZGM1YTI0OWJhMGI4NTk3ZjRiNGJjMjUxOTdmNjdiOGYifX0.ueUv3IYQPXgOUzRrBIxMH1eWkHUzVt6Um04ZWO7dhoj0SHlVA4ZS82glDfG0Njjn0fRc4UlpZGKz8j0aJJKABg"
  ]
}
```

Congratulations! You've just created a complete invoice written in GOBL, including digital signatures.

Take a look at some of the data that has been generated, a few observations:

- Everything has been embedded inside a GOBL Envelope, with a head and signature.
- Each invoice line has been updated with a `sum` and `total`, where the total is the sum with discounts applied.
- A `totals` property has been added to the invoice, and details all the taxes that were applied to the calculations made from the lines.

## Test PDF

[Invopop](https://invopop.com), the creators of GOBL, offer a free service to be able to test what your invoices look like as a PDFs. We do not recommend this for production use as it does not come with the same service guarantees, but is great for testing.

Create a new `invoice.json` file:

```bash
gobl envelop -t bill.Invoice -i invoice.yaml > invoice.json
```

Now use curl to send it to the testing service and dump the output:

```bash
curl -X POST -F envelope=@invoice.json https://es-pdf.invopop.com/api > output.pdf
```

And open it either from the command line if supported by your system, or directly in your favorite PDF reader:

```bash
open output.pdf
```

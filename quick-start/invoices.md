# Invoices

For this guide, we assume you already have already successfully followed in the steps in the CLI Quick Start.

Creating an invoice is one of the most useful features of GOBL as it demonstrates the full power of the library. For the sake of the examples on this page, we're using Spain as the base region, but everything here should be supported with minor modifications in any other supported country.

## Prepare Source

To get started we're going to need some base data. Take the following yaml, modify as you see fit, and output to a file like `invoice.yaml`:

```yaml
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
      - percent: "10%"
    taxes:
      - cat: "VAT"
        rate: "standard"
```

Take note of some of the key details there:

- `supplier` identifies the issuer of the invoice, and the `tax_id` specifically sets the origin country for this invoice.
- `customer` includes the minimum amount of data possible for a customer.
- `lines` defines a simple list of items, including the tax codes, in this case the category is "VAT" at the "standard" rate.
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
    "uuid": "0699dd94-d296-11ec-9e60-02f7bbd929bc",
    "dig": {
      "alg": "sha256",
      "val": "c0d4efca94e2980a90f725a98d16ff5f7a48e8deb932a436da24b54b2a76aa3d"
    }
  },
  "doc": {
    "$schema": "https://gobl.org/draft-0/bill/invoice",
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
            "percent": "10%",
            "amount": "100.00"
          }
        ],
        "taxes": [
          {
            "cat": "VAT",
            "rate": "standard"
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
                "key": "standard",
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
      "tax": "189.00",
      "total_with_tax": "1089.00",
      "payable": "1089.00"
    }
  },
  "sigs": [
    "eyJhbGciOiJFUzI1NiIsImtpZCI6IjlmNWM5MmUyLWIwMDUtNGIyZi04MTExLTUzYTVlMmJmNzZiNSJ9.eyJ1dWlkIjoiMDY5OWRkOTQtZDI5Ni0xMWVjLTllNjAtMDJmN2JiZDkyOWJjIiwiZGlnIjp7ImFsZyI6InNoYTI1NiIsInZhbCI6ImMwZDRlZmNhOTRlMjk4MGE5MGY3MjVhOThkMTZmZjVmN2E0OGU4ZGViOTMyYTQzNmRhMjRiNTRiMmE3NmFhM2QifX0.zDA-foD7axWp-FY0rz5tdjkAOaxy_GHHDGN_GuV-aS1U7TAe56pX3K3RjBMAAPc1UMA4JRirb3eehU9jUuvIbg"
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
curl -X POST -F envelope=@invoice.json https://pdf.invopop.com/api > output.pdf
```

And open it either from the command line if supported by your system, or directly in your favorite PDF reader:

```bash
open output.pdf
```

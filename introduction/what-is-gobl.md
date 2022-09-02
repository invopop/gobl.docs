# What is GOBL?

GOBL stands for "Go Business Language", and is three things in one:

- a [Go library](https://github.com/invopop/gobl), [CLI tool, and microservice](https://github.com/invopop/gobl.cli) used to build and validate business documents in JSON,
- a [repository of global taxes and validation rules](https://github.com/invopop/gobl/tree/main/regions) defined in code and output into JSON, and,
- a [JSON Schema](https://github.com/invopop/gobl/tree/main/build/schemas) for easy creation and sharing of GOBL data.

The base component of GOBL is an [Envelope](../core/envelopes.md). This acts as a wrapper around the contents or [Document](../core/documents.md) with the actual payload to be read and used.

All structures in GOBL are provided with their [JSON Schema](https://json-schema.org/) definitions generated from the structures in Go.

JSON Schemas allows us to easily generate structures and definitions of GOBL in other languages, so we also currently support:

- [gobl.ruby](https://github.com/invopop/gobl.ruby) - read GOBL generated files in formal Ruby classes.

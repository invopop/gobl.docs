# What is GOBL?

GOBL stands for "Go Business Language", and is three things in one:

* a Go library, CLI tool, and microservice used to build and validate business documents in JSON,
* a repository of global taxes and validation rules defined in code and output into JSON, and,
* a JSON Schema for easy creation and sharing of GOBL data.

The base component of GOBL is an [Envelope](../core/envelopes.md). This acts as a wrapper around the contents or [Document](../core/documents.md) with the actual payload to be read and used.

All structures in GOBL are provided with their [JSON Schema](https://json-schema.org) definitions generated from the structures in Go.

JSON Schemas allows us to easily generate structures and definitions of GOBL in other languages, so we also currently support:

* [gobl.ruby](https://github.com/invopop/gobl.ruby) - read GOBL generated files in formal Ruby classes.

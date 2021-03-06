# Envelopes

"GOBL Envelope" is the name we've given to the structure that wraps around a document, adding headers and seals (signatures), just like an envelope in real-life. There are four key parts to an envelope in GOBL:

* Schema (`$schema`) - Where the definition for the GOBL envelope can be found, usually something like `https://gobl.org/draft-0/envelope`.
* Header (`head`) - Meta data that describes the included document, including the document's hash.
* Document (`doc`) - The actual payload of the envelope, like an invoice.
* Signatures (`sigs`) -  A set of [JSON Web Signatures](https://en.wikipedia.org/wiki/JSON\_Web\_Signature) that can be used to verify the headers included in the signature, and thus that the document has not been modified.

We'll go through each of this in a bit more detail in the following chapters.

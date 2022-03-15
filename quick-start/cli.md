---
description: Getting started with the GOBL Command Line Interface.
---

# CLI

The GOBL CLI is one of the easiest ways to get to grips with using and understanding some of the underlying concepts.

For the sake of the examples here, we'll be using the `note` package's `Message` type.

### Installation

Until we have some pre-built binaries, you'll need to have a [working Go environment](https://go.dev/doc/install).

Install the GOBL CLI tool with:

```
go install github.com/invopop/gobl.cli/cmd/gobl
```

Check it's working:

```
gobl version
```

### Keys

Valid GOBL envelopes require signatures. The CLI makes this process trivial, but you do need to have generated a [JSON Web Key](https://datatracker.ietf.org/doc/html/rfc7517). The `keygen` command will create a key pair inside the `~/.gobl` directory by running:

```shell
gobl keygen
```

Check the contents of the key:

```
cat ~/.gobl/id_es256.jwk | json_pp
```

Outputs something like:

```
{
   "d" : "iCGEXA1Yn6y7DnUdUbmng8IPmx-yXukDtmCC2XGppjY",
   "x" : "T737Xx74Wacl8hrdG0SqEucY_02yuNwku4C-ANDu5MM",
   "kid" : "69b22998-e434-41f2-b957-e8ac885f487d",
   "y" : "kRWPf8nraKep8FXzKds5JPnk36LJpuqF84x8TAjFPoI",
   "use" : "sig",
   "kty" : "EC",
   "crv" : "P-256",
   "alg" : "ES256"
}
```

**IMPORTANT**: A private key like the one above but intended for use should never be shared.

The GOBL CLI generates a second _public_ key which can be shared:

```
cat ~/.gobl/id_es256.pub.jwk | json_pp
```

Outputs:

```
{
   "y" : "kRWPf8nraKep8FXzKds5JPnk36LJpuqF84x8TAjFPoI",
   "x" : "T737Xx74Wacl8hrdG0SqEucY_02yuNwku4C-ANDu5MM",
   "use" : "sig",
   "crv" : "P-256",
   "kid" : "69b22998-e434-41f2-b957-e8ac885f487d",
   "kty" : "EC",
   "alg" : "ES256"
}
```

If anyone ever needs to verify the source of a GOBL Envelope that you signed, simply send them a copy or provide them access to your public key.

### Building a Message

The [notes.Message](https://github.com/invopop/gobl/blob/main/note/message.go) type is great for setting up a simple test. For this tutorial open your text editor and a simple YAML file called `message.yaml` that looks like:

```
title: "Test Message"
content: |-
  We hope you like this test message!
```

You could also create the base document directly in JSON, but we find YAML much easier for creating contents by hand.

Now send the document to the gobl `envelop` command, indicating that we want to insert a document of type `note.Message` into a new envelope:

```
gobl envelop -i -t note.Message ./message.yaml | json_pp
```

You should get something similar to the following:

```json
{
  "$schema": "https://gobl.org/draft-0/envelope",
  "head": {
    "uuid": "8e69dd09-9adb-11ec-82ed-665181255c0a",
    "dig": {
      "alg": "sha256",
      "val": "45ac3115c8569a1789e58af8d0dc91ef3baa1fb71daaf38f5aef94f82b4d0033"
    }
  },
  "doc": {
    "$schema": "https://gobl.org/draft-0/note/message",
    "title": "Test Message",
    "content": "We hope you like this test message!"
  },
  "sigs": [
    "eyJhbGciOiJFUzI1NiIsImtpZCI6ImE2YzM5MjBkLThlMzMtNDE1OS1iMzM3LTllNzQ2MTcxNmRmMSJ9.eyJ1dWlkIjoiOGU2OWRkMDktOWFkYi0xMWVjLTgyZWQtNjY1MTgxMjU1YzBhIiwiZGlnIjp7ImFsZyI6InNoYTI1NiIsInZhbCI6IjQ1YWMzMTE1Yzg1NjlhMTc4OWU1OGFmOGQwZGM5MWVmM2JhYTFmYjcxZGFhZjM4ZjVhZWY5NGY4MmI0ZDAwMzMifX0.VV9LRGEVPoO-tnOS-j6ItUEvYNcaQ1CbwCMN3qJorZXV3ON51wzalRuzJxulPnlFPtohWd_gc2Mf81MDIAK47Q"
  ]
}
```

The envelop command performed a few tasks to create a complete envelope:

1. Validated the message object to check it contains the required fields, and added the message schema: `https://gobl.org/draft-0/note/message`.
2. Wrapped the document inside an envelope defined with the JSON schema `https://gobl.org/draft-0/envelope`.
3. Generated a `head` element containing a unique UUIDv1 for the envelope and a digest of the document (`doc`) contents.
4. Added a [JSON Web Signature](https://datatracker.ietf.org/doc/html/rfc7515) to the `sigs` array.

> **TIP:** you can checkout the contents of the JSON Web Signature by copying and pasting it on the website [jwt.io](https://jwt.io).

It's also possible "complete" a GOBL envelope, especially useful where you want to add additional data to the envelope's headers. For example, given a message definition like the following:

```yaml
head:
  draft: true
  tags:
    - "sample"
doc:
  $schema: "https://gobl.org/draft-0/note/message"
  title: "Test Message"
  content: |-
    We hope you like this test message!
```

Running the command:

```bash
gobl build -i ./message.draft.yaml
```

Will result in the following output, note the lack of the `sigs` property which are not included in draft envelopes:

```json
{
  "$schema": "https://gobl.org/draft-0/envelope",
  "head": {
    "uuid": "8e69dd09-9adb-11ec-82ed-665181255c0a",
    "tags": ["sample"],
    "draft": true,
    "dig": {
      "alg": "sha256",
      "val": "45ac3115c8569a1789e58af8d0dc91ef3baa1fb71daaf38f5aef94f82b4d0033"
    }
  },
  "doc": {
    "$schema": "https://gobl.org/draft-0/note/message",
    "title": "Test Message",
    "content": "We hope you like this test message!"
  }
}
```

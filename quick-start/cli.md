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
cat ~/.gobl/id_es256 | json_pp
```

Outputs something like:

```
{
   "kty" : "EC",
   "x" : "eQBI2X3ZfSFRb7hcuBb0Mq994JMJ0sVvelKybpXjXYo",
   "crv" : "P-256",
   "kid" : "bf51d6e0-7b6d-49af-9cce-c86719ede284",
   "use" : "sig",
   "y" : "9H-a68g_Rl_Oq9nsO5rhtXpr-8Lx27zHMUlx3Rr-oDo",
   "d" : "d7dxIHe3pvqndjofgUFTEr7BH4fJLd8znOlIQmcpnk8",
   "alg" : "ES256"
}
```

### Building a Message

The [notes.Message](https://github.com/invopop/gobl/blob/main/note/message.go) type is great for setting up a simple test. For this tutorial open your text editor and a simple YAML file called `message.yaml` that looks like:

```
doc:
  $schema: "https://gobl.org/draft-0/note/message"
  title: "Test Message"
  content: |-
    We hope you like this test message!
```

You could also create the base document directly in JSON, but we find YAML much easier for creating contents by hand.

Notice the `$schema` property. This lets the builder know what type of document we want to create.

Now send the document to the gobl `build` command:

```
gobl build ./message.yaml | json_pp
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

The build command performed a few tasks to create a complete envelope:

1. &#x20;Validated the message object to check it contains the required fields.
2. Added the envelope's JSON schema, in this case: `https://gobl.org/draf`t`-0/envelope`
3. Generated a `head` element containing a unique UUIDv1 for the envelope and a digest of the document (`doc`) contents.
4. Added a [JSON Web Signature](https://datatracker.ietf.org/doc/html/rfc7515) to the `sigs` array.


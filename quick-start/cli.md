---
description: Getting started with the GOBL Command Line Interface.
---

# CLI

The GOBL CLI is a useful tool to get to grips with using and understanding some of the underlying concepts.

For the sake of the examples here, we'll be using the `note` package's `Message` type.

### Installation

Our recommendation is to head over to the [GOBL CLI releases page](https://github.com/invopop/gobl.cli/releases) and download the latest version for your platform and copy the binary to a directory of your choosing.

If you already have a [working Go environment](https://go.dev/doc/install) you may also find it easy to install the latest version with:

```
go install github.com/invopop/gobl.cli/cmd/gobl@latest
```

Once downloaded and installed or built, check it's working:

```
gobl version
```

If you're using a packaged version, you should get version information back like:

```json
{
  "version": "0.46.0",
  "gobl": "v0.30.2",
  "date": "2022-09-02T12:37:32Z"
}
```

### Building a Message

The [notes.Message](https://github.com/invopop/gobl/blob/main/note/message.go) type is great for setting up a simple test. For this tutorial open your text editor and a simple JSON file called `message.json` that looks like:

```json
{
  "$schema": "https://gobl.org/draft-0/note/message",
  "title": "Test Message",
  "content": "We hope you like this test message!"
}
```

If you're using an editor like VSCode which has built-in support for JSON Schemas, it may already have performed some pre-validations on the document which will all be positive for this simple example.

The GOBL CLI also supports YAML input, but in general, thanks to the schemas and powerful text editors, we find it a bit easier to write files in JSON.

Now send the document to the gobl `build` command with the `--envelope` and `--draft` flags indicating that we want a draft envelope of the message. The `-i` flag produces prettier output:

```
gobl build -i --envelop --draft ./message.json
```

You should get something similar to the following:

```json
{
  "$schema": "https://gobl.org/draft-0/envelope",
  "head": {
    "uuid": "dda69a91-2ad0-11ed-a33d-3e7e00ce5635",
    "dig": {
      "alg": "sha256",
      "val": "45ac3115c8569a1789e58af8d0dc91ef3baa1fb71daaf38f5aef94f82b4d0033"
    },
    "draft": true
  },
  "doc": {
    "$schema": "https://gobl.org/draft-0/note/message",
    "title": "Test Message",
    "content": "We hope you like this test message!"
  }
}
```

The original message has now been placed into a GOBL Envelope with a header that allows us to ensure that the contents of the document cannot be modified without creating a new digest.

### Keys

GOBL has built in support for digital signatures using [JSON Web Keys](https://datatracker.ietf.org/doc/html/rfc7517). The CLI makes this process trivial, but you do need to have generated a private key. The `keygen` command will create a key pair inside the `~/.gobl` directory by running:

```shell
gobl keygen
```

Check the contents of the key:

```
cat ~/.gobl/id_es256.jwk | jq
```

Outputs something like:

```
{
   "use" : "sig",
   "kty" : "EC",
   "kid" : "69b22998-e434-41f2-b957-e8ac885f487d",
   "crv" : "P-256",
   "alg" : "ES256",
   "x" : "T737Xx74Wacl8hrdG0SqEucY_02yuNwku4C-ANDu5MM",
   "y" : "kRWPf8nraKep8FXzKds5JPnk36LJpuqF84x8TAjFPoI",
   "d" : "iCGEXA1Yn6y7DnUdUbmng8IPmx-yXukDtmCC2XGppjY"
}
```

**IMPORTANT**: Private keys should never be shared! The GOBL CLI generates a second _public_ key which can be shared with others to validate a document is from you:

```
cat ~/.gobl/id_es256.pub.jwk | json_pp
```

Outputs:

```
{
   "use" : "sig",
   "kty" : "EC",
   "kid" : "69b22998-e434-41f2-b957-e8ac885f487d",
   "crv" : "P-256",
   "alg" : "ES256",
   "x" : "T737Xx74Wacl8hrdG0SqEucY_02yuNwku4C-ANDu5MM",
   "y" : "kRWPf8nraKep8FXzKds5JPnk36LJpuqF84x8TAjFPoI"
}
```

If anyone ever needs to verify the source of a GOBL Envelope that you signed, simply send them a copy or provide them access to your public key.

### Signing

Now we have a private key, we can sign the original message. Run the following command:

```bash
gobl sign -i message.json
```

The output produced should be something like:

```json
{
  "$schema": "https://gobl.org/draft-0/envelope",
  "head": {
    "uuid": "c7eac0a3-2ad2-11ed-964a-3e7e00ce5635",
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
    "eyJhbGciOiJFUzI1NiIsImtpZCI6IjBhMjg2MDAwLTM2MGEtNGU2Ni04MWFhLTU2ZDQ0YmI4ZjEwNyJ9.eyJ1dWlkIjoiYzdlYWMwYTMtMmFkMi0xMWVkLTk2NGEtM2U3ZTAwY2U1NjM1IiwiZGlnIjp7ImFsZyI6InNoYTI1NiIsInZhbCI6IjQ1YWMzMTE1Yzg1NjlhMTc4OWU1OGFmOGQwZGM5MWVmM2JhYTFmYjcxZGFhZjM4ZjVhZWY5NGY4MmI0ZDAwMzMifX0.JsLXd3TkKOwuy0KOXJVG8atIShlNZb1vbLglVO8PDZnZbwGVyWRE_i7y85lVKvSby-j0rwU9wgleIxcGj9tE4g"
  ]
}
```

Essentially the `doc` and `head` fields are identical to the original message, but we've now added the `sigs` array at the end. Combined with your public key, anyone can easily verify the contents of your message where indeed signed with your private key.

If you're interested, you can check the contents of the signature here: [jwt.io](https://jwt.io).

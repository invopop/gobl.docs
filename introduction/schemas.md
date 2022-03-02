# Schemas

GOBL defines all structures in Go and uses the [jsonschema](https://github.com/invopop/jsonschema) package to automatically convert them into valid [JSON Schema](https://json-schema.org).

For example, take the following `Message` struct found inside the GOBL [note package](https://github.com/invopop/gobl/tree/main/note):

```go
// Message represents the minimum possible contents for a GoBL document type. This is
// mainly meant to be used for testing purposes.
type Message struct {
	// Summary of the message content
	Title string `json:"title,omitempty" jsonschema:"title=Title"`
	// Details of what exactly this message wants to communicate
	Content string `json:"content" jsonschema:"title=Content"`
	// Any additional semi-structured data that might be useful.
	Meta org.Meta `json:"meta,omitempty" jsonschema:"title=Meta Data"`
}
```

When converted to JSON Schema, it takes the following form:

```json
{
  "$schema": "http://json-schema.org/draft/2020-12/schema",
  "$id": "https://gobl.org/draft-0/note/message",
  "$ref": "#/$defs/Message",
  "$defs": {
    "Message": {
      "properties": {
        "title": {
          "type": "string",
          "title": "Title",
          "description": "Summary of the message content"
        },
        "content": {
          "type": "string",
          "title": "Content",
          "description": "Details of what exactly this message wants to communicate"
        },
        "meta": {
          "patternProperties": {
            ".*": {
              "type": "string"
            }
          },
          "type": "object",
          "title": "Meta Data",
          "description": "Any additional semi-structured data that might be useful."
        }
      },
      "type": "object",
      "required": [
        "content"
      ],
      "description": "Message represents the minimum possible contents for a GoBL document type."
    }
  },
  "$comment": "Generated with GOBL v0.18.0"
}
```


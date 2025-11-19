# GOBL Documentation

The content and configuration powering the documentation available at [docs.gobl.org](https://docs.gobl.org)

### 👩‍💻 Development

Install the [Mintlify CLI](https://www.npmjs.com/package/mintlify) to preview the documentation changes locally. To install, use the following command

```
npm i mintlify -g
```

Run the following command at the root of your documentation (where mint.json is)

```
mintlify dev
```

### Generated Content

There are two sources:

 * Internal script for catalagues, addons, and regime definitions.
 * Ruby generate tool (`gobl.generate`) for transforming schemas.

For the Ruby Generate tool, see that repo directly for details.

For the internal scripts, simply run:

```bash
go build ./cmd/generate && ./generate
```


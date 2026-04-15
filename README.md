# GOBL Documentation

The content and configuration powering the documentation available at [docs.gobl.org](https://docs.gobl.org)

### 👩‍💻 Development

Install the [Mintlify CLI](https://www.mintlify.com/docs/installation) to preview the documentation changes locally. To install, use the following command

```
npm i mint -g
```

Run the following command at the root of your documentation (where mint.json is)

```
mint dev
```

### Generated Content

There are two sources:

 * Internal script for catalagues, addons, and regime definitions.
 * Ruby generator tool (`gobl.generator`) for transforming schemas.

For the Ruby Generate tool, see that repo directly for details. Generally, you can run:

```bash
rm -rf ./draft-0
../gobl.generator/bin/generate -l markdown -i ../gobl/data/schemas -o ./draft-0
```

For the internal scripts, simply run:

```bash
go build ./cmd/generate && ./generate
```

After running any generation script, you'll need to double check that everything is linked to from the `docs.json` file.


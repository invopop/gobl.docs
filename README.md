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

Schema, catalogue, addon, and regime pages are produced by the internal Go
generator in `cmd/generate`. Schemas are read from the `data.Content` embedded
FS of the GOBL module pinned in `go.mod`, so no sibling checkout of the gobl
repo is required.

To regenerate, run:

```bash
go build ./cmd/generate && ./generate
```

The generator wipes `./draft-0` on each run so schemas removed upstream don't
linger as orphaned pages. Treat the entire `./draft-0` directory as
generator-owned — don't hand-edit files under it, as any local changes will be
overwritten on the next regeneration.

After running the generator, double check that everything is linked to from the `docs.json` file.


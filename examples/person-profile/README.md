# Generic person-profile example

This public fixture demonstrates local profile authoring with an admitted local
personality library. It contains no customer material.

From this directory, compose the compiled example:

```text
agent-compose compose request.kdl --out ./bundles
```

The summary prints the immutable `<bundle-dir>` beneath `./bundles`. Inspect,
verify, compare, and export it with:

```text
agent-compose verify <bundle-dir>
agent-compose describe <bundle-dir>
agent-compose diff <bundle-dir> <bundle-dir>
agent-compose bundle export <bundle-dir> --out ./caption-workbench.tar.gz
```

Inspect the complete effective profile:

```text
agent-compose catalog personalities --person-source . --personality-library ../shared-personality-library
agent-compose catalog personalities --person-source . --personality-library ../shared-personality-library --query supportive --json
agent-compose catalog roles --person-source . --personality-library ../shared-personality-library --json
agent-compose catalog seats --person-source . --personality-library ../shared-personality-library --role bulk-captioner --json
agent-compose catalog expressions --json
```

Exercise the remaining generated surfaces:

```text
agent-compose evaluation --person-source . --personality-library ../shared-personality-library --role bulk-captioner --seat chatbot-sonnet-low --format yaml
agent-compose overlay --person-source . --personality-library ../shared-personality-library --role bulk-captioner --seat chatbot-sonnet-low --expression available --json
agent-compose roster --person-source . --personality-library ../shared-personality-library --out ./roster
agent-compose palette-data --person-source . --personality-library ../shared-personality-library --out ./palette.json
```

The executable test suite runs the same profile through compose, verification,
export, describe, evaluation, overlay, roster, palette, v3 and v4 snapshots,
and every catalogue projection. It also proves the arbitrary
`chatbot-sonnet-low` seat, `they` pronouns, package-local plus shared-library
meld, single-personality meld, unused catalogue personality, role skills, and
copy-contract provenance.

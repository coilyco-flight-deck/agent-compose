# Personality palette explorer

The repository ships a local visual explorer for the canonical personality
colors and each role's complete meld. It preserves color as expression only.
Authority, safety, and completion remain outside personality.

## Run it

From the repository root:

```sh
just palette-serve
```

Ward generates `web/personality-palette/public/palette.json` from the embedded
person source, installs the pinned browser toolchain, and starts Vite on its
local address. The JSON, dependency directory, and production build stay
uncommitted.

The remaining lifecycle verbs are:

* `just palette-build` - type-check and build the production assets.
* `just palette-test` - build and verify the generated artifact.
* `just palette-tidy` - reconcile `package-lock.json`.

`just test` includes the palette test, so every release validates the
committed explorer.

## Data ownership

The embedded KDL person source remains the only owner of personality colors,
identity primitives, role membership, role order, and boundary inputs. The hidden
`palette-data` command projects that source into versioned JSON. The Go color
package derives each role's boundary before the browser sees it.

The TypeScript layer owns only visual presentation metadata such as friendly
color names, short associations, and spectrum ordering. Startup validation
fails visibly if that presentation list drifts from the canonical catalog.

## Interaction

The explorer provides:

* the full sixteen-personality spectrum with emblem, motif, form, and sound
* role filters with complete personality melds
* component colors and the derived role boundary
* day and night previews
* spectrum and alphabetical ordering
* one-click copying for component and melded colors
* responsive layout and reduced-motion support

The app is a framework-free Vite and TypeScript project under
`web/personality-palette`. It is a local source tool, not a deployment target.

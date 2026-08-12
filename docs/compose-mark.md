# The agent-compose mark

A spool of thread wound with twill cloth, its flanges crossing the ring. It is
a sibling of the coilyco org avatars and shares their ink, mint, and lilac.

The files live in [assets/mark/README.md](../assets/mark/README.md).

## What ships in assets/mark

- `agent-compose.svg` - the canon mark on the 400 avatar canvas. Its opaque
  ink field is part of the mark, so it stands alone on any background.
- `agent-compose-{400,256,128}.png` - coin rasters, transparent outside a disc
  just inside the ring, so each drops onto any surface without dark corners.
- `agent-compose-favicon-{64,32,16}.{svg,png}` - small sizes, each drawn on its
  own pixel grid. These keep the opaque field rather than the coin mask.

Do not resize the canon for a favicon. Nothing in it lands on a pixel boundary
at small sizes, so every edge blurs across two pixels. Use the size-specific
files, or regenerate.

## Geometry

On the 400 canvas: flange half-width 124 before the cut, cut by the ring circle
at r 171.5, thickness 32, taper 26, ink halo 7. Core half-width 72, half-height
96. Twill at cell 26. Mint ring r 165.5 stroke 12, lilac ring r 153 stroke 13.

Three numbers are load-bearing rather than cosmetic.

- The core half-width equals `flange_half_width - taper`, so the core's edges
  continue the line the flange taper ends on. Narrower reads thin, and wider
  leaves the thread standing proud of the core.
- The flanges are drawn after the rings, over a 7-wide ink halo. Drawn before
  them a mint flange merges into the mint ring, and the crossing then reads as
  a mistake rather than as a decision.
- The flange is cut by the ring circle so its ends follow the curve. Past
  r 171.5 it would leave the ring and the coin mask would clip it.

## Regenerating

The generator is `scripts/marks/agent_compose_mark.py` in `agentic-os-xxx`, and
its canon output is pixel-identical to what ships here. Its comments carry the
constraints that silently change the mark when they are broken.

Two forms are still outstanding. The website canvas is a redraw at 500 with an
ink filter rather than a resize, and the lockup form over the coilyco S has not
been drawn.

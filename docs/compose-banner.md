# The agent-compose banner

A field of point-twill cloth with the [spool mark](compose-mark.md) on it,
drawn in the same construction language as the mark: flat fill, every edge on
an axis or at 45 degrees, ink with mint and lilac.

The file lives in [assets/banner/README.md](../assets/banner/README.md).

## What ships in assets/banner

- `agent-compose-banner.jpg` - 2560 by 1280, the only form.

One file, not four. The earlier convention called for a 1280 beside the 2560
and a form without the mark, on the theory that some surfaces already show the
avatar. No surface in use does: a social preview, a link card and a README
header all present the image alone. The 2560 downscales cleanly because every
pixel in it is drawn rather than photographed, and at 384 KB it sits inside
GitHub's 1 MB social-preview cap. Draw another form when a surface asks for it.

JPEG rather than PNG. The field is thousands of flat cells, which PNG handles
well, but the type carries a blurred halo that dithers visibly at any palette
small enough to help.

## The field

Point twill: a twill draft folded on both axes, so the cloth reads as nested
diamonds rather than as a texture. Herringbone folds one axis and goes to noise
at banner size.

On the 1280 form, from which the 2560 scales exactly: cell 32 with a 3 seam,
the fold at 6 cells, lattice contour 2 cells thick, secondary contours every 6.
Ground weight 0.10, lattice 0.50, secondary 0.22. Light rises to the right in
four quantized steps from 0.3 to 1.0, and the pattern fades to a quarter over
the outer 45 percent of the height at the top and bottom edges.

Those last two are why the banner sits in a README rather than on it. A first
cut ran the lattice to 0.8 and hard to every edge, which reads well on its own
and reads as a slab dropped onto the page. A banner is judged where it lands.

Four numbers there are load-bearing rather than cosmetic.

- The lattice is an integer contour of the folded draft, `zigzag(i) + zigzag(j) == fold`. A threshold on a normalised depth widens into a blob at every diamond vertex, because the fold's gradient goes flat there.
- The contour is 2 cells thick. A 1-cell contour on a 45 degree stair touches only at the corners, so it reads as a dotted line, and the seam between picks finishes it off.
- Each thread darkens along ink to its own shadow to its hue, mint through `#103a3f` and lilac through `#38275c`, with the shadow at 0.35 of the ramp. Interpolating straight from the shadow leaves the ground at a mid violet and the whole field glows.
- Picks draw as merged floats rather than as separate cells. A 2/2 twill's visible unit is a float two cells long, and boxing each cell breaks it into a pixel grid.

## Type

`agent-compose // $ acompose` over the tagline, lilac with a mint separator,
set on a centred dark halo rather than an offset drop shadow. An offset implies
a light direction the mark does not have and protects one side of each glyph.

The cloth itself carries the legibility: the pattern thins under the two lines
of type the way a woven label interrupts a sett, its edge dithered by a cell so
the quiet still reads as cloth. Quieting the whole lockup instead empties the
banner, since that box runs four fifths of the width, and the mark needs no
help.

## Regenerating

The generator is `scripts/banners/agent_compose_banner.py` in `agentic-os-xxx`,
alongside the mark generator it draws the spool from. Its working record and
the directions that failed on the way here are in that repo's
`kai-comfyui-agentic` skill.

The banner is not set as the repository's social preview. That is a setting
rather than a file, and it stays an operator action.

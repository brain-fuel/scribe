# scribe: Design Specification

Date: 2026-07-01
Status: Approved
Repo: `~/matt/goforge.dev/scribe`, module `goforge.dev/scribe`, GitHub `brain-fuel/scribe`
License: MIT

## What

A 2D drawing engine in Go, later ported to Rune, in the lineage of QuickDraw,
Quartz, and PostScript. It is predicated on Terry Lambert's account of what made
QuickDraw graphics good:

1. Coordinates are located at the intersections between pixels (the corner
   grid), not on pixel centers.
2. Rounded corners are produced by table-driven code emitting Bezier curves,
   whose start, end, and control points land between pixels.
3. Together these give clean antialiased edges and pixelation-free scaling.

**Metaphor (goforge naming world):** a metalworker's scriber scratches
dimensionless layout lines on metal before any cut is made. Layout lines are
corner-grid geometry: marks between material, not on it.

## Foundational model

### Corner-grid coordinates

Coordinates name the infinitely thin grid lines between pixels. Pixel `(x, y)`
is the unit square from point `(x, y)` to point `(x+1, y+1)`.

Load-bearing consequences:

- Rect `(0, 0, 4, 4)` covers exactly 16 pixels; width is exactly
  `right - left`. No off-by-one errors anywhere in the geometry layer.
- Scaling a rect by 2 covers exactly 4x the pixels. Scaling is exact; there is
  no pixelation drift.
- A stroke centered on a grid line straddles adjacent pixels symmetrically, so
  coverage-based antialiasing renders it correctly with no special cases.
- Bezier endpoints and control points land between pixels, which is what makes
  corners smooth at every scale.

### Numbers and determinism

- Public API uses `float64`.
- The core uses only IEEE-exact operations (`+ - * / sqrt`). **No trig in any
  hot path.**
- Circles, arcs, and round joins/caps come from **table-driven Bezier control
  coefficients** (Lambert's actual mechanism):
  - Circular corners: quarter-circle cubic with kappa = 4/3 * (sqrt(2) - 1).
  - Continuous corners (modern Apple squircle, G2 curvature continuity):
    precomputed control-point table.
- The rasterizer quantizes to fixed-point subpixels (1/256) and accumulates
  coverage in integers, so output is **byte-identical across platforms**. This
  matches the goforge divergence-lock ethos and eases the later Rune port
  (exact numeric tower).

## Architecture (layers and packages)

```
geom/      Point, Rect, Affine transform; corner-grid semantics live here
path/      Path (MoveTo / LineTo / CubicTo / Close), adaptive flattening,
           stroke-to-fill conversion (joins and caps built from the same
           corner tables), RoundRect with both corner styles
raster/    coverage-based scanline antialiasing (signed-area accumulation,
           FreeType-smooth family); nonzero and even-odd fill rules;
           produces an alpha mask
paint/     solid color in v1 (gradients later); premultiplied-alpha
           compositing into image.RGBA
scribe.go  Canvas: QuickDraw-flavored top-level API (SetPen, Fill, Stroke)
           rendering to image.RGBA
dl/        display list: drawing ops as data. THE portable IR: serializable
           and deterministic. This is what the Rune port targets (an
           inductive datatype interpreted by a fold).
ps/        postfix mini-language, the PostScript homage:
           "0 0 512 512 114 roundrect fill" -> dl -> raster
svg/       SVG backend (exact geometry preserved; no rasterization)
term/      TUI backend: kitty graphics protocol, then sixel, then
           half-block-cell fallback
gui/       minimal surface: raster plus blit example. Heavy dependencies stay
           outside the core module path.
icon/      presets: Apple icon grid, corner-radius ratio ~22.37%, standard
           size ladder
cmd/scribe CLI: `scribe render icon.scr -o icon.png --size 1024`
```

**Zero dependencies in the core.** Standard library only (`image`,
`image/png`). The GUI example may use one dependency, quarantined outside the
core packages.

### Unit boundaries

Each package answers: what it does, how you use it, what it depends on.

- `geom` depends on nothing.
- `path` depends on `geom`.
- `raster` depends on `geom` and `path` (consumes flattened paths).
- `paint` depends on `raster` (consumes alpha masks).
- `dl` depends on `geom`/`path` types only; backends (`raster`+`paint`, `svg`,
  `term`) each interpret `dl` independently.
- `ps` produces `dl`; it never touches a backend directly.

## Milestones

| Milestone | Ships | Proves |
|-----------|-------|--------|
| M1 v0.1 | geom + path + raster + paint + PNG + CLI | Lambert model end to end: a rounded-rect icon, pixel-perfect and scale-exact |
| M2 v0.2 | dl + ps | drawing as data; Rune-port shape locked |
| M3 v0.3 | svg + term | one display list rendering to vector and terminal |
| M4 v0.4 | gui + icon | interactive surface plus asset toolkit |
| M5 | Rune port spec | dl as an inductive type; rasterizer as a pure fold or host-op accelerated (bible-ETL pattern) |

Each milestone is its own plan -> implement -> verify cycle, worked in small
green increments (additive first, delete after).

## Testing strategy

- **Golden PNGs**: byte-identical, committed to the repo.
- **Analytic coverage**: a pixel half-covered by an edge lying on a grid line
  must have alpha exactly 128; known geometric areas are checked exactly.
- **Scale-invariance property**: geometry scaled by 2 and rendered at 2x size
  must structurally match the 1x reference.
- **Cross-check** (from M3): SVG output rasterized by a reference tool must
  match scribe's own raster within a perceptual tolerance.
- TDD throughout; conformance tests are consumer-driven.

## Error handling

- Geometry and path construction are total: no invalid states are
  representable (e.g. a Path builder that cannot emit a curve before a
  MoveTo).
- Rasterization clips to the target rect; out-of-bounds geometry is not an
  error.
- `ps` parsing returns structured diagnostics (line, column, expected/got) in
  the human-grade-errors style used across goforge tools.

## Non-goals (v1)

- Text/font rendering (kerning and greeking are acknowledged in the lineage
  but out of scope until the core is proven).
- Gradients, images-as-paint, filters, blend modes beyond source-over.
- GPU acceleration.
- Windowing toolkit; `gui/` is a surface plus an example, not a framework.

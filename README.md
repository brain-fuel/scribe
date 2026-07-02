# scribe

A 2D drawing engine in the lineage of QuickDraw, Quartz, and PostScript.

A metalworker's scriber scratches dimensionless layout lines on metal
before any cut is made. scribe draws the same way: coordinates name the
grid lines between pixels, not the pixels themselves. Corners are
table-driven Bezier curves. Rendering is deterministic: the same input
produces byte-identical pixels on every platform.

Spec: docs/superpowers/specs/2026-07-01-scribe-design.md

## Use

    go run goforge.dev/scribe/cmd/scribe@latest icon -size 1024 -o icon.png

Library:

    c := scribe.NewCanvas(256, 256)
    c.Fill(path.RoundRect(geom.RectXYWH(16, 16, 224, 224), 50, path.Continuous), color)
    c.SavePNG("out.png")

Drawing language (PostScript homage):

    scribe render drawing.scr -size 512 -o out.png

    % drawing.scr
    #ff6a00 setcolor
    0 0 512 512 114.7 continuous roundrect
    fill

Backends: one display list renders everywhere.

    scribe render drawing.scr -svg out.svg          # exact vector SVG
    scribe render drawing.scr -term                 # in your terminal
    scribe render drawing.scr -term -protocol sixel # kitty|sixel|halfblock

## The model

- Coordinates name grid lines between pixels. Pixel (x, y) is the unit
  square from (x, y) to (x+1, y+1). Rect(0,0,4,4) covers exactly 16
  pixels.
- Corners are table-driven cubic Beziers: circular (kappa table) or
  continuous (Apple's G2 squircle table, via PaintCode).
- Coverage antialiasing with exact analytic areas: a pixel half-covered
  by an edge on a grid line has alpha exactly 128.
- Deterministic: identical input gives byte-identical PNGs everywhere.

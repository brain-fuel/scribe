# scribe

A 2D drawing engine in the lineage of QuickDraw, Quartz, and PostScript.

A metalworker's scriber scratches dimensionless layout lines on metal
before any cut is made. scribe draws the same way: coordinates name the
grid lines between pixels, not the pixels themselves. Corners are
table-driven Bezier curves. Rendering is deterministic: the same input
produces byte-identical pixels on every platform.

Spec: docs/superpowers/specs/2026-07-01-scribe-design.md

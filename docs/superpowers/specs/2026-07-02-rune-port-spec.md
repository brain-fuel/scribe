# scribe M5: Rune Port Specification

Status: Spec only (implementation happens in the rune repo)
Depends on: scribe v0.2.0+ (the dl IR and ps codec are the port surface)

## What ports

Not the Go code: the display list. `dl.Program` is the drawing. A rune
scribe is a new interpreter for the same IR, plus the ps text codec so
both worlds exchange drawings as plain text.

## Why rune is a natural fit

1. The IR is an inductive datatype: 20 constructors, no functions, no
   mutation. `Program = List Op`. Interpreters are folds.
2. Every constant in scribe is a decimal literal: the kappa table
   (0.5522847498307936) and the PaintCode continuous-corner table are
   rationals. In rune's exact Q tower the whole geometry pipeline
   (affine transforms, de Casteljau splits, shoelace areas, coverage)
   is EXACT arithmetic: no FMA blocking, no rounding discipline, no
   cross-platform drift by construction. The Go rasterizer works to
   stay deterministic; the rune one cannot be anything else.
3. Coverage is exactly rational, so `alpha = floor(cov * 256)` is an
   exact integer operation. The corner-grid model plus Q makes the
   half-covered-pixel-is-exactly-128 property a theorem, not a test.

## Layers (in dependency order)

- **L1: IR + codec.** `Op` as an inductive type, `parse : Bytes ->
  Result Diagnostic (List Op)`, `print : List Op -> Bytes`. Round-trip
  law `parse (print p) = ok p` is a property test (and a provable
  lemma later). Diagnostics follow rune's human-grade error style
  (line:col, expected/got), same positions as the Go parser.
- **L2: geometry.** DONE (rune v3.357.0, listings/ch562_scribe_geom.rune):
  exact rationals, the corner tables as literal fractions, builders, and a
  worklist flattener; the epsilon assertions of the Go tests are refl proofs
  (exact extent, exact clamped bounds, exact radial band). Stroking deferred
  as planned. Point/Affine over Q, path builder, corner tables,
  adaptive flattening. Flatness test compares squared distances:
  stays in Q, no square roots. (The Go stroke uses sqrt for unit
  normals; the rune stroker offsets with squared-length comparisons
  or defers stroking to a later phase: fills alone prove the model.)
- **L3: rasterizer.** DONE (rune v3.360.0, listings/ch563_scribe_raster.rune):
  font-rs signed-area accumulation in exact rationals; half-coverage =
  alpha 128 is a refl fact; see acceptance 3. Signed-area accumulation as a pure fold over
  edges into a per-row cell map, then a prefix-sum fold to alpha.
  Total, exact, slow: acceptance on small canvases (up to 64x64).
- **L4: host acceleration.** The bible-port pattern: keep the data
  plane total in rune, push the hot loop behind a host op.
  `rasterFill : List Polygon -> Int -> Int -> Bytes` (alpha mask) and
  the existing write-trio host ops for PNG output. The pure L3
  rasterizer stays as the reference the host op must match bit for
  bit (same divergence-lock discipline as the bible ETL).
- **L5: divergence lock.** A shared corpus of .scr files (start with
  scribe's testdata/demo.scr and the golden scenes). CI renders each
  with Go scribe and rune scribe on every backend in the lock; PNGs
  must be byte-identical. This extends the existing 8-way lock
  pattern to a second domain.

## Stroke width caveat

`dl.Stroke` scales pen width by sqrt(abs(det CTM)), the one
irrational operation in the pipeline. Options, decided at L2 time:
(a) restrict the lock corpus to uniform scales where det is a perfect
square of a rational; (b) define width scaling as a Q approximation
of the square root at fixed precision (deterministic, documented);
(c) carry width as width-squared until rasterization. Default
recommendation: (b) with 1/256 precision, matching the subpixel grid.

## Non-goals

- Porting the Go packages function by function.
- SVG/terminal backends in rune (the IR ports; backends can wait).
- Performance parity: L3 is a reference semantics, L4 is the fast path.

## Acceptance

1. L1 round-trip property green in rune REPL and compiled (REPL
   integration is a mandatory acceptance step for rune features).
   DONE (rune v3.356.0, listings/ch561_scribe_dl.rune): the round-trip
   law is a refl proof on an all-constructor program, the codec runs
   compiled (emit gate prints demo.scr's op count), and the whole
   listing loads line by line in rune repl (TestREPLScribeDL).
2. demo.scr parsed by rune, printed, re-parsed: fixed point.
   DONE (same chapter: demoParses and rtDemo, both refl).
3. A 32x32 roundrect fill rendered by L3 matches Go scribe's PNG
   byte for byte.
   DONE: 1024 of 1024 alphas identical (Go tools/lockgen fixture vs
   ch563; permanent gate TestScribeLockL3 in the rune harness). Go's
   disciplined float64 and rune's exact arithmetic agree exactly.
4. Lock CI stage running the corpus on the standard backend set.

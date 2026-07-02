// Package raster converts flattened paths into antialiased coverage
// masks. It uses signed-area accumulation (the font-rs algorithm):
// each edge deposits per-cell area deltas; a per-row prefix sum
// recovers exact analytic coverage.
//
// Determinism: input vertices snap to a 1/256 subpixel grid, and every
// product feeding an addition in the inner loop is wrapped in an
// explicit float64 conversion, which the Go spec defines as blocking
// fused-multiply-add contraction. The same polygons therefore produce
// byte-identical masks on every architecture.
package raster

import (
	"image"
	"math"

	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
)

// FillRule selects how winding numbers map to coverage.
type FillRule uint8

const (
	// NonZero fills where the winding number is nonzero.
	NonZero FillRule = iota
	// EvenOdd fills where the winding number is odd.
	EvenOdd
)

// Rasterizer accumulates polygons and renders coverage masks for a
// w x h pixel canvas covering corner-grid coordinates [0,w] x [0,h].
type Rasterizer struct {
	w, h int
	acc  []float64 // stride w+1: one guard column
}

// New returns a Rasterizer for a w x h canvas.
func New(w, h int) *Rasterizer {
	return &Rasterizer{w: w, h: h, acc: make([]float64, (w+1)*h)}
}

// Reset clears all accumulated coverage.
func (r *Rasterizer) Reset() {
	for i := range r.acc {
		r.acc[i] = 0
	}
}

// snap quantizes a coordinate to the 1/256 subpixel grid. Rounding is
// exact: division by a power of two.
func snap(v float64) float64 { return math.Round(v*256) / 256 }

// AddPolyline adds a subpath. Open polylines are implicitly closed
// (fill semantics). Polylines with fewer than 3 points add nothing.
func (r *Rasterizer) AddPolyline(pl path.Polyline) {
	pts := pl.Pts
	if len(pts) < 3 {
		return
	}
	prev := geom.Pt(snap(pts[0].X), snap(pts[0].Y))
	first := prev
	for _, q := range pts[1:] {
		cur := geom.Pt(snap(q.X), snap(q.Y))
		r.line(prev, cur)
		prev = cur
	}
	r.line(prev, first)
}

// line deposits the signed-area contribution of one edge.
func (r *Rasterizer) line(p0, p1 geom.Point) {
	if p0.Y == p1.Y {
		return
	}
	dir := 1.0
	if p0.Y > p1.Y {
		dir = -1.0
		p0, p1 = p1, p0
	}
	dxdy := (p1.X - p0.X) / (p1.Y - p0.Y)
	x := p0.X
	yStart := p0.Y
	if yStart < 0 {
		x = p0.X + float64(-p0.Y*dxdy)
		yStart = 0
	}
	yEnd := p1.Y
	if yEnd > float64(r.h) {
		yEnd = float64(r.h)
	}
	if yStart >= yEnd {
		return
	}
	stride := r.w + 1
	wf := float64(r.w)
	y0 := int(math.Floor(yStart))
	y1 := int(math.Ceil(yEnd))
	for y := y0; y < y1; y++ {
		rowTop := math.Max(float64(y), yStart)
		rowBot := math.Min(float64(y+1), yEnd)
		dy := rowBot - rowTop
		xNext := x + float64(dxdy*dy)
		d := dy * dir
		x0, x1 := x, xNext
		if x0 > x1 {
			x0, x1 = x1, x0
		}
		// Clamp horizontally. Area beyond the left edge lands in
		// column 0 (winding must still count); beyond the right edge
		// in the guard column (invisible).
		if x0 < 0 {
			x0 = 0
		}
		if x1 < 0 {
			x1 = 0
		}
		if x0 > wf {
			x0 = wf
		}
		if x1 > wf {
			x1 = wf
		}
		row := r.acc[y*stride : (y+1)*stride]
		x0f := math.Floor(x0)
		x0i := int(x0f)
		x1c := math.Ceil(x1)
		x1i := int(x1c)
		if x1i <= x0i+1 {
			// Single cell (plus spill into the next).
			xmf := float64(0.5*(x0+x1)) - x0f
			row[x0i] += float64(d * (1 - xmf))
			if x0i+1 < stride {
				row[x0i+1] += float64(d * xmf)
			}
		} else {
			s := 1 / (x1 - x0)
			x0fr := x0 - x0f
			a0 := float64(0.5*s) * float64((1-x0fr)*(1-x0fr))
			x1fr := x1 - x1c + 1
			am := float64(0.5*s) * float64(x1fr*x1fr)
			row[x0i] += float64(d * a0)
			if x1i == x0i+2 {
				row[x0i+1] += float64(d * (1 - a0 - am))
			} else {
				a1 := float64(s * (1.5 - x0fr))
				row[x0i+1] += float64(d * (a1 - a0))
				for xi := x0i + 2; xi < x1i-1; xi++ {
					row[xi] += float64(d * s)
				}
				a2 := a1 + float64(float64(x1i-x0i-3)*s)
				row[x1i-1] += float64(d * (1 - a2 - am))
			}
			if x1i < stride {
				row[x1i] += float64(d * am)
			}
		}
		x = xNext
	}
}

// Mask renders accumulated coverage to an alpha mask using rule.
// alpha = min(255, floor(coverage * 256)), so half coverage is
// exactly 128.
func (r *Rasterizer) Mask(rule FillRule) *image.Alpha {
	m := image.NewAlpha(image.Rect(0, 0, r.w, r.h))
	stride := r.w + 1
	for y := 0; y < r.h; y++ {
		acc := 0.0
		for x := 0; x < r.w; x++ {
			acc += r.acc[y*stride+x]
			var cov float64
			if rule == NonZero {
				cov = math.Abs(acc)
				if cov > 1 {
					cov = 1
				}
			} else {
				t := math.Mod(math.Abs(acc), 2)
				if t <= 1 {
					cov = t
				} else {
					cov = 2 - t
				}
			}
			a := int(cov * 256)
			if a > 255 {
				a = 255
			}
			m.Pix[y*m.Stride+x] = uint8(a)
		}
	}
	return m
}

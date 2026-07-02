// Package geom provides corner-grid geometry. Coordinates name the
// infinitely thin grid lines between pixels: pixel (x, y) is the unit
// square from Pt(x, y) to Pt(x+1, y+1). Rect width is exactly
// Max.X - Min.X, scaling is exact, and edges on grid lines straddle
// pixels symmetrically.
package geom

import (
	"image"
	"math"
)

// Point is a location on the corner grid.
type Point struct{ X, Y float64 }

// Pt is shorthand for Point{x, y}.
func Pt(x, y float64) Point { return Point{x, y} }

func (p Point) Add(q Point) Point   { return Point{p.X + q.X, p.Y + q.Y} }
func (p Point) Sub(q Point) Point   { return Point{p.X - q.X, p.Y - q.Y} }
func (p Point) Mul(k float64) Point { return Point{p.X * k, p.Y * k} }

// Lerp linearly interpolates from p to q at parameter t.
func (p Point) Lerp(q Point, t float64) Point {
	return Point{p.X + (q.X-p.X)*t, p.Y + (q.Y-p.Y)*t}
}

// Rect is an axis-aligned rectangle on the corner grid.
type Rect struct{ Min, Max Point }

// RectXYWH builds a Rect from origin and size.
func RectXYWH(x, y, w, h float64) Rect {
	return Rect{Point{x, y}, Point{x + w, y + h}}
}

func (r Rect) W() float64 { return r.Max.X - r.Min.X }
func (r Rect) H() float64 { return r.Max.Y - r.Min.Y }

// OuterPixels returns the smallest pixel rectangle containing r.
func (r Rect) OuterPixels() image.Rectangle {
	return image.Rect(
		int(math.Floor(r.Min.X)), int(math.Floor(r.Min.Y)),
		int(math.Ceil(r.Max.X)), int(math.Ceil(r.Max.Y)),
	)
}

// Affine is a 2D affine transform mapping (x, y) to
// (A*x + C*y + E, B*x + D*y + F).
type Affine struct{ A, B, C, D, E, F float64 }

// Identity returns the identity transform.
func Identity() Affine { return Affine{A: 1, D: 1} }

// Translate returns a translation by (tx, ty).
func Translate(tx, ty float64) Affine { return Affine{A: 1, D: 1, E: tx, F: ty} }

// Scale returns a scale by (sx, sy) about the origin.
func Scale(sx, sy float64) Affine { return Affine{A: sx, D: sy} }

// Mul composes transforms: m.Mul(n) applies n first, then m.
func (m Affine) Mul(n Affine) Affine {
	return Affine{
		A: m.A*n.A + m.C*n.B,
		B: m.B*n.A + m.D*n.B,
		C: m.A*n.C + m.C*n.D,
		D: m.B*n.C + m.D*n.D,
		E: m.A*n.E + m.C*n.F + m.E,
		F: m.B*n.E + m.D*n.F + m.F,
	}
}

// Apply transforms p.
func (m Affine) Apply(p Point) Point {
	return Point{m.A*p.X + m.C*p.Y + m.E, m.B*p.X + m.D*p.Y + m.F}
}

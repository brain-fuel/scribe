package path

import (
	"math"
	"testing"

	"goforge.dev/scribe/geom"
)

// Continuous corners span 1.52866483 * radius along each edge.
func TestContinuousExtent(t *testing.T) {
	const rad = 10.0
	p := RoundRect(geom.RectXYWH(0, 0, 100, 100), rad, Continuous)
	pts := p.Points()
	// First point is MoveTo at (x0 + extent*r, y0).
	want := geom.Pt(1.52866483*rad, 0)
	if math.Abs(pts[0].X-want.X) > 1e-9 || pts[0].Y != 0 {
		t.Errorf("start = %v, want %v", pts[0], want)
	}
}

// The shape must stay inside the rect and be 4-fold rotationally
// symmetric about the rect center.
func TestContinuousSymmetryAndContainment(t *testing.T) {
	const rad = 8.0
	p := RoundRect(geom.RectXYWH(0, 0, 64, 64), rad, Continuous)
	polys := p.Flatten(0.01)
	pts := polys[0].Pts
	// containment
	for _, pt := range pts {
		if pt.X < -1e-9 || pt.Y < -1e-9 || pt.X > 64+1e-9 || pt.Y > 64+1e-9 {
			t.Fatalf("point %v escapes rect", pt)
		}
	}
	// Symmetry: rotating any flattened point 90 degrees about (32,32)
	// lands within tolerance of the flattened POLYLINE (segment
	// distance, not vertex distance: flattening chords can be ~0.8
	// apart at tol 0.01, so vertex-to-vertex comparison is wrong).
	rot := func(q geom.Point) geom.Point {
		return geom.Pt(64-q.Y, q.X)
	}
	segDist := func(q, a, b geom.Point) float64 {
		ab := b.Sub(a)
		l2 := ab.X*ab.X + ab.Y*ab.Y
		t0 := 0.0
		if l2 > 0 {
			t0 = ((q.X-a.X)*ab.X + (q.Y-a.Y)*ab.Y) / l2
			if t0 < 0 {
				t0 = 0
			} else if t0 > 1 {
				t0 = 1
			}
		}
		c := a.Lerp(b, t0)
		return math.Hypot(q.X-c.X, q.Y-c.Y)
	}
	for _, pt := range pts {
		r := rot(pt)
		best := math.Inf(1)
		for i := range pts {
			d := segDist(r, pts[i], pts[(i+1)%len(pts)])
			if d < best {
				best = d
			}
		}
		// both boundaries are within 0.01 of the true curve
		if best > 0.025 {
			t.Fatalf("rotated point %v -> %v off boundary by %.4f", pt, r, best)
		}
	}
}

// Radius clamp: a continuous roundrect on a rect too small for the
// extent must still fit exactly.
func TestContinuousClamp(t *testing.T) {
	p := RoundRect(geom.RectXYWH(0, 0, 10, 10), 100, Continuous)
	b := p.Bounds()
	if b.Min != geom.Pt(0, 0) || b.Max != geom.Pt(10, 10) {
		t.Errorf("clamped bounds = %v", b)
	}
}

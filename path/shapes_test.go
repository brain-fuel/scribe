package path

import (
	"math"
	"testing"

	"goforge.dev/scribe/geom"
)

func TestRectPath(t *testing.T) {
	p := RectPath(geom.RectXYWH(1, 2, 3, 4))
	polys := p.Flatten(0.1)
	if len(polys) != 1 || !polys[0].Closed || len(polys[0].Pts) != 4 {
		t.Fatalf("rect flatten = %+v", polys)
	}
}

// Circle from the kappa table: all flattened points within kappa
// approximation error of the true radius.
func TestCircle(t *testing.T) {
	c := geom.Pt(50, 50)
	const r = 40.0
	p := Circle(c, r)
	polys := p.Flatten(0.01)
	if len(polys) != 1 || !polys[0].Closed {
		t.Fatalf("circle flatten shape: %d polys", len(polys))
	}
	for _, pt := range polys[0].Pts {
		radius := math.Hypot(pt.X-c.X, pt.Y-c.Y)
		if math.Abs(radius-r) > 0.01+3e-4*r {
			t.Errorf("point %v radius %.4f", pt, radius)
		}
	}
}

func TestRoundRectDegenerate(t *testing.T) {
	// radius 0 degenerates to plain rect
	p := RoundRect(geom.RectXYWH(0, 0, 10, 10), 0, Circular)
	b := p.Bounds()
	if b.Min != geom.Pt(0, 0) || b.Max != geom.Pt(10, 10) {
		t.Errorf("r=0 bounds = %v", b)
	}
	// radius > min(w,h)/2 clamps: bounds still exact
	p2 := RoundRect(geom.RectXYWH(0, 0, 10, 4), 100, Circular)
	b2 := p2.Bounds()
	if b2.Min != geom.Pt(0, 0) || b2.Max != geom.Pt(10, 4) {
		t.Errorf("clamped bounds = %v", b2)
	}
}

// The rounded rect must be inside the rect, and its corner arc must
// stay within kappa error of the true corner circle.
func TestRoundRectCircularCorner(t *testing.T) {
	const rad = 4.0
	p := RoundRect(geom.RectXYWH(0, 0, 20, 20), rad, Circular)
	polys := p.Flatten(0.01)
	center := geom.Pt(20-rad, rad) // top-right corner circle center
	for _, pt := range polys[0].Pts {
		if pt.X < -1e-9 || pt.Y < -1e-9 || pt.X > 20+1e-9 || pt.Y > 20+1e-9 {
			t.Errorf("point %v escapes rect", pt)
		}
		// points in the top-right corner square must lie on the arc
		if pt.X >= 20-rad && pt.Y <= rad {
			radius := math.Hypot(pt.X-center.X, pt.Y-center.Y)
			if math.Abs(radius-rad) > 0.01+3e-4*rad {
				t.Errorf("corner point %v radius %.5f, want %v", pt, radius, rad)
			}
		}
	}
}

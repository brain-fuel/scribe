package path

import (
	"math"
	"testing"

	"goforge.dev/scribe/geom"
)

func TestFlattenLines(t *testing.T) {
	var p Path
	p.MoveTo(geom.Pt(0, 0))
	p.LineTo(geom.Pt(4, 0))
	p.LineTo(geom.Pt(4, 4))
	p.Close()
	p.MoveTo(geom.Pt(10, 10))
	p.LineTo(geom.Pt(11, 10))
	polys := p.Flatten(0.1)
	if len(polys) != 2 {
		t.Fatalf("polylines = %d, want 2", len(polys))
	}
	if !polys[0].Closed || polys[1].Closed {
		t.Errorf("closed flags = %v, %v", polys[0].Closed, polys[1].Closed)
	}
	if len(polys[0].Pts) != 3 {
		t.Errorf("closed poly pts = %v", polys[0].Pts)
	}
	if len(polys[1].Pts) != 2 {
		t.Errorf("open poly pts = %v", polys[1].Pts)
	}
}

// A cubic that is already a straight line flattens to few points.
func TestFlattenDegenerateCubic(t *testing.T) {
	var p Path
	p.MoveTo(geom.Pt(0, 0))
	p.CubicTo(geom.Pt(1, 0), geom.Pt(2, 0), geom.Pt(3, 0))
	polys := p.Flatten(0.1)
	if len(polys[0].Pts) > 3 {
		t.Errorf("straight cubic over-subdivided: %d points", len(polys[0].Pts))
	}
}

// Quarter-circle cubic (kappa) flattened at tol must stay within
// tol + kappa approximation error of the true circle.
func TestFlattenQuarterCircleAccuracy(t *testing.T) {
	const r = 100.0
	const kappa = 0.5522847498307936
	var p Path
	p.MoveTo(geom.Pt(r, 0))
	p.CubicTo(geom.Pt(r, r*kappa), geom.Pt(r*kappa, r), geom.Pt(0, r))
	const tol = 0.05
	polys := p.Flatten(tol)
	if len(polys[0].Pts) < 8 {
		t.Fatalf("too few points for r=100 arc: %d", len(polys[0].Pts))
	}
	// kappa cubic max radial error is about 2.7e-4 * r
	maxErr := tol + 3e-4*r
	for _, pt := range polys[0].Pts {
		radius := math.Hypot(pt.X, pt.Y)
		if math.Abs(radius-r) > maxErr {
			t.Errorf("point %v radius %.4f deviates > %.4f", pt, radius, maxErr)
		}
	}
}

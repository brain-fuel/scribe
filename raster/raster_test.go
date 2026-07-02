package raster

import (
	"bytes"
	"testing"

	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
)

func poly(closed bool, pts ...geom.Point) path.Polyline {
	return path.Polyline{Pts: pts, Closed: closed}
}

func rectPoly(x, y, w, h float64) path.Polyline {
	return poly(true,
		geom.Pt(x, y), geom.Pt(x+w, y), geom.Pt(x+w, y+h), geom.Pt(x, y+h))
}

func TestFillFullPixels(t *testing.T) {
	r := New(4, 4)
	r.AddPolyline(rectPoly(1, 1, 2, 2))
	m := r.Mask(NonZero)
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			want := uint8(0)
			if x >= 1 && x < 3 && y >= 1 && y < 3 {
				want = 255
			}
			if got := m.AlphaAt(x, y).A; got != want {
				t.Errorf("(%d,%d) = %d, want %d", x, y, got, want)
			}
		}
	}
}

// Grid-line edge through pixel middle: exactly half coverage, alpha 128.
func TestHalfCoverageHorizontal(t *testing.T) {
	r := New(4, 3)
	r.AddPolyline(rectPoly(0, 0, 4, 2.5))
	m := r.Mask(NonZero)
	for x := 0; x < 4; x++ {
		if got := m.AlphaAt(x, 1).A; got != 255 {
			t.Errorf("full row (%d,1) = %d", x, got)
		}
		if got := m.AlphaAt(x, 2).A; got != 128 {
			t.Errorf("half row (%d,2) = %d, want 128", x, got)
		}
	}
}

func TestHalfCoverageVertical(t *testing.T) {
	r := New(2, 1)
	r.AddPolyline(rectPoly(0.5, 0, 1, 1))
	m := r.Mask(NonZero)
	if got := m.AlphaAt(0, 0).A; got != 128 {
		t.Errorf("(0,0) = %d, want 128", got)
	}
	if got := m.AlphaAt(1, 0).A; got != 128 {
		t.Errorf("(1,0) = %d, want 128", got)
	}
}

// Diagonal edge: triangle (0,0),(1,0),(0,1) half-covers pixel (0,0).
func TestDiagonalHalfPixel(t *testing.T) {
	r := New(1, 1)
	r.AddPolyline(poly(true, geom.Pt(0, 0), geom.Pt(1, 0), geom.Pt(0, 1)))
	m := r.Mask(NonZero)
	if got := m.AlphaAt(0, 0).A; got != 128 {
		t.Errorf("alpha = %d, want 128", got)
	}
}

// Quarter coverage.
func TestQuarterCoverage(t *testing.T) {
	r := New(1, 1)
	r.AddPolyline(rectPoly(0, 0, 0.5, 0.5))
	m := r.Mask(NonZero)
	if got := m.AlphaAt(0, 0).A; got != 64 {
		t.Errorf("alpha = %d, want 64", got)
	}
}

// Fill rules: two same-direction overlapping rects.
func TestFillRules(t *testing.T) {
	build := func() *Rasterizer {
		r := New(6, 2)
		r.AddPolyline(rectPoly(0, 0, 4, 2))
		r.AddPolyline(rectPoly(2, 0, 4, 2))
		return r
	}
	nz := build().Mask(NonZero)
	if got := nz.AlphaAt(3, 1).A; got != 255 {
		t.Errorf("nonzero overlap = %d, want 255", got)
	}
	eo := build().Mask(EvenOdd)
	if got := eo.AlphaAt(3, 1).A; got != 0 {
		t.Errorf("evenodd overlap = %d, want 0", got)
	}
	if got := eo.AlphaAt(1, 1).A; got != 255 {
		t.Errorf("evenodd single = %d, want 255", got)
	}
}

// Geometry escaping the canvas clips without artifacts: winding from
// off-canvas left edges must still count.
func TestClipping(t *testing.T) {
	r := New(2, 2)
	r.AddPolyline(rectPoly(-10, -10, 20, 20)) // covers whole canvas
	m := r.Mask(NonZero)
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			if got := m.AlphaAt(x, y).A; got != 255 {
				t.Errorf("(%d,%d) = %d, want 255", x, y, got)
			}
		}
	}
}

// Same input twice: byte-identical masks.
func TestDeterminism(t *testing.T) {
	render := func() []byte {
		r := New(32, 32)
		r.AddPolyline(poly(true,
			geom.Pt(1.13, 2.71), geom.Pt(30.99, 4.5),
			geom.Pt(17.2, 30.001), geom.Pt(2.0001, 20.5)))
		return r.Mask(NonZero).Pix
	}
	if !bytes.Equal(render(), render()) {
		t.Error("two renders differ")
	}
}

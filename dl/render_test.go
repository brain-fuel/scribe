package dl

import (
	"bytes"
	"image/color"
	"testing"

	"goforge.dev/scribe"
	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
)

var red = color.RGBA{R: 255, A: 255}

func pix(c *scribe.Canvas) []byte { return c.Image().Pix }

// A dl program must produce byte-identical pixels to the equivalent
// direct Canvas calls.
func TestRenderEquivalentToCanvas(t *testing.T) {
	prog := Program{
		SetColor{red},
		MoveTo{geom.Pt(1, 1)},
		LineTo{geom.Pt(9, 1)},
		CubicTo{geom.Pt(11, 1), geom.Pt(11, 9), geom.Pt(9, 9)},
		LineTo{geom.Pt(1, 9)},
		ClosePath{},
		Fill{},
	}
	got := scribe.NewCanvas(12, 12)
	Render(prog, got)

	want := scribe.NewCanvas(12, 12)
	var p path.Path
	p.MoveTo(geom.Pt(1, 1))
	p.LineTo(geom.Pt(9, 1))
	p.CubicTo(geom.Pt(11, 1), geom.Pt(11, 9), geom.Pt(9, 9))
	p.LineTo(geom.Pt(1, 9))
	p.Close()
	want.Fill(&p, red)

	if !bytes.Equal(pix(got), pix(want)) {
		t.Error("dl render differs from direct canvas calls")
	}
}

// fill consumes the current path: a second fill paints nothing.
func TestPaintConsumesPath(t *testing.T) {
	prog := Program{
		SetColor{red},
		MoveTo{geom.Pt(0, 0)}, LineTo{geom.Pt(4, 0)},
		LineTo{geom.Pt(4, 4)}, LineTo{geom.Pt(0, 4)}, ClosePath{},
		Fill{},
		SetColor{color.RGBA{G: 255, A: 255}},
		Fill{}, // empty path: no-op
	}
	c := scribe.NewCanvas(4, 4)
	Render(prog, c)
	if got := c.Image().RGBAAt(2, 2); got != red {
		t.Errorf("second fill repainted: %v", got)
	}
}

func TestStrokeAndState(t *testing.T) {
	prog := Program{
		SetColor{red},
		SetLineWidth{2},
		SetCap{path.ButtCap},
		SetJoin{path.BevelJoin},
		MoveTo{geom.Pt(2, 5)}, LineTo{geom.Pt(8, 5)},
		Stroke{},
	}
	got := scribe.NewCanvas(10, 10)
	Render(prog, got)

	want := scribe.NewCanvas(10, 10)
	var p path.Path
	p.MoveTo(geom.Pt(2, 5))
	p.LineTo(geom.Pt(8, 5))
	want.Stroke(&p, path.Pen{Width: 2, Cap: path.ButtCap, Join: path.BevelJoin}, red)

	if !bytes.Equal(pix(got), pix(want)) {
		t.Error("dl stroke differs from direct canvas stroke")
	}
}

// Default color is opaque black, default width 1.
func TestDefaults(t *testing.T) {
	prog := Program{
		MoveTo{geom.Pt(0, 0)}, LineTo{geom.Pt(2, 0)},
		LineTo{geom.Pt(2, 2)}, LineTo{geom.Pt(0, 2)}, ClosePath{},
		Fill{},
	}
	c := scribe.NewCanvas(2, 2)
	Render(prog, c)
	if got := c.Image().RGBAAt(1, 1); (got != color.RGBA{A: 255}) {
		t.Errorf("default fill = %v, want opaque black", got)
	}
}

func TestNewPathDiscards(t *testing.T) {
	prog := Program{
		SetColor{red},
		MoveTo{geom.Pt(0, 0)}, LineTo{geom.Pt(4, 0)},
		LineTo{geom.Pt(4, 4)}, LineTo{geom.Pt(0, 4)}, ClosePath{},
		NewPath{},
		Fill{},
	}
	c := scribe.NewCanvas(4, 4)
	Render(prog, c)
	if got := c.Image().RGBAAt(2, 2); (got != color.RGBA{}) {
		t.Errorf("newpath did not discard: %v", got)
	}
}

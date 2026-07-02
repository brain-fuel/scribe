package dl

import (
	"bytes"
	"testing"

	"goforge.dev/scribe"
	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
)

// Shape ops must match direct path construction exactly.
func TestShapesEquivalent(t *testing.T) {
	prog := Program{
		SetColor{red},
		RoundRect{geom.RectXYWH(2, 2, 28, 28), 6, path.Continuous},
		Fill{},
		Circle{geom.Pt(16, 16), 8},
		FillEO{},
	}
	got := scribe.NewCanvas(32, 32)
	Render(prog, got)

	want := scribe.NewCanvas(32, 32)
	want.Fill(path.RoundRect(geom.RectXYWH(2, 2, 28, 28), 6, path.Continuous), red)
	want.FillEvenOdd(path.Circle(geom.Pt(16, 16), 8), red)

	if !bytes.Equal(pix(got), pix(want)) {
		t.Error("shape ops differ from direct construction")
	}
}

// translate/scale apply to later ops; save/restore bracket state.
func TestTransformAndSaveRestore(t *testing.T) {
	prog := Program{
		SetColor{red},
		Save{},
		Translate{4, 4},
		Scale{2, 2},
		Rect{geom.RectXYWH(0, 0, 2, 2)}, // device: (4,4)-(8,8)
		Fill{},
		Restore{},
		Rect{geom.RectXYWH(0, 0, 2, 2)}, // device: (0,0)-(2,2)
		Fill{},
	}
	got := scribe.NewCanvas(10, 10)
	Render(prog, got)

	want := scribe.NewCanvas(10, 10)
	want.Fill(path.RectPath(geom.RectXYWH(4, 4, 4, 4)), red)
	want.Fill(path.RectPath(geom.RectXYWH(0, 0, 2, 2)), red)

	if !bytes.Equal(pix(got), pix(want)) {
		t.Error("transform or save/restore semantics wrong")
	}
}

// Stroke width is user-space: under scale 2 a width-1 stroke is 2 wide.
func TestStrokeWidthScales(t *testing.T) {
	prog := Program{
		SetColor{red},
		Scale{2, 2},
		SetLineWidth{1},
		MoveTo{geom.Pt(1, 2.5)}, LineTo{geom.Pt(4, 2.5)},
		Stroke{},
	}
	got := scribe.NewCanvas(10, 10)
	Render(prog, got)

	want := scribe.NewCanvas(10, 10)
	var p path.Path
	p.MoveTo(geom.Pt(2, 5))
	p.LineTo(geom.Pt(8, 5))
	want.Stroke(&p, path.Pen{Width: 2}, red)

	if !bytes.Equal(pix(got), pix(want)) {
		t.Error("stroke width did not scale with CTM")
	}
}

// Restore with empty stack is a no-op, not a crash.
func TestRestoreUnderflowTotal(t *testing.T) {
	c := scribe.NewCanvas(2, 2)
	Render(Program{Restore{}, Restore{}, Clear{red}}, c)
	if got := c.Image().RGBAAt(0, 0); got != red {
		t.Errorf("after restores, clear = %v", got)
	}
}

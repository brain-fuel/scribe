package ps

import (
	"image/color"
	"reflect"
	"testing"

	"goforge.dev/scribe/dl"
	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
)

func fullProgram() dl.Program {
	return dl.Program{
		dl.Clear{C: color.RGBA{R: 1, G: 2, B: 3, A: 0xFF}},
		dl.Save{},
		dl.Translate{X: 2, Y: 3},
		dl.Scale{X: 1.5, Y: 1.5},
		dl.SetColor{C: color.RGBA{R: 0xAB, G: 0xCD, B: 0xEF, A: 0x80}},
		dl.SetLineWidth{W: 2.25},
		dl.SetCap{C: path.SquareCap},
		dl.SetJoin{J: path.RoundJoin},
		dl.MoveTo{P: geom.Pt(1, 2)},
		dl.LineTo{P: geom.Pt(3, 4)},
		dl.CubicTo{C1: geom.Pt(5, 6), C2: geom.Pt(7, 8), P: geom.Pt(9, 10)},
		dl.ClosePath{},
		dl.Stroke{},
		dl.Restore{},
		dl.NewPath{},
		dl.Rect{R: geom.RectXYWH(1, 1, 4, 4)},
		dl.FillEO{},
		dl.Circle{C: geom.Pt(5, 5), Radius: 2},
		dl.Fill{},
		dl.RoundRect{R: geom.RectXYWH(0, 0, 8, 8), Radius: 2, Style: path.Continuous},
		dl.Fill{},
	}
}

func TestRoundTrip(t *testing.T) {
	orig := fullProgram()
	text := Print(orig)
	back, err := Parse(text)
	if err != nil {
		t.Fatalf("Parse(Print(prog)): %v\ntext:\n%s", err, text)
	}
	if !reflect.DeepEqual(back, orig) {
		t.Errorf("round trip mismatch\ntext:\n%s\ngot  %#v\nwant %#v", text, back, orig)
	}
}

func TestPrintCanonical(t *testing.T) {
	got := Print(dl.Program{
		dl.SetColor{C: color.RGBA{R: 0xFF, G: 0x6A, A: 0xFF}},
		dl.SetColor{C: color.RGBA{R: 0xFF, A: 0x80}},
		dl.RoundRect{R: geom.RectXYWH(0, 0, 512, 512), Radius: 114.7, Style: path.Continuous},
		dl.Fill{},
	})
	want := "#ff6a00 setcolor\n" +
		"#ff000080 setcolor\n" +
		"0 0 512 512 114.7 continuous roundrect\n" +
		"fill\n"
	if got != want {
		t.Errorf("Print = %q, want %q", got, want)
	}
}

func TestPrintEmpty(t *testing.T) {
	if got := Print(nil); got != "" {
		t.Errorf("Print(nil) = %q", got)
	}
}

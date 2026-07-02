package ps

import (
	"image/color"
	"reflect"
	"testing"

	"goforge.dev/scribe/dl"
	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
)

func TestParseBasics(t *testing.T) {
	src := `
% a drawing
#ff6a00 setcolor
0 0 512 512 114.7 continuous roundrect
fill
`
	prog, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	want := dl.Program{
		dl.SetColor{C: color.RGBA{R: 0xFF, G: 0x6A, B: 0x00, A: 0xFF}},
		dl.RoundRect{R: geom.RectXYWH(0, 0, 512, 512), Radius: 114.7, Style: path.Continuous},
		dl.Fill{},
	}
	if !reflect.DeepEqual(prog, want) {
		t.Errorf("prog = %#v\nwant %#v", prog, want)
	}
}

func TestParseAllOps(t *testing.T) {
	src := `#102030ff clear
gsave
2 3 translate 1.5 1.5 scale
#00ff00 setcolor 2 setlinewidth round setcap bevel setjoin
1 2 moveto 3 4 lineto 5 6 7 8 9 10 curveto closepath
stroke
grestore
newpath
1 1 4 4 rect eofill
5 5 2 circle fill
0 0 8 8 2 circular roundrect fill`
	prog, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	want := dl.Program{
		dl.Clear{C: color.RGBA{R: 0x10, G: 0x20, B: 0x30, A: 0xFF}},
		dl.Save{},
		dl.Translate{X: 2, Y: 3},
		dl.Scale{X: 1.5, Y: 1.5},
		dl.SetColor{C: color.RGBA{G: 0xFF, A: 0xFF}},
		dl.SetLineWidth{W: 2},
		dl.SetCap{C: path.RoundCap},
		dl.SetJoin{J: path.BevelJoin},
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
		dl.RoundRect{R: geom.RectXYWH(0, 0, 8, 8), Radius: 2, Style: path.Circular},
		dl.Fill{},
	}
	if !reflect.DeepEqual(prog, want) {
		t.Errorf("prog mismatch\ngot  %#v\nwant %#v", prog, want)
	}
}

func TestParseNegativeAndScientific(t *testing.T) {
	prog, err := Parse(`-1.5 2e2 moveto`)
	if err != nil {
		t.Fatal(err)
	}
	want := dl.Program{dl.MoveTo{P: geom.Pt(-1.5, 200)}}
	if !reflect.DeepEqual(prog, want) {
		t.Errorf("prog = %#v", prog)
	}
}

func TestParseEmpty(t *testing.T) {
	prog, err := Parse("  % nothing but a comment\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(prog) != 0 {
		t.Errorf("prog = %#v, want empty", prog)
	}
}

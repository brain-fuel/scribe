package svg

import (
	"bytes"
	"encoding/xml"
	"image/color"
	"strings"
	"testing"

	"goforge.dev/scribe/dl"
	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
)

func encode(t *testing.T, prog dl.Program) string {
	t.Helper()
	var buf bytes.Buffer
	if err := Encode(&buf, prog, 64, 64); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestWellFormedXML(t *testing.T) {
	out := encode(t, dl.Program{
		dl.Clear{C: color.RGBA{R: 16, G: 18, B: 26, A: 255}},
		dl.SetColor{C: color.RGBA{R: 255, G: 106, A: 255}},
		dl.RoundRect{R: geom.RectXYWH(8, 8, 48, 48), Radius: 10, Style: path.Continuous},
		dl.Fill{},
	})
	d := xml.NewDecoder(strings.NewReader(out))
	for {
		_, err := d.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			t.Fatalf("not well-formed: %v\n%s", err, out)
		}
	}
	if !strings.HasPrefix(out, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"`) {
		t.Errorf("header wrong:\n%s", out)
	}
}

// Curves must survive as C commands (exact geometry, no flattening).
func TestCurvesSurvive(t *testing.T) {
	out := encode(t, dl.Program{
		dl.RoundRect{R: geom.RectXYWH(0, 0, 64, 64), Radius: 16, Style: path.Circular},
		dl.Fill{},
	})
	if !strings.Contains(out, "C") {
		t.Error("no cubic commands in output")
	}
	if strings.Count(out, "C") < 4 {
		t.Errorf("want at least 4 cubics (4 corners), got %d", strings.Count(out, "C"))
	}
}

func TestFillAttributes(t *testing.T) {
	out := encode(t, dl.Program{
		dl.SetColor{C: color.RGBA{R: 0x33, G: 0x66, B: 0x99, A: 0x80}},
		dl.Rect{R: geom.RectXYWH(1, 1, 4, 4)},
		dl.FillEO{},
	})
	for _, want := range []string{
		`fill="#336699"`, `fill-rule="evenodd"`, `fill-opacity="0.501961"`,
		`M 1 1`, `L 5 1`, `Z`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

// Transforms bake into coordinates; stroke width scales like dl.Render.
func TestStrokeTransformed(t *testing.T) {
	out := encode(t, dl.Program{
		dl.SetColor{C: color.RGBA{R: 255, A: 255}},
		dl.Scale{X: 2, Y: 2},
		dl.SetLineWidth{W: 3},
		dl.SetCap{C: path.RoundCap},
		dl.SetJoin{J: path.RoundJoin},
		dl.MoveTo{P: geom.Pt(1, 2)},
		dl.LineTo{P: geom.Pt(5, 2)},
		dl.Stroke{},
	})
	for _, want := range []string{
		`M 2 4`, `L 10 4`, `stroke-width="6"`,
		`stroke-linecap="round"`, `stroke-linejoin="round"`,
		`fill="none"`, `stroke="#ff0000"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

// Paint consumes the path: second fill emits no second path element.
func TestPaintConsumes(t *testing.T) {
	out := encode(t, dl.Program{
		dl.Rect{R: geom.RectXYWH(0, 0, 2, 2)},
		dl.Fill{},
		dl.Fill{},
	})
	if got := strings.Count(out, "<path"); got != 1 {
		t.Errorf("path elements = %d, want 1\n%s", got, out)
	}
}

func TestClear(t *testing.T) {
	out := encode(t, dl.Program{dl.Clear{C: color.RGBA{R: 1, G: 2, B: 3, A: 255}}})
	if !strings.Contains(out, `<rect width="64" height="64" fill="#010203"/>`) {
		t.Errorf("clear wrong:\n%s", out)
	}
}

package scribe

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
)

func scenes() map[string]*Canvas {
	m := map[string]*Canvas{}

	c := NewCanvas(64, 64)
	c.Fill(path.Circle(geom.Pt(32, 32), 24), color.RGBA{R: 220, G: 90, B: 30, A: 255})
	m["circle"] = c

	c = NewCanvas(128, 128)
	c.Fill(path.RoundRect(geom.RectXYWH(8, 8, 112, 112), 24, path.Circular),
		color.RGBA{R: 30, G: 90, B: 220, A: 255})
	m["roundrect_circular"] = c

	c = NewCanvas(128, 128)
	c.Fill(path.RoundRect(geom.RectXYWH(8, 8, 112, 112), 24, path.Continuous),
		color.RGBA{R: 30, G: 160, B: 90, A: 255})
	m["roundrect_continuous"] = c

	c = NewCanvas(96, 96)
	var zig path.Path
	zig.MoveTo(geom.Pt(12, 72))
	zig.LineTo(geom.Pt(36, 24))
	zig.LineTo(geom.Pt(60, 72))
	zig.LineTo(geom.Pt(84, 24))
	c.Stroke(&zig, path.Pen{Width: 8, Cap: path.RoundCap, Join: path.RoundJoin},
		color.RGBA{R: 120, G: 60, B: 200, A: 255})
	m["stroke_zigzag"] = c

	return m
}

func TestGolden(t *testing.T) {
	regen := os.Getenv("REGEN_GOLDEN") != ""
	for name, c := range scenes() {
		golden := filepath.Join("testdata", "golden", name+".png")
		if regen {
			if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := c.SavePNG(golden); err != nil {
				t.Fatal(err)
			}
			t.Logf("regenerated %s", golden)
			continue
		}
		// Byte comparison of encoded PNGs. Comparing decoded pixels
		// against the in-memory premultiplied image is wrong: PNG
		// stores non-premultiplied color, and unpremultiplying is
		// lossy (plus or minus one) on semi-transparent AA pixels.
		// Byte equality is also the stronger determinism claim.
		wantBytes, err := os.ReadFile(golden)
		if err != nil {
			t.Fatalf("%s: %v (run REGEN_GOLDEN=1 go test to create)", name, err)
		}
		var buf bytes.Buffer
		if err := c.WritePNG(&buf); err != nil {
			t.Fatal(err)
		}
		if bytes.Equal(buf.Bytes(), wantBytes) {
			continue
		}
		// Diagnostics: decode both, count differing pixels.
		want, err := png.Decode(bytes.NewReader(wantBytes))
		if err != nil {
			t.Fatalf("%s: golden bytes differ and golden is undecodable: %v", name, err)
		}
		got, err := png.Decode(bytes.NewReader(buf.Bytes()))
		if err != nil {
			t.Fatal(err)
		}
		wb, gb := want.Bounds(), got.Bounds()
		if wb != gb {
			t.Errorf("%s: bounds %v != %v", name, gb, wb)
			continue
		}
		diff := 0
		for y := gb.Min.Y; y < gb.Max.Y; y++ {
			for x := gb.Min.X; x < gb.Max.X; x++ {
				r1, g1, b1, a1 := got.At(x, y).RGBA()
				r2, g2, b2, a2 := want.At(x, y).RGBA()
				if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
					diff++
				}
			}
		}
		t.Errorf("%s: PNG bytes differ from golden (%d pixels differ)", name, diff)
	}
}

// Scale invariance: rendering geometry scaled 2x on a 2x canvas, then
// box-downsampling, must closely match the 1x render. The corner-grid
// model makes geometry scaling exact; only AA sampling differs.
func TestScaleInvariance(t *testing.T) {
	render := func(scale float64) *image.RGBA {
		s := int(128 * scale)
		c := NewCanvas(s, s)
		p := path.RoundRect(geom.RectXYWH(8*scale, 8*scale, 112*scale, 112*scale),
			24*scale, path.Continuous)
		c.Fill(p, color.RGBA{R: 255, A: 255})
		return c.Image()
	}
	one := render(1)
	two := render(2)
	var sumDiff, maxDiff, n int
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			// box downsample 2x2 alpha
			var a int
			for dy := 0; dy < 2; dy++ {
				for dx := 0; dx < 2; dx++ {
					a += int(two.RGBAAt(2*x+dx, 2*y+dy).A)
				}
			}
			a = (a + 2) / 4
			d := a - int(one.RGBAAt(x, y).A)
			if d < 0 {
				d = -d
			}
			sumDiff += d
			if d > maxDiff {
				maxDiff = d
			}
			n++
		}
	}
	if maxDiff > 32 {
		t.Errorf("max per-pixel diff = %d, want <= 32", maxDiff)
	}
	if mean := float64(sumDiff) / float64(n); mean > 1.0 {
		t.Errorf("mean abs diff = %.3f, want <= 1.0", mean)
	}
}

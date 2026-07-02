package scribe

import (
	"bytes"
	"image/color"
	"image/png"
	"testing"

	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
)

var red = color.RGBA{R: 255, A: 255}

func TestCanvasFillRect(t *testing.T) {
	c := NewCanvas(4, 4)
	c.Fill(path.RectPath(geom.RectXYWH(1, 1, 2, 2)), red)
	img := c.Image()
	if got := img.RGBAAt(2, 2); got != red {
		t.Errorf("inside = %v", got)
	}
	if got := img.RGBAAt(0, 0); (got != color.RGBA{}) {
		t.Errorf("outside = %v", got)
	}
}

func TestCanvasClear(t *testing.T) {
	c := NewCanvas(2, 2)
	white := color.RGBA{255, 255, 255, 255}
	c.Clear(white)
	if got := c.Image().RGBAAt(1, 1); got != white {
		t.Errorf("clear = %v", got)
	}
}

// Antialiased composite: half-covered pixel over white background
// blends source over destination.
func TestCanvasAABlend(t *testing.T) {
	c := NewCanvas(1, 1)
	c.Clear(color.RGBA{255, 255, 255, 255})
	c.Fill(path.RectPath(geom.RectXYWH(0, 0, 1, 0.5)), color.RGBA{A: 255}) // black half
	got := c.Image().RGBAAt(0, 0)
	// coverage 128/255: R = 255 - 255*128/255 = 127
	if got.R != 127 || got.A != 255 {
		t.Errorf("blend = %v, want R=127 A=255", got)
	}
}

func TestCanvasStroke(t *testing.T) {
	c := NewCanvas(10, 10)
	var p path.Path
	p.MoveTo(geom.Pt(2, 5))
	p.LineTo(geom.Pt(8, 5))
	c.Stroke(&p, path.Pen{Width: 2, Cap: path.ButtCap}, red)
	img := c.Image()
	// stroke of width 2 centered on grid line y=5 fully covers pixel
	// rows 4 and 5 between x=2 and x=8
	if got := img.RGBAAt(5, 4); got != red {
		t.Errorf("row 4 = %v", got)
	}
	if got := img.RGBAAt(5, 5); got != red {
		t.Errorf("row 5 = %v", got)
	}
	if got := img.RGBAAt(5, 3); (got != color.RGBA{}) {
		t.Errorf("row 3 = %v, want empty", got)
	}
}

func TestWritePNGDeterministic(t *testing.T) {
	render := func() []byte {
		c := NewCanvas(16, 16)
		c.Fill(path.Circle(geom.Pt(8, 8), 6), red)
		var buf bytes.Buffer
		if err := c.WritePNG(&buf); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	}
	a, b := render(), render()
	if !bytes.Equal(a, b) {
		t.Error("PNG bytes differ between renders")
	}
	img, err := png.Decode(bytes.NewReader(a))
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != 16 {
		t.Errorf("decoded width = %d", img.Bounds().Dx())
	}
}

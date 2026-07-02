package gui

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"testing"

	"goforge.dev/scribe/dl"
	"goforge.dev/scribe/geom"
)

type fakeSurface struct{ last *image.RGBA }

func (f *fakeSurface) Blit(img *image.RGBA) { f.last = img }

var _ Surface = (*fakeSurface)(nil)

func movingDot(frame int) dl.Program {
	return dl.Program{
		dl.Clear{C: color.RGBA{A: 255}},
		dl.SetColor{C: color.RGBA{R: 255, A: 255}},
		dl.Circle{C: geom.Pt(4+float64(frame*4), 8), Radius: 3},
		dl.Fill{},
	}
}

func TestWriteGIF(t *testing.T) {
	var buf bytes.Buffer
	o := GIFOptions{W: 24, H: 16, Frames: 3, DelayCS: 5}
	if err := WriteGIF(&buf, o, movingDot); err != nil {
		t.Fatal(err)
	}
	g, err := gif.DecodeAll(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Image) != 3 {
		t.Fatalf("frames = %d, want 3", len(g.Image))
	}
	for i, fr := range g.Image {
		if fr.Bounds().Dx() != 24 || fr.Bounds().Dy() != 16 {
			t.Errorf("frame %d bounds = %v", i, fr.Bounds())
		}
		if g.Delay[i] != 5 {
			t.Errorf("frame %d delay = %d", i, g.Delay[i])
		}
	}
	// The dot moves: frame 0 and frame 2 differ.
	if bytes.Equal(g.Image[0].Pix, g.Image[2].Pix) {
		t.Error("frames identical; animation did not animate")
	}
}

func TestWriteGIFDeterministic(t *testing.T) {
	render := func() []byte {
		var buf bytes.Buffer
		if err := WriteGIF(&buf, GIFOptions{W: 16, H: 16, Frames: 2, DelayCS: 10}, movingDot); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	}
	if !bytes.Equal(render(), render()) {
		t.Error("GIF output not deterministic")
	}
}

// Package gui is scribe's surface layer: the integration point
// between display lists and anything that can show pixels. It is a
// surface, not a framework.
//
// Integrating with a windowing toolkit is one Blit call per frame.
// With ebiten (github.com/hajimehoshi/ebiten), for example:
//
//	func (g *game) Draw(screen *ebiten.Image) {
//	    c := scribe.NewCanvas(w, h)
//	    dl.Render(g.frame(), c)
//	    screen.WritePixels(c.Image().Pix)
//	}
//
// scribe stays zero-dependency; the toolkit lives in your module.
// For a testable motion target without a display, WriteGIF renders a
// frame function to an animated GIF deterministically.
package gui

import (
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"io"

	"goforge.dev/scribe"
	"goforge.dev/scribe/dl"
)

// Surface is anything that can display a frame.
type Surface interface {
	Blit(img *image.RGBA)
}

// GIFOptions sizes an animation.
type GIFOptions struct {
	W, H    int
	Frames  int
	DelayCS int // per-frame delay in centiseconds
}

// WriteGIF renders draw(0..Frames-1) through the display-list
// interpreter into an animated GIF. Quantization uses the Plan9
// palette with Floyd-Steinberg dithering; output is deterministic.
func WriteGIF(w io.Writer, o GIFOptions, drawFrame func(frame int) dl.Program) error {
	anim := &gif.GIF{}
	for i := 0; i < o.Frames; i++ {
		c := scribe.NewCanvas(o.W, o.H)
		dl.Render(drawFrame(i), c)
		pal := image.NewPaletted(image.Rect(0, 0, o.W, o.H), palette.Plan9)
		draw.FloydSteinberg.Draw(pal, pal.Bounds(), c.Image(), image.Point{})
		anim.Image = append(anim.Image, pal)
		anim.Delay = append(anim.Delay, o.DelayCS)
	}
	return gif.EncodeAll(w, anim)
}

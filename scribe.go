// Package scribe is a 2D drawing engine in the lineage of QuickDraw,
// Quartz, and PostScript. Coordinates name the grid lines between
// pixels: a Canvas of size w x h covers corner-grid coordinates
// [0,w] x [0,h], and pixel (x, y) is the unit square from (x, y) to
// (x+1, y+1). Rendering is deterministic: identical input produces
// byte-identical pixels on every platform.
package scribe

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"os"

	"goforge.dev/scribe/paint"
	"goforge.dev/scribe/path"
	"goforge.dev/scribe/raster"
)

// FlattenTol is the curve flattening tolerance in pixels used by
// Canvas operations.
const FlattenTol = 0.05

// Canvas is a pixel canvas with corner-grid coordinates.
type Canvas struct {
	img *image.RGBA
	ras *raster.Rasterizer
}

// NewCanvas returns a transparent w x h canvas.
func NewCanvas(w, h int) *Canvas {
	return &Canvas{
		img: image.NewRGBA(image.Rect(0, 0, w, h)),
		ras: raster.New(w, h),
	}
}

// Image returns the backing image (shared, not a copy).
func (c *Canvas) Image() *image.RGBA { return c.img }

// Clear fills the whole canvas with col.
func (c *Canvas) Clear(col color.Color) {
	draw.Draw(c.img, c.img.Bounds(), paint.Solid(col), image.Point{}, draw.Src)
}

// Fill fills p with col using the nonzero winding rule.
func (c *Canvas) Fill(p *path.Path, col color.Color) {
	c.fill(p, col, raster.NonZero)
}

// FillEvenOdd fills p with col using the even-odd rule.
func (c *Canvas) FillEvenOdd(p *path.Path, col color.Color) {
	c.fill(p, col, raster.EvenOdd)
}

func (c *Canvas) fill(p *path.Path, col color.Color, rule raster.FillRule) {
	c.ras.Reset()
	for _, pl := range p.Flatten(FlattenTol) {
		c.ras.AddPolyline(pl)
	}
	mask := c.ras.Mask(rule)
	draw.DrawMask(c.img, c.img.Bounds(), paint.Solid(col), image.Point{},
		mask, image.Point{}, draw.Over)
}

// WritePNG encodes the canvas as PNG.
func (c *Canvas) WritePNG(w io.Writer) error { return png.Encode(w, c.img) }

// SavePNG writes the canvas to a PNG file.
func (c *Canvas) SavePNG(name string) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	if err := c.WritePNG(f); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

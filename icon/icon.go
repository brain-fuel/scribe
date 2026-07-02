// Package icon renders app-icon assets from presets: rounded-rect
// plates on the corner grid, the Apple corner-radius ratio, and the
// standard size ladder.
package icon

import (
	"fmt"
	"image/color"
	"path/filepath"

	"goforge.dev/scribe"
	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
)

// AppleRadiusRatio approximates the Apple app-icon corner radius as a
// fraction of icon size.
const AppleRadiusRatio = 0.2237

// AppleSizes is the standard icon size ladder.
var AppleSizes = []int{16, 32, 64, 128, 256, 512, 1024}

// Spec describes an icon plate. Radius <= 0 means the Apple ratio of
// the size.
type Spec struct {
	Size   int
	Radius float64
	Style  path.CornerStyle
	Fill   color.RGBA
}

// Plate renders the icon plate for s.
func Plate(s Spec) *scribe.Canvas {
	r := s.Radius
	if r <= 0 {
		r = AppleRadiusRatio * float64(s.Size)
	}
	c := scribe.NewCanvas(s.Size, s.Size)
	sz := float64(s.Size)
	c.Fill(path.RoundRect(geom.RectXYWH(0, 0, sz, sz), r, s.Style), s.Fill)
	return c
}

// WriteSet writes <name>_<size>.png into dir for each size, returning
// the paths in size order. An explicit base radius scales
// proportionally with size; a zero radius stays the Apple ratio at
// every size.
func WriteSet(dir, name string, base Spec, sizes []int) ([]string, error) {
	var out []string
	for _, sz := range sizes {
		sp := base
		if sp.Radius > 0 && base.Size > 0 {
			sp.Radius = base.Radius * float64(sz) / float64(base.Size)
		}
		sp.Size = sz
		p := filepath.Join(dir, fmt.Sprintf("%s_%d.png", name, sz))
		if err := Plate(sp).SavePNG(p); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

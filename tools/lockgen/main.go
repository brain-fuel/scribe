// lockgen renders the L3/L5 cross-language lock instance with the Go engine
// and prints its raw alpha mask, one value per line: canvas 32x32,
// RoundRect(4,4,24,24) radius 6 continuous, FlattenTol, nonzero rule.
package main

import (
	"fmt"

	"goforge.dev/scribe"
	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
	"goforge.dev/scribe/raster"
)

func main() {
	p := path.RoundRect(geom.RectXYWH(4, 4, 24, 24), 6, path.Continuous)
	r := raster.New(32, 32)
	for _, pl := range p.Flatten(scribe.FlattenTol) {
		r.AddPolyline(pl)
	}
	m := r.Mask(raster.NonZero)
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			fmt.Println(m.AlphaAt(x, y).A)
		}
	}
}

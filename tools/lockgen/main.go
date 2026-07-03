// lockgen renders the cross-language lock fixtures with the Go engine.
//
// No arguments: the L3 instance (32x32 RoundRect(4,4,24,24) r6 continuous
// raw mask, one alpha per line).
//
// With .scr file arguments (the L5 corpus): each scene is parsed, its per-
// paint-op masks are rendered via dl.Masks at 32x32, and every alpha of
// every mask prints in order, one per line, scenes and masks concatenated.
package main

import (
	"fmt"
	"os"

	"goforge.dev/scribe"
	"goforge.dev/scribe/dl"
	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
	"goforge.dev/scribe/ps"
	"goforge.dev/scribe/raster"
)

func main() {
	if len(os.Args) > 1 {
		for _, f := range os.Args[1:] {
			src, err := os.ReadFile(f)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			prog, err := ps.Parse(string(src))
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s:%v\n", f, err)
				os.Exit(1)
			}
			for _, m := range dl.Masks(prog, 32, 32) {
				for y := 0; y < 32; y++ {
					for x := 0; x < 32; x++ {
						fmt.Println(m.AlphaAt(x, y).A)
					}
				}
			}
		}
		return
	}
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

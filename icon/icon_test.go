package icon

import (
	"bytes"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"goforge.dev/scribe"
	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
)

var blue = color.RGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xFF}

func TestPlateBasics(t *testing.T) {
	c := Plate(Spec{Size: 64, Style: path.Continuous, Fill: blue})
	img := c.Image()
	if img.Bounds().Dx() != 64 {
		t.Fatalf("size = %v", img.Bounds())
	}
	if got := img.RGBAAt(32, 32); got != blue {
		t.Errorf("center = %v", got)
	}
	if got := img.RGBAAt(0, 0); (got != color.RGBA{}) {
		t.Errorf("corner = %v, want transparent", got)
	}
}

// Radius <= 0 means the Apple ratio: byte-identical to rendering the
// plate directly with ratio*size.
func TestPlateDefaultRadius(t *testing.T) {
	got := Plate(Spec{Size: 128, Style: path.Continuous, Fill: blue})

	want := scribe.NewCanvas(128, 128)
	want.Fill(path.RoundRect(geom.RectXYWH(0, 0, 128, 128), AppleRadiusRatio*128, path.Continuous), blue)

	if !bytes.Equal(got.Image().Pix, want.Image().Pix) {
		t.Error("default radius is not AppleRadiusRatio * size")
	}
}

func TestWriteSet(t *testing.T) {
	dir := t.TempDir()
	paths, err := WriteSet(dir, "icon", Spec{Style: path.Continuous, Fill: blue}, AppleSizes)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != len(AppleSizes) {
		t.Fatalf("paths = %d, want %d", len(paths), len(AppleSizes))
	}
	for i, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			t.Fatal(err)
		}
		img, err := png.Decode(f)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		if img.Bounds().Dx() != AppleSizes[i] {
			t.Errorf("%s: size %d, want %d", p, img.Bounds().Dx(), AppleSizes[i])
		}
	}
	if filepath.Base(paths[0]) != "icon_16.png" {
		t.Errorf("first path = %s", paths[0])
	}
}

// An explicit radius scales proportionally across the set: size 200
// with base radius 20 at base size 100 equals a direct radius-40 render.
func TestWriteSetExplicitRadiusScales(t *testing.T) {
	dir := t.TempDir()
	base := Spec{Size: 100, Radius: 20, Style: path.Circular, Fill: blue}
	paths, err := WriteSet(dir, "x", base, []int{200})
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	got, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	want := Plate(Spec{Size: 200, Radius: 40, Style: path.Circular, Fill: blue})
	for _, xy := range [][2]int{{0, 0}, {100, 100}, {20, 3}, {199, 199}} {
		_, _, _, ga := got.At(xy[0], xy[1]).RGBA()
		wc := want.Image().RGBAAt(xy[0], xy[1])
		if (ga == 0) != (wc.A == 0) {
			t.Errorf("pixel %v transparency mismatch", xy)
		}
	}
}

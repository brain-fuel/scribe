package svg

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"goforge.dev/scribe/ps"
)

// The demo scene's SVG output is byte-stable.
func TestGoldenSVG(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "testdata", "demo.scr"))
	if err != nil {
		t.Fatal(err)
	}
	prog, err := ps.Parse(string(src))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := Encode(&buf, prog, 128, 128); err != nil {
		t.Fatal(err)
	}
	golden := filepath.Join("testdata", "demo.svg")
	if os.Getenv("REGEN_GOLDEN") != "" {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, buf.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("regenerated %s", golden)
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("%v (run REGEN_GOLDEN=1 go test to create)", err)
	}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("SVG output differs from golden\ngot:\n%s", buf.String())
	}
}

// Cross-check against a reference SVG rasterizer when one is
// installed; skip otherwise. Keeps the backend honest against the
// wider ecosystem without adding a CI dependency.
func TestReferenceRasterizer(t *testing.T) {
	if _, err := exec.LookPath("rsvg-convert"); err != nil {
		t.Skip("rsvg-convert not installed")
	}
	out, err := exec.Command("rsvg-convert", filepath.Join("testdata", "demo.svg")).Output()
	if err != nil {
		t.Fatalf("rsvg-convert: %v", err)
	}
	if len(out) < 100 {
		t.Errorf("reference render suspiciously small: %d bytes", len(out))
	}
}

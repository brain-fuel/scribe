package path

import (
	"testing"

	"goforge.dev/scribe/geom"
)

func TestBuilder(t *testing.T) {
	var p Path
	p.MoveTo(geom.Pt(0, 0))
	p.LineTo(geom.Pt(4, 0))
	p.CubicTo(geom.Pt(5, 0), geom.Pt(6, 1), geom.Pt(6, 2))
	p.Close()
	wantVerbs := []Verb{VMoveTo, VLineTo, VCubicTo, VClose}
	gotVerbs := p.Verbs()
	if len(gotVerbs) != len(wantVerbs) {
		t.Fatalf("verbs = %v", gotVerbs)
	}
	for i := range wantVerbs {
		if gotVerbs[i] != wantVerbs[i] {
			t.Errorf("verb[%d] = %v, want %v", i, gotVerbs[i], wantVerbs[i])
		}
	}
	if n := len(p.Points()); n != 5 {
		t.Errorf("points = %d, want 5", n)
	}
}

// Totality: ops with no open subpath must not create invalid states.
// LineTo with no subpath ensures one at its target; CubicTo at its
// first control point (HTML canvas convention). Close with no subpath
// is a no-op.
func TestBuilderTotality(t *testing.T) {
	var p Path
	p.Close() // no-op
	if len(p.Verbs()) != 0 {
		t.Fatalf("Close on empty path emitted verbs: %v", p.Verbs())
	}
	var q Path
	q.LineTo(geom.Pt(3, 4))
	if v := q.Verbs(); len(v) != 2 || v[0] != VMoveTo || v[1] != VLineTo {
		t.Fatalf("LineTo on empty path = %v, want implicit MoveTo", v)
	}
	if q.Points()[0] != geom.Pt(3, 4) {
		t.Errorf("implicit MoveTo at %v, want (3,4)", q.Points()[0])
	}
	var c Path
	c.CubicTo(geom.Pt(1, 0), geom.Pt(2, 0), geom.Pt(3, 0))
	if v := c.Verbs(); len(v) != 2 || v[0] != VMoveTo {
		t.Fatalf("CubicTo on empty path = %v, want implicit MoveTo", v)
	}
	if c.Points()[0] != geom.Pt(1, 0) {
		t.Errorf("implicit MoveTo at %v, want (1,0)", c.Points()[0])
	}
}

func TestTransformAndBounds(t *testing.T) {
	var p Path
	p.MoveTo(geom.Pt(1, 1))
	p.LineTo(geom.Pt(3, 2))
	p.Transform(geom.Scale(2, 2))
	if p.Points()[1] != geom.Pt(6, 4) {
		t.Errorf("Transform: %v", p.Points()[1])
	}
	b := p.Bounds()
	if b.Min != geom.Pt(2, 2) || b.Max != geom.Pt(6, 4) {
		t.Errorf("Bounds = %v", b)
	}
}

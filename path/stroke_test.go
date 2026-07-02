package path

import (
	"testing"

	"goforge.dev/scribe/geom"
)

// Helper: does the stroke fill path, flattened, contain point q?
// Winding-number test over the flattened contours.
func strokeCovers(t *testing.T, p *Path, pen Pen, q geom.Point) bool {
	t.Helper()
	s := Stroke(p, pen, 0.01)
	w := 0.0
	for _, pl := range s.Flatten(0.01) {
		pts := pl.Pts
		n := len(pts)
		for i := 0; i < n; i++ {
			a, b := pts[i], pts[(i+1)%n]
			if (a.Y <= q.Y) != (b.Y <= q.Y) {
				xCross := a.X + (q.Y-a.Y)/(b.Y-a.Y)*(b.X-a.X)
				if xCross > q.X {
					if b.Y > a.Y {
						w++
					} else {
						w--
					}
				}
			}
		}
	}
	return w != 0
}

func hline() *Path {
	var p Path
	p.MoveTo(geom.Pt(2, 5))
	p.LineTo(geom.Pt(8, 5))
	return &p
}

func TestStrokeSegmentBody(t *testing.T) {
	pen := Pen{Width: 2, Cap: ButtCap, Join: BevelJoin}
	if !strokeCovers(t, hline(), pen, geom.Pt(5, 5.9)) {
		t.Error("body below centerline not covered")
	}
	if !strokeCovers(t, hline(), pen, geom.Pt(5, 4.1)) {
		t.Error("body above centerline not covered")
	}
	if strokeCovers(t, hline(), pen, geom.Pt(5, 6.5)) {
		t.Error("outside width covered")
	}
	if strokeCovers(t, hline(), pen, geom.Pt(1.5, 5)) {
		t.Error("butt cap extends past endpoint")
	}
}

func TestStrokeCaps(t *testing.T) {
	round := Pen{Width: 2, Cap: RoundCap, Join: BevelJoin}
	if !strokeCovers(t, hline(), round, geom.Pt(1.3, 5)) {
		t.Error("round cap missing")
	}
	if strokeCovers(t, hline(), round, geom.Pt(0.9, 5)) {
		t.Error("round cap too long")
	}
	square := Pen{Width: 2, Cap: SquareCap, Join: BevelJoin}
	if !strokeCovers(t, hline(), square, geom.Pt(1.2, 5.8)) {
		t.Error("square cap corner missing")
	}
}

func elbow() *Path {
	var p Path
	p.MoveTo(geom.Pt(2, 2))
	p.LineTo(geom.Pt(8, 2))
	p.LineTo(geom.Pt(8, 8))
	return &p
}

// The join must cover the elbow region on the outer side; without a
// join the two quads leave a notch.
func TestStrokeJoins(t *testing.T) {
	for _, join := range []Join{BevelJoin, RoundJoin} {
		pen := Pen{Width: 2, Cap: ButtCap, Join: join}
		// Point in the outer notch: outside both quads (x > 8 and
		// y < 2), inside the bevel triangle (8,2),(8,1),(9,2) whose
		// interior satisfies y >= x-7, and inside the unit disc at
		// (8,2). Near the triangle centroid.
		q := geom.Pt(8.33, 1.67)
		if !strokeCovers(t, elbow(), pen, q) {
			t.Errorf("join %v: notch %v not covered", join, q)
		}
		// the joint pixel itself
		if !strokeCovers(t, elbow(), pen, geom.Pt(8, 2)) {
			t.Errorf("join %v: vertex not covered", join)
		}
	}
}

// Closed polylines get joins at the wrap vertex and no caps.
func TestStrokeClosed(t *testing.T) {
	var p Path
	p.MoveTo(geom.Pt(2, 2))
	p.LineTo(geom.Pt(8, 2))
	p.LineTo(geom.Pt(8, 8))
	p.LineTo(geom.Pt(2, 8))
	p.Close()
	pen := Pen{Width: 2, Cap: ButtCap, Join: RoundJoin}
	if !strokeCovers(t, &p, pen, geom.Pt(1.45, 1.45)) {
		t.Error("wrap join not covered")
	}
	if !strokeCovers(t, &p, pen, geom.Pt(5, 1.5)) {
		t.Error("top edge not covered")
	}
	if strokeCovers(t, &p, pen, geom.Pt(5, 5)) {
		t.Error("interior covered; stroke leaked")
	}
}

// Dot: single-point subpath with round cap.
func TestStrokeDot(t *testing.T) {
	var p Path
	p.MoveTo(geom.Pt(5, 5))
	pen := Pen{Width: 4, Cap: RoundCap}
	if !strokeCovers(t, &p, pen, geom.Pt(6.5, 5)) {
		t.Error("dot missing")
	}
	if strokeCovers(t, &p, pen, geom.Pt(7.5, 5)) {
		t.Error("dot too big")
	}
	butt := Pen{Width: 4, Cap: ButtCap}
	if strokeCovers(t, &p, butt, geom.Pt(5, 5)) {
		t.Error("butt dot should be empty")
	}
}

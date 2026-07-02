package path

import "goforge.dev/scribe/geom"

// kappa is the cubic Bezier control distance approximating a quarter
// circle of unit radius: 4/3 * (sqrt(2) - 1). Max radial error is
// about 2.7e-4 of the radius. This is the table-driven mechanism
// QuickDraw used for rounded corners: no trig, just coefficients.
const kappa = 0.5522847498307936

// CornerStyle selects the corner table for RoundRect.
type CornerStyle uint8

const (
	// Circular corners: a single quarter-circle cubic per corner.
	Circular CornerStyle = iota
	// Continuous corners: Apple's G2 continuous-curvature corner
	// (the modern icon squircle), three cubics and a short line per
	// corner, coefficients recovered by PaintCode.
	Continuous
)

// cornerSeg is one segment of a unit-radius corner table. Coordinates
// are (in, out): "in" is distance from the corner point back along the
// incoming edge, "out" is distance along the outgoing edge.
type cornerSeg struct {
	line bool
	pts  [3][2]float64 // line: pts[0] only; cubic: c1, c2, to
}

// circularCorner: one quarter-circle cubic from (1,0) to (0,1) about
// center (1,1).
var circularCorner = struct {
	extent float64
	segs   []cornerSeg
}{
	extent: 1,
	segs: []cornerSeg{
		{pts: [3][2]float64{{1 - kappa, 0}, {0, 1 - kappa}, {0, 1}}},
	},
}

// emitCorner appends a corner to p. k is the corner point, u the unit
// direction of travel into the corner, v the unit direction leaving
// it, r the corner radius. A table point (a, b) maps to
// k - u*(a*r) + v*(b*r). The caller must already be at the mapped
// table start point (extent, 0).
func emitCorner(p *Path, segs []cornerSeg, k, u, v geom.Point, r float64) {
	mp := func(c [2]float64) geom.Point {
		return k.Sub(u.Mul(c[0] * r)).Add(v.Mul(c[1] * r))
	}
	for _, s := range segs {
		if s.line {
			p.LineTo(mp(s.pts[0]))
		} else {
			p.CubicTo(mp(s.pts[0]), mp(s.pts[1]), mp(s.pts[2]))
		}
	}
}

// RectPath returns the closed rectangular path of r.
func RectPath(r geom.Rect) *Path {
	var p Path
	p.MoveTo(r.Min)
	p.LineTo(geom.Pt(r.Max.X, r.Min.Y))
	p.LineTo(r.Max)
	p.LineTo(geom.Pt(r.Min.X, r.Max.Y))
	p.Close()
	return &p
}

// Circle returns a closed circle path built from four kappa cubics.
func Circle(center geom.Point, radius float64) *Path {
	c, r := center, radius
	k := kappa * r
	var p Path
	p.MoveTo(geom.Pt(c.X+r, c.Y))
	p.CubicTo(geom.Pt(c.X+r, c.Y+k), geom.Pt(c.X+k, c.Y+r), geom.Pt(c.X, c.Y+r))
	p.CubicTo(geom.Pt(c.X-k, c.Y+r), geom.Pt(c.X-r, c.Y+k), geom.Pt(c.X-r, c.Y))
	p.CubicTo(geom.Pt(c.X-r, c.Y-k), geom.Pt(c.X-k, c.Y-r), geom.Pt(c.X, c.Y-r))
	p.CubicTo(geom.Pt(c.X+k, c.Y-r), geom.Pt(c.X+r, c.Y-k), geom.Pt(c.X+r, c.Y))
	p.Close()
	return &p
}

// cornerTable returns the table and edge extent for a style.
func cornerTable(style CornerStyle) (float64, []cornerSeg) {
	if style == Continuous {
		return continuousExtent(), continuousSegs()
	}
	return circularCorner.extent, circularCorner.segs
}

// contExtent is how far the continuous corner reaches along each edge,
// in units of the corner radius.
const contExtent = 1.52866483

// continuousCornerSegs is Apple's continuous-curvature corner as
// recovered from CoreGraphics output by PaintCode ("Drawing Code for
// iOS 7 Rounded Rectangles"). Unit corner radius, (in, out)
// coordinates. Three cubics with one short line between the first
// and second. This is literally table-driven Bezier corner code.
var continuousCornerSegs = []cornerSeg{
	{pts: [3][2]float64{
		{1.08849323, 0}, {0.86840689, 0}, {0.66993427, 0.06549600},
	}},
	{line: true, pts: [3][2]float64{
		{0.63149399, 0.07491100}, {}, {},
	}},
	{pts: [3][2]float64{
		{0.37282392, 0.16905899}, {0.16906013, 0.37282392}, {0.07491100, 0.63149399},
	}},
	{pts: [3][2]float64{
		{0, 0.86840689}, {0, 1.08849323}, {0, 1.52866483},
	}},
}

func continuousExtent() float64   { return contExtent }
func continuousSegs() []cornerSeg { return continuousCornerSegs }

// RoundRect returns a rounded rectangle. radius is the corner radius;
// it is clamped so corners never overlap. radius <= 0 yields RectPath(r).
func RoundRect(r geom.Rect, radius float64, style CornerStyle) *Path {
	if radius <= 0 {
		return RectPath(r)
	}
	extent, segs := cornerTable(style)
	short := r.W()
	if r.H() < short {
		short = r.H()
	}
	if m := short / (2 * extent); radius > m {
		radius = m
	}
	e := extent * radius
	x0, y0, x1, y1 := r.Min.X, r.Min.Y, r.Max.X, r.Max.Y
	right := geom.Pt(1, 0)
	down := geom.Pt(0, 1)
	left := geom.Pt(-1, 0)
	up := geom.Pt(0, -1)
	var p Path
	p.MoveTo(geom.Pt(x0+e, y0)) // end of top-left corner
	p.LineTo(geom.Pt(x1-e, y0))
	emitCorner(&p, segs, geom.Pt(x1, y0), right, down, radius)
	p.LineTo(geom.Pt(x1, y1-e))
	emitCorner(&p, segs, geom.Pt(x1, y1), down, left, radius)
	p.LineTo(geom.Pt(x0+e, y1))
	emitCorner(&p, segs, geom.Pt(x0, y1), left, up, radius)
	p.LineTo(geom.Pt(x0, y0+e))
	emitCorner(&p, segs, geom.Pt(x0, y0), up, right, radius)
	p.Close()
	return &p
}

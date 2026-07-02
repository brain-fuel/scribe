// Package path builds and processes vector paths: move/line/cubic
// verbs on corner-grid coordinates, adaptive flattening, shapes with
// table-driven Bezier corners, and stroke-to-fill conversion.
package path

import "goforge.dev/scribe/geom"

// Verb is a path construction operation.
type Verb uint8

const (
	VMoveTo Verb = iota // 1 point
	VLineTo             // 1 point
	VCubicTo            // 3 points: c1, c2, to
	VClose              // 0 points
)

// Path is a sequence of verbs over points. The zero value is an empty
// path, ready to use.
type Path struct {
	verbs []Verb
	pts   []geom.Point
	open  bool // a subpath is open
}

// Verbs returns the verb sequence (shared slice; do not mutate).
func (p *Path) Verbs() []Verb { return p.verbs }

// Points returns the point sequence (shared slice; do not mutate).
func (p *Path) Points() []geom.Point { return p.pts }

// MoveTo starts a new subpath at pt.
func (p *Path) MoveTo(pt geom.Point) {
	p.verbs = append(p.verbs, VMoveTo)
	p.pts = append(p.pts, pt)
	p.open = true
}

// ensure opens a subpath at pt if none is open (totality: no op is an
// error; this is the HTML canvas convention).
func (p *Path) ensure(pt geom.Point) {
	if !p.open {
		p.MoveTo(pt)
	}
}

// LineTo appends a line segment to pt.
func (p *Path) LineTo(pt geom.Point) {
	p.ensure(pt)
	p.verbs = append(p.verbs, VLineTo)
	p.pts = append(p.pts, pt)
}

// CubicTo appends a cubic Bezier with control points c1, c2 ending at to.
func (p *Path) CubicTo(c1, c2, to geom.Point) {
	p.ensure(c1)
	p.verbs = append(p.verbs, VCubicTo)
	p.pts = append(p.pts, c1, c2, to)
}

// Close closes the current subpath. No-op if none is open.
func (p *Path) Close() {
	if !p.open {
		return
	}
	p.verbs = append(p.verbs, VClose)
	p.open = false
}

// Transform applies m to every point in place.
func (p *Path) Transform(m geom.Affine) {
	for i := range p.pts {
		p.pts[i] = m.Apply(p.pts[i])
	}
}

// Bounds returns the control-point bounding box (loose for curves:
// contains the curve, may be larger). Zero Rect for an empty path.
func (p *Path) Bounds() geom.Rect {
	if len(p.pts) == 0 {
		return geom.Rect{}
	}
	b := geom.Rect{Min: p.pts[0], Max: p.pts[0]}
	for _, pt := range p.pts[1:] {
		if pt.X < b.Min.X {
			b.Min.X = pt.X
		}
		if pt.Y < b.Min.Y {
			b.Min.Y = pt.Y
		}
		if pt.X > b.Max.X {
			b.Max.X = pt.X
		}
		if pt.Y > b.Max.Y {
			b.Max.Y = pt.Y
		}
	}
	return b
}

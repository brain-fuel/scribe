// Package svg renders a display list to SVG. Geometry stays exact:
// cubics are emitted as C commands, never flattened. Interpretation
// semantics (CTM at construction time, user-space stroke width,
// paint consumes path) match dl.Render, so the SVG and the raster
// depict the same drawing.
package svg

import (
	"fmt"
	"image/color"
	"io"
	"math"
	"strconv"
	"strings"

	"goforge.dev/scribe/dl"
	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
)

func num(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }

func colAttr(name string, c color.RGBA) string {
	s := fmt.Sprintf(` %s="#%02x%02x%02x"`, name, c.R, c.G, c.B)
	if c.A != 0xFF {
		s += fmt.Sprintf(` %s-opacity="%s"`, name,
			strconv.FormatFloat(float64(c.A)/255, 'g', 6, 64))
	}
	return s
}

type gstate struct {
	ctm geom.Affine
	col color.RGBA
	pen path.Pen
}

// Encode writes prog as an SVG document of the given pixel size.
func Encode(w io.Writer, prog dl.Program, width, height int) error {
	st := gstate{ctm: geom.Identity(), col: color.RGBA{A: 255}, pen: path.Pen{Width: 1}}
	var stack []gstate
	var d strings.Builder

	pt := func(p geom.Point) string {
		q := st.ctm.Apply(p)
		return num(q.X) + " " + num(q.Y)
	}
	seg := func(cmd string, pts ...geom.Point) {
		if d.Len() > 0 {
			d.WriteByte(' ')
		}
		d.WriteString(cmd)
		for _, p := range pts {
			d.WriteByte(' ')
			d.WriteString(pt(p))
		}
	}
	appendShape := func(sp *path.Path) {
		pts := sp.Points()
		pi := 0
		for _, v := range sp.Verbs() {
			switch v {
			case path.VMoveTo:
				seg("M", pts[pi])
				pi++
			case path.VLineTo:
				seg("L", pts[pi])
				pi++
			case path.VCubicTo:
				seg("C", pts[pi], pts[pi+1], pts[pi+2])
				pi += 3
			case path.VClose:
				seg("Z")
			}
		}
	}

	var out strings.Builder
	fmt.Fprintf(&out, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">`+"\n",
		width, height, width, height)

	for _, op := range prog {
		switch o := op.(type) {
		case dl.MoveTo:
			seg("M", o.P)
		case dl.LineTo:
			seg("L", o.P)
		case dl.CubicTo:
			seg("C", o.C1, o.C2, o.P)
		case dl.ClosePath:
			seg("Z")
		case dl.NewPath:
			d.Reset()
		case dl.Rect:
			appendShape(path.RectPath(o.R))
		case dl.Circle:
			appendShape(path.Circle(o.C, o.Radius))
		case dl.RoundRect:
			appendShape(path.RoundRect(o.R, o.Radius, o.Style))
		case dl.SetColor:
			st.col = o.C
		case dl.SetLineWidth:
			st.pen.Width = o.W
		case dl.SetCap:
			st.pen.Cap = o.C
		case dl.SetJoin:
			st.pen.Join = o.J
		case dl.Save:
			stack = append(stack, st)
		case dl.Restore:
			if n := len(stack); n > 0 {
				st = stack[n-1]
				stack = stack[:n-1]
			}
		case dl.Translate:
			st.ctm = st.ctm.Mul(geom.Translate(o.X, o.Y))
		case dl.Scale:
			st.ctm = st.ctm.Mul(geom.Scale(o.X, o.Y))
		case dl.Fill:
			if d.Len() > 0 {
				fmt.Fprintf(&out, `<path d="%s"%s/>`+"\n", d.String(), colAttr("fill", st.col))
			}
			d.Reset()
		case dl.FillEO:
			if d.Len() > 0 {
				fmt.Fprintf(&out, `<path d="%s" fill-rule="evenodd"%s/>`+"\n", d.String(), colAttr("fill", st.col))
			}
			d.Reset()
		case dl.Stroke:
			if d.Len() > 0 {
				width := st.pen.Width * math.Sqrt(math.Abs(st.ctm.A*st.ctm.D-st.ctm.B*st.ctm.C))
				capStr := "butt"
				switch st.pen.Cap {
				case path.RoundCap:
					capStr = "round"
				case path.SquareCap:
					capStr = "square"
				}
				join := "bevel"
				if st.pen.Join == path.RoundJoin {
					join = "round"
				}
				fmt.Fprintf(&out,
					`<path d="%s" fill="none"%s stroke-width="%s" stroke-linecap="%s" stroke-linejoin="%s"/>`+"\n",
					d.String(), colAttr("stroke", st.col), num(width), capStr, join)
			}
			d.Reset()
		case dl.Clear:
			fmt.Fprintf(&out, `<rect width="%d" height="%d"%s/>`+"\n", width, height, colAttr("fill", o.C))
		}
	}
	out.WriteString("</svg>\n")
	_, err := io.WriteString(w, out.String())
	return err
}

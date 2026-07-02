// Package ps is scribe's PostScript-homage postfix language and the
// canonical text codec for dl.Program. Values (numbers, colors,
// names) push onto a stack; operators pop them. Parsing statically
// simulates the stack, so arity and type errors surface at parse
// time with line:col positions and the interpreter stays total.
package ps

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"goforge.dev/scribe/dl"
	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
)

// ParseError is a diagnostic with a 1-based source position.
type ParseError struct {
	Line, Col int
	Msg       string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%d:%d: %s", e.Line, e.Col, e.Msg)
}

type token struct {
	text      string
	line, col int
}

func tokenize(src string) []token {
	var toks []token
	line := 1
	col := 1
	i := 0
	for i < len(src) {
		c := src[i]
		switch {
		case c == '\n':
			line++
			col = 1
			i++
		case c == ' ' || c == '\t' || c == '\r':
			col++
			i++
		case c == '%':
			for i < len(src) && src[i] != '\n' {
				i++
			}
		default:
			start := i
			startCol := col
			for i < len(src) && !strings.ContainsRune(" \t\r\n%", rune(src[i])) {
				i++
				col++
			}
			toks = append(toks, token{src[start:i], line, startCol})
		}
	}
	return toks
}

// value is a typed stack entry during static simulation.
type value struct {
	tok  token
	num  float64
	col  color.RGBA
	name string
	kind kind
}

type kind uint8

const (
	kNum kind = iota
	kColor
	kName
)

func (k kind) String() string {
	switch k {
	case kNum:
		return "number"
	case kColor:
		return "color"
	default:
		return "name"
	}
}

var names = map[string]bool{
	"butt": true, "round": true, "square": true,
	"bevel": true, "circular": true, "continuous": true,
}

// parser carries the static-simulation state.
type parser struct {
	stack  []value
	gsaves int
}

// Parse compiles src to a display list. The first error is returned
// as a *ParseError.
func Parse(src string) (dl.Program, error) {
	var prog dl.Program
	p := &parser{}
	for _, t := range tokenize(src) {
		if n, err := strconv.ParseFloat(t.text, 64); err == nil {
			p.stack = append(p.stack, value{tok: t, num: n, kind: kNum})
			continue
		}
		if strings.HasPrefix(t.text, "#") {
			c, err := parseColor(t.text)
			if err != nil {
				return nil, &ParseError{t.line, t.col, err.Error()}
			}
			p.stack = append(p.stack, value{tok: t, col: c, kind: kColor})
			continue
		}
		if names[t.text] {
			p.stack = append(p.stack, value{tok: t, name: t.text, kind: kName})
			continue
		}
		op, perr := p.applyOp(t)
		if perr != nil {
			return nil, perr
		}
		prog = append(prog, op)
	}
	if len(p.stack) > 0 {
		v := p.stack[len(p.stack)-1]
		return nil, &ParseError{v.tok.line, v.tok.col,
			fmt.Sprintf("leftover %s %q not consumed by any operator", v.kind, v.tok.text)}
	}
	return prog, nil
}

func (p *parser) pop(t token, want kind) (value, *ParseError) {
	if len(p.stack) == 0 {
		return value{}, &ParseError{t.line, t.col,
			fmt.Sprintf("%s: missing %s argument", t.text, want)}
	}
	v := p.stack[len(p.stack)-1]
	p.stack = p.stack[:len(p.stack)-1]
	if v.kind != want {
		return value{}, &ParseError{t.line, t.col,
			fmt.Sprintf("%s: want %s, got %s %q", t.text, want, v.kind, v.tok.text)}
	}
	return v, nil
}

func (p *parser) popNums(t token, n int) ([]float64, *ParseError) {
	out := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		v, err := p.pop(t, kNum)
		if err != nil {
			return nil, err
		}
		out[i] = v.num
	}
	return out, nil
}

func (p *parser) applyOp(t token) (dl.Op, *ParseError) {
	switch t.text {
	case "moveto":
		n, err := p.popNums(t, 2)
		if err != nil {
			return nil, err
		}
		return dl.MoveTo{P: geom.Pt(n[0], n[1])}, nil
	case "lineto":
		n, err := p.popNums(t, 2)
		if err != nil {
			return nil, err
		}
		return dl.LineTo{P: geom.Pt(n[0], n[1])}, nil
	case "curveto":
		n, err := p.popNums(t, 6)
		if err != nil {
			return nil, err
		}
		return dl.CubicTo{C1: geom.Pt(n[0], n[1]), C2: geom.Pt(n[2], n[3]), P: geom.Pt(n[4], n[5])}, nil
	case "closepath":
		return dl.ClosePath{}, nil
	case "newpath":
		return dl.NewPath{}, nil
	case "rect":
		n, err := p.popNums(t, 4)
		if err != nil {
			return nil, err
		}
		return dl.Rect{R: geom.RectXYWH(n[0], n[1], n[2], n[3])}, nil
	case "circle":
		n, err := p.popNums(t, 3)
		if err != nil {
			return nil, err
		}
		return dl.Circle{C: geom.Pt(n[0], n[1]), Radius: n[2]}, nil
	case "roundrect":
		style, err := p.pop(t, kName)
		if err != nil {
			return nil, err
		}
		var cs path.CornerStyle
		switch style.name {
		case "circular":
			cs = path.Circular
		case "continuous":
			cs = path.Continuous
		default:
			return nil, &ParseError{style.tok.line, style.tok.col,
				fmt.Sprintf("roundrect: corner style must be circular or continuous, got %q", style.name)}
		}
		n, err := p.popNums(t, 5)
		if err != nil {
			return nil, err
		}
		return dl.RoundRect{R: geom.RectXYWH(n[0], n[1], n[2], n[3]), Radius: n[4], Style: cs}, nil
	case "setcolor":
		v, err := p.pop(t, kColor)
		if err != nil {
			return nil, err
		}
		return dl.SetColor{C: v.col}, nil
	case "setlinewidth":
		n, err := p.popNums(t, 1)
		if err != nil {
			return nil, err
		}
		return dl.SetLineWidth{W: n[0]}, nil
	case "setcap":
		v, err := p.pop(t, kName)
		if err != nil {
			return nil, err
		}
		switch v.name {
		case "butt":
			return dl.SetCap{C: path.ButtCap}, nil
		case "round":
			return dl.SetCap{C: path.RoundCap}, nil
		case "square":
			return dl.SetCap{C: path.SquareCap}, nil
		}
		return nil, &ParseError{v.tok.line, v.tok.col,
			fmt.Sprintf("setcap: want butt, round or square, got %q", v.name)}
	case "setjoin":
		v, err := p.pop(t, kName)
		if err != nil {
			return nil, err
		}
		switch v.name {
		case "bevel":
			return dl.SetJoin{J: path.BevelJoin}, nil
		case "round":
			return dl.SetJoin{J: path.RoundJoin}, nil
		}
		return nil, &ParseError{v.tok.line, v.tok.col,
			fmt.Sprintf("setjoin: want bevel or round, got %q", v.name)}
	case "gsave":
		p.gsaves++
		return dl.Save{}, nil
	case "grestore":
		if p.gsaves == 0 {
			return nil, &ParseError{t.line, t.col, "grestore without matching gsave"}
		}
		p.gsaves--
		return dl.Restore{}, nil
	case "translate":
		n, err := p.popNums(t, 2)
		if err != nil {
			return nil, err
		}
		return dl.Translate{X: n[0], Y: n[1]}, nil
	case "scale":
		n, err := p.popNums(t, 2)
		if err != nil {
			return nil, err
		}
		return dl.Scale{X: n[0], Y: n[1]}, nil
	case "fill":
		return dl.Fill{}, nil
	case "eofill":
		return dl.FillEO{}, nil
	case "stroke":
		return dl.Stroke{}, nil
	case "clear":
		v, err := p.pop(t, kColor)
		if err != nil {
			return nil, err
		}
		return dl.Clear{C: v.col}, nil
	}
	return nil, &ParseError{t.line, t.col, fmt.Sprintf("unknown word %q", t.text)}
}

func parseColor(s string) (color.RGBA, error) {
	hex := s[1:]
	if len(hex) != 6 && len(hex) != 8 {
		return color.RGBA{}, fmt.Errorf("color %q: want #rrggbb or #rrggbbaa", s)
	}
	var b [4]byte
	b[3] = 0xFF
	for i := 0; i < len(hex)/2; i++ {
		v, err := strconv.ParseUint(hex[2*i:2*i+2], 16, 8)
		if err != nil {
			return color.RGBA{}, fmt.Errorf("color %q: bad hex digits", s)
		}
		b[i] = byte(v)
	}
	return color.RGBA{R: b[0], G: b[1], B: b[2], A: b[3]}, nil
}

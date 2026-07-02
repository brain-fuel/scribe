// Command scribe renders drawings to PNG.
//
//	scribe icon -size 1024 -style continuous -fill '#FF6A00' -o icon.png
package main

import (
	"flag"
	"fmt"
	"image/color"
	"io"
	"os"

	"goforge.dev/scribe"
	"goforge.dev/scribe/dl"
	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
	"goforge.dev/scribe/ps"
	"goforge.dev/scribe/svg"
	"goforge.dev/scribe/term"
)

// appleRadiusRatio approximates the Apple app-icon corner radius as a
// fraction of icon size.
const appleRadiusRatio = 0.2237

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "scribe:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: scribe icon|render [flags]")
	}
	switch args[0] {
	case "icon":
		return iconCmd(args[1:])
	case "render":
		return renderCmd(args[1:])
	default:
		return fmt.Errorf("unknown command %q (want: icon or render)", args[0])
	}
}

func renderCmd(args []string) error {
	fs := flag.NewFlagSet("render", flag.ContinueOnError)
	w := fs.Int("w", 256, "canvas width in pixels")
	h := fs.Int("h", 256, "canvas height in pixels")
	size := fs.Int("size", 0, "square canvas size (overrides -w and -h)")
	out := fs.String("o", "out.png", "output PNG file")
	svgOut := fs.String("svg", "", "also write SVG to this file")
	termOut := fs.Bool("term", false, "write the render to the terminal")
	protocol := fs.String("protocol", "auto", "terminal protocol: auto|kitty|sixel|halfblock")
	// flag package stops at the first non-flag arg, so accept both
	// "render FILE -flags" and "render -flags FILE".
	var file string
	rest := args
	if len(rest) > 0 && rest[0] != "" && rest[0][0] != '-' {
		file = rest[0]
		rest = rest[1:]
	}
	if err := fs.Parse(rest); err != nil {
		return err
	}
	if file == "" {
		if fs.NArg() < 1 {
			return fmt.Errorf("render: missing input file (use - for stdin)")
		}
		file = fs.Arg(0)
	}
	var src []byte
	var err error
	if file == "-" {
		src, err = io.ReadAll(os.Stdin)
	} else {
		src, err = os.ReadFile(file)
	}
	if err != nil {
		return err
	}
	prog, err := ps.Parse(string(src))
	if err != nil {
		return fmt.Errorf("%s:%v", file, err)
	}
	if *size > 0 {
		*w, *h = *size, *size
	}
	if *svgOut != "" {
		f, err := os.Create(*svgOut)
		if err != nil {
			return err
		}
		if err := svg.Encode(f, prog, *w, *h); err != nil {
			f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		fmt.Printf("wrote %s (%dx%d, %d ops)\n", *svgOut, *w, *h, len(prog))
	}
	oSet := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "o" {
			oSet = true
		}
	})
	needRaster := oSet || *termOut || (*svgOut == "" && !*termOut)
	if !needRaster {
		return nil
	}
	c := scribe.NewCanvas(*w, *h)
	dl.Render(prog, c)
	if oSet || (*svgOut == "" && !*termOut) {
		if err := c.SavePNG(*out); err != nil {
			return err
		}
		fmt.Printf("wrote %s (%dx%d, %d ops)\n", *out, *w, *h, len(prog))
	}
	if *termOut {
		p := term.HalfBlock
		switch *protocol {
		case "auto":
			p = term.Detect(os.Getenv)
		case "kitty":
			p = term.Kitty
		case "sixel":
			p = term.Sixel
		case "halfblock":
			p = term.HalfBlock
		default:
			return fmt.Errorf("unknown protocol %q (want auto, kitty, sixel or halfblock)", *protocol)
		}
		if err := term.Write(os.Stdout, c.Image(), p); err != nil {
			return err
		}
	}
	return nil
}

func iconCmd(args []string) error {
	fs := flag.NewFlagSet("icon", flag.ContinueOnError)
	size := fs.Int("size", 1024, "icon size in pixels")
	radius := fs.Float64("radius", -1, "corner radius in pixels (-1: Apple ratio)")
	style := fs.String("style", "continuous", "corner style: circular or continuous")
	fill := fs.String("fill", "#FF6A00", "fill color #RRGGBB or #RRGGBBAA")
	out := fs.String("o", "icon.png", "output PNG file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	var cs path.CornerStyle
	switch *style {
	case "circular":
		cs = path.Circular
	case "continuous":
		cs = path.Continuous
	default:
		return fmt.Errorf("unknown style %q (want circular or continuous)", *style)
	}
	col, err := parseHexColor(*fill)
	if err != nil {
		return err
	}
	r := *radius
	if r < 0 {
		r = appleRadiusRatio * float64(*size)
	}
	s := float64(*size)
	c := scribe.NewCanvas(*size, *size)
	c.Fill(path.RoundRect(geom.RectXYWH(0, 0, s, s), r, cs), col)
	if err := c.SavePNG(*out); err != nil {
		return err
	}
	fmt.Printf("wrote %s (%dx%d, style %s, radius %.1f)\n", *out, *size, *size, *style, r)
	return nil
}

func parseHexColor(s string) (color.RGBA, error) {
	if len(s) > 0 && s[0] == '#' {
		s = s[1:]
	}
	var c color.RGBA
	c.A = 0xFF
	switch len(s) {
	case 6:
		if _, err := fmt.Sscanf(s, "%02x%02x%02x", &c.R, &c.G, &c.B); err != nil {
			return c, fmt.Errorf("bad color %q", s)
		}
	case 8:
		if _, err := fmt.Sscanf(s, "%02x%02x%02x%02x", &c.R, &c.G, &c.B, &c.A); err != nil {
			return c, fmt.Errorf("bad color %q", s)
		}
	default:
		return c, fmt.Errorf("bad color %q (want #RRGGBB or #RRGGBBAA)", s)
	}
	return c, nil
}

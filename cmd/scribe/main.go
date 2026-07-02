// Command scribe renders drawings to PNG.
//
//	scribe icon -size 1024 -style continuous -fill '#FF6A00' -o icon.png
package main

import (
	"flag"
	"fmt"
	"image/color"
	"os"

	"goforge.dev/scribe"
	"goforge.dev/scribe/geom"
	"goforge.dev/scribe/path"
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
		return fmt.Errorf("usage: scribe icon [flags]")
	}
	switch args[0] {
	case "icon":
		return iconCmd(args[1:])
	default:
		return fmt.Errorf("unknown command %q (want: icon)", args[0])
	}
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

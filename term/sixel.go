package term

import (
	"fmt"
	"image"
	"io"
)

// sixLevel quantizes an 8-bit channel to 6 levels.
func sixLevel(v uint8) int { return (int(v)*5 + 127) / 255 }

// writeSixel encodes img as DEC sixel with a fixed 216-color cube.
func writeSixel(w io.Writer, img image.Image) error {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Quantize the whole image to palette indexes.
	idx := make([]int, width*height)
	used := make([]bool, 216)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b := overBlack(img, bounds.Min.X+x, bounds.Min.Y+y)
			i := sixLevel(r)*36 + sixLevel(g)*6 + sixLevel(b)
			idx[y*width+x] = i
			used[i] = true
		}
	}

	var out []byte
	out = append(out, "\x1bPq"...)
	out = fmt.Appendf(out, "\"1;1;%d;%d", width, height)
	// Palette definitions in ascending register order.
	for i, u := range used {
		if !u {
			continue
		}
		r6 := i / 36
		g6 := (i / 6) % 6
		b6 := i % 6
		out = fmt.Appendf(out, "#%d;2;%d;%d;%d", i, r6*100/5, g6*100/5, b6*100/5)
	}
	// Bands of 6 rows.
	for top := 0; top < height; top += 6 {
		firstColor := true
		for ci, u := range used {
			if !u {
				continue
			}
			// Build the row of sixel chars for this color.
			row := make([]byte, width)
			any := false
			for x := 0; x < width; x++ {
				var mask byte
				for dy := 0; dy < 6 && top+dy < height; dy++ {
					if idx[(top+dy)*width+x] == ci {
						mask |= 1 << dy
					}
				}
				row[x] = 63 + mask
				if mask != 0 {
					any = true
				}
			}
			if !any {
				continue
			}
			if !firstColor {
				out = append(out, '$') // carriage return, same band
			}
			firstColor = false
			out = fmt.Appendf(out, "#%d", ci)
			// RLE emit.
			for x := 0; x < width; {
				run := 1
				for x+run < width && row[x+run] == row[x] {
					run++
				}
				if run >= 4 {
					out = fmt.Appendf(out, "!%d%c", run, row[x])
				} else {
					for k := 0; k < run; k++ {
						out = append(out, row[x])
					}
				}
				x += run
			}
		}
		if top+6 < height {
			out = append(out, '-')
		}
	}
	out = append(out, "\x1b\\"...)
	_, err := w.Write(out)
	return err
}

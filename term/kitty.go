package term

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"io"
)

const kittyChunk = 4096

// writeKitty transmits img as a PNG via the kitty graphics protocol,
// chunked APC escapes, transmit-and-display.
func writeKitty(w io.Writer, img image.Image) error {
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		return err
	}
	data := base64.StdEncoding.EncodeToString(pngBuf.Bytes())
	first := true
	for len(data) > 0 {
		n := len(data)
		if n > kittyChunk {
			n = kittyChunk
		}
		chunk := data[:n]
		data = data[n:]
		keys := ""
		if first {
			keys = "a=T,f=100,"
			first = false
		}
		m := "m=0"
		if len(data) > 0 {
			m = "m=1"
		}
		if _, err := io.WriteString(w, "\x1b_G"+keys+m+";"+chunk+"\x1b\\"); err != nil {
			return err
		}
	}
	return nil
}

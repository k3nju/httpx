package httpx

import (
	"bytes"
	"io"
)

func parseStartLine(line []byte) ([]byte, []byte, []byte, bool) {
	parts := bytes.SplitN(line, []byte(" "), 3)
	if len(parts) != 3 {
		return nil, nil, nil, false
	}

	return parts[0], parts[1], parts[2], true
}

func joinByteSlices(bs ...[]byte) []byte {
	s := 0
	for _, b := range bs {
		s += len(b)
	}

	t := make([]byte, 0, s)
	for _, b := range bs {
		t = append(t, b...)
	}

	return t
}

func trimAsFieldValue(s []byte) []byte {
	// TODO: consider convert directory to s
	d := make([]byte, len(s))
	i := 0
	for _, b := range s {
		if b == 0x21 || (0x23 <= b && b <= 0x7e) {
			d[i] = b
			i++
		}
	}

	return d[:i]
}

func trimAsToken(s []byte) []byte {
	// TODO: consider convert directory to s
	d := make([]byte, len(s))
	i := 0
	for _, b := range s {
		switch b {
		// 0x22 = '
		case '!', '#', '$', '%', '&', 0x22, '*',
			'+', '-', '.', '^', '_', '`', '|', '~':
			d[i] = b
			i++
			continue
		}
		if (0x30 <= b && b <= 0x39) || (0x41 <= b && b <= 0x5a) || (0x61 <= b && b <= 0x7a) {
			d[i] = b
			i++
		}
	}

	return d[:i]
}

func writeAll(w io.Writer, bs ...[]byte) (int64, error) {
	var t int64

	for _, b := range bs {
		for len(b) > 0 {
			n, err := w.Write(b)
			if n > 0 {
				b = b[n:]
				t += int64(n)
			}
			if err != nil {
				return t, err
			}
		}
	}

	return t, nil
}

package task

import (
	"bytes"
	"io"
	"strings"
	"time"
)

func PrintTreePeriodically(ctx Context, period time.Duration, verboseLevel int32) {
	block := [32]byte{}
	var text []byte
	buf := bytes.Buffer{}
	buf.Grow(256)

	ticker := time.NewTicker(period)
	for running := true; running; {
		select {
		case <-ticker.C:
			{
				PrintContextTree(ctx, &buf, verboseLevel)
				changed := false
				R := buf.Len()
				if R != len(text) {
					if cap(text) < R {
						text = make([]byte, R, (R+0x1FF)&^0x1FF)
					} else {
						text = text[:R]
					}
					changed = true
				}
				for pos := 0; pos < R; {
					n, _ := buf.Read(block[:])
					if n == 0 {
						break
					}
					if !changed {
						changed = !bytes.Equal(block[:n], text[pos:pos+n])
					}
					if changed {
						copy(text[pos:], block[:n])
					}
					pos += n
				}
				if changed {
					ctx.Info(verboseLevel, string(text))
				}
				buf.Reset()
			}
		case <-ctx.Closing():
			running = false
		}
	}
	ticker.Stop()
}

// Writes pretty debug state info of a given verbosity level.
// If out == nil, the text output is instead directed to this context's logger.Info()
func PrintContextTree(ctx Context, out io.Writer, verboseLevel int32) {
	buf := new(strings.Builder)
	buf.WriteString("\n")

	var prefixBuf [256]rune
	printContextTree(ctx, buf, 0, prefixBuf[:0], true)
	outStr := buf.String()
	if out != nil {
		out.Write([]byte(outStr))
	} else {
		ctx.Info(verboseLevel, outStr)
	}
}

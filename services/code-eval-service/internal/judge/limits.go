package judge

import (
	"bytes"
	"fmt"
	"os"
)

const (
	maxProcs       = 64
	maxOpenFiles   = 256
	maxFileBytes   = 64 << 20
	maxOutputBytes = 1 << 20
	cpuGraceSec    = 1

	compileMemoryMB = 256
)

func limitsEnabled() bool {
	return os.Getenv("ENVIRONMENT") != "test"
}

func wrapLimits(run []string, memoryLimitMB, timeLimitMS int, skipAddressLimit bool) []string {
	if !limitsEnabled() {
		return run
	}

	cpuSec := (timeLimitMS+999)/1000 + cpuGraceSec

	wrapper := []string{
		"prlimit",
		fmt.Sprintf("--cpu=%d", cpuSec),
		fmt.Sprintf("--fsize=%d", int64(maxFileBytes)),
		fmt.Sprintf("--nproc=%d", maxProcs),
		fmt.Sprintf("--nofile=%d", maxOpenFiles),
	}

	if !skipAddressLimit {
		wrapper = append(wrapper, fmt.Sprintf("--as=%d", int64(memoryLimitMB)<<20))
	}

	wrapper = append(wrapper, "--")
	return append(wrapper, run...)
}

type capWriter struct {
	buf   bytes.Buffer
	limit int
}

func (w *capWriter) Write(p []byte) (int, error) {
	if room := w.limit - w.buf.Len(); room > 0 {
		if len(p) > room {
			w.buf.Write(p[:room])
		} else {
			w.buf.Write(p)
		}
	}
	return len(p), nil
}

func (w *capWriter) String() string { return w.buf.String() }

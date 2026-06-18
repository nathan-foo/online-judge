package judge

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"
)

func normalize(s string) string { return strings.TrimRight(s, " \t\r\n") }

func runPythonCase(dir string, tc TestCase, timeLimitMS int) TestResult {
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(timeLimitMS)*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, "python3", "main.py")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(tc.Stdin)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start).Milliseconds()

	res := TestResult{ID: tc.ID, TimeMS: elapsed}

	switch {
	case ctx.Err() == context.DeadlineExceeded:
		res.Status = TestTimeLimitExceeded
	case err != nil:
		res.Status = TestRuntimeError
	case normalize(stdout.String()) == normalize(tc.ExpectedStdout):
		res.Status = TestAccepted
	default:
		res.Status = TestWrongAnswer
	}

	return res
}

package judge

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const compileTimeout = 10 * time.Second

type langSpec struct {
	filename         string
	compile          []string
	run              []string
	skipAddressLimit bool
	memFlag          string
	memEnv           string
}

var langSpecs = map[Language]langSpec{
	LanguagePython: {filename: "main.py", run: []string{"python3", "main.py"}},
	LanguageC: {
		filename: "main.c",
		compile:  []string{"gcc", "-O2", "-std=c17", "main.c", "-o", "main", "-lm"},
		run:      []string{"./main"},
	},
	LanguageCPP: {
		filename: "main.cpp",
		compile:  []string{"g++", "-O2", "-std=c++17", "main.cpp", "-o", "main"},
		run:      []string{"./main"},
	},
	LanguageJava: {
		filename:         "Main.java",
		compile:          []string{"javac", "Main.java"},
		run:              []string{"java", "Main"},
		memFlag:          "-Xmx%dm",
		skipAddressLimit: true,
	},
	LanguageJavaScript: {
		filename:         "main.js",
		run:              []string{"node", "main.js"},
		memFlag:          "--max-old-space-size=%d",
		skipAddressLimit: true,
	},
	LanguageGo: {
		filename:         "main.go",
		compile:          []string{"go", "build", "-o", "main", "main.go"},
		run:              []string{"./main"},
		memEnv:           "GOMEMLIMIT=%dMiB",
		skipAddressLimit: true,
	},
	LanguageTypeScript: {
		filename:         "main.ts",
		compile:          []string{"esbuild", "main.ts", "--outfile=main.js", "--platform=node", "--format=cjs", "--log-level=error"},
		run:              []string{"node", "main.js"},
		memFlag:          "--max-old-space-size=%d",
		skipAddressLimit: true,
	},
	LanguageKotlin: {
		filename:         "main.kt",
		compile:          []string{"kotlinc", "main.kt", "-include-runtime", "-d", "main.jar"},
		run:              []string{"java", "-jar", "main.jar"},
		memFlag:          "-Xmx%dm",
		skipAddressLimit: true,
	},
	LanguageRust: {
		filename: "main.rs",
		compile:  []string{"rustc", "-O", "main.rs", "-o", "main"},
		run:      []string{"./main"},
	},
	LanguageCSharp: {
		filename:         "main.cs",
		compile:          []string{"mcs", "main.cs"},
		run:              []string{"mono", "main.exe"},
		skipAddressLimit: true,
	},
}

func runArgv(spec langSpec, memMB int) []string {
	if spec.memFlag == "" {
		return spec.run
	}
	out := make([]string, 0, len(spec.run)+1)
	out = append(out, spec.run[0], fmt.Sprintf(spec.memFlag, memMB))
	return append(out, spec.run[1:]...)
}

func runEnv(spec langSpec, memMB int) []string {
	if spec.memEnv == "" {
		return nil
	}
	return []string{fmt.Sprintf(spec.memEnv, memMB)}
}

func normalize(s string) string { return strings.TrimRight(s, " \t\r\n") }

func compile(dir string, spec langSpec) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), compileTimeout)
	defer cancel()

	limited := wrapLimits(spec.compile, compileMemoryMB, int(compileTimeout/time.Millisecond), spec.skipAddressLimit)
	cmd := exec.CommandContext(ctx, limited[0], limited[1:]...)
	cmd.Dir = dir

	stderr := capWriter{limit: maxOutputBytes}
	cmd.Stderr = &stderr

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return "compilation timed out", nil
	}
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return stderr.String(), nil
		}
		return "", err
	}
	return "", nil
}

func runCase(dir string, spec langSpec, tc TestCase, timeLimitMS, memoryLimitMB int) TestResult {
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Duration(timeLimitMS)*time.Millisecond)
	defer cancel()

	argv := wrapLimits(runArgv(spec, memoryLimitMB), memoryLimitMB, timeLimitMS, spec.skipAddressLimit)
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(tc.Stdin)
	if env := runEnv(spec, memoryLimitMB); env != nil {
		cmd.Env = append(os.Environ(), env...)
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	cmd.WaitDelay = 2 * time.Second

	stdout := capWriter{limit: maxOutputBytes}
	stderr := capWriter{limit: maxOutputBytes}
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
		if oomKilled(cmd.ProcessState) {
			res.Status = TestMemoryLimitExceeded
		} else {
			res.Status = TestRuntimeError
		}
	case normalize(stdout.String()) == normalize(tc.ExpectedStdout):
		res.Status = TestAccepted
	default:
		res.Status = TestWrongAnswer
	}

	return res
}

func oomKilled(ps *os.ProcessState) bool {
	if ps == nil {
		return false
	}
	ws, ok := ps.Sys().(syscall.WaitStatus)
	return ok && ws.Signaled() && ws.Signal() == syscall.SIGKILL
}

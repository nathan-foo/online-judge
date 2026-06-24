package judge

import (
	"fmt"
	"os"
	"path/filepath"
)

func RunTests(req EvalRequest) (results []TestResult, compileErr string, err error) {
	spec, ok := langSpecs[req.Language]
	if !ok {
		return nil, "", fmt.Errorf("unsupported language: %s", req.Language)
	}

	dir, err := os.MkdirTemp("", "eval-*")

	if err != nil {
		return nil, "", err
	}

	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, spec.filename), []byte(req.SourceCode), 0o644); err != nil {
		return nil, "", err
	}

	if len(spec.compile) > 0 {
		cerr, err := compile(dir, spec)
		if err != nil {
			return nil, "", err
		}
		if cerr != "" {
			return nil, cerr, nil
		}
	}

	results = make([]TestResult, 0, len(req.TestCases))

	for _, tc := range req.TestCases {
		results = append(results, runCase(dir, spec, tc, req.TimeLimitMS, req.MemoryLimitMB))
	}

	return results, "", nil
}

func Aggregate(req EvalRequest, results []TestResult, compileErr string, err error) EvalResult {
	out := EvalResult{
		AttemptID:    req.AttemptID,
		ProblemID:    req.ProblemID,
		Total:        len(req.TestCases),
		CompileError: compileErr,
		Results:      results,
	}

	if err != nil {
		out.Status = StatusError
		return out
	}

	out.Status = StatusDone

	for _, r := range results {
		if r.Status == TestAccepted {
			out.Passed++
		}
	}
	if out.Total > 0 && out.Passed == out.Total {
		out.PointsAwarded = req.Points
	}

	return out
}

// For testing
func Run(req EvalRequest) EvalResult {
	results, compileErr, err := RunTests(req)
	return Aggregate(req, results, compileErr, err)
}

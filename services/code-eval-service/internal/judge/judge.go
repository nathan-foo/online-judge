package judge

import (
	"os"
	"path/filepath"
)

// Exec Agent
func RunTests(req EvalRequest) (results []TestResult, compileErr string, err error) {
	dir, err := os.MkdirTemp("", "eval-*")

	if err != nil {
		return nil, "", err
	}

	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "main.py"), []byte(req.SourceCode), 0o644); err != nil {
		return nil, "", err
	}

	results = make([]TestResult, 0, len(req.TestCases))

	for _, tc := range req.TestCases {
		results = append(results, runPythonCase(dir, tc, req.TimeLimitMS))
	}

	return results, "", nil
}

// Eval Service
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

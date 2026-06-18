package judge

// Exec Agent
func RunTests(req EvalRequest) (results []TestResult, compileErr string, err error) {
	return
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

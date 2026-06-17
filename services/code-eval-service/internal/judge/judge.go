package judge

// Exec Agent
func RunTests(req EvalRequest) (results []TestResult, compileErr string, err error) {
	return
}

// Eval Service
func Aggregate(req EvalRequest, results []TestResult, compileErr string, err error) EvalResult {
	return EvalResult{}
}

// For testing
func Run(req EvalRequest) EvalResult {
	results, compileErr, err := RunTests(req)
	return Aggregate(req, results, compileErr, err)
}

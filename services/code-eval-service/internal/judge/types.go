package judge

type Language string

const (
	LanguagePython     Language = "python"
	LanguageJava       Language = "java"
	LanguageCPP        Language = "cpp"
	LanguageJavaScript Language = "javascript"
)

type TestCase struct {
	ID             string `json:"id"`
	Stdin          string `json:"stdin"`
	ExpectedStdout string `json:"expected_stdout"`
	IsExample      bool   `json:"is_example"`
}

type EvalRequest struct {
	AttemptID     string     `json:"attempt_id"`
	ProblemID     string     `json:"problem_id"`
	Language      Language   `json:"language"`
	SourceCode    string     `json:"source_code"`
	TestCases     []TestCase `json:"test_cases"`
	TimeLimitMS   int        `json:"time_limit_ms"`
	MemoryLimitMB int        `json:"memory_limit_mb"`
	Points        int        `json:"points"`
}

type Status string

const (
	StatusDone  Status = "done"
	StatusError Status = "error"
)

type TestStatus string

const (
	TestAccepted            TestStatus = "accepted"
	TestWrongAnswer         TestStatus = "wrong_answer"
	TestTimeLimitExceeded   TestStatus = "time_limit_exceeded"
	TestMemoryLimitExceeded TestStatus = "memory_limit_exceeded"
	TestRuntimeError        TestStatus = "runtime_error"
)

type TestResult struct {
	ID     string     `json:"id"`
	Status TestStatus `json:"status"`
	TimeMS int64      `json:"time_ms"`
}

type EvalResult struct {
	AttemptID     string       `json:"attempt_id"`
	ProblemID     string       `json:"problem_id"`
	Status        Status       `json:"status"`
	Passed        int          `json:"passed"`
	Total         int          `json:"total"`
	PointsAwarded int          `json:"points_awarded"`
	CompileError  string       `json:"compile_error,omitempty"`
	Results       []TestResult `json:"results,omitempty"`
}

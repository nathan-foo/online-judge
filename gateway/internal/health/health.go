package health

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

type Check struct {
	Name string
	Fn   func(ctx context.Context) error
}

type Checker struct {
	checks []Check
}

func NewChecker(checks ...Check) *Checker {
	return &Checker{checks: checks}
}

type statusResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

func (c *Checker) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, statusResponse{Status: "ok"})
}

func (c *Checker) Readyz(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string, len(c.checks))
	ready := true

	for _, check := range c.checks {
		if err := check.Fn(r.Context()); err != nil {
			checks[check.Name] = err.Error()
			ready = false
			log.Warn().Err(err).Str("check", check.Name).Msg("readiness check failed")
		} else {
			checks[check.Name] = "ok"
		}
	}

	status := http.StatusOK
	resp := statusResponse{Status: "ok", Checks: checks}
	if !ready {
		status = http.StatusServiceUnavailable
		resp.Status = "unavailable"
	}

	writeJSON(w, status, resp)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

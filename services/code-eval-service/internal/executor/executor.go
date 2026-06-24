package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/nathan-foo/online-judge/services/code-eval-service/internal/judge"
)

const (
	compileBudget = 15 * time.Second
	perCaseGrace  = 2 * time.Second
	requestMargin = 5 * time.Second
)

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{http: &http.Client{}}
}

func (c *Client) Run(ctx context.Context, url string, req judge.EvalRequest) (judge.ExecResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, budget(req))
	defer cancel()

	body, err := json.Marshal(req)
	if err != nil {
		return judge.ExecResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url+"/exec", bytes.NewReader(body))
	if err != nil {
		return judge.ExecResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return judge.ExecResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return judge.ExecResponse{}, fmt.Errorf("exec-agent %s: status %d", url, resp.StatusCode)
	}

	var out judge.ExecResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return judge.ExecResponse{}, err
	}
	return out, nil
}

func budget(req judge.EvalRequest) time.Duration {
	perCase := time.Duration(req.TimeLimitMS)*time.Millisecond + perCaseGrace
	return compileBudget + time.Duration(len(req.TestCases))*perCase + requestMargin
}

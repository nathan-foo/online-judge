package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/nathan-foo/online-judge/gateway/internal/config"
)

type Proxy struct {
	testProxy *httputil.ReverseProxy
}

func NewProxy(cfg *config.Config) (*Proxy, error) {
	testUrl, err := url.Parse(cfg.Endpoints.TEST_SERVICE_URL)
	if err != nil {
		return nil, err
	}
	return &Proxy{
		testProxy: httputil.NewSingleHostReverseProxy(testUrl),
	}, nil
}

func (p *Proxy) TestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.testProxy.ServeHTTP(w, r)
	})
}

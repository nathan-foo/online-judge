package proxy

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/nathan-foo/online-judge/gateway/internal/config"
)

type Proxy struct {
	testProxy *httputil.ReverseProxy
}

func NewProxy(endpointConfig config.EndpointConfig) (*Proxy, error) {
	testUrl, err := url.Parse(endpointConfig.TestServiceUrl)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     90 * time.Second,

		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,

		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}

	return &Proxy{
		testProxy: &httputil.ReverseProxy{
			Rewrite: func(r *httputil.ProxyRequest) {
				r.SetURL(testUrl)
			},
			Transport: transport,
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadGateway)
				json.NewEncoder(w).Encode(map[string]string{"error": "service temporarily unavailable"})
			},
		},
	}, nil
}

func (p *Proxy) TestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.testProxy.ServeHTTP(w, r)
	})
}

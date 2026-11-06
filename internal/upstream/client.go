// Package upstream forwards gateway requests to the configured provider endpoint.
package upstream

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hermeto/tollgate/internal/config"
	"github.com/hermeto/tollgate/internal/pipeline"
)

type Client struct {
	base       *url.URL
	credential string
	http       *http.Client
}

func New(c config.UpstreamConfig) (*Client, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{MaxIdleConns: 256, MaxIdleConnsPerHost: 256, IdleConnTimeout: time.Duration(c.Timeouts.Idle), ResponseHeaderTimeout: time.Duration(c.Timeouts.ResponseHeader), ForceAttemptHTTP2: true, DialContext: (&net.Dialer{Timeout: time.Duration(c.Timeouts.Connect)}).DialContext}
	return &Client{base: u, credential: c.CredentialEnv, http: &http.Client{Transport: transport}}, nil
}

func (c *Client) Forward(ctx context.Context, req *pipeline.Request) (*http.Response, error) {
	u := *c.base
	u.Path = strings.TrimRight(c.base.Path, "/") + req.Path
	body := io.NopCloser(bytes.NewReader(req.Body))
	r, err := http.NewRequestWithContext(ctx, req.Method, u.String(), body)
	if err != nil {
		return nil, err
	}
	copyHeaders(r.Header, req.Header)
	copyHeaders(r.Header, req.UpstreamHeader)
	if c.credential != "" {
		if key := os.Getenv(c.credential); key != "" {
			r.Header.Set("Authorization", "Bearer "+key)
		}
	}
	return c.http.Do(r)
}

func copyHeaders(dst, src http.Header) {
	for k, values := range src {
		if hopByHop(k) || strings.EqualFold(k, "Authorization") || strings.EqualFold(k, "X-Api-Key") {
			continue
		}
		for _, v := range values {
			dst.Add(k, v)
		}
	}
}
func hopByHop(k string) bool {
	switch strings.ToLower(k) {
	case "connection", "keep-alive", "proxy-connection", "transfer-encoding", "upgrade", "te", "trailer":
		return true
	}
	return false
}

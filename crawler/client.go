package crawler

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

type URL string

type Result http.Response

var (
	ResultOK   = Result{Status: "OK"}
	ResultFAIL = Result{Status: "FAIL"}
)

func (r Result) content() (string, error) {
	if r.StatusCode < 200 || r.StatusCode >= 300 || r.Body == nil {
		return "", nil
	}
	var b bytes.Buffer
	if _, err := io.Copy(&b, r.Body); err != nil {
		return "", err
	}
	return b.String(), nil
}

func (u URL) String() string {
	return string(u)
}

type HTTPClient interface {
	Request(ctx context.Context, url URL) (Result, error)
}

type httpClient struct {
	client http.Client
}

func (c httpClient) Request(ctx context.Context, url URL) (Result, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return ResultFAIL, err
	}
	r, err := c.client.Do(req)
	if err != nil {
		return ResultFAIL, err
	}
	return Result(*r), nil
}

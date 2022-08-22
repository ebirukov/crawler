package crawler

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"crawler/log"
)

type URL string

func (u URL) hash() string {
	b := sha1.Sum([]byte(u))
	return hex.EncodeToString(b[:])
}

type URLs []URL

func (u URLs) hash() string {
	return URL(fmt.Sprintf("%v", u)).hash()
}

type Result struct {
	Status        string
	StatusCode    int
	Body          io.ReadCloser
	ContentLength int64
} //http.Response

func NewResult(r *http.Response) Result {
	return Result{
		Status:        r.Status,
		StatusCode:    r.StatusCode,
		Body:          r.Body,
		ContentLength: r.ContentLength,
	}
}

var (
	ResultOK     = Result{Status: "200 OK", StatusCode: 200, Body: http.NoBody}
	ResultFAIL   = Result{Status: "500 Internal Server Error", StatusCode: 500, Body: http.NoBody}
	ResultCANCEL = Result{}
)

func (r Result) content() (string, error) {
	if r.StatusCode < 200 || r.StatusCode >= 300 || r.Body == nil {
		return "", nil
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(r.Body)
	var b bytes.Buffer
	if _, err := io.Copy(&b, r.Body); err != nil {
		return "", err
	}
	return b.String(), nil
}

func (u URL) String() string {
	return string(u)
}

func WorkerHandler(client http.Client, metrics Metrics) WorkerFunc {
	logger := log.Adapter(log.Printer)
	return func(ctx context.Context, url URL) Result {
		if ctx.Err() != nil {
			metrics.IncRequestTimeout()
			return ResultCANCEL
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
		if err != nil {
			logger.Log(fmt.Sprintf("request handler: %v", err))
			return ResultFAIL
		}
		r, err := client.Do(req)
		if err != nil {
			//logger.Log(fmt.Sprintf("crawler handler: %v", err))
			metrics.IncRequestTimeout()
			return ResultFAIL
		}
		return NewResult(r)
	}
}

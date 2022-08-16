package crawler

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var MockWorkFn = func(ctx context.Context, url URL) Result {
	n := rand.Intn(50)
	select {
	case <-time.After(time.Duration(n * 1000)):
		return Result{Status: "OK"}
	case <-ctx.Done():
		return Result{Status: "FAIL"}
	}
}

func generateURLs(n int) []URL {
	urls := make([]URL, n)
	for i := 0; i < len(urls); i++ {
		urls[i] = URL(fmt.Sprintf("https://example.com?q=%d", i))
	}
	return urls
}

func BenchmarkWorker(b *testing.B) {

	pool := NewWorker(MockWorkFn, 100, 0)
	pool.logger = nil
	urls := generateURLs(100)
	var o Result
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out := pool.SubmitTasks(urls)
		for o = range out {
			if o.Status != "OK" {
				b.Fail()
			}
		}
	}
}

package crawler

import (
	"context"
)

type processor struct {
	client HTTPClient
	worker Worker
}

type Worker interface {
	Shutdown()
	SubmitTasks(urls []URL) <-chan Result
}

func New(client HTTPClient, worker Worker) *processor {
	return &processor{client: client, worker: worker}
}

func (p *processor) Work(ctx context.Context, url URL) Result {
	r, err := p.client.Request(ctx, url)
	if err != nil {
		return Result{Status: "FAIL"}
	}
	return r
}

func (p *processor) walk(urls []URL) ([]Result, error) {
	results := make([]Result, 0)
	for r := range p.worker.SubmitTasks(urls) {
		r.content()
		results = append(results, r)
	}
	return results, nil
}

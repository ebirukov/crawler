package crawler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"crawler/log"
)

type WorkerFunc func(ctx context.Context, url URL) Result

type worker struct {
	workFn  WorkerFunc
	limiter chan struct{}
	cancel  context.CancelFunc
	ctx     context.Context
	logger  log.Logger
}

func NewWorker(workFn WorkerFunc, rateLimit int, timeout time.Duration) *worker {
	ctx, cancel := context.WithCancel(context.Background())
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	}
	pool := &worker{
		workFn:  workFn,
		limiter: make(chan struct{}, rateLimit),
		cancel:  cancel,
		ctx:     ctx,
		logger:  log.Adapter(log.Printer),
	}
	//worker.start()
	return pool
}

func (p *worker) Shutdown() {
	p.cancel()
}

func (p *worker) SubmitTasks(urls []URL) <-chan Result {
	results := make(chan Result, len(p.limiter))
	wg := new(sync.WaitGroup)
	wg.Add(len(urls))
	for i, url := range urls {
		task := url
		th := i
		go func(url URL, th int) {
			defer func() {
				wg.Done()
				<-p.limiter
			}()
			select {
			case <-p.ctx.Done():
				return
			default:
				p.limiter <- struct{}{}
				break
			}
			p.log(th, task)
			r := p.workFn(p.ctx, task)
			results <- r
			return
		}(task, th)
	}
	go func() {
		defer close(results)
		wg.Wait()
	}()
	return results
}

func (p *worker) log(th int, task URL) {
	if p.logger == nil {
		return
	}
	p.logger.Log(fmt.Sprintf("th %d: ready to Work for task %s", th, task))
}

package crawler

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"crawler/log"
)

type workerV2 struct {
	workFn  WorkerFunc
	limiter chan struct{}
	results chan Result
	*sync.WaitGroup
	shutdown    bool
	cancel      context.CancelFunc
	ctx         context.Context
	logger      log.Logger
	execTimeout time.Duration
	timer       *time.Timer
	metrics     Metrics
}

func NewWorkerV2(workFn WorkerFunc, rateLimit int, timeout time.Duration, execTimeout time.Duration, metrics Metrics) *workerV2 {
	ctx, cancel := context.WithCancel(context.Background())
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	}
	pool := &workerV2{
		workFn:      workFn,
		limiter:     make(chan struct{}, rateLimit),
		cancel:      cancel,
		WaitGroup:   new(sync.WaitGroup),
		ctx:         ctx,
		logger:      log.Adapter(log.Printer),
		results:     make(chan Result),
		execTimeout: execTimeout,
		metrics:     metrics,
	}
	pool.timer = time.AfterFunc(execTimeout, pool.GracefulShutdown)
	return pool
}

func (p *workerV2) Shutdown() {
	p.cancel()
	p.GracefulShutdown()
}

func (p *workerV2) GracefulShutdown() {
	p.logger.Log("start graceful shutdown")
	p.shutdown = true
	p.logger.Log("wait for pool is complete")
	p.Wait()
	close(p.results)
	close(p.limiter)
	p.timer.Stop()
	p.logger.Log("pool was complete")
}

func (p *workerV2) SubmitTasks(urls []URL) <-chan Result {
	if len(urls) == 0 {
		return p.results
	}
	if p.shutdown {
		p.metrics.IncSkipped(len(urls))
		return p.results
	}
	p.Add(len(urls))
	for i, url := range urls {
		task := url
		th := i
		go func(url URL, th int) {
			defer func() {
				p.Done()
			}()

			p.metrics.IncSubmitted()
			for p.ctx.Err() == nil {
				select {
				case <-p.ctx.Done():
					p.log("cancelled", th, task)
					return
				case p.limiter <- struct{}{}:
					result := p.workFn(p.ctx, task)
					<-p.limiter
					p.results <- result
					p.metrics.IncProcessed()
					return
				default:
					runtime.Gosched()
				}
			}
		}(task, th)
	}
	return p.results
}

func (p *workerV2) log(message string, th int, task URL) {
	if p.logger == nil {
		return
	}
	p.logger.Log(fmt.Sprintf("th %d: %s for task %s", th, message, task))
}

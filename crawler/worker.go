package crawler

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"crawler/log"
)

type WorkerFunc func(ctx context.Context, url URL) Result

type WaitTask struct {
	id string
	*sync.WaitGroup
}

type worker struct {
	workFn      WorkerFunc
	limiter     chan struct{}
	results     chan Result
	wait        chan WaitTask
	shutdown    bool
	cancel      context.CancelFunc
	ctx         context.Context
	logger      log.Logger
	execTimeout time.Duration
	metrics     Metrics
}

func NewWorker(workFn WorkerFunc, rateLimit int, timeout time.Duration, execTimeout time.Duration) *worker {
	ctx, cancel := context.WithCancel(context.Background())
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	}
	pool := &worker{
		workFn:      workFn,
		limiter:     make(chan struct{}, rateLimit),
		cancel:      cancel,
		wait:        make(chan WaitTask),
		ctx:         ctx,
		logger:      log.Adapter(log.Printer),
		results:     make(chan Result, rateLimit),
		execTimeout: execTimeout,
	}
	go pool.waitComplete()
	return pool
}

func (p *worker) Shutdown() {
	p.shutdown = true
	p.cancel()
}

func (p *worker) GracefulShutdown() {
	p.shutdown = true
}

func (p *worker) SubmitTasks(urls []URL) <-chan Result {
	if len(urls) == 0 {
		return p.results
	}
	if p.shutdown {
		p.logger.Log("not submitted. Await shutdown")
		return p.results
	}
	wg := WaitTask{id: URLs(urls).hash(), WaitGroup: new(sync.WaitGroup)}
	wg.Add(len(urls))
	for i, url := range urls {
		task := url
		th := i
		go func(url URL, th int) {
			defer func() {
				wg.Done()
			}()

			for {
				select {
				case <-p.ctx.Done():
					p.log("cancelled", th, task)
					return
				case p.limiter <- struct{}{}:
					r := p.workFn(p.ctx, task)
					<-p.limiter
					p.log("retrieve result", th, task)
					for {
						if p.ctx.Err() != nil {
							return
						}
						select {
						case <-p.ctx.Done():
							return
						case p.results <- r:
							return
						default:
							runtime.Gosched()
						}
					}
				default:
					runtime.Gosched()
					break
				}
			}
		}(task, th)
	}
	go func() {
		p.logger.Log(fmt.Sprintf("submit works %s", wg.id))
		for {
			if p.ctx.Err() != nil {
				return
			}
			select {
			case <-p.ctx.Done():
				return
			case p.wait <- wg:
				return
			default:
				runtime.Gosched()
			}
		}
	}()
	return p.results
}

func (p *worker) waitComplete() {
	t := time.NewTicker(p.execTimeout)
	defer t.Stop()

	stop := make(chan struct{})

	go func() {
		<-t.C
		stop <- struct{}{}
		p.logger.Log("read timeout")
		return
	}()

	for {
		select {
		case wg := <-p.wait:
			t.Reset(p.execTimeout)
			wg.Wait()
			p.logger.Log(fmt.Sprintf("complete of works %s", wg.id))
			continue
		case <-stop:
			p.logger.Log("stop by read timeout")
			goto END
		}
	}
END:
	close(p.wait)
	p.logger.Log("await while a works was completed")
	for wg := range p.wait {
		wg.Wait()
		p.logger.Log(fmt.Sprintf("complete of works %s", wg.id))
	}
	close(p.results)
}

func (p *worker) log(message string, th int, task URL) {
	if p.logger == nil {
		return
	}
	p.logger.Log(fmt.Sprintf("th %d: %s for task %s", th, message, task))
}

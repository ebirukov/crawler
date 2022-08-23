package crawler

import (
	"fmt"
	"sync/atomic"
	"time"

	"crawler/log"
)

type metrics struct {
	proc      uint64
	skip      uint64
	submit    uint64
	rtimeout  uint64
	duplicate uint64
	ticker    *time.Ticker
	logger    log.Logger
}

func NewMetrics(logger log.Logger) *metrics {
	m := &metrics{
		logger: logger,
	}
	m.ticker = time.NewTicker(time.Second)
	go func() {
		for range m.ticker.C {
			m.logger.Log(
				fmt.Sprintf("\n \u001B[31m requests with timeout: %d \n wait processing: %d \n processed: %d \u001B[0m \n",
					m.rtimeout, m.submit-m.proc, m.proc))
		}
	}()
	//m.timer = time.AfterFunc(1*time.Second, m.Print)
	return m
}

func (m *metrics) Stop() {
	m.ticker.Stop()
}

func (m *metrics) IncProcessed() {
	atomic.AddUint64(&m.proc, 1)
}

func (m *metrics) IncSkipped(cnt int) {
	atomic.AddUint64(&m.skip, uint64(cnt))
}

func (m *metrics) IncSubmitted() {
	atomic.AddUint64(&m.submit, 1)
}

func (m *metrics) IncRequestTimeout() {
	atomic.AddUint64(&m.rtimeout, 1)
}

func (m *metrics) IncDuplicate() {
	atomic.AddUint64(&m.duplicate, 1)
}

func (m *metrics) Print() {
	m.logger.Log(
		fmt.Sprintf("duplicated urls: %d \n requests with timeout: %d \n submitted: %d \n processed: %d \n skipped: %d \n",
			m.duplicate, m.rtimeout, m.submit, m.proc, m.skip),
	)
}

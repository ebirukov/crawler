package crawler

import (
	"fmt"
	"io"
	_ "net/http/pprof"
	"net/url"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"golang.org/x/net/html"
)

type processor struct {
	worker  Worker
	metrics Metrics
	stopped bool
}

type Worker interface {
	Shutdown()
	SubmitTasks(urls []URL) <-chan Result
}

type Metrics interface {
	IncProcessed()
	IncSkipped(cnt int)
	IncSubmitted()
	IncRequestTimeout()
	IncDuplicate()
}

func New(worker Worker, metrics Metrics) *processor {
	p := &processor{worker: worker, metrics: metrics}
	time.AfterFunc(2*time.Minute, func() {
		p.stopped = true
	})
	return p
}

func (p *processor) Walk(urls []URL) (int, error) {
	bFilter := bloom.NewWithEstimates(10000000, 0.0001)
	//parsedUrls := make([]string, 0)
	out := p.worker.SubmitTasks(urls)
	for r := range out {
		r := r
		go func() {
			pu := ExtractLinks(r.Body)
			urls := make([]URL, len(pu))
			for i, url := range pu {
				if bFilter.Test([]byte(url)) {
					p.metrics.IncDuplicate()
					continue
				}
				bFilter.Add([]byte(url))
				urls[i] = URL(url)
			}
			if !p.stopped {
				p.worker.SubmitTasks(urls)
			}
		}()
	}
	actualFpRate := bloom.EstimateFalsePositiveRate(bFilter.Cap(), bFilter.K(), 10000000)
	println(actualFpRate)
	return 0, nil
}

func ExtractLinks(body io.ReadCloser) []string {
	urls := make([]string, 0)
	if body == nil {
		return urls
	}
	defer func() {
		if err := body.Close(); err != nil {
			panic(err)
		}
	}()
	tokenizer := html.NewTokenizer(body)
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			if tokenizer.Err() == io.EOF {
				return urls
			}
			fmt.Printf("Error next token: %v", tokenizer.Err())
			return urls
		}
		tag, hasAttr := tokenizer.TagName()
		if string(tag) != "a" {
			continue
		}
		//fmt.Printf("Tag: %v\n", string(tag))
		if hasAttr {
			for {
				attrKey, attrValue, moreAttr := tokenizer.TagAttr()
				if string(attrKey) != "href" {
					break
				}
				url, err := url.Parse(string(attrValue))
				if err != nil || url.Scheme == "" {
					break
				}
				//fmt.Printf("Attr: %v\n", string(attrKey))
				urls = append(urls, url.String())
				//fmt.Printf("Attr: %v\n", url)
				//fmt.Printf("Shema: %v\n", url.Scheme)
				//fmt.Printf("Attr: %v\n", moreAttr)
				if !moreAttr {
					break
				}
			}
		}
	}
}

package crawler

import (
	"fmt"
	"io"
	"net/url"

	"golang.org/x/net/html"
)

type processor struct {
	worker Worker
}

type Worker interface {
	Shutdown()
	SubmitTasks(urls []URL) <-chan Result
}

func New(worker Worker) *processor {
	return &processor{worker: worker}
}

func (p *processor) Walk(urls []URL) ([]Result, error) {
	results := make([]Result, 0)
	for r := range p.worker.SubmitTasks(urls) {
		p.ExtractLinks(r.Body)
		results = append(results, r)
	}
	return results, nil
}

func (p *processor) ExtractLinks(body io.ReadCloser) {
	if body == nil {
		return
	}
	defer body.Close()
	tokenizer := html.NewTokenizer(body)
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			if tokenizer.Err() == io.EOF {
				return
			}
			fmt.Printf("Error next token: %v", tokenizer.Err())
			return
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
				fmt.Printf("Attr: %v\n", url)
				//fmt.Printf("Shema: %v\n", url.Scheme)
				//fmt.Printf("Attr: %v\n", moreAttr)
				if !moreAttr {
					break
				}
			}
		}
	}
}

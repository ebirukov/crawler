package crawler

import (
	"fmt"
	"io"
	_ "net/http/pprof"
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

func (p *processor) Walk(urls []URL) (int, error) {

	//parsedUrls := make([]string, 0)
	out := p.worker.SubmitTasks(urls)
	for r := range out {
		r := r
		go func() {
			pu := ExtractLinks(r.Body)
			if len(pu) > 5 {
				pu = pu[:5]
			}
			urls := make([]URL, len(pu))
			for i, url := range pu {
				urls[i] = URL(url)
			}
			p.worker.SubmitTasks(urls)
		}()
	}
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

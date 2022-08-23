package main

import (
	"net/http"
	"time"

	"crawler/crawler"
	"crawler/log"
)

func main() {
	//go http.ListenAndServe("localhost:8080", nil)
	cl := http.Client{Timeout: 2000 * time.Millisecond}
	m := crawler.NewMetrics(log.Adapter(log.Printer))
	defer m.Stop()
	w := crawler.NewWorkerV2(crawler.WorkerHandler(cl, m), 1000, 300*time.Second, 300*time.Second, m)
	c := crawler.New(w, m)
	_, _ = c.Walk([]crawler.URL{"https://ru.wikipedia.org/wiki/%D0%92%D0%B8%D0%BA%D0%B8%D0%BF%D0%B5%D0%B4%D0%B8%D1%8F", "https://habr.com", "https://google.com", "https://habr.com/ru/post/571374/", "https://ru.wikipedia.org"})
	m.Print()
}

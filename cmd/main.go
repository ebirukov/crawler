package main

import (
	"net/http"
	"time"

	"crawler/crawler"
)

func main() {
	cl := http.Client{Timeout: 1 * time.Second}
	w := crawler.NewWorker(crawler.WorkerHandler(cl), 10, time.Second)
	c := crawler.New(w)
	c.Walk([]crawler.URL{"https://ru.wikipedia.org/wiki/%D0%92%D0%B8%D0%BA%D0%B8%D0%BF%D0%B5%D0%B4%D0%B8%D1%8F", "https://habr.com", "https://google.com", "https://habr.com/ru/post/571374/", "https://ru.wikipedia.org"})
}

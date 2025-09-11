package worker

import (
	"log"
	"net/http"
	"time"
)

type Endpoint struct {
	URL      string
	Interval time.Duration
}

func Start(endpoints []Endpoint) {
	for _, ep := range endpoints {
		go func(ep Endpoint) {
			ticker := time.NewTicker(ep.Interval)
			for range ticker.C {
				go func(url string) {
					resp, err := http.Get(url)
					if err != nil {
						log.Printf("error requesting %s: %v", url, err)
						return
					}
					log.Printf("GET %s -> %s", url, resp.Status)
					resp.Body.Close()
				}(ep.URL)
			}
		}(ep)
	}
}

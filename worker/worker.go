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
				resp, err := http.Get(ep.URL)
				if err != nil {
					log.Printf("error requesting %s: %v", ep.URL, err)
					continue
				}
				log.Printf("GET %s -> %s", ep.URL, resp.Status)
				resp.Body.Close()
			}
		}(ep)
	}
}

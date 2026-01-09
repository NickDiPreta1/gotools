package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"time"
)

func main() {
	requests := flag.Int("requests", 50, "How many requests to send")
	workers := flag.Int("workers", 10, "How many workers to use")
	url := flag.String("url", "", "Target URL to stress test")

	flag.Parse()

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	jobsChan := make(chan struct{})
	resultsChan := make(chan Result)

	start := time.Now()

	for i := 0; i < *workers; i++ {
		go worker(context.Background(), client, *url, jobsChan, resultsChan)
	}

	go func() {
		for i := 0; i < *requests; i++ {
			jobsChan <- struct{}{}
		}
		close(jobsChan)
	}()

	var results []Result

	for i := 0; i < *requests; i++ {
		res := <-resultsChan
		results = append(results, res)
	}

	close(resultsChan)

	duration := time.Since(start)

	var success, failed int
	var totalLatency time.Duration

	for _, r := range results {
		if r.Error != nil || r.Status < 200 || r.Status >= 300 {
			failed++
		} else {
			success++
		}
		totalLatency += r.Latency
	}

	rps := float64(*requests) / duration.Seconds()
	avgLatency := totalLatency / time.Duration(len(results))

	fmt.Printf("Requests:    %d\n", *requests)
	fmt.Printf("Success:     %d\n", success)
	fmt.Printf("Failed:      %d\n", failed)
	fmt.Printf("Duration:    %s\n", duration.Round(time.Millisecond))
	fmt.Printf("RPS:         %.2f\n", rps)
	fmt.Printf("Avg Latency: %s\n", avgLatency.Round(time.Millisecond))
}

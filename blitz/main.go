package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"slices"
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
	var errs int

	for i := 1; i <= *requests; i++ {
		res := <-resultsChan
		if res.Error != nil {
			errs++
		}
		results = append(results, res)
		duration := time.Since(start)
		rps := float64(i) / duration.Seconds()
		fmt.Printf("Running: %d/%d | %.2f req/s | Errors: %d\r", i, *requests, rps, errs)
	}
	fmt.Printf("\rRunning: %d/%d | %.2f req/s | Errors: %d\n", *requests, float64(*requests)/time.Since(start).Seconds(), errs)

	close(resultsChan)

	duration := time.Since(start)

	var success, failed int
	var totalLatency time.Duration
	var latencyList []time.Duration

	for _, r := range results {
		if r.Error != nil || r.Status < 200 || r.Status >= 300 {
			failed++
		} else {
			success++
		}
		latencyList = append(latencyList, r.Latency)
		totalLatency += r.Latency
	}

	rps := float64(*requests) / duration.Seconds()

	fmt.Printf("Requests:    %d\n", *requests)
	fmt.Printf("Success:     %d\n", success)
	fmt.Printf("Failed:      %d\n", failed)
	fmt.Printf("Duration:    %s\n", duration.Round(time.Millisecond))
	fmt.Printf("RPS:         %.2f\n", rps)

	if len(latencyList) > 0 {
		slices.Sort(latencyList)
		avgLatency := totalLatency / time.Duration(len(latencyList))

		p50Idx := len(latencyList) * 50 / 100
		p95Idx := len(latencyList) * 95 / 100
		p99Idx := len(latencyList) * 99 / 100

		// Clamp to valid range
		if p50Idx >= len(latencyList) {
			p50Idx = len(latencyList) - 1
		}
		if p95Idx >= len(latencyList) {
			p95Idx = len(latencyList) - 1
		}
		if p99Idx >= len(latencyList) {
			p99Idx = len(latencyList) - 1
		}

		fmt.Printf("Latency:\n")
		fmt.Printf("  Min:         %s\n", latencyList[0].Round(time.Millisecond))
		fmt.Printf("  Avg:         %s\n", avgLatency.Round(time.Millisecond))
		fmt.Printf("  P50:         %s\n", latencyList[p50Idx].Round(time.Millisecond))
		fmt.Printf("  P95:         %s\n", latencyList[p95Idx].Round(time.Millisecond))
		fmt.Printf("  P99:         %s\n", latencyList[p99Idx].Round(time.Millisecond))
		fmt.Printf("  Max:         %s\n", latencyList[len(latencyList)-1].Round(time.Millisecond))
	} else {
		fmt.Printf("Latency:     no successful requests\n")
	}
}

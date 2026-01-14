package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/NickDiPreta/gokit/cli"
)

func main() {
	requests := flag.Int("requests", 50, "How many requests to send")
	workers := flag.Int("workers", 10, "How many workers to use")
	url := flag.String("url", "", "Target URL to stress test")
	rate := flag.Int("rate", 0, "Set the maximum requests per second")

	flag.Parse()

	if *url == "" {
		fmt.Println(cli.Error("Error: URL is required"))
		flag.Usage()
		return
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	jobsChan := jobGenerator(*requests, *rate)
	resultsChan := make(chan Result)

	start := time.Now()

	for i := 0; i < *workers; i++ {
		go worker(context.Background(), client, *url, jobsChan, resultsChan)
	}

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
		fmt.Printf("Running: %d/%d | %.2f req/s | Errors: %d\r",
			i, *requests, rps, errs)
	}
	fmt.Println() // Clear the progress line

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

	// Summary Section
	fmt.Println("\n" + cli.Bold + "=== SUMMARY ===" + cli.Reset)
	summaryTable := cli.NewTable("Metric", "Value")
	summaryTable.AddRow("Total Requests", fmt.Sprintf("%d", *requests))
	summaryTable.AddRow("Successful", cli.Success(fmt.Sprintf("%d", success)))
	summaryTable.AddRow("Failed", cli.Error(fmt.Sprintf("%d", failed)))
	summaryTable.AddRow("Duration", duration.Round(time.Millisecond).String())
	summaryTable.AddRow("Requests/sec", fmt.Sprintf("%.2f", rps))
	summaryTable.Render()

	// Latency Section
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

		fmt.Println("\n" + cli.Bold + "=== LATENCY ===" + cli.Reset)
		latencyTable := cli.NewTable("Percentile", "Duration")
		latencyTable.AddRow("Min", latencyList[0].Round(time.Millisecond).String())
		latencyTable.AddRow("Average", avgLatency.Round(time.Millisecond).String())
		latencyTable.AddRow("P50 (Median)", latencyList[p50Idx].Round(time.Millisecond).String())
		latencyTable.AddRow("P95", latencyList[p95Idx].Round(time.Millisecond).String())
		latencyTable.AddRow("P99", latencyList[p99Idx].Round(time.Millisecond).String())
		latencyTable.AddRow("Max", latencyList[len(latencyList)-1].Round(time.Millisecond).String())
		latencyTable.Render()
	} else {
		fmt.Println("\n" + cli.Error("No successful requests"))
	}

	fmt.Println() // Final blank line for spacing
}

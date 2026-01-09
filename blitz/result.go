package main

import (
	"context"
	"io"
	"net/http"
	"time"
)

type Result struct {
	Status    int
	Latency   time.Duration
	Error     error
	Timestamp time.Time
}

func worker(ctx context.Context, client *http.Client, url string, jobs <-chan struct{}, results chan<- Result) {
	for range jobs {
		results <- makeRequest(ctx, client, url)
	}

}

// helper function used to make the http request so we can close the body cleanly
// don't want to risk leaving open in range loop
func makeRequest(ctx context.Context, client *http.Client, url string) Result {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{
			Error:     err,
			Timestamp: time.Now(),
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return Result{
			Error:     err,
			Timestamp: time.Now(),
		}
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return Result{
		Status:    resp.StatusCode,
		Latency:   time.Since(start),
		Timestamp: time.Now(),
	}
}

package main

import "time"

func jobGenerator(count, rate int) <-chan struct{} {
	jobsChan := make(chan struct{})

	var ticker *time.Ticker
	if rate > 0 {
		ticker = time.NewTicker(time.Second / time.Duration(rate))
	}

	go func() {
		if ticker != nil {
			defer ticker.Stop()
		}

		for i := 0; i < count; i++ {
			if ticker != nil {
				<-ticker.C
			}
			jobsChan <- struct{}{}
		}
		close(jobsChan)
	}()

	return jobsChan
}

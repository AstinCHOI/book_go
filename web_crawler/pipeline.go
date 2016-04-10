package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(n int, done <-chan struct{}, jobs <-chan int, c chan<- string) {
	for j := range jobs {
		select {
			case c <- fmt.Sprintf("Worker: %d, Job: %d", n, j):
			case <- done:
				return
		}
	}
}

func main() {
	jobs := make(chan int) // requesting jobs
	done := make(chan struct{}) // sending stop signal to jobs
	c := make(chan string) // saving results

	var wg sync.WaitGroup
	const numWorkers = 5
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func(n int) {
			worker(n, done, jobs, c)
			wg.Done()
		}(i)
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	go func() {
		for i := 0; i < 10; i++ {
			jobs <- i
			time.Sleep(10 * time.Millisecond)
		}
		close(done)
		close(jobs)
	}()

	for r := range c {
		fmt.Println(r)
	}
}
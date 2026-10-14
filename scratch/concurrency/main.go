// Listing 8: starting goroutines and sending data over a channel.
// Run with: go run ./scratch/concurrency
package main

import (
	"fmt"
	"sync"
)

func main() {
	results := make(chan int) // unbuffered channel of ints
	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(n int) { // "go" starts a goroutine
			defer wg.Done()
			results <- n * n // send blocks until main receives
		}(i)
	}
	go func() { wg.Wait(); close(results) }() // close the channel when all workers finish
	for r := range results {                  // receive until the channel is closed
		fmt.Println(r)
	}
}

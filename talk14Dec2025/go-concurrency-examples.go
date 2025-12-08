package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// Go Concurrency at Scale: Lessons from a Kubernetes Platform
// Live Code Examples
// ============================================================================

// ============================================================================
// EXAMPLE 1: Worker Pool Pattern
// ============================================================================

// BAD: Uncontrolled concurrency
func processItemsUncontrolled(items []int) {
	for _, item := range items {
		go func(i int) {
			// Simulate work
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("Processed: %d\n", i)
		}(item)
	}
	// Problem: No way to wait for completion!
	// Problem: If items = 10000, we spawn 10000 goroutines!
}

// GOOD: Worker Pool with bounded concurrency
func processItemsWithWorkerPool(items []int, batchSize int) {
	totalItems := len(items)

	for i := 0; i < totalItems; {
		// Calculate remaining batch
		remainingBatch := totalItems - i
		if remainingBatch < batchSize {
			batchSize = remainingBatch
		}

		var wg sync.WaitGroup
		for j := 0; j < batchSize; j++ {
			wg.Add(1)
			index := i + j

			go func(idx int) {
				defer wg.Done()
				// Simulate work
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				fmt.Printf("Processed item %d\n", items[idx])
			}(index)
		}

		wg.Wait() // Wait for current batch to complete
		i += batchSize
	}
}

// ============================================================================
// EXAMPLE 2: Fan-Out/Fan-In Pattern
// ============================================================================

type Result struct {
	Source string
	Data   interface{}
	Err    error
}

// Fetch data from multiple sources in parallel
func fetchFromMultipleSources(ctx context.Context) ([]Result, error) {
	var wg sync.WaitGroup
	results := make([]Result, 3)

	// Fan-out: Launch parallel operations
	wg.Add(3)

	// Fetch from Database
	go func() {
		defer wg.Done()
		time.Sleep(200 * time.Millisecond) // Simulate DB query
		results[0] = Result{
			Source: "database",
			Data:   "DB data",
			Err:    nil,
		}
	}()

	// Fetch from Cache
	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond) // Simulate cache lookup
		results[1] = Result{
			Source: "cache",
			Data:   "Cached data",
			Err:    nil,
		}
	}()

	// Fetch from External API
	go func() {
		defer wg.Done()
		time.Sleep(300 * time.Millisecond) // Simulate API call
		results[2] = Result{
			Source: "api",
			Data:   "API data",
			Err:    nil,
		}
	}()

	// Fan-in: Wait for all results
	wg.Wait()

	return results, nil
}

// ============================================================================
// EXAMPLE 3: Thread-Safe Result Collection with sync.Map
// ============================================================================

func processWithSyncMap(items []int) map[int]string {
	var wg sync.WaitGroup
	resultMap := &sync.Map{} // Thread-safe map

	for _, item := range items {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// Simulate processing
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			result := fmt.Sprintf("Result for %d", i)

			// Thread-safe store
			resultMap.Store(i, result)
		}(item)
	}

	wg.Wait()

	// Convert sync.Map to regular map
	regularMap := make(map[int]string)
	resultMap.Range(func(key, value interface{}) bool {
		regularMap[key.(int)] = value.(string)
		return true
	})

	return regularMap
}

// ============================================================================
// EXAMPLE 4: Atomic Counters for Thread-Safe Counting
// ============================================================================

func countWithAtomic(items []int, batchSize int) (processed, failed uint64) {
	var processedCount uint64 = 0
	var failedCount uint64 = 0

	totalItems := len(items)

	for i := 0; i < totalItems; {
		remainingBatch := totalItems - i
		if remainingBatch < batchSize {
			batchSize = remainingBatch
		}

		var wg sync.WaitGroup
		for j := 0; j < batchSize; j++ {
			wg.Add(1)
			index := i + j

			go func(idx int) {
				defer wg.Done()

				// Simulate processing with random success/failure
				time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

				if rand.Float32() > 0.2 { // 80% success rate
					atomic.AddUint64(&processedCount, 1)
				} else {
					atomic.AddUint64(&failedCount, 1)
				}
			}(index)
		}

		wg.Wait()
		i += batchSize
	}

	return processedCount, failedCount
}

// ============================================================================
// EXAMPLE 5: Context Cancellation
// ============================================================================

func processWithCancellation(ctx context.Context, items []int) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	for _, item := range items {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			// Check if context is cancelled
			select {
			case <-ctx.Done():
				fmt.Printf("Item %d: cancelled\n", i)
				return
			default:
			}

			// Simulate work
			for j := 0; j < 10; j++ {
				select {
				case <-ctx.Done():
					fmt.Printf("Item %d: cancelled during processing\n", i)
					return
				default:
					time.Sleep(50 * time.Millisecond)
				}
			}

			fmt.Printf("Item %d: completed\n", i)
		}(item)
	}

	// Wait for all goroutines
	wg.Wait()

	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

// ============================================================================
// EXAMPLE 6: Select Statement for Non-Blocking Operations
// ============================================================================

type Message struct {
	ID      int
	Content string
}

func broadcastWithSelect(messages []Message, clients []chan Message) {
	for _, msg := range messages {
		for _, client := range clients {
			select {
			case client <- msg:
				// Message sent successfully
				fmt.Printf("Sent message %d to client\n", msg.ID)
			default:
				// Client channel is full - skip this slow client
				fmt.Printf("Skipped slow client for message %d\n", msg.ID)
			}
		}
	}
}

// ============================================================================
// EXAMPLE 7: Graceful Shutdown Pattern
// ============================================================================

type Server struct {
	shutdown chan bool
	done     chan bool
}

func (s *Server) Start() {
	go s.run()
}

func (s *Server) run() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.shutdown:
			fmt.Println("Shutting down gracefully...")
			// Cleanup operations here
			s.done <- true
			return

		case <-ticker.C:
			fmt.Println("Server tick...")
		}
	}
}

func (s *Server) Stop() {
	s.shutdown <- true
	<-s.done // Wait for graceful shutdown
	fmt.Println("Server stopped")
}

// ============================================================================
// MAIN: Demo all patterns
// ============================================================================

func main() {
	fmt.Println("=== Go Concurrency Patterns Demo ===\n")

	// Example 1: Worker Pool
	fmt.Println("1. Worker Pool Pattern:")
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	processItemsWithWorkerPool(items, 3)
	fmt.Println()

	// Example 2: Fan-Out/Fan-In
	fmt.Println("2. Fan-Out/Fan-In Pattern:")
	start := time.Now()
	results, _ := fetchFromMultipleSources(context.Background())
	fmt.Printf("Fetched %d results in %v\n", len(results), time.Since(start))
	for _, r := range results {
		fmt.Printf("  - %s: %v\n", r.Source, r.Data)
	}
	fmt.Println()

	// Example 3: sync.Map
	fmt.Println("3. Thread-Safe Result Collection:")
	resultMap := processWithSyncMap([]int{1, 2, 3, 4, 5})
	fmt.Printf("Collected %d results\n", len(resultMap))
	fmt.Println()

	// Example 4: Atomic Counters
	fmt.Println("4. Atomic Counters:")
	processed, failed := countWithAtomic([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 3)
	fmt.Printf("Processed: %d, Failed: %d\n", processed, failed)
	fmt.Println()

	// Example 5: Context Cancellation
	fmt.Println("5. Context Cancellation:")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	processWithCancellation(ctx, []int{1, 2, 3})
	fmt.Println()

	// Example 6: Graceful Shutdown
	fmt.Println("6. Graceful Shutdown:")
	server := &Server{
		shutdown: make(chan bool),
		done:     make(chan bool),
	}
	server.Start()
	time.Sleep(3 * time.Second)
	server.Stop()
}

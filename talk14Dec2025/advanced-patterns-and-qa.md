# Advanced Patterns & Q&A Preparation
## Go Concurrency at Scale: Lessons from a Kubernetes Platform

## Advanced Patterns (If Time Permits)

### 1. errgroup Pattern

**When to use:** When you need coordinated error handling across goroutines

```go
import "golang.org/x/sync/errgroup"

func processWithErrGroup(ctx context.Context, items []string) error {
    g, ctx := errgroup.WithContext(ctx)
    
    for _, item := range items {
        item := item // Capture loop variable
        g.Go(func() error {
            return processItem(ctx, item)
        })
    }
    
    // Wait for all goroutines and return first error
    if err := g.Wait(); err != nil {
        return err
    }
    
    return nil
}
```

**Benefits:**
- Automatic error propagation
- Context cancellation on first error
- Cleaner than manual error channel management

**Real-world use case in Devtron:**
- Deploying to multiple environments
- If one fails, cancel all others
- Return the error to user immediately

---

### 2. Semaphore Pattern (Weighted Worker Pool)

**When to use:** Different tasks require different resources

```go
import "golang.org/x/sync/semaphore"

func processWithSemaphore(ctx context.Context, items []Item) error {
    // Allow max 10 concurrent operations
    sem := semaphore.NewWeighted(10)
    
    for _, item := range items {
        // Acquire semaphore (blocks if limit reached)
        if err := sem.Acquire(ctx, 1); err != nil {
            return err
        }
        
        go func(i Item) {
            defer sem.Release(1)
            processItem(i)
        }(item)
    }
    
    // Wait for all to complete
    if err := sem.Acquire(ctx, 10); err != nil {
        return err
    }
    
    return nil
}
```

**Advanced: Weighted semaphore**
```go
// Heavy tasks take more "weight"
sem := semaphore.NewWeighted(100)

// Light task: weight 1
sem.Acquire(ctx, 1)

// Heavy task: weight 10
sem.Acquire(ctx, 10)
```

---

### 3. Pipeline Pattern

**When to use:** Multi-stage processing

```go
func pipeline(ctx context.Context, input []int) <-chan int {
    // Stage 1: Generate
    gen := func() <-chan int {
        out := make(chan int)
        go func() {
            defer close(out)
            for _, n := range input {
                select {
                case out <- n:
                case <-ctx.Done():
                    return
                }
            }
        }()
        return out
    }
    
    // Stage 2: Square
    square := func(in <-chan int) <-chan int {
        out := make(chan int)
        go func() {
            defer close(out)
            for n := range in {
                select {
                case out <- n * n:
                case <-ctx.Done():
                    return
                }
            }
        }()
        return out
    }
    
    // Stage 3: Filter
    filter := func(in <-chan int) <-chan int {
        out := make(chan int)
        go func() {
            defer close(out)
            for n := range in {
                if n%2 == 0 {
                    select {
                    case out <- n:
                    case <-ctx.Done():
                        return
                    }
                }
            }
        }()
        return out
    }
    
    // Connect stages
    return filter(square(gen()))
}
```

---

### 4. Rate Limiter with Token Bucket

```go
import "golang.org/x/time/rate"

type RateLimitedProcessor struct {
    limiter *rate.Limiter
}

func NewRateLimitedProcessor(rps int) *RateLimitedProcessor {
    return &RateLimitedProcessor{
        limiter: rate.NewLimiter(rate.Limit(rps), rps),
    }
}

func (p *RateLimitedProcessor) Process(ctx context.Context, items []string) error {
    for _, item := range items {
        // Wait for rate limiter
        if err := p.limiter.Wait(ctx); err != nil {
            return err
        }
        
        // Process item
        if err := processItem(item); err != nil {
            return err
        }
    }
    return nil
}

// Burst handling
func (p *RateLimitedProcessor) ProcessWithBurst(ctx context.Context, items []string) error {
    for _, item := range items {
        // Allow burst of up to 10 items
        if err := p.limiter.WaitN(ctx, 1); err != nil {
            return err
        }
        
        go processItem(item)
    }
    return nil
}
```

---

### 5. Circuit Breaker Pattern

**When to use:** Protect against cascading failures

```go
type CircuitBreaker struct {
    maxFailures int
    resetTime   time.Duration
    failures    int
    lastFailure time.Time
    state       string // "closed", "open", "half-open"
    mu          sync.Mutex
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    // Check if circuit is open
    if cb.state == "open" {
        if time.Since(cb.lastFailure) > cb.resetTime {
            cb.state = "half-open"
        } else {
            return fmt.Errorf("circuit breaker is open")
        }
    }
    
    // Execute function
    err := fn()
    
    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()
        
        if cb.failures >= cb.maxFailures {
            cb.state = "open"
        }
        return err
    }
    
    // Success - reset
    cb.failures = 0
    cb.state = "closed"
    return nil
}
```

---

## Common Questions & Answers

### Q1: When should I use channels vs WaitGroup?

**A:** 
- **Use WaitGroup when:** You just need to wait for goroutines to complete
- **Use channels when:** You need to pass data between goroutines

```go
// WaitGroup: Just coordination
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    doWork()
}()
wg.Wait()

// Channels: Data passing
results := make(chan Result)
go func() {
    results <- doWork()
}()
result := <-results
```

---

### Q2: How do I handle errors from goroutines?

**A:** Three approaches:

**1. Error channel:**
```go
errChan := make(chan error, len(items))
for _, item := range items {
    go func(i Item) {
        if err := process(i); err != nil {
            errChan <- err
        }
    }(item)
}

// Collect errors
for i := 0; i < len(items); i++ {
    if err := <-errChan; err != nil {
        log.Error(err)
    }
}
```

**2. errgroup (recommended):**
```go
g, ctx := errgroup.WithContext(ctx)
for _, item := range items {
    item := item
    g.Go(func() error {
        return process(ctx, item)
    })
}
return g.Wait() // Returns first error
```

**3. Collect all errors:**
```go
var mu sync.Mutex
var errs []error

var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(i Item) {
        defer wg.Done()
        if err := process(i); err != nil {
            mu.Lock()
            errs = append(errs, err)
            mu.Unlock()
        }
    }(item)
}
wg.Wait()
```

---

### Q3: What's the difference between buffered and unbuffered channels?

**A:**

**Unbuffered (synchronous):**
```go
ch := make(chan int)
ch <- 1  // Blocks until someone receives
```

**Buffered (asynchronous up to capacity):**
```go
ch := make(chan int, 10)
ch <- 1  // Doesn't block (until buffer is full)
```

**When to use buffered:**
- Producer/consumer with different speeds
- Prevent goroutine blocking
- Known maximum queue size

**When to use unbuffered:**
- Synchronization point needed
- Guaranteed delivery
- Simple coordination

---

### Q4: How do I prevent goroutine leaks?

**A:** Common causes and solutions:

**1. Channel never closed:**
```go
// BAD
ch := make(chan int)
go func() {
    for v := range ch {  // Waits forever if ch never closed
        process(v)
    }
}()

// GOOD
ch := make(chan int)
go func() {
    for v := range ch {
        process(v)
    }
}()
// ... later
close(ch)  // Goroutine exits
```

**2. Context not propagated:**
```go
// BAD
go func() {
    for {
        doWork()  // Never exits
    }
}()

// GOOD
go func() {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            doWork()
        }
    }
}()
```

---

### Q5: What's the overhead of goroutines?

**A:**
- **Memory:** ~2KB per goroutine (stack)
- **Creation time:** ~1-2 microseconds
- **Context switch:** ~200 nanoseconds

**Practical limits:**
- Can easily run 10,000+ goroutines
- Seen production systems with 100,000+ goroutines
- Real limit is usually memory, not goroutines themselves

**But:** Always use worker pools for:
- External API calls (rate limiting)
- Database connections (connection pool limits)
- File I/O (OS limits)

---

### Q6: How do I debug concurrent code?

**A:** Tools and techniques:

**1. Race detector:**
```bash
go test -race
go run -race main.go
```

**2. Logging with goroutine ID:**
```go
import "runtime"

func getGoroutineID() uint64 {
    // ... implementation
}

log.Printf("[goroutine %d] Processing item", getGoroutineID())
```

**3. Deadlock detection:**
```go
// Go runtime detects deadlocks automatically
// Will panic with "fatal error: all goroutines are asleep - deadlock!"
```

**4. pprof for goroutine profiling:**
```go
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Visit http://localhost:6060/debug/pprof/goroutine
```

---

### Q7: Should I use sync.Mutex or channels?

**A:** "Share memory by communicating, don't communicate by sharing memory"

**Use Mutex when:**
- Protecting simple state (counters, flags)
- Short critical sections
- Performance critical (mutex is faster)

```go
var mu sync.Mutex
var counter int

mu.Lock()
counter++
mu.Unlock()
```

**Use Channels when:**
- Passing ownership of data
- Coordinating goroutines
- Implementing patterns (pipeline, fan-out)

```go
results := make(chan Result)
go func() {
    results <- computeResult()
}()
result := <-results
```

---

### Q8: How do I implement timeouts?

**A:** Use context.WithTimeout:

```go
// Simple timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result, err := doWork(ctx)
if err == context.DeadlineExceeded {
    log.Println("Operation timed out")
}

// Per-operation timeout
func processWithTimeout(item Item) error {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()
    
    done := make(chan error, 1)
    go func() {
        done <- process(item)
    }()
    
    select {
    case err := <-done:
        return err
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

---

### Q9: What's the best batch size for worker pools?

**A:** It depends on:

**CPU-bound tasks:**
- Batch size = Number of CPU cores
- `runtime.NumCPU()`

**I/O-bound tasks:**
- Batch size = 10-100 (experiment!)
- Depends on:
  - External service rate limits
  - Connection pool size
  - Memory constraints

**Devtron's approach:**
```go
// Configurable batch sizes
type Config struct {
    CIAutoTriggerBatchSize int  // Default: 5
    ResourceFetchBatchSize int  // Default: 10
    EnforcerBatchSize      int  // Default: 10
}
```

**Rule of thumb:**
- Start with 10
- Measure performance
- Increase until diminishing returns
- Monitor resource usage

---

### Q10: How do I test concurrent code?

**A:** Strategies:

**1. Use race detector:**
```go
func TestConcurrent(t *testing.T) {
    // Run with: go test -race
    var wg sync.WaitGroup
    counter := 0
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter++ // Race detector will catch this!
        }()
    }
    wg.Wait()
}
```

**2. Stress testing:**
```go
func TestStress(t *testing.T) {
    for i := 0; i < 1000; i++ {
        t.Run(fmt.Sprintf("iteration-%d", i), func(t *testing.T) {
            t.Parallel()
            // Your concurrent code here
        })
    }
}
```

**3. Table-driven tests:**
```go
func TestWorkerPool(t *testing.T) {
    tests := []struct {
        name      string
        items     int
        batchSize int
        want      int
    }{
        {"small", 10, 2, 10},
        {"large", 1000, 10, 1000},
        {"edge", 1, 1, 1},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := processWithWorkerPool(tt.items, tt.batchSize)
            if result != tt.want {
                t.Errorf("got %d, want %d", result, tt.want)
            }
        })
    }
}
```

---

## Performance Benchmarking

### How to benchmark concurrent code:

```go
func BenchmarkSequential(b *testing.B) {
    items := generateItems(100)
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        for _, item := range items {
            process(item)
        }
    }
}

func BenchmarkConcurrent(b *testing.B) {
    items := generateItems(100)
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        var wg sync.WaitGroup
        for _, item := range items {
            wg.Add(1)
            go func(i Item) {
                defer wg.Done()
                process(i)
            }(item)
        }
        wg.Wait()
    }
}

func BenchmarkWorkerPool(b *testing.B) {
    items := generateItems(100)
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        processWithWorkerPool(items, 10)
    }
}
```

Run with:
```bash
go test -bench=. -benchmem -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

---

## Additional Resources

### Books:
- "Concurrency in Go" by Katherine Cox-Buday
- "The Go Programming Language" by Donovan & Kernighan

### Talks:
- "Go Concurrency Patterns" by Rob Pike
- "Advanced Go Concurrency Patterns" by Sameer Ajmani

### Tools:
- Race detector: `go test -race`
- pprof: Profiling tool
- Delve: Go debugger with goroutine support

### Practice:
- Implement your own worker pool
- Build a rate limiter
- Create a concurrent web scraper
- Contribute to Devtron!


# Go Concurrency at Scale - Quick Reference Card

## Pattern Cheat Sheet

### 1. Worker Pool (Bounded Concurrency)

**When:** Processing many items, need to control resource usage

**Code:**
```go
batchSize := 10
for i := 0; i < len(items); {
    remainingBatch := len(items) - i
    if remainingBatch < batchSize {
        batchSize = remainingBatch
    }
    
    var wg sync.WaitGroup
    for j := 0; j < batchSize; j++ {
        wg.Add(1)
        index := i + j
        go func(idx int) {
            defer wg.Done()
            process(items[idx])
        }(index)
    }
    wg.Wait()
    i += batchSize
}
```

**Key Points:**
- ✅ Bounded concurrency
- ✅ Predictable resource usage
- ✅ No goroutine explosion

---

### 2. Fan-Out/Fan-In

**When:** Parallel independent operations, aggregate results

**Code:**
```go
var wg sync.WaitGroup
results := make([]Result, 3)

wg.Add(3)

go func() {
    defer wg.Done()
    results[0] = fetchFromDB()
}()

go func() {
    defer wg.Done()
    results[1] = fetchFromCache()
}()

go func() {
    defer wg.Done()
    results[2] = fetchFromAPI()
}()

wg.Wait()
// Use results
```

**Key Points:**
- ✅ Parallel execution
- ✅ Wait for all results
- ✅ Faster than sequential

---

### 3. Thread-Safe Result Collection

**When:** Multiple goroutines need to store results

**Option A: sync.Map**
```go
var wg sync.WaitGroup
resultMap := &sync.Map{}

for _, item := range items {
    wg.Add(1)
    go func(i Item) {
        defer wg.Done()
        result := process(i)
        resultMap.Store(i.ID, result)
    }(item)
}
wg.Wait()
```

**Option B: Mutex**
```go
var wg sync.WaitGroup
var mu sync.Mutex
results := make(map[int]Result)

for _, item := range items {
    wg.Add(1)
    go func(i Item) {
        defer wg.Done()
        result := process(i)
        mu.Lock()
        results[i.ID] = result
        mu.Unlock()
    }(item)
}
wg.Wait()
```

**Key Points:**
- ✅ Thread-safe
- ✅ No race conditions
- sync.Map: Better for many reads
- Mutex: Better for simple cases

---

### 4. Atomic Counters

**When:** Simple counters, flags, no complex state

**Code:**
```go
var processed uint64 = 0
var failed uint64 = 0

var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(i Item) {
        defer wg.Done()
        if process(i) {
            atomic.AddUint64(&processed, 1)
        } else {
            atomic.AddUint64(&failed, 1)
        }
    }(item)
}
wg.Wait()

fmt.Printf("Processed: %d, Failed: %d\n", processed, failed)
```

**Key Points:**
- ✅ Lock-free
- ✅ Fast
- ✅ Simple
- ❌ Only for basic types

---

### 5. Context Cancellation

**When:** Need to cancel operations, timeouts, HTTP requests

**Code:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

for _, item := range items {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        process(ctx, item)
    }
}
```

**HTTP Handler:**
```go
ctx, cancel := context.WithCancel(r.Context())
defer cancel()

// Detect client disconnect
if cn, ok := w.(http.CloseNotifier); ok {
    go func(done <-chan struct{}, closed <-chan bool) {
        select {
        case <-done:
        case <-closed:
            cancel()
        }
    }(ctx.Done(), cn.CloseNotify())
}

result, err := service.Process(ctx, data)
```

**Key Points:**
- ✅ Graceful cancellation
- ✅ Timeout support
- ✅ Resource cleanup

---

### 6. Select Statement (Non-Blocking)

**When:** Multiple channel operations, timeouts, defaults

**Code:**
```go
select {
case msg := <-ch1:
    // Received from ch1
case ch2 <- value:
    // Sent to ch2
case <-time.After(1 * time.Second):
    // Timeout
case <-ctx.Done():
    // Cancelled
default:
    // Non-blocking: do this if nothing else ready
}
```

**Key Points:**
- ✅ Non-blocking operations
- ✅ Timeout handling
- ✅ Multiple channels

---

### 7. errgroup (Error Handling)

**When:** Need coordinated error handling

**Code:**
```go
import "golang.org/x/sync/errgroup"

g, ctx := errgroup.WithContext(ctx)

for _, item := range items {
    item := item // Capture
    g.Go(func() error {
        return process(ctx, item)
    })
}

if err := g.Wait(); err != nil {
    return err // First error
}
```

**Key Points:**
- ✅ Automatic error propagation
- ✅ Context cancellation on error
- ✅ Cleaner than manual channels

---

### 8. Semaphore (Weighted Concurrency)

**When:** Different tasks need different resources

**Code:**
```go
import "golang.org/x/sync/semaphore"

sem := semaphore.NewWeighted(10)

for _, item := range items {
    if err := sem.Acquire(ctx, 1); err != nil {
        return err
    }
    
    go func(i Item) {
        defer sem.Release(1)
        process(i)
    }(item)
}

// Wait for all
if err := sem.Acquire(ctx, 10); err != nil {
    return err
}
```

**Key Points:**
- ✅ Weighted resources
- ✅ Context-aware
- ✅ Flexible limits

---

## Common Pitfalls & Solutions

### Pitfall 1: Goroutine Leak
```go
// ❌ BAD: Goroutine never exits
go func() {
    for {
        doWork()
    }
}()

// ✅ GOOD: Exit condition
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

### Pitfall 2: Forgot wg.Done()
```go
// ❌ BAD: Deadlock!
wg.Add(1)
go func() {
    doWork()
    // Forgot wg.Done()!
}()
wg.Wait() // Waits forever

// ✅ GOOD: Always defer
wg.Add(1)
go func() {
    defer wg.Done()
    doWork()
}()
wg.Wait()
```

---

### Pitfall 3: Race Condition
```go
// ❌ BAD: Race condition
var counter int
for i := 0; i < 100; i++ {
    go func() {
        counter++ // RACE!
    }()
}

// ✅ GOOD: Atomic
var counter uint64
for i := 0; i < 100; i++ {
    go func() {
        atomic.AddUint64(&counter, 1)
    }()
}
```

---

### Pitfall 4: Closing Closed Channel
```go
// ❌ BAD: Panic!
close(ch)
close(ch) // Panic!

// ✅ GOOD: Close once
var once sync.Once
once.Do(func() {
    close(ch)
})
```

---

### Pitfall 5: Loop Variable Capture
```go
// ❌ BAD: All goroutines see last value
for _, item := range items {
    go func() {
        process(item) // Wrong item!
    }()
}

// ✅ GOOD: Capture variable
for _, item := range items {
    item := item // Capture
    go func() {
        process(item)
    }()
}

// ✅ ALSO GOOD: Pass as parameter
for _, item := range items {
    go func(i Item) {
        process(i)
    }(item)
}
```

---

## Decision Tree

### Should I use concurrency?

```
Is the operation I/O bound? (network, disk, database)
├─ YES → Use concurrency ✅
└─ NO → Is it CPU-bound?
    ├─ YES → Use worker pool with runtime.NumCPU() workers
    └─ NO → Don't use concurrency (keep it simple)

How many operations?
├─ < 10 → Simple goroutines with WaitGroup
├─ 10-100 → Worker pool (batch size: 10)
├─ 100-1000 → Worker pool (batch size: 10-50)
└─ > 1000 → Worker pool + consider streaming/pagination

Do I need error handling?
├─ YES, stop on first error → Use errgroup
├─ YES, collect all errors → Use error channel or mutex
└─ NO → Use WaitGroup

Do I need to collect results?
├─ YES, simple map → Use sync.Map or mutex
├─ YES, complex aggregation → Use channels
└─ NO → Just use WaitGroup

Do I need cancellation?
├─ YES → Use context.Context
└─ NO → Still use context for future-proofing

Do I need rate limiting?
├─ YES, external API → Use worker pool or rate.Limiter
├─ YES, database → Use worker pool (match connection pool size)
└─ NO → Still use worker pool for safety
```

---

## Performance Guidelines

### Batch Sizes

| Workload Type | Recommended Batch Size |
|---------------|------------------------|
| CPU-bound | `runtime.NumCPU()` |
| Database queries | 5-10 (match connection pool) |
| HTTP API calls | 10-50 (check rate limits) |
| Kubernetes API | 5-10 (avoid throttling) |
| File I/O | 10-20 |

### Memory Considerations

| Item | Memory per Unit |
|------|-----------------|
| Goroutine | ~2KB (initial stack) |
| Channel (unbuffered) | ~96 bytes |
| Channel (buffered) | ~96 bytes + (element size × capacity) |
| sync.WaitGroup | ~12 bytes |
| sync.Mutex | ~8 bytes |

---

## Testing Checklist

- [ ] Run with race detector: `go test -race`
- [ ] Test with high iteration count (1000+)
- [ ] Test with different batch sizes
- [ ] Test cancellation scenarios
- [ ] Test error handling
- [ ] Benchmark: `go test -bench=.`
- [ ] Profile: `go test -cpuprofile=cpu.prof`
- [ ] Check for goroutine leaks

---

## Debugging Commands

```bash
# Race detection
go test -race ./...
go run -race main.go

# Benchmarking
go test -bench=. -benchmem

# CPU profiling
go test -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof
go tool pprof mem.prof

# Goroutine profiling (in running app)
curl http://localhost:6060/debug/pprof/goroutine

# Trace
go test -trace=trace.out
go tool trace trace.out
```

---

## Import Statements

```go
import (
    "context"
    "sync"
    "sync/atomic"
    "time"
    
    // Extended packages
    "golang.org/x/sync/errgroup"
    "golang.org/x/sync/semaphore"
    "golang.org/x/time/rate"
)
```

---

## Real-World Examples from Devtron

| Pattern | File | Use Case |
|---------|------|----------|
| Worker Pool | `pkg/workflow/dag/WorkflowDagExecutor.go` | CI auto-trigger |
| Fan-Out/Fan-In | `api/restHandler/app/pipeline/configure/PipelineConfigRestHandler.go` | Fetch CI/CD status |
| Rate Limiting | `pkg/k8s/K8sCommonService.go` | K8s resource batch fetch |
| Context Cancel | `api/cluster/ClusterRestHandler.go` | HTTP handlers |
| SSE Broker | `api/sse/Broker.go` | Real-time updates |
| Atomic Counters | `pkg/appStore/installedApp/service/FullMode/resource/ResourceTreeService.go` | Hibernation check |
| Thread-Safe Map | `pkg/cluster/ClusterService.go` | Cluster connection test |

---

## Quick Tips

1. **Always defer wg.Done()** - Prevents deadlocks
2. **Capture loop variables** - Avoid closure bugs
3. **Use context everywhere** - Future-proof your code
4. **Start with WaitGroup** - Simplest coordination
5. **Measure before optimizing** - Don't guess
6. **Use race detector** - Catch bugs early
7. **Limit concurrency** - Worker pools are your friend
8. **Handle errors** - Don't ignore goroutine errors
9. **Close channels** - Prevent goroutine leaks
10. **Test thoroughly** - Concurrent bugs are sneaky

---

## One-Liners

```go
// Wait for goroutines
var wg sync.WaitGroup; wg.Add(1); go func() { defer wg.Done(); work() }(); wg.Wait()

// Timeout
ctx, cancel := context.WithTimeout(ctx, 5*time.Second); defer cancel()

// Atomic increment
atomic.AddUint64(&counter, 1)

// Non-blocking send
select { case ch <- value: default: }

// Non-blocking receive
select { case v := <-ch: default: }

// Thread-safe map store
var m sync.Map; m.Store(key, value)

// Thread-safe map load
if v, ok := m.Load(key); ok { use(v) }
```

---

## Remember

> "Concurrency is about dealing with lots of things at once.
> Parallelism is about doing lots of things at once."
> — Rob Pike

**Start simple. Add complexity only when needed. Measure everything.**

---

Print this card and keep it handy! 📋


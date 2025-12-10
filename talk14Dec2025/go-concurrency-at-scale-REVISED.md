# Go Concurrency at Scale: Lessons from a Kubernetes Platform

**Duration:** 35 minutes
**Audience:** Intermediate to Advanced Go Developers
**Focus:** Deep dive into scaling Go concurrency in production

---

## Talk Outline

### 1. Introduction: The Scale Problem (3 mins)
### 2. The Evolution: From Naive to Scalable (8 mins)
### 3. Pattern 1: Worker Pool (Bounded Concurrency) (10 mins)
### 4. Pattern 2: Fan-Out/Fan-In (Parallel Aggregation) (10 mins)
### 5. Key Takeaways & Production Lessons (4 mins)

---

## 1. Introduction: The Scale Problem (3 mins)

### About Devtron

**What is Devtron?**
- Open-source Kubernetes deployment platform
- Manages 1000+ applications across 100+ clusters
- Handles 10,000+ deployments daily
- Processes thousands of CI/CD pipelines concurrently

**Our Concurrency Challenges:**
- Triggering 100+ CI pipelines simultaneously
- Fetching resources from 100+ Kubernetes clusters in parallel
- Processing thousands of webhook events
- Real-time status updates for thousands of deployments

---
### What This Talk Covers

We'll explore the **evolution of concurrency patterns** and **2 critical patterns** for scaling:

**The Evolution:**
- Approach 1: Naive (spawn goroutine per task) → Crashes
- Approach 2: sync.WaitGroup (wait for completion) → Still crashes at scale
- Approach 3: Worker Pool (bounded concurrency) → Scales!

**Pattern 1: Bounded Concurrency (Worker Pools)**
- How to process 10,000 items with controlled concurrency
- Line-by-line code walkthrough
- Production metrics: Before/After

**Pattern 2: Fan-Out/Fan-In (Parallel Aggregation)**
- How to fetch from 100+ sources in parallel and safely combine results
- Why we use sync.Map vs regular map + mutex vs pre-allocated slice
- Real examples: Cluster connection, workflow status fetching

---

## 2. The Evolution: From Naive to Scalable (8 mins)

### Real-World Analogy: The Restaurant Story

**Imagine you own a restaurant that just got popular:**

**Week 1: Small restaurant (10 customers/day)**
- You cook all orders yourself
- Works perfectly fine!
- Customers are happy

**Week 2: Featured on TV (1000 customers/day)**
- You try the same approach: cook all orders yourself
- Problem: You're overwhelmed!
- Customers wait hours for food
- Kitchen runs out of ingredients
- You collapse from exhaustion

**This is exactly what happened to our system:**
- **Week 1** = Development environment (10 apps)
- **Week 2** = Production environment (1000+ apps)
- **You cooking alone** = Simple goroutines without control
- **Overwhelmed kitchen** = System crashes

Let's see how we evolved our approach, just like a restaurant evolves from a small kitchen to a professional operation.

---

### The Critical Question

**Scenario:** You need to process 1000 tasks concurrently (trigger pipelines, send emails, process files, etc.)

**Question:** How do you implement this in Go?

Let's see 3 approaches and understand why only one scales.

---

### Approach 1: Naive - Spawn Goroutine Per Task (❌ FAILS)

**Restaurant Analogy:**
- 1000 customers walk in
- You hire 1000 chefs on the spot
- Your kitchen has space for only 10 chefs
- Result: Chaos! Chefs bumping into each other, kitchen on fire!

**The Code:**

```go
// Approach 1: Naive - Just spawn goroutines
func ProcessTasksNaive(tasks []Task) {
    for _, task := range tasks {
        go func(t Task) {
            processTask(t)  // Do the work
        }(task)
    }
    // Function returns immediately!
}
```

**What happens:**

```
Main goroutine:
  ├─→ Spawns 1000 goroutines
  └─→ Returns immediately (doesn't wait!)

Result: Function returns before any task completes!
```

**Problems:**

1. **❌ No waiting** - Function returns before tasks complete
2. **❌ No error handling** - Can't collect errors
3. **❌ No result collection** - Can't get results back
4. **❌ Unbounded concurrency** - All 1000 goroutines run at once

**When this fails:**
- 1000 tasks: Might work (if tasks are light)
- 10,000 tasks: Probably crashes (OOM, resource exhaustion)
- 100,000 tasks: Definitely crashes

**Real example:**
```go
// This is what we did initially at Devtron
for _, pipeline := range pipelines {  // 100 pipelines
    go func(p Pipeline) {
        impl.triggerCiPipeline(p)  // Each spawns ~50 more goroutines
    }(pipeline)
}
// Result: 100 × 50 = 5,000 goroutines → OOM crash!
```

---

### Approach 2: sync.WaitGroup - Wait for Completion (⚠️ BETTER, BUT STILL FAILS AT SCALE)

**Restaurant Analogy:**
- 1000 customers walk in
- You hire 1000 chefs
- You wait for all chefs to finish before closing the restaurant
- Problem: Still 1000 chefs in a kitchen built for 10!
- Result: Better than before (you wait), but kitchen still overwhelmed!

**The Code:**

```go
// Approach 2: Use sync.WaitGroup to wait for completion
func ProcessTasksWithWaitGroup(tasks []Task) error {
    var wg sync.WaitGroup

    for _, task := range tasks {
        wg.Add(1)  // Increment counter

        go func(t Task) {
            defer wg.Done()  // Decrement counter when done
            processTask(t)
        }(task)
    }

    wg.Wait()  // Wait for all goroutines to complete
    return nil
}
```

**What happens:**

```
Main goroutine:
  ├─→ Spawns 1000 goroutines (all at once)
  ├─→ Waits for all to complete
  └─→ Returns after all complete

Result: Function waits, but still spawns 1000 goroutines!
```

**What's better:**
- ✅ **Waits for completion** - Function doesn't return early
- ✅ **Proper cleanup** - defer wg.Done() ensures counter is decremented
- ✅ **Synchronization** - Main goroutine waits for all workers

**What's still wrong:**
- ❌ **Still unbounded** - All 1000 goroutines run simultaneously
- ❌ **Resource exhaustion** - Can still crash with OOM, DB exhaustion, API rate limits
- ❌ **No control** - Can't limit concurrency

**The key insight:**

> **sync.WaitGroup solves the "waiting" problem, but NOT the "too many goroutines" problem!**

**When this fails:**

| Tasks | Goroutines | Result |
|-------|------------|--------|
| 10 | 10 | ✅ Works fine |
| 100 | 100 | ⚠️ Might work (depends on resources) |
| 1,000 | 1,000 | ❌ Likely fails (DB connections, API limits) |
| 10,000 | 10,000 | ❌ Definitely crashes (OOM, thrashing) |

**Real example at Devtron:**

```go
// We tried this next - added WaitGroup
var wg sync.WaitGroup

for _, pipeline := range pipelines {  // 100 pipelines
    wg.Add(1)
    go func(p Pipeline) {
        defer wg.Done()
        impl.triggerCiPipeline(p)  // Each needs DB connection
    }(pipeline)
}

wg.Wait()

// Result:
// - ✅ Function waits for completion
// - ❌ Still spawns 100 goroutines at once
// - ❌ Database: "pq: sorry, too many clients already" (max 100 connections)
// - ❌ Kubernetes API: 429 Too Many Requests (rate limit exceeded)
```

**The problem:**
- We have 100 database connections available
- We spawn 100 goroutines
- Each goroutine tries to get a DB connection
- But other parts of the application also need connections!
- Result: Connection pool exhausted → Crash

---

### Approach 3: Worker Pool - Bounded Concurrency (✅ SCALES!)

**Restaurant Analogy:**
- 1000 customers walk in
- You hire only 5 professional chefs (your kitchen capacity)
- First 5 customers: 5 chefs cook → customers served
- Next 5 customers: Same 5 chefs cook → customers served
- Continue in batches of 5 until all 1000 customers served
- Result: Organized, efficient, no chaos!

**The Code:**

```go
// Approach 3: Worker Pool - Control concurrency with batching
func ProcessTasksWithWorkerPool(tasks []Task) error {
    batchSize := 5  // Only 5 goroutines at a time!

    for i := 0; i < len(tasks); {
        // Calculate batch size (last batch might be smaller)
        remainingTasks := len(tasks) - i
        if remainingTasks < batchSize {
            batchSize = remainingTasks
        }

        var wg sync.WaitGroup

        // Launch only batchSize goroutines
        for j := 0; j < batchSize; j++ {
            wg.Add(1)
            index := i + j

            go func(idx int) {
                defer wg.Done()
                processTask(tasks[idx])
            }(index)
        }

        wg.Wait()  // Wait for this batch to complete
        i += batchSize  // Move to next batch
    }

    return nil
}
```

**What happens:**

```
Main goroutine:
  ├─→ Batch 1: Spawns 5 goroutines → Waits for completion
  ├─→ Batch 2: Spawns 5 goroutines → Waits for completion
  ├─→ Batch 3: Spawns 5 goroutines → Waits for completion
  |   ... (200 batches for 1000 tasks)
  └─→ Returns after all batches complete

Result: Max 5 goroutines at any time, processes all 1000 tasks!
```

**What's better:**
- ✅ **Bounded concurrency** - Never more than 5 goroutines at once
- ✅ **Predictable resource usage** - Max 5 DB connections, 5 API calls
- ✅ **Scales to any number** - 1000 tasks? 10,000? 1,000,000? Same max 5 goroutines
- ✅ **Tunable** - Adjust batchSize based on your constraints

**When this works:**

| Tasks | Max Goroutines | Result |
|-------|----------------|--------|
| 10 | 5 | ✅ Works |
| 100 | 5 | ✅ Works |
| 1,000 | 5 | ✅ Works |
| 10,000 | 5 | ✅ Works |
| 1,000,000 | 5 | ✅ Works (takes longer, but doesn't crash) |

---

### Side-by-Side Comparison

**Processing 1000 tasks:**

| Approach | Goroutines | Waits? | Scales? | Use Case |
|----------|------------|--------|---------|----------|
| **Naive** | 1000 | ❌ No | ❌ No | Never use |
| **sync.WaitGroup** | 1000 | ✅ Yes | ❌ No | Small tasks only (< 100) |
| **Worker Pool** | 5 | ✅ Yes | ✅ Yes | Production at scale |

---

### The Key Difference: sync.WaitGroup vs Worker Pool

**This is the critical insight:**

> **sync.WaitGroup is a TOOL, not a PATTERN.**
>
> - **Approach 2** uses sync.WaitGroup to wait for 1000 goroutines
> - **Approach 3** uses sync.WaitGroup to wait for 5 goroutines (per batch)
>
> **Both use sync.WaitGroup, but Worker Pool limits how many goroutines exist at once!**

**Analogy:**

**Approach 2 (sync.WaitGroup only):**
- Restaurant gets 1000 orders
- Hires 1000 chefs immediately
- Manager waits for all chefs to finish
- **Problem:** Kitchen is too crowded, runs out of ingredients, chaos!

**Approach 3 (Worker Pool):**
- Restaurant gets 1000 orders
- Hires only 5 chefs
- Chefs process orders in batches
- Manager waits for each batch
- **Result:** Controlled, efficient, scalable!

---

### When to Use Each Approach

**Use Approach 2 (sync.WaitGroup only) when:**
- ✅ Small number of tasks (< 50)
- ✅ Tasks are very lightweight (no DB, no API calls)
- ✅ No resource constraints
- ✅ Example: Processing items in memory, simple calculations

**Use Approach 3 (Worker Pool) when:**
- ✅ Large number of tasks (100+)
- ✅ Tasks use external resources (DB, API, network)
- ✅ Resource constraints exist (connection pools, rate limits)
- ✅ Need predictable resource usage
- ✅ **This is what you need in production!**

---

### Deep Dive: Why sync.WaitGroup?

**Question:** Why use `sync.WaitGroup` instead of channels?

**Alternative 1: Using Channels**

```go
// ❌ More complex, unnecessary overhead
done := make(chan bool, batchSize)

for j := 0; j < batchSize; j++ {
    go func(idx int) {
        impl.triggerCiPipeline(pipelines[idx])
        done <- true  // Send completion signal
    }(i + j)
}

// Wait for all goroutines
for j := 0; j < batchSize; j++ {
    <-done  // Receive completion signal
}
```

**Problems with channels:**
1. Need to create buffered channel (size = batchSize)
2. Need to send signal after work
3. Need to receive exactly batchSize times
4. More allocations (channel creation)
5. More complex to read

**Alternative 2: Using sync.WaitGroup**

```go
// ✅ Simpler, clearer intent
var wg sync.WaitGroup

for j := 0; j < batchSize; j++ {
    wg.Add(1)
    go func(idx int) {
        defer wg.Done()
        impl.triggerCiPipeline(pipelines[idx])
    }(i + j)
}

wg.Wait()  // Clear intent: "wait for group"
```

**Why WaitGroup wins:**
1. ✅ Clear semantic intent: "wait for group of goroutines"
2. ✅ Less code, easier to read
3. ✅ No channel allocation overhead
4. ✅ defer wg.Done() is idiomatic
5. ✅ No need to count receives

**When to use channels instead:**
- When you need to pass data between goroutines
- When you need to implement producer-consumer pattern
- When you need select with multiple channels

**Our case:** We just need to wait for completion → WaitGroup is perfect

---

### Key Takeaways: Worker Pool Pattern

**What we learned:**
1. ✅ **Batching controls concurrency** - Only N goroutines at a time
2. ✅ **sync.WaitGroup for waiting** - Simpler than channels for this use case
3. ✅ **Configurable batch size** - Tune based on resource limits
4. ✅ **Avoid loop variable capture** - Pass index as parameter

**When to use Worker Pool:**
- ✅ Processing large number of independent tasks
- ✅ Resource constraints exist (DB connections, API limits)
- ✅ Need predictable, controlled resource usage
- ✅ Tasks are I/O-bound (network, database, file operations)

**Production Impact at Devtron:**
- **Before:** System crashes with 100+ concurrent triggers
- **After:** Stable processing of 1000+ triggers with batch size = 5
- **Result:** 100% success rate, predictable resource usage

---

## 4. Pattern 2: Fan-Out/Fan-In (Parallel Aggregation) (10 mins)

### What is Fan-Out/Fan-In?

**Fan-Out:** Distribute work to multiple goroutines running in parallel
**Fan-In:** Collect results from all goroutines into a single place

**The Pattern:**
```
Input → Fan-Out → [Worker 1, Worker 2, Worker 3, ...] → Fan-In → Combined Result
```

**Use Cases:**
- Fetching data from multiple sources (databases, APIs, clusters)
- Parallel processing with result aggregation
- Scatter-gather pattern

**Key Difference from Worker Pool:**
- **Worker Pool:** Process many tasks in controlled batches (bounded concurrency)
- **Fan-Out/Fan-In:** Process N tasks in parallel, collect N results (often unbounded, but N is small)

---

### Real-World Analogy: The Library Research

**Imagine you're a student researching for a paper:**

You need information from 2 different sections of the library:
- **Section A:** History books (5 minutes to walk there, find book, and return)
- **Section B:** Science books (5 minutes to walk there, find book, and return)

**Option 1: Sequential (You do it alone)**
- Walk to Section A, get history book → 5 minutes
- Walk to Section B, get science book → 5 minutes
- **Total time:** 10 minutes

**Option 2: Parallel (You and your friend)**
- You walk to Section A for history book → 5 minutes
- Your friend walks to Section B for science book → 5 minutes (at the same time!)
- You both meet back at the table
- **Total time:** 5 minutes (50% faster!)

**This is exactly what happens in our system:**
- **Section A** = CI workflow status query
- **Section B** = CD workflow status query
- **You and your friend** = 2 goroutines running in parallel
- **Meeting back at table** = sync.WaitGroup waiting for both to complete

---
### Real-World Example : Fetching Workflow Status

**Context:** User opens the application dashboard and needs to see CI and CD workflow status.

**The Problem:**
- Need to fetch CI workflow status (calls database, ~500ms)
- Need to fetch CD workflow status (calls database, ~500ms)
- Sequential: 500ms + 500ms = 1000ms (too slow!)
- User sees loading spinner for 1 second

**The Solution:** Fetch both in parallel! (Like the library example)

---

### Code Walkthrough: Parallel Status Fetching

**File:** `api/restHandler/app/pipeline/configure/PipelineConfigRestHandler.go`
**Function:** `FetchWorkflowStatus`

**Real production code from Devtron:**

```go
// This function is called when user opens the application dashboard
// It needs to show both CI and CD workflow status

func (handler *PipelineConfigRestHandlerImpl) FetchWorkflowStatus(
    w http.ResponseWriter,
    r *http.Request,
) {
    // Get appId from request
    vars := mux.Vars(r)
    appId, _ := strconv.Atoi(vars["app-id"])

    // STEP 1: Declare variables to store results
    // Why separate variables: Each goroutine will write to its own variable
    var ciWorkflowStatus []*pipelineConfig.CiWorkflowStatus
    var cdWorkflowStatus []*pipelineConfig.CdWorkflowStatus
    var err error
    var err1 error

    // STEP 2: Create WaitGroup for 2 goroutines
    // Why 2: We're launching exactly 2 goroutines (CI and CD)
    wg := sync.WaitGroup{}
    wg.Add(2)  // Increment counter by 2

    // STEP 3: Launch goroutine to fetch CI status
    go func() {
        // STEP 3a: Ensure Done() is called
        // Why defer: Guarantees execution even if panic
        defer wg.Done()

        // STEP 3b: Fetch CI workflow status
        // This function:
        // - Queries ci_workflow table
        // - Joins with ci_pipeline table
        // - Aggregates status for all CI pipelines
        // Takes ~500ms
        ciWorkflowStatus, err = handler.ciHandler.FetchCiStatusForTriggerView(appId)
    }()

    // STEP 4: Launch goroutine to fetch CD status
    go func() {
        // STEP 4a: Ensure Done() is called
        defer wg.Done()

        // STEP 4b: Fetch CD workflow status
        // This function:
        // - Queries cd_workflow table
        // - Joins with cd_pipeline table
        // - Aggregates status for all CD pipelines
        // Takes ~500ms
        cdWorkflowStatus, err1 = handler.cdHandler.FetchAppWorkflowStatusForTriggerView(appId)
    }()

    // STEP 5: Wait for both goroutines to complete
    // Why Wait(): We need both results before responding to user
    // At this point, both goroutines are running in parallel
    wg.Wait()

    // STEP 6: Check for errors
    if err != nil {
        handler.Logger.Errorw("service err, FetchAppWorkflowStatusForTriggerView",
            "err", err, "appId", appId)
        common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
        return
    }

    // STEP 7: Combine results and return
    triggerWorkflowStatus := pipelineConfig.TriggerWorkflowStatus{
        CiWorkflowStatus: ciWorkflowStatus,
        CdWorkflowStatus: cdWorkflowStatus,
    }

    common.WriteJsonResp(w, nil, triggerWorkflowStatus, http.StatusOK)
}
```

**Results:**
- **Before (Sequential):** 500ms + 500ms = 1000ms
- **After (Parallel):** max(500ms, 500ms) = 500ms
- **Speedup:** 2x faster!
- **User experience:** Dashboard loads in half the time

**Why this works:**
- CI and CD fetches are independent (no shared state)
- Both can run simultaneously
- We wait for both to complete before responding
- Simple, effective, scales well

---
### Key Takeaways: Fan-Out/Fan-In Pattern

**What we learned:**
1. ✅ **Fan-Out:** Launch multiple goroutines in parallel for independent tasks
2. ✅ **Fan-In:** Collect results safely using sync.Map or pre-allocated slice
3. ✅ **sync.Map for dynamic keys** - When you don't know keys in advance
4. ✅ **Pre-allocated slice for indexed results** - When you know indices
5. ✅ **atomic operations for counters** - Fast, lock-free increments

**When to use Fan-Out/Fan-In:**
- ✅ Fetching from multiple independent sources (APIs, databases, clusters)
- ✅ Need to combine/aggregate results
- ✅ Can tolerate partial failures
- ✅ N is relatively small (< 1000) or use with batching

**Production Impact at Devtron:**
- **Workflow status:** 1000ms → 500ms (2x faster)
- **Cluster connection:** 20s → 300ms (66x faster)
- **Resource fetching:** 5s → 1s (5x faster)

---

## 5. Key Takeaways & Production Lessons (4 mins)

### The Evolution Summary

**We covered 3 approaches to concurrent task processing:**

| Approach | Code Pattern | Goroutines | Waits? | Scales? | When to Use |
|----------|--------------|------------|--------|---------|-------------|
| **1. Naive** | `go func()` in loop | N (all tasks) | ❌ No | ❌ No | Never in production |
| **2. sync.WaitGroup** | `wg.Add(1)` + `go func()` + `wg.Wait()` | N (all tasks) | ✅ Yes | ❌ No | Small N (< 50), no resources |
| **3. Worker Pool** | Batching + `wg.Add(1)` + `go func()` + `wg.Wait()` | B (batch size) | ✅ Yes | ✅ Yes | **Production at scale** |

**Key Insight:**
> **sync.WaitGroup is a synchronization tool, not a scaling solution.**
>
> - Approach 2 uses it to wait for N goroutines (unbounded)
> - Approach 3 uses it to wait for B goroutines per batch (bounded)
>
> **Worker Pool = Batching + sync.WaitGroup**

---

### Pattern Summary

**Pattern 1: Worker Pool (Bounded Concurrency)**

**When to use:**
- Processing large number of items (1000+)
- External resource limits (database, API)
- Need predictable resource usage

**Key techniques:**
- Batch processing with fixed size
- sync.WaitGroup for coordination
- Pass loop variables correctly

**Production impact:**
- Crash → Stable
- 0% success → 100% success
- Predictable resource usage

---

**Pattern 2: Fan-Out/Fan-In (Parallel Aggregation)**

**When to use:**
- Fetching from multiple independent sources
- Need to combine results
- Can tolerate partial failures

**Key techniques:**
- sync.Map for thread-safe result collection
- Pre-allocated slice for indexed results
- atomic operations for counters
- Error handling per goroutine

**Production impact:**
- 20s → 300ms (66x faster)
- Better user experience
- Graceful degradation

---

### Best Practices from Production

**1. Always Use defer wg.Done()**
```go
// ✅ ALWAYS
go func() {
    defer wg.Done()  // Guarantees execution
    doWork()
}()
```

**Why:** If doWork() panics, Done() is still called. Without defer, deadlock.

---

**2. Pass Loop Variables Correctly**
```go
// ✅ CORRECT
for i, item := range items {
    go func(idx int, it Item) {
        process(idx, it)
    }(i, item)
}
```

**Why:** Avoid loop variable capture bug.

---

**3. Choose the Right Approach - Decision Tree**

```
Do you need to process multiple tasks concurrently?
│
├─ No → Just use sequential processing
│
└─ Yes → How many tasks?
    │
    ├─ Small (< 50) AND no external resources (DB, API)
    │   → Use sync.WaitGroup (Approach 2)
    │   → Simple, fast, no need for batching
    │
    └─ Large (100+) OR uses external resources
        → Use Worker Pool (Approach 3)
        → Bounded concurrency, predictable resources

        How to choose batch size?
        │
        ├─ External API rate limit exists?
        │   → batch_size = rate_limit / calls_per_task
        │
        ├─ Database connection pool limit?
        │   → batch_size = available_connections / 2
        │
        ├─ Memory constrained?
        │   → batch_size = available_memory / memory_per_goroutine
        │
        └─ No constraints?
            → Start with batch_size = num_cpu_cores
            → Monitor and tune based on metrics
```

**Examples:**

| Scenario | Tasks | Resources | Approach | Batch Size |
|----------|-------|-----------|----------|------------|
| Process 20 files in memory | 20 | None | sync.WaitGroup | N/A |
| Send 100 emails via API (10 QPS limit) | 100 | API | Worker Pool | 5 |
| Trigger 1000 CI pipelines (DB + K8s API) | 1000 | DB + API | Worker Pool | 5-10 |
| Fetch from 100 K8s clusters | 100 | Network | Fan-Out/Fan-In | 100 (all) |

---

### Final Thoughts

**Goroutines are cheap, but not free:**
- Each goroutine: ~2KB stack + heap allocations
- Context switching overhead
- Resource contention (database, API, memory)

**Scale requires discipline:**
- Bounded concurrency (worker pools)
- Safe state management (sync.Map, atomic)
- Proper error handling
- Monitoring and tuning

**Production lessons:**
- Start simple, measure, optimize
- Stability > Speed
- Monitor everything
- Plan for failures

---

## Questions?

**Resources:**
- Devtron GitHub: https://github.com/devtron-labs/devtron
- Go Concurrency Patterns: https://go.dev/blog/pipelines
- sync package: https://pkg.go.dev/sync
- atomic package: https://pkg.go.dev/sync/atomic

**Thank you!** 🚀



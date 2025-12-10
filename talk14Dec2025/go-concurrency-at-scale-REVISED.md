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

### Real Production Impact at Devtron

**Approach 2 (sync.WaitGroup only):**
```go
var wg sync.WaitGroup
for _, pipeline := range pipelines {  // 100 pipelines
    wg.Add(1)
    go func(p Pipeline) {
        defer wg.Done()
        impl.triggerCiPipeline(p)
    }(pipeline)
}
wg.Wait()
```

**Result:**
- ❌ Application crash (OOM)
- ❌ Database: "too many clients"
- ❌ Kubernetes API: 429 errors
- ❌ Success rate: 0%

**Approach 3 (Worker Pool):**
```go
batchSize := 5
for i := 0; i < len(pipelines); {
    var wg sync.WaitGroup
    for j := 0; j < batchSize; j++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            impl.triggerCiPipeline(pipelines[idx])
        }(i + j)
    }
    wg.Wait()
    i += batchSize
}
```

**Result:**
- ✅ No crashes
- ✅ Stable resource usage
- ✅ All pipelines complete
- ✅ Success rate: 100%

---

## 3. Pattern 1: Worker Pool - Bounded Concurrency (10 mins)

### Recap: Why Worker Pool?

We just saw 3 approaches:
1. **Naive** - No waiting, unbounded → ❌ Crashes
2. **sync.WaitGroup** - Waits, but still unbounded → ❌ Crashes at scale
3. **Worker Pool** - Waits AND bounded → ✅ Scales

Now let's dive deep into the Worker Pool implementation with production code.

---

### The Real Production Problem

**Scenario:** Auto-trigger CI pipelines for 100+ applications when a base image is updated

**Constraints we discovered:**
- Database connection pool: 100 connections (shared with other services)
- Kubernetes API rate limit: 50 QPS
- Available memory: Limited (running in container)
- CPU cores: 8
- Other services also using these resources!

**The Question:** What batch size should we use?

**The Math:**
- Total pipelines: 100
- Available DB connections: ~50 (other services use 50)
- Safe batch size: 5-10 (leave margin for spikes)
- Our choice: **5** (conservative, prioritizes stability)

**Why batch size = 5?**
- Database: Use max 5 out of 50 available (90% headroom)
- Kubernetes API: 5 concurrent requests << 50 QPS limit
- Memory: 5 goroutines × ~100KB = 500KB (negligible)
- CPU: 5 goroutines on 8 cores = good utilization, no thrashing

---

### Real-World Analogy: The Restaurant Kitchen

**Imagine you're managing a restaurant kitchen:**

You just finished preparing a large catering order (the main dish). Now you need to prepare 50 side dishes that go with it.

**Option 1: Hire 50 chefs at once**
- You call 50 chefs to start cooking simultaneously
- Problem: Your kitchen has only 10 stoves
- Result: 40 chefs standing around waiting, chaos, bumping into each other
- Your kitchen is overwhelmed!

**Option 2: Batch cooking with 5 chefs**
- You call 5 chefs to cook the first 5 side dishes
- When they finish, you call the next 5 chefs for the next 5 dishes
- Continue until all 50 dishes are done
- Result: Organized, efficient, no chaos!

**This is exactly what happens in our system:**
- **50 side dishes** = 50 child CD pipelines to trigger
- **Kitchen stoves** = Database connections (limited to 100)
- **Kitchen space** = Kubernetes API rate limit (50 requests/second)
- **Batch of 5 chefs** = Batch size of 5 concurrent triggers

---

### The Technical Problem: Auto-Triggering Child Pipelines

**Context:** When a CI build completes successfully, we need to automatically trigger all dependent CD pipelines.

**The Problem:**
- A single CI pipeline can have 10-50 child CD pipelines
- Each child pipeline trigger:
  - Creates database records (CdWorkflow, CdWorkflowRunner)
  - Calls Kubernetes API to create Argo Workflow
  - Sends notifications (Slack, email)
  - Updates pipeline status

**What happens if we trigger all at once?**
- 50 child pipelines × 3 DB operations each = 150 concurrent DB connections
- Our DB pool has only 100 connections → **Pool exhausted!**
- 50 Kubernetes API calls in < 1 second
- K8s API limit is 50 QPS → **Rate limit exceeded!**
- Result: System crashes, pipelines fail

**The Solution: Batch Processing (like our kitchen!)**
- Process 5 pipelines at a time (configurable batch size)
- Wait for batch to complete before starting next batch
- Result: Controlled resource usage, no crashes

---

### Code Walkthrough: Line-by-Line Explanation

**File:** `pkg/workflow/dag/WorkflowDagExecutor.go`
**Function:** `HandleCiSuccessEvent` (simplified for clarity)

**Real production code from Devtron:**

```go
// This function is called when a CI build completes successfully
// It auto-triggers all child CD pipelines that depend on this CI artifact

func (impl *WorkflowDagExecutorImpl) HandleCiSuccessEvent(
    triggerContext triggerBean.TriggerContext,
    ciArtifactArr []*repository.CiArtifact,  // CI artifacts to process
    async bool,
    userId int32,
) (int, error) {

    // STEP 1: Get batch size from configuration
    // Why configurable: Different environments have different resource limits
    // Production: 1-5, Staging: 10, Local: 20
    // Env var: CI_SUCCESS_AUTO_TRIGGER_BATCH_SIZE
    batchSize := impl.config.CIAutoTriggerBatchSize
    if batchSize <= 0 {
        batchSize = 1  // Safety: Minimum 1
    }

    totalCIArtifactCount := len(ciArtifactArr)

    // STEP 2: Log start time for monitoring
    // Why: Track performance, alert if too slow
    start := time.Now()
    impl.logger.Infow("Started: auto trigger for children Stage/CD pipelines",
        "Artifact count", totalCIArtifactCount)

    // STEP 3: Process artifacts in batches
    // Why batching: Control concurrency, prevent resource exhaustion
    for i := 0; i < totalCIArtifactCount; {

        // STEP 3a: Calculate remaining items in this iteration
        // Why: Last batch might be smaller than batchSize
        // Example: 23 artifacts, batch size 5 → batches of 5,5,5,5,3
        remainingBatch := totalCIArtifactCount - i
        if remainingBatch < batchSize {
            batchSize = remainingBatch
        }

        // STEP 3b: Create WaitGroup for this batch
        // Why sync.WaitGroup: Need to wait for all goroutines in this batch
        // Alternative: Channels - more complex, unnecessary for just waiting
        var wg sync.WaitGroup

        // STEP 3c: Launch workers for current batch
        for j := 0; j < batchSize; j++ {

            // STEP 3c-i: Increment WaitGroup counter BEFORE launching goroutine
            // Why before: Prevents race condition
            // If we Add() inside goroutine, Wait() might be called before Add()
            // Result: Wait() returns immediately, doesn't wait for goroutines
            wg.Add(1)

            // STEP 3c-ii: Calculate index for this worker
            // Why separate variable: Avoid loop variable capture bug
            // If we use i+j directly in goroutine, value changes in next iteration
            // All goroutines would see the final value of i+j
            index := i + j

            // STEP 3c-iii: Define the work function
            // Why separate function: Allows us to pass index as parameter
            runnableFunc := func(index int) {

                // STEP 3c-iii-a: Ensure Done() is called
                // Why defer: Guarantees Done() even if panic occurs
                // Without defer: If handleCiSuccessEvent panics, Done() never called
                // Result: wg.Wait() blocks forever (deadlock)
                defer wg.Done()

                // STEP 3c-iii-b: Get the artifact for this index
                ciArtifact := ciArtifactArr[index]

                // STEP 3c-iii-c: Do the actual work
                // This function:
                // - Finds all child CD pipelines for this CI artifact
                // - For each child pipeline:
                //   * Creates CdWorkflow record in database
                //   * Creates CdWorkflowRunner record
                //   * Calls Kubernetes API to create Argo Workflow
                //   * Sends notification (Slack, email)
                //   * Updates pipeline status
                err := impl.handleCiSuccessEvent(triggerContext, ciArtifact, async, userId)
                if err != nil {
                    impl.logger.Errorw("error on handle ci success event",
                        "ciArtifactId", ciArtifact.Id, "err", err)
                }
            }

            // STEP 3c-iv: Execute the function in a goroutine
            // Why asyncRunnable.Execute: Provides goroutine pool management
            // Alternative: go runnableFunc(index) - works, but no pool management
            impl.asyncRunnable.Execute(func() { runnableFunc(index) })
        }

        // STEP 3d: Wait for current batch to complete
        // Why wait here: Ensures we don't start next batch until current finishes
        // This is what gives us "bounded concurrency"
        // At this point, exactly batchSize goroutines are running
        wg.Wait()

        // STEP 3e: Move to next batch
        i += batchSize
    }

    // STEP 4: Log completion time
    // Why: Monitor performance, track how long batch processing takes
    impl.logger.Debugw("Completed: auto trigger for children Stage/CD pipelines",
        "Time taken", time.Since(start).Seconds())

    return buildArtifact.Id, nil
}
```

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

### Real-World Analogy: The Multi-City Weather Check

**Imagine you're a weather reporter:**

You need to check the current temperature in 100 different cities for your report.

**Option 1: Call each city one by one**
- Call City 1, wait for answer (2 seconds)
- Call City 2, wait for answer (2 seconds)
- ... repeat 100 times
- **Total time:** 100 × 2 seconds = 200 seconds (3+ minutes!)

**Option 2: Call all cities at once**
- You have 100 assistants
- Each assistant calls one city simultaneously
- All calls happen at the same time
- **Total time:** 2 seconds (100x faster!)

**But there's a problem:**
- Where do you collect all the answers?
- 100 assistants trying to write on the same notepad → chaos!
- They'll overwrite each other's notes

**The Solution: Numbered slots**
- Give each assistant a numbered slot on a board
- Assistant 1 writes in slot 1, Assistant 2 in slot 2, etc.
- No conflicts, everyone writes to their own slot!

**This is exactly what happens in our system:**
- **100 cities** = 100 Kubernetes clusters to check
- **100 assistants** = 100 goroutines
- **Numbered slots** = Pre-allocated slice or sync.Map
- **No conflicts** = Thread-safe result collection

---

### Real-World Example 1: Fetching Workflow Status

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

### Real-World Analogy: The Hotel Room Inspection

**Imagine you're a hotel manager:**

You have 100 hotel rooms to inspect before guests arrive. Each inspection takes 2 minutes.

**Option 1: Inspect rooms one by one**
- Room 101 → 2 minutes
- Room 102 → 2 minutes
- ... repeat 100 times
- **Total time:** 100 × 2 minutes = 200 minutes (3+ hours!)

**Option 2: Hire 100 inspectors**
- Each inspector checks one room simultaneously
- All inspections happen at the same time
- **Total time:** 2 minutes (100x faster!)

**But there's a problem:**
- How do you collect the inspection results?
- 100 inspectors trying to write in the same logbook → chaos!
- Some results might get lost or overwritten

**The Solution: A thread-safe logbook (sync.Map)**
- Each inspector has a unique room number
- They write their result with the room number as the key
- The logbook is designed to handle multiple people writing at once
- No results get lost!

**This is exactly what happens in our system:**
- **100 hotel rooms** = 100 Kubernetes clusters
- **2 minutes per inspection** = 200ms connection test
- **100 inspectors** = 100 goroutines
- **Thread-safe logbook** = sync.Map
- **Room number as key** = Cluster ID as key

---

### Real-World Example 2: Connecting to Multiple Clusters

**Context:** Devtron manages 100+ Kubernetes clusters. We need to check connection status for all clusters.

**The Problem:**
- Each cluster connection test takes ~200ms (network call to K8s API)
- Sequential: 100 clusters × 200ms = 20 seconds
- This runs as a cron job every 5 minutes
- 20 seconds is too slow!

**The Solution:** Connect to all clusters in parallel with safe result collection! (Like the hotel inspection)

---

### Code Walkthrough: Parallel Cluster Connection

**File:** `pkg/cluster/ClusterService.go`
**Function:** `ConnectClustersInBatch`

**Real production code from Devtron:**

```go
// This function is called by a cron job every 5 minutes
// It checks connection status for all clusters

func (impl *ClusterServiceImpl) ConnectClustersInBatch(
    clusters []*bean.ClusterBean,  // All clusters to check
    clusterExistInDb bool,
) {

    // STEP 1: Create WaitGroup for all clusters
    // Why: We need to wait for ALL cluster connections to complete
    var wg sync.WaitGroup

    // STEP 2: Create thread-safe map for results
    // Why sync.Map: Multiple goroutines will write results concurrently
    // Alternative: regular map + mutex (more code, manual locking)
    respMap := &sync.Map{}

    // STEP 3: Launch goroutine for each cluster
    // Why goroutine per cluster: Clusters are independent, can connect in parallel
    for idx := range clusters {
        cluster := clusters[idx]

        // STEP 3a: Skip virtual clusters
        // Why: Virtual clusters don't have real K8s API endpoints
        if cluster.IsVirtualCluster {
            impl.updateConnectionStatusForVirtualCluster(respMap, cluster.Id, cluster.ClusterName)
            continue
        }

        // STEP 3b: Increment WaitGroup
        // Why before goroutine: Prevent race condition
        wg.Add(1)

        // STEP 3c: Define the work function
        // Why separate function: Allows us to pass idx and cluster as parameters
        runnableFunc := func(idx int, cluster *bean.ClusterBean) {

            // STEP 3c-i: Ensure Done() is called
            // Why defer: Guarantees execution even if panic
            defer wg.Done()

            // STEP 3c-ii: Get cluster configuration
            clusterConfig := cluster.GetClusterConfig()

            // STEP 3c-iii: Try to connect to cluster
            // This makes network call to Kubernetes API
            // Takes ~200ms per cluster
            _, _, k8sClientSet, err := impl.K8sUtil.GetK8sConfigAndClients(clusterConfig)

            if err != nil {
                // STEP 3c-iv: Store error in sync.Map
                // Why sync.Map.Store: Thread-safe write operation
                // Multiple goroutines writing different keys simultaneously
                respMap.Store(cluster.Id, err)
                return
            }

            // STEP 3c-v: Get cluster ID
            id := cluster.Id
            if !clusterExistInDb {
                id = idx
            }

            // STEP 3c-vi: Check cluster health and store result
            // This calls /livez endpoint on K8s API
            impl.GetAndUpdateConnectionStatusForOneCluster(k8sClientSet, id, respMap)
        }

        // STEP 3d: Execute the function in a goroutine
        // Why asyncRunnable.Execute: Provides goroutine pool management
        impl.asyncRunnable.Execute(func() { runnableFunc(idx, cluster) })
    }

    // STEP 4: Wait for all goroutines to complete
    // Why Wait(): We need all results before proceeding
    // At this point, all 100 goroutines are running in parallel
    wg.Wait()

    // STEP 5: Handle errors and update database
    // This iterates over sync.Map and updates cluster connection status
    impl.HandleErrorInClusterConnections(clusters, respMap, clusterExistInDb)
}
```

**Results:**
- **Before (Sequential):** 100 clusters × 200ms = 20 seconds
- **After (Parallel):** max(200ms) ≈ 300ms (including overhead)
- **Speedup:** 66x faster!
- **Cron job:** Completes in < 1 second instead of 20 seconds

**Why sync.Map?**
- 100 goroutines writing results simultaneously
- Each writes to a different key (cluster.Id)
- sync.Map is optimized for this pattern
- No manual mutex management needed

---

### Real-World Analogy: The Package Delivery Service

**Imagine you're a delivery company manager:**

You have 50 packages to deliver across the city. Each delivery takes 10 minutes.

**Option 1: One driver delivers all packages**
- Package 1 → 10 minutes
- Package 2 → 10 minutes
- ... repeat 50 times
- **Total time:** 50 × 10 minutes = 500 minutes (8+ hours!)

**Option 2: Hire 50 drivers at once**
- All 50 drivers start delivering simultaneously
- Problem: You only have 5 delivery trucks!
- Result: 45 drivers standing around waiting for trucks

**Option 3: Batch delivery with 5 drivers**
- First batch: 5 drivers deliver packages 1-5 → 10 minutes
- Second batch: Same 5 drivers deliver packages 6-10 → 10 minutes
- Continue until all 50 packages delivered
- **Total time:** 10 batches × 10 minutes = 100 minutes
- **Result:** 5x faster than one driver, no wasted resources!

**But there's a problem:**
- How do you track which packages were delivered successfully?
- 5 drivers trying to update the same delivery log → confusion!

**The Solution: Pre-assigned delivery slots**
- You have a board with 50 numbered slots (1 to 50)
- Driver delivering package #7 writes result in slot #7
- Driver delivering package #23 writes result in slot #23
- No conflicts because each driver writes to a different slot!

**This is exactly what happens in our system:**
- **50 packages** = 50 Kubernetes resources to fetch
- **5 delivery trucks** = Batch size of 5 (K8s API limit)
- **5 drivers per batch** = 5 goroutines at a time
- **50 numbered slots** = Pre-allocated slice with 50 positions
- **No conflicts** = Each goroutine writes to its own index

---

### Real-World Example 3: Batch Resource Fetching

**Context:** User wants to see resources (Pods, Deployments, Services) from multiple namespaces.

**The Problem:**
- Need to fetch 50 different resources
- Each fetch takes ~100ms (K8s API call)
- Sequential: 50 × 100ms = 5 seconds
- User sees loading spinner for 5 seconds

**The Solution:** Fetch resources in batches with bounded concurrency! (Like the package delivery)

---

### Code Walkthrough: Batch Resource Fetching

**File:** `pkg/k8s/K8sCommonService.go`
**Function:** `getManifestsByBatch`

**Real production code from Devtron:**

```go
// This function fetches multiple Kubernetes resources in batches
// Used when user views resources in the dashboard

func (impl *K8sCommonServiceImpl) getManifestsByBatch(
    ctx context.Context,
    requests []bean5.ResourceRequestBean,  // Resources to fetch
) []bean5.BatchResourceResponse {

    // STEP 1: Get batch size from configuration
    // Why configurable: Different environments have different K8s API limits
    // Production: 5, Staging: 10
    batchSize := impl.K8sApplicationServiceConfig.BatchSize

    requestsLength := len(requests)

    // STEP 2: Pre-allocate result slice
    // Why pre-allocate: Avoid race conditions when writing results
    // Each goroutine writes to its own index, no conflicts
    res := make([]bean5.BatchResourceResponse, requestsLength)

    // STEP 3: Process requests in batches
    for i := 0; i < requestsLength; {

        // STEP 3a: Calculate remaining requests
        // Why: Last batch might be smaller than batchSize
        remainingBatch := requestsLength - i
        if remainingBatch < batchSize {
            batchSize = remainingBatch
        }

        // STEP 3b: Create WaitGroup for this batch
        var wg sync.WaitGroup

        // STEP 3c: Launch workers for current batch
        for j := 0; j < batchSize; j++ {

            // STEP 3c-i: Increment WaitGroup
            wg.Add(1)

            // STEP 3c-ii: Define the work function
            runnableFunc := func(index int) {

                // STEP 3c-ii-a: Create response object
                resp := bean5.BatchResourceResponse{}

                // STEP 3c-ii-b: Fetch resource from Kubernetes
                // This makes API call to K8s cluster
                // Takes ~100ms per resource
                response, err := impl.GetResource(ctx, &requests[index])

                if response != nil {
                    resp.ManifestResponse = response.ManifestResponse
                }
                resp.Err = err

                // STEP 3c-ii-c: Store result at specific index
                // Why safe: Each goroutine writes to different index
                // No race condition, no need for mutex or sync.Map
                res[index] = resp

                // STEP 3c-ii-d: Decrement WaitGroup
                // Why not defer: Simple function, no panic risk
                wg.Done()
            }

            // STEP 3c-iii: Calculate index
            index := i + j

            // STEP 3c-iv: Execute in goroutine
            impl.asyncRunnable.Execute(func() { runnableFunc(index) })
        }

        // STEP 3d: Wait for this batch to complete
        wg.Wait()

        // STEP 3e: Move to next batch
        i += batchSize
    }

    // STEP 4: Return all results
    return res
}
```

**Results:**
- **Before (Sequential):** 50 resources × 100ms = 5 seconds
- **After (Batch size 5):** 10 batches × 100ms = 1 second
- **Speedup:** 5x faster!
- **User experience:** Resources load in 1 second instead of 5

**Key Insight: Pre-allocated Slice vs sync.Map**

**Why pre-allocated slice works here:**
- Each goroutine writes to a different index
- No two goroutines write to the same index
- No race condition, no need for sync.Map

**When to use sync.Map:**
- When you don't know the keys in advance
- When multiple goroutines might write to the same key
- When you need dynamic key-value storage

**Our case:** We know the indices (0 to requestsLength-1), so pre-allocated slice is simpler and faster!

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

**4. Choose the Right Synchronization Primitive**

| Use Case | Tool | Why |
|----------|------|-----|
| Wait for goroutines | sync.WaitGroup | Simple, clear intent |
| Thread-safe map | sync.Map | Built-in safety, optimized |
| Simple counters | atomic.AddUint64 | Fast, no locks |
| Complex state | sync.Mutex | Full control |
| Pass data | channels | Communication |

---

**5. Monitor and Tune**

```go
// Expose metrics
metrics.RecordConcurrentGoroutines(count)
metrics.RecordProcessingTime(duration)
metrics.RecordErrorRate(errors / total)
```

**Why:**
- Understand actual behavior in production
- Tune batch sizes based on real data
- Alert on anomalies

---

**6. Start Conservative, Then Optimize**

- Start with small batch size (5-10)
- Monitor resource usage
- Gradually increase if safe
- Measure impact of changes

**Why:** Stability > Speed. Crashes are worse than slow.

---

### Common Pitfalls to Avoid

**1. Forgetting to Wait**
```go
// ❌ BAD: Function returns before goroutines finish
for _, item := range items {
    go process(item)
}
return  // Goroutines still running!
```

**2. Not Handling Panics**
```go
// ✅ GOOD: Recover from panics
go func() {
    defer func() {
        if r := recover(); r != nil {
            log.Error("Panic:", r)
        }
    }()
    defer wg.Done()
    doWork()
}()
```

**3. Unbounded Goroutine Creation**
```go
// ❌ BAD: Can create millions of goroutines
for _, item := range millionItems {
    go process(item)
}
```

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

**Requirements:**
- Fetch Pod, Deployment, Service status from each cluster
- Combine results from all clusters
- Display on dashboard in < 2 seconds
- Handle failures gracefully (some clusters might be down)

**Challenges:**
- Each cluster fetch takes ~200ms
- Sequential: 100 clusters × 200ms = 20 seconds ❌
- Need parallel fetching
- Need to safely collect results from multiple goroutines
- Need to handle partial failures

---

### The Solution: Fan-Out/Fan-In with Safe Result Collection

**Core Idea:**
1. **Fan-Out:** Launch goroutines for each cluster (parallel fetching)
2. **Fan-In:** Collect results safely from all goroutines
3. **Safe State:** Use sync.Map for thread-safe result storage

**Architecture:**

```
Main Goroutine
      |
      |--- FAN-OUT (Launch 100 goroutines) --->
      |
      ├─→ Goroutine 1: Fetch Cluster 1 → Store in sync.Map
      ├─→ Goroutine 2: Fetch Cluster 2 → Store in sync.Map
      ├─→ Goroutine 3: Fetch Cluster 3 → Store in sync.Map
      |   ... (100 goroutines total)
      |
      |--- FAN-IN (Wait for all, read sync.Map) --->
      |
    Combine results → Return
```

---

### Code Walkthrough: Line-by-Line Explanation

**File:** `pkg/cluster/ClusterService.go`

```go
package cluster

import (
    "sync"           // Why: sync.WaitGroup for coordination, sync.Map for safe storage
    "sync/atomic"    // Why: Thread-safe counter operations
)

// FetchClusterResourcesInParallel fetches resources from all clusters concurrently
func (impl *ClusterServiceImpl) FetchClusterResourcesInParallel(
    clusters []*bean.ClusterBean,
) map[int]*ClusterResources {

    // STEP 1: Create thread-safe map for results
    // Why sync.Map: Multiple goroutines will write results concurrently
    // Alternative: regular map + mutex (more code, manual locking)
    var resultsMap sync.Map

    // STEP 2: Create atomic counter for tracking
    // Why atomic: Multiple goroutines will increment concurrently
    // Alternative: mutex + regular int (slower, more complex)
    var successCount uint64
    var failureCount uint64

    // STEP 3: Create WaitGroup for all goroutines
    // Why: We need to wait for ALL clusters to be processed
    var wg sync.WaitGroup

    // STEP 4: Fan-Out - Launch goroutine for each cluster
    for _, cluster := range clusters {

        // STEP 4a: Increment WaitGroup
        // Why before goroutine: Prevent race condition
        wg.Add(1)

        // STEP 4b: Launch goroutine
        // Why goroutine per cluster: Clusters are independent, can fetch in parallel
        go func(c *bean.ClusterBean) {
            // STEP 4c: Ensure Done() is called
            // Why defer: Guarantees execution even on panic
            defer wg.Done()

            // STEP 4d: Fetch resources from this cluster
            // This makes network call to Kubernetes API
            // Takes ~200ms per cluster
            resources, err := impl.fetchResourcesFromCluster(c)

            if err != nil {
                // STEP 4e: Handle error - store error in results
                // Why store error: We want to show which clusters failed
                resultsMap.Store(c.Id, &ClusterResources{
                    ClusterId: c.Id,
                    Error:     err,
                })

                // STEP 4f: Increment failure counter atomically
                // Why atomic.AddUint64: Thread-safe increment
                // Multiple goroutines might fail simultaneously
                // Without atomic: Race condition, incorrect count
                atomic.AddUint64(&failureCount, 1)
                return
            }

            // STEP 4g: Store successful result
            // Why sync.Map.Store: Thread-safe write operation
            // Multiple goroutines writing different keys simultaneously
            resultsMap.Store(c.Id, resources)

            // STEP 4h: Increment success counter atomically
            atomic.AddUint64(&successCount, 1)

        }(cluster)  // Why pass cluster: Avoid loop variable capture
    }

    // STEP 5: Fan-In - Wait for all goroutines to complete
    // Why Wait(): We need all results before proceeding
    // At this point, all 100 goroutines are running in parallel
    wg.Wait()

    // STEP 6: Convert sync.Map to regular map
    // Why: sync.Map is for concurrent writes, now we're done writing
    // Regular map is easier to work with for single-threaded read
    finalResults := make(map[int]*ClusterResources)

    // STEP 7: Iterate over sync.Map
    // Why Range: Only way to iterate over sync.Map
    resultsMap.Range(func(key, value interface{}) bool {
        clusterId := key.(int)
        resources := value.(*ClusterResources)
        finalResults[clusterId] = resources
        return true  // Continue iteration
    })

    // STEP 8: Log metrics
    // Why: Monitor success/failure rates for alerting
    impl.logger.Infow("Cluster resources fetched",
        "total", len(clusters),
        "success", atomic.LoadUint64(&successCount),
        "failure", atomic.LoadUint64(&failureCount),
    )

    return finalResults
}
```

---

### Deep Dive: Why sync.Map Instead of Regular Map?

**The Problem with Regular Map:**

```go
// ❌ DANGER: Race condition!
results := make(map[int]*ClusterResources)

for _, cluster := range clusters {
    go func(c *bean.ClusterBean) {
        resources := fetchResources(c)

        // RACE CONDITION: Multiple goroutines writing to map
        results[c.Id] = resources  // 💥 CRASH: concurrent map writes

    }(cluster)
}
```

**What happens:**
- Go runtime detects concurrent map writes
- **Panic:** `fatal error: concurrent map writes`
- Application crashes

**Why does this happen?**
- Maps in Go are not thread-safe
- Internal structure can be corrupted by concurrent writes
- Go runtime actively detects this and panics (better than silent corruption)

---

### Solution 1: Regular Map + Mutex

```go
// ✅ Works, but more code
results := make(map[int]*ClusterResources)
var mutex sync.Mutex  // Need to manage mutex manually

for _, cluster := range clusters {
    go func(c *bean.ClusterBean) {
        resources := fetchResources(c)

        // Lock before write
        mutex.Lock()
        results[c.Id] = resources
        mutex.Unlock()  // Don't forget to unlock!

    }(cluster)
}
```

**Pros:**
- ✅ Works correctly
- ✅ Familiar pattern

**Cons:**
- ❌ More code (Lock/Unlock)
- ❌ Easy to forget Unlock (deadlock)
- ❌ Can't use defer (performance overhead)
- ❌ Contention on single mutex (all goroutines wait)

---

### Solution 2: sync.Map (Our Choice)

```go
// ✅ Simpler, built-in thread safety
var results sync.Map

for _, cluster := range clusters {
    go func(c *bean.ClusterBean) {
        resources := fetchResources(c)

        // Thread-safe by default
        results.Store(c.Id, resources)

    }(cluster)
}
```

**Pros:**
- ✅ Thread-safe by default
- ✅ No manual locking
- ✅ Optimized for concurrent writes
- ✅ Less code, clearer intent

**Cons:**
- ❌ Type-unsafe (interface{} keys and values)
- ❌ Need type assertions when reading
- ❌ Can't use range loop (need Range method)

**When to use sync.Map:**
- ✅ Multiple goroutines writing different keys
- ✅ Write-heavy or read-heavy workloads (optimized for both)
- ✅ Don't want to manage mutexes manually

**When to use regular map + mutex:**
- ✅ Need type safety
- ✅ Small number of writes
- ✅ Complex operations (read-modify-write)

**Our case:** 100 goroutines writing different keys → sync.Map is perfect

---

### Deep Dive: Why atomic.AddUint64 for Counters?

**The Problem with Regular Increment:**

```go
// ❌ DANGER: Race condition!
var successCount uint64

for _, cluster := range clusters {
    go func(c *bean.ClusterBean) {
        resources := fetchResources(c)

        // RACE CONDITION: Read-modify-write is not atomic
        successCount++  // Equivalent to: successCount = successCount + 1

    }(cluster)
}
```

**What happens:**
- Goroutine 1 reads: successCount = 5
- Goroutine 2 reads: successCount = 5 (before G1 writes)
- Goroutine 1 writes: successCount = 6
- Goroutine 2 writes: successCount = 6 (overwrites G1's write!)
- **Result:** Lost update, incorrect count

**Why does this happen?**
- `successCount++` is actually 3 operations:
  1. Read current value
  2. Add 1
  3. Write new value
- These 3 operations are not atomic
- Another goroutine can interleave between them

---

### Solution 1: Mutex

```go
// ✅ Works, but slower
var successCount uint64
var mutex sync.Mutex

for _, cluster := range clusters {
    go func(c *bean.ClusterBean) {
        resources := fetchResources(c)

        mutex.Lock()
        successCount++
        mutex.Unlock()

    }(cluster)
}
```

**Pros:**
- ✅ Correct

**Cons:**
- ❌ Slower (mutex overhead)
- ❌ Contention (all goroutines wait for lock)
- ❌ More code

---

### Solution 2: atomic.AddUint64 (Our Choice)

```go
// ✅ Fast and correct
var successCount uint64

for _, cluster := range clusters {
    go func(c *bean.ClusterBean) {
        resources := fetchResources(c)

        // Atomic increment - single CPU instruction
        atomic.AddUint64(&successCount, 1)

    }(cluster)
}
```

**Pros:**
- ✅ Thread-safe
- ✅ Fast (single CPU instruction, no locks)
- ✅ No contention
- ✅ Less code

**Cons:**
- ❌ Only works for simple operations (add, load, store, swap)
- ❌ Can't do complex operations

**How atomic operations work:**
- Uses CPU-level atomic instructions (e.g., LOCK XADD on x86)
- Hardware guarantees atomicity
- Much faster than mutex (no context switch)

**When to use atomic:**
- ✅ Simple counters
- ✅ Flags (atomic.LoadUint32, atomic.StoreUint32)
- ✅ High contention scenarios

**When to use mutex:**
- ✅ Complex operations (read-modify-write with logic)
- ✅ Protecting multiple variables together
- ✅ Need to hold lock across function calls

**Our case:** Simple counter increment → atomic is perfect

---

### Production Results

**Before (Sequential Fetching):**
- 100 clusters fetched one by one
- **Total time:** 100 × 200ms = 20 seconds
- **User experience:** Dashboard takes 20 seconds to load ❌
- **Throughput:** 5 clusters/second

**After (Parallel Fetching with sync.Map):**
- 100 clusters fetched in parallel
- **Total time:** max(200ms) ≈ 300ms (including overhead)
- **User experience:** Dashboard loads in < 1 second ✅
- **Throughput:** 333 clusters/second
- **Speedup:** 66x faster!

**Key Metrics:**
- **Latency:** 20s → 300ms (98.5% reduction)
- **Concurrency:** 100 goroutines running simultaneously
- **Memory usage:** ~20MB (100 goroutines × ~200KB each)
- **CPU usage:** 60% during fetch (good utilization)
- **Success rate:** 98% (some clusters might be down)

**Handling Failures:**
- 2 out of 100 clusters down
- **Result:** Show 98 clusters, mark 2 as unavailable
- **User experience:** Partial data is better than no data
- **Reliability:** System doesn't fail if one cluster is down

---

### Real-World Considerations

**1. Timeout Handling**

```go
// Add timeout to prevent hanging on slow clusters
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resources, err := impl.fetchResourcesFromCluster(ctx, c)
```

**Why:**
- Some clusters might be slow or unresponsive
- Without timeout: One slow cluster delays entire operation
- With timeout: Fail fast, show partial results

**2. Error Aggregation**

```go
// Collect all errors for debugging
type ClusterResources struct {
    ClusterId int
    Pods      []Pod
    Error     error  // Store error for this cluster
}
```

**Why:**
- Need to know which clusters failed
- Show partial results to user
- Log errors for debugging

**3. Rate Limiting (If Needed)**

```go
// If we have 1000+ clusters, use semaphore
semaphore := make(chan struct{}, 50)  // Max 50 concurrent

for _, cluster := range clusters {
    semaphore <- struct{}{}  // Acquire
    go func(c *bean.ClusterBean) {
        defer func() { <-semaphore }()  // Release
        defer wg.Done()
        // ... fetch resources ...
    }(cluster)
}
```

**Why:**
- Too many concurrent connections might overwhelm network
- Semaphore limits concurrency
- Trade-off: Speed vs resource usage

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

**Pattern 1: Bounded Concurrency (Worker Pools)**

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

**Pattern 2: Parallel Aggregation**

**When to use:**
- Fetching from multiple independent sources
- Need to combine results
- Can tolerate partial failures

**Key techniques:**
- sync.Map for thread-safe result collection
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
| Fetch from 100 K8s clusters | 100 | Network | Parallel Aggregation | 100 (all) |

---

**4. Choose the Right Synchronization Primitive**

| Use Case | Tool | Why |
|----------|------|-----|
| Wait for goroutines | sync.WaitGroup | Simple, clear intent |
| Thread-safe map | sync.Map | Built-in safety, optimized |
| Simple counters | atomic.AddUint64 | Fast, no locks |
| Complex state | sync.Mutex | Full control |
| Pass data | channels | Communication |

---

**5. Monitor and Tune**

```go
// Expose metrics
metrics.RecordConcurrentGoroutines(count)
metrics.RecordProcessingTime(duration)
metrics.RecordErrorRate(errors / total)
```

**Why:**
- Understand actual behavior in production
- Tune batch sizes based on real data
- Alert on anomalies

---

**6. Start Conservative, Then Optimize**

- Start with small batch size (5-10)
- Monitor resource usage
- Gradually increase if safe
- Measure impact of changes

**Why:** Stability > Speed. Crashes are worse than slow.

---

### Common Pitfalls to Avoid

**1. Forgetting to Wait**
```go
// ❌ BAD: Function returns before goroutines finish
for _, item := range items {
    go process(item)
}
return  // Goroutines still running!
```

**2. Not Handling Panics**
```go
// ✅ GOOD: Recover from panics
go func() {
    defer func() {
        if r := recover(); r != nil {
            log.Error("Panic:", r)
        }
    }()
    defer wg.Done()
    doWork()
}()
```

**3. Unbounded Goroutine Creation**
```go
// ❌ BAD: Can create millions of goroutines
for _, item := range millionItems {
    go process(item)
}
```

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


**Common Mistake:**

```go
// ❌ BUG: All goroutines see the same value
for j := 0; j < batchSize; j++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        // BUG: 'j' is shared across all goroutines
        // By the time goroutine runs, j might have changed
        impl.triggerCiPipeline(pipelines[i + j])
    }()
}
```

**What happens:**
- Loop runs: j = 0, 1, 2, 3, 4
- Goroutines are scheduled (not executed immediately)
- By the time goroutines execute, loop has finished
- All goroutines see j = 5 (final value)
- **Result:** We process pipeline[i+5] five times, skip others

**The Fix:**

```go
// ✅ CORRECT: Pass loop variable as parameter
for j := 0; j < batchSize; j++ {
    wg.Add(1)
    index := i + j  // Create new variable
    go func(idx int) {  // Pass as parameter
        defer wg.Done()
        // Each goroutine gets its own copy of idx
        impl.triggerCiPipeline(pipelines[idx])
    }(index)
}
```

**Why this works:**
- Function parameters are passed by value
- Each goroutine gets its own copy of `idx`
- Values are captured at the time of goroutine creation

**Alternative (Go 1.22+):**
```go
// ✅ Also correct in Go 1.22+
// Loop variables are now per-iteration by default
for j := 0; j < batchSize; j++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        impl.triggerCiPipeline(pipelines[i + j])
    }()
}
```

---

### Production Results

**Before (Unbounded Concurrency):**
- 100 pipelines triggered simultaneously
- **Result:** Application crash (OOM)
- **Time to failure:** ~5 seconds
- **Pipelines completed:** 0
- **Success rate:** 0%

**After (Worker Pool, Batch Size = 5):**
- 100 pipelines processed in 20 batches
- **Result:** All completed successfully
- **Total time:** ~2 seconds
- **Pipelines completed:** 100
- **Success rate:** 100%
- **Memory usage:** Stable (~50MB)
- **Database connections:** Max 5 concurrent
- **CPU usage:** 40% (good utilization)

**Key Metrics:**
- **Throughput:** 50 pipelines/second
- **Latency:** ~20ms per pipeline
- **Resource efficiency:** 10x better memory usage
- **Reliability:** 0 crashes in 6 months

---

### Tuning Batch Size

**How to choose the right batch size?**

**Factors to consider:**

1. **External API Rate Limits**
   - Kubernetes API: 50 QPS
   - If each pipeline makes 5 API calls
   - Max concurrent pipelines: 50 / 5 = 10
   - **Safe batch size:** 5-8 (with margin)

2. **Database Connection Pool**
   - Max connections: 100
   - Other services using: ~50
   - Available for this operation: 50
   - **Safe batch size:** 10-20

3. **Memory Constraints**
   - Available memory: 1GB
   - Each goroutine: ~2KB stack + ~100KB heap
   - Max goroutines: 1GB / 102KB ≈ 10,000
   - **Not a constraint for batch size < 100**

4. **CPU Cores**
   - Available cores: 8
   - Optimal for CPU-bound: 8-16 goroutines
   - Our work is I/O-bound (network, database)
   - **Can use higher batch size:** 20-50

**Our Choice: Batch Size = 5**
- Conservative approach
- Prioritizes stability over speed
- Leaves headroom for other operations
- Easy to increase if needed

**Monitoring:**
```go
// We expose metrics to tune batch size
metrics.RecordBatchProcessingTime(duration)
metrics.RecordConcurrentGoroutines(batchSize)
metrics.RecordDatabaseConnections(activeConnections)
```



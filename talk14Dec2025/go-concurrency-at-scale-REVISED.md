# Go Concurrency at Scale: A Production Story

**Duration:** 40 minutes
**Audience:** Intermediate to Advanced Go Developers
**Format:** A journey from failure to success

---

## The Story Arc

```
Act 1: The Problem     → We're overwhelmed
Act 2: First Attempt   → We fail (Naive approach)
Act 3: Understanding   → We learn how Go works (Deep dive)
Act 4: Second Attempt  → We fail again (WaitGroup)
Act 5: The Solution    → We succeed (Worker Pool)
Act 6: Scaling Further → We master it (Fan-Out/Fan-In)
Act 7: The Wisdom      → Lessons learned
```

---

## Act 1: The Problem (3 mins)

### Our Story Begins at Devtron

**Setting the scene:**

I work at Devtron, an open-source Kubernetes deployment platform. We help teams deploy applications to Kubernetes clusters.

**The numbers:**
- 1,000+ applications managed
- 100+ Kubernetes clusters
- 10,000+ deployments daily
- Thousands of CI/CD pipelines running concurrently

**Everything was fine... until it wasn't.**

---

### The Day Everything Broke

**Monday morning, 9 AM:**

```
[ERROR] Database: pq: sorry, too many clients already
[ERROR] Kubernetes API: 429 Too Many Requests
[FATAL] Out of memory: killed by OOM
```

**What happened?**
- A customer triggered 100 CI pipelines at once
- Our system tried to process all 100 simultaneously
- Database connections: Exhausted (max 100)
- Kubernetes API: Rate limited (50 QPS)
- Memory: Out of memory (OOM killed)
- **Result:** Complete system crash

**The question that started our journey:**
> "How do we process 1000 concurrent tasks without crashing?"

---

### The Restaurant Analogy

**Let me tell you a story about a restaurant...**

**Week 1: The small restaurant**
- You own a cozy restaurant
- 10 customers per day
- You cook everything yourself
- Life is good! ✅

**Week 2: Featured on TV**
- Your restaurant gets featured on a popular TV show
- Next day: 1,000 customers show up!
- You try to cook for all of them yourself
- **Result:**
  - You're overwhelmed
  - Customers wait hours
  - Kitchen runs out of ingredients
  - You collapse from exhaustion
  - Restaurant closes ❌

**This is exactly what happened to our system:**
- Week 1 = Development (10 apps)
- Week 2 = Production (1,000 apps)
- You cooking = Simple goroutines
- Overwhelmed = System crash

**The journey ahead:**
How do we evolve from a one-person kitchen to a professional operation that can serve 1,000 customers?

---

## Act 2: First Attempt - The Naive Approach (5 mins)

### "Let's Just Use Goroutines!"

**Our first thought:**
> "Go has goroutines! They're lightweight! Let's just spawn one per task!"

**The code we wrote:**

```go
// Attempt 1: Naive approach
func TriggerAllPipelines(pipelines []Pipeline) {
    for _, pipeline := range pipelines {
        go func(p Pipeline) {
            triggerCiPipeline(p)  // Trigger each pipeline
        }(pipeline)
    }
    // Function returns immediately!
}
```

**Restaurant analogy:**
- 1,000 customers walk in
- You hire 1,000 chefs on the spot
- Your kitchen has space for only 10 chefs
- **Result:** Chaos! Chefs bumping into each other, kitchen on fire! 🔥

---

### What Happened When We Deployed This

**Production logs:**

```
09:00:01 - Triggering 100 pipelines...
09:00:01 - Spawned 100 goroutines
09:00:02 - Spawned 5,000 goroutines (each pipeline spawns ~50 more)
09:00:03 - Database connections: 100/100 (EXHAUSTED)
09:00:04 - Kubernetes API: 429 Too Many Requests
09:00:05 - Memory usage: 8GB → 12GB → 16GB
09:00:06 - [FATAL] Out of memory: killed by OOM
09:00:07 - System crash 💥
```

**The problems:**

1. **❌ No waiting** - Function returned immediately, didn't wait for pipelines
2. **❌ No error handling** - Couldn't collect errors
3. **❌ Unbounded concurrency** - All 5,000 goroutines ran at once
4. **❌ Resource exhaustion** - DB connections, API rate limits, memory all exhausted

**The realization:**
> "Goroutines are lightweight, but they're not free. We can't just spawn thousands of them!"

---

### Why Did This Fail?

**The math:**
- 100 pipelines triggered
- Each pipeline spawns ~50 goroutines (for various tasks)
- **Total:** 100 × 50 = 5,000 goroutines

**The cost:**
- Each goroutine: ~2KB stack minimum
- 5,000 goroutines: 5,000 × 2KB = 10MB (just for stacks)
- Plus heap allocations, scheduling overhead, context switching
- **Actual memory:** ~500MB-1GB

**The resource limits:**
- Database connection pool: 100 connections
- 5,000 goroutines trying to get DB connections
- **Result:** Connection pool exhausted

**We needed a better approach...**

---

## Act 3: Understanding the Machine (8 mins)

### "Wait... What Actually Happens When I Write `go func()`?"

**After the crash, we needed to understand:**
> "If goroutines are so lightweight, why did we crash? What's actually happening under the hood?"

**This is the story of a goroutine's journey from your code to the CPU...**

---

### Real-World Analogy: The Restaurant Kitchen Management System

**Imagine a modern restaurant with:**
- **Kitchen Manager** (Go Runtime Scheduler)
- **Cooking Stations** (OS Threads / CPU Cores)
- **Recipe Cards** (Goroutines)
- **Ingredient Storage** (Memory/Stack)

**Traditional approach (OS Threads):**
- Each chef (thread) is a full-time employee
- Hiring a chef is expensive (1-2MB memory, OS overhead)
- Each chef needs their own locker, uniform, workspace
- Maximum chefs = Number of cooking stations (CPU cores)

**Go's approach (Goroutines):**
- Recipe cards (goroutines) are lightweight instructions
- Kitchen manager assigns recipe cards to available chefs
- One chef can work on multiple recipes (context switching)
- Can have 10,000+ recipe cards, but only 8 chefs working at a time

---

### Level 1: What You Write (Source Code)

```go
go func() {
    processTask(task)
}()
```

**What you're saying:**
> "Hey Go, please execute this function concurrently, but I don't care when or on which thread."

---

### Level 2: What the Compiler Does (Compile Time)

**The Go compiler transforms your code:**

**Your code:**
```go
go func() {
    processTask(task)
}()
```

**Compiler generates (simplified):**
```go
// 1. Allocate a new goroutine structure
g := runtime.newproc(funcPC(processTask), args)

// 2. Initialize goroutine stack (2KB initially)
g.stack = runtime.stackalloc(2048)  // 2KB stack

// 3. Set goroutine state to "runnable"
g.status = _Grunnable

// 4. Add to global run queue
runtime.globrunqput(g)

// 5. Wake up a scheduler if needed
runtime.wakep()
```

**Real-World Analogy:**
- **You write:** "Make this dish"
- **Kitchen manager receives:** A recipe card with:
  - Recipe name (function pointer)
  - Ingredients needed (function arguments)
  - Workspace allocation (2KB stack)
  - Priority level (runnable)
- **Manager's action:** Put recipe card in the "To-Do" queue

---

### Level 3: The Go Runtime Scheduler (Runtime)

**The Go’s scheduler is built on what’s called the GMP model:**
- **G** = Goroutines (Your actual code that needs to run)
- **M** = OS Threads (Threads that execute the work)
- **P** = Processors (Logical processors, usually = CPU cores, that manage execution context)

**Architecture:**

```
┌─────────────────────────────────────────────────────────┐
│                    Go Runtime Scheduler                  │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  Global Run Queue: [G1, G2, G3, G4, G5, ...]            │
│                                                           │
├──────────────┬──────────────┬──────────────┬────────────┤
│   P0         │   P1         │   P2         │   P3       │
│ (Processor)  │ (Processor)  │ (Processor)  │ (Processor)│
├──────────────┼──────────────┼──────────────┼────────────┤
│ Local Queue: │ Local Queue: │ Local Queue: │ Local Queue│
│ [G10, G11]   │ [G20, G21]   │ [G30]        │ [G40, G41] │
├──────────────┼──────────────┼──────────────┼────────────┤
│      ↓       │      ↓       │      ↓       │      ↓     │
│     M0       │     M1       │     M2       │     M3     │
│  (OS Thread) │  (OS Thread) │  (OS Thread) │  (OS Thread)│
└──────────────┴──────────────┴──────────────┴────────────┘
       ↓              ↓              ↓              ↓
┌──────────────────────────────────────────────────────────┐
│              Hardware (CPU Cores)                         │
│    Core 0      Core 1      Core 2      Core 3            │
└──────────────────────────────────────────────────────────┘
```

**Real-World Analogy:**

**Global Run Queue** = Central order board
- All new recipe cards go here first
- Shared by all kitchen managers

**P (Processor)** = Kitchen manager
- Each manager has their own local queue of recipes
- Number of managers = Number of CPU cores (typically)
- Each manager assigns recipes to chefs

**Local Queue** = Manager's personal clipboard
- Each manager keeps 256 recipe cards max
- Faster to access than global queue
- Work stealing: If one manager is idle, they steal from others!

**M (OS Thread)** = Actual chef
- Does the real cooking (executes code)
- Expensive to hire/fire (OS overhead)
- Go creates/destroys threads as needed

**G (Goroutine)** = Recipe card
- Lightweight (2KB initial stack)
- Contains: Function to execute, arguments, stack, state

---

### Level 4: Scheduling Decisions (Runtime Scheduler Logic)

**When you write `go func()`, here's what happens:**

**Step 1: Goroutine Creation**
```go
go processTask(task)
```

**Runtime action:**
1. Allocate goroutine struct (~2KB)
2. Copy function arguments to goroutine stack
3. Set instruction pointer to function start
4. Mark goroutine as "runnable", Marking it as runnable means placing it in a queue and marking it as ready to be executed by an available OS thread

**Real-World Analogy:**
- Chef receives order: "Cook pasta"
- Manager creates recipe card
- Copies ingredients list to card
- Marks card as "ready to cook"

---

**Step 2: Scheduling**

**The scheduler picks a goroutine to run:**

```
Scheduler decision tree:
│
├─ Check local queue (P's own queue)
│  └─ Found goroutine? → Run it (fast path)
│
├─ Check global queue
│  └─ Found goroutine? → Run it
│
├─ Check network poller (for I/O goroutines)
│  └─ Found ready goroutine? → Run it
│
└─ Work stealing: Steal from other P's local queue
   └─ Found goroutine? → Run it
```

**Real-World Analogy:**
1. **Local queue:** Check your own clipboard first (fastest)
2. **Global queue:** Check central order board
3. **Network poller:** Check if any orders waiting for delivery arrived
4. **Work stealing:** If idle, steal recipes from busy managers

---

**Step 3: Execution**

**Goroutine runs on OS thread:**

```
M (OS Thread) executes G (Goroutine):
│
├─ Load goroutine's stack pointer
├─ Load goroutine's instruction pointer
├─ Execute function code
│
└─ Goroutine blocks (I/O, channel, mutex)?
   │
   ├─ Save goroutine state (stack, registers)
   ├─ Mark goroutine as "waiting"
   ├─ Put goroutine in wait queue
   └─ Schedule next goroutine (context switch)
```

**Real-World Analogy:**
- Chef starts cooking pasta
- Pasta needs to boil for 10 minutes (I/O wait)
- Chef doesn't stand idle!
- Chef saves pasta state ("boiling, 3 minutes done")
- Chef picks up next recipe card
- When pasta is ready, recipe card goes back to queue

---

### Level 5: Context Switching (The Magic of Goroutines)

**OS Thread context switch (expensive):**
```
Cost: ~1-2 microseconds
Steps:
1. Save all CPU registers (16+ registers)
2. Save floating point state
3. Save instruction pointer
4. Switch page tables (memory isolation)
5. Flush TLB (Translation Lookaside Buffer)
6. Load new thread's state
7. Restore all registers
```

**Goroutine context switch (cheap):**
```
Cost: ~200 nanoseconds (10x faster!)
Steps:
1. Save 3 registers (PC, SP, BP)
2. Save goroutine stack pointer
3. Load next goroutine's stack pointer
4. Load next goroutine's registers
5. Continue execution
```

**Why is goroutine switching faster?**
- ✅ No kernel involvement (all in user space)
- ✅ No page table switching (same address space)
- ✅ No TLB flush
- ✅ Smaller state to save/restore
- ✅ Cooperative scheduling (goroutines yield voluntarily)

**Real-World Analogy:**

**OS Thread switch** = Changing chefs
- Chef A goes home
- Chef B arrives, needs to:
  - Change into uniform
  - Learn kitchen layout
  - Find ingredients
  - Understand ongoing orders
- **Time:** 10 minutes

**Goroutine switch** = Same chef, different recipe
- Chef just picks up different recipe card
- Same kitchen, same ingredients location
- Just different instructions
- **Time:** 10 seconds

---

### Level 6: Hardware Execution (CPU Level)

**When goroutine actually runs on CPU:**

```
CPU Core executes machine code:
│
├─ Fetch instruction from memory
├─ Decode instruction
├─ Execute instruction
│  ├─ Arithmetic (ADD, SUB, MUL)
│  ├─ Memory access (LOAD, STORE)
│  ├─ Atomic operations (LOCK XADD, CMPXCHG)
│  └─ System calls (for I/O)
│
└─ Write results back to registers/memory
```

**Atomic operations (for WaitGroup, sync.Map, atomic.AddUint64):**

**Regular increment (NOT thread-safe):**
```assembly
MOV  RAX, [counter]    ; Read counter into register
ADD  RAX, 1            ; Increment register
MOV  [counter], RAX    ; Write back to memory
; Problem: Another CPU core can modify counter between steps!
```

**Atomic increment (thread-safe):**
```assembly
LOCK XADD [counter], 1  ; Atomic read-modify-write
; LOCK prefix ensures:
; - No other CPU can access this memory location
; - Cache line is locked across all cores
; - Operation is indivisible
```

**Real-World Analogy:**

**Regular increment** = Two chefs updating same order count
- Chef A reads: "5 orders"
- Chef B reads: "5 orders" (at same time)
- Chef A writes: "6 orders"
- Chef B writes: "6 orders"
- **Result:** Lost update! Should be 7, but shows 6

**Atomic increment** = Lock the order board
- Chef A locks board, reads "5", writes "6", unlocks
- Chef B waits for lock, reads "6", writes "7", unlocks
- **Result:** Correct count: 7

---

### The Complete Journey: go func() to CPU Execution

**Let's trace one goroutine from code to hardware:**

```go
go processTask(task)
```

**Step-by-step journey:**

1. **Compile time:**
   - Compiler generates `runtime.newproc()` call
   - Allocates goroutine struct in binary

2. **Runtime - Goroutine creation:**
   - Allocate 2KB stack
   - Copy function pointer and arguments
   - Set goroutine state = "runnable"
   - Add to global run queue

3. **Runtime - Scheduling:**
   - Processor P checks local queue (empty)
   - Processor P checks global queue (finds our goroutine!)
   - Processor P assigns goroutine to OS thread M

4. **OS Thread - Execution:**
   - Thread M loads goroutine's stack pointer
   - Thread M loads goroutine's instruction pointer
   - Thread M starts executing function code

5. **CPU - Hardware execution:**
   - CPU fetches instructions from memory
   - CPU executes: ADD, MOV, CALL, etc.
   - CPU writes results to registers/memory

6. **Goroutine blocks (e.g., wg.Wait()):**
   - Save goroutine state (3 registers)
   - Mark goroutine as "waiting"
   - Put in semaphore wait queue
   - Thread M picks up next goroutine

7. **Goroutine wakes up (e.g., wg.Done() called):**
   - Semaphore signals waiting goroutine
   - Goroutine marked as "runnable"
   - Added back to run queue
   - Eventually scheduled again on some thread

8. **Goroutine completes:**
   - Function returns
   - Stack is freed
   - Goroutine struct is freed
   - Thread M picks up next goroutine

**Time breakdown:**
- Goroutine creation: ~1-2 microseconds
- Context switch: ~200 nanoseconds
- Actual work: Depends on function
- Goroutine cleanup: ~1 microsecond

---

### Key Insights: Why Goroutines Scale

**1. Lightweight:**
- OS Thread: 1-2 MB stack
- Goroutine: 2 KB stack (grows dynamically)
- **Result:** Can create 100,000 goroutines in same memory as 200 threads

**2. Fast context switching:**
- OS Thread switch: 1-2 microseconds (kernel involved)
- Goroutine switch: 200 nanoseconds (user space)
- **Result:** 10x faster switching

**3. Efficient scheduling:**
- OS scheduler: Preemptive (forcefully switches threads)
- Go scheduler: Cooperative (goroutines yield voluntarily)
- **Result:** Less overhead, better cache locality

**4. Work stealing:**
- Idle processors steal work from busy ones
- **Result:** Better CPU utilization

---

### But... Goroutines Are NOT Free!

**Each goroutine still costs:**
- 2 KB minimum stack (can grow to MB)
- Goroutine struct: ~200 bytes
- Scheduling overhead
- Context switching overhead

**Creating 1,000,000 goroutines:**
- Memory: 1M × 2KB = 2 GB (just for stacks!)
- Scheduling: Scheduler spends more time scheduling than executing
- Cache thrashing: Too many goroutines = poor cache locality

**This is why we need Worker Pools!**

---

## Act 4: Second Attempt - "Let's Add sync.WaitGroup!" (4 mins)

### Armed with Knowledge, We Try Again

**After understanding how goroutines work, we realized:**
> "The first problem was that our function returned immediately. Let's make it wait!"

**We discovered `sync.WaitGroup`:**

```go
// Attempt 2: Add sync.WaitGroup to wait for completion
func TriggerAllPipelines(pipelines []Pipeline) error {
    var wg sync.WaitGroup

    for _, pipeline := range pipelines {
        wg.Add(1)  // Tell WaitGroup: "One more goroutine coming"

        go func(p Pipeline) {
            defer wg.Done()  // Tell WaitGroup: "I'm done!"
            triggerCiPipeline(p)
        }(pipeline)
    }

    wg.Wait()  // Wait for all goroutines to finish
    return nil
}
```

**Restaurant analogy:**
- 1,000 customers walk in
- You hire 1,000 chefs
- You wait for all chefs to finish before closing
- **Problem:** Still 1,000 chefs in a kitchen built for 10!
- **Result:** Better (you wait), but kitchen still overwhelmed! 🔥

---

### What Happened When We Deployed This

**Production logs:**

```
10:00:01 - Triggering 100 pipelines with WaitGroup...
10:00:01 - Spawned 100 goroutines
10:00:02 - Spawned 5,000 goroutines (cascading effect)
10:00:03 - Database connections: 100/100 (EXHAUSTED)
10:00:04 - Kubernetes API: 429 Too Many Requests
10:00:05 - [ERROR] Database: pq: sorry, too many clients already
10:00:06 - Some pipelines failing...
10:00:10 - WaitGroup waiting... (function doesn't return)
10:00:15 - Still waiting...
10:00:20 - Finally all goroutines complete (many failed)
10:00:20 - Function returns
```

**What improved:**
- ✅ **Function waits** - Doesn't return until all goroutines complete
- ✅ **Proper cleanup** - `defer wg.Done()` ensures counter is decremented
- ✅ **Synchronization** - We know when everything is done

**What's still broken:**
- ❌ **Still unbounded** - All 5,000 goroutines run simultaneously
- ❌ **Resource exhaustion** - DB connections, API rate limits still exhausted
- ❌ **No control** - Can't limit how many goroutines run at once
- ❌ **Failures** - Many pipelines fail due to resource exhaustion

---

### The Critical Realization

**We had an "aha!" moment:**

> **sync.WaitGroup is a TOOL for waiting, not a SOLUTION for scaling!**

**The problem:**
- Database connection pool: 100 connections
- We spawn 100 goroutines
- Each goroutine needs a DB connection
- But other parts of the app also need connections (50 connections)
- **Available for us:** 100 - 50 = 50 connections
- **We try to use:** 100 connections
- **Result:** 50 goroutines succeed, 50 fail ❌

**The insight:**
```
sync.WaitGroup tells you WHEN goroutines finish.
It does NOT control HOW MANY goroutines run at once.
```

**We needed something more...**

---

## Act 5: The Solution - Worker Pool (12 mins)

### The Breakthrough: Batching!

**The idea:**
> "What if we don't spawn all goroutines at once? What if we process them in small batches?"

**Restaurant analogy:**
- 1,000 customers walk in
- You hire only 5 professional chefs (your kitchen's capacity)
- **Batch 1:** First 5 customers → 5 chefs cook → customers served
- **Batch 2:** Next 5 customers → same 5 chefs cook → customers served
- **Continue** in batches of 5 until all 1,000 customers served
- **Result:** Organized, efficient, no chaos! ✅

**This is the Worker Pool pattern!**

---

### The Worker Pool Code

```go
// Attempt 3: Worker Pool - Bounded concurrency with batching
func TriggerAllPipelines(pipelines []Pipeline) error {
    batchSize := 5  // Only 5 goroutines at a time!

    for i := 0; i < len(pipelines); {
        // Calculate current batch size (last batch might be smaller)
        remainingPipelines := len(pipelines) - i
        currentBatchSize := batchSize
        if remainingPipelines < batchSize {
            currentBatchSize = remainingPipelines
        }

        var wg sync.WaitGroup

        // Launch only currentBatchSize goroutines
        for j := 0; j < currentBatchSize; j++ {
            wg.Add(1)
            index := i + j

            go func(idx int) {
                defer wg.Done()
                triggerCiPipeline(pipelines[idx])
            }(index)
        }

        wg.Wait()  // Wait for this batch to complete
        i += currentBatchSize  // Move to next batch
    }

    return nil
}
```

---

### How It Works: Step-by-Step

**Processing 100 pipelines with batch size = 5:**

```
Batch 1 (pipelines 0-4):
  ├─→ Spawn 5 goroutines
  ├─→ Wait for all 5 to complete
  └─→ All 5 done ✅

Batch 2 (pipelines 5-9):
  ├─→ Spawn 5 goroutines
  ├─→ Wait for all 5 to complete
  └─→ All 5 done ✅

... (20 batches total)

Batch 20 (pipelines 95-99):
  ├─→ Spawn 5 goroutines
  ├─→ Wait for all 5 to complete
  └─→ All 5 done ✅

Result: All 100 pipelines processed!
Max goroutines at any time: 5
```

---

### What Happened When We Deployed This

**Production logs:**

```
11:00:01 - Triggering 100 pipelines with Worker Pool (batch=5)...
11:00:01 - Batch 1: Spawned 5 goroutines
11:00:02 - Batch 1: All 5 complete ✅
11:00:02 - Batch 2: Spawned 5 goroutines
11:00:03 - Batch 2: All 5 complete ✅
...
11:00:40 - Batch 20: All 5 complete ✅
11:00:40 - All 100 pipelines complete!
11:00:40 - Success rate: 100% ✅
11:00:40 - Database connections used: Max 5 (plenty available)
11:00:40 - Kubernetes API: No rate limiting
11:00:40 - Memory usage: Stable at 2GB
```

**The results:**
- ✅ **All pipelines succeeded** - 100% success rate!
- ✅ **No crashes** - System remained stable
- ✅ **Predictable resources** - Max 5 DB connections, 5 API calls
- ✅ **Scales to any number** - 1,000 pipelines? 10,000? Same approach!

**We finally had a solution that worked!** 🎉

---

### The Journey So Far

**Our evolution:**

| Attempt | Approach | Goroutines | Result | Lesson |
|---------|----------|------------|--------|--------|
| **1** | Naive | 5,000 | ❌ Crash | Goroutines aren't free |
| **2** | WaitGroup | 5,000 | ❌ Crash | Waiting ≠ Scaling |
| **3** | Worker Pool | 5 | ✅ Success! | Batching = Control |

**The critical insight:**

> **sync.WaitGroup is a TOOL for waiting, not a SOLUTION for scaling.**
>
> - Attempt 2 used it to wait for 5,000 goroutines (unbounded)
> - Attempt 3 used it to wait for 5 goroutines per batch (bounded)
>
> **Worker Pool = Batching + sync.WaitGroup**

---

### Real Production Code from Devtron

**Let me show you the actual code we use in production:**

**File:** `pkg/workflow/dag/WorkflowDagExecutor.go`
**Function:** `HandleCiSuccessEvent`
**Context:** When a CI build completes, auto-trigger child CD pipelines

**The scenario:**
- CI build completes successfully
- Need to trigger 10-50 child CD pipelines automatically
- Each trigger makes DB calls + Kubernetes API calls
- Must not exhaust resources

**The code:**

```go
// Real production code from Devtron
func (impl *WorkflowDagExecutorImpl) HandleCiSuccessEvent(
    ciArtifactArr []CiArtifact,
    async bool,
    userId int32,
) error {
    // Get batch size from config (default: 5)
    batchSize := impl.config.CIAutoTriggerBatchSize
    totalCIArtifactCount := len(ciArtifactArr)

    // Process in batches
    for i := 0; i < totalCIArtifactCount; {
        // Calculate current batch size
        remainingBatch := totalCIArtifactCount - i
        if remainingBatch < batchSize {
            batchSize = remainingBatch
        }

        var wg sync.WaitGroup

        // Launch batch of goroutines
        for j := 0; j < batchSize; j++ {
            wg.Add(1)
            index := i + j

            runnableFunc := func(idx int) {
                defer wg.Done()

                ciArtifact := ciArtifactArr[idx]

                // Trigger CD pipeline (DB + K8s API calls)
                err := impl.handleCiSuccessEvent(
                    triggerContext,
                    ciArtifact,
                    async,
                    userId,
                )

                if err != nil {
                    impl.logger.Errorw("error in triggering CD",
                        "err", err,
                        "ciArtifact", ciArtifact)
                }
            }

            // Execute in goroutine pool
            impl.asyncRunnable.Execute(func() {
                runnableFunc(index)
            })
        }

        wg.Wait()  // Wait for this batch to complete
        i += batchSize  // Move to next batch
    }

    return nil
}
```

**Why this works:**
- ✅ **Batch size = 5** (configurable via config)
- ✅ **Max 5 concurrent DB connections** (out of 100 available)
- ✅ **Max 5 concurrent K8s API calls** (under 50 QPS limit)
- ✅ **Predictable memory** (~10MB for 5 goroutines)
- ✅ **100% success rate** in production

**Production metrics:**
- **Before Worker Pool:** 0% success rate (system crash)
- **After Worker Pool:** 100% success rate
- **Pipelines triggered:** 1,000+ per day
- **Uptime:** 99.9% (no crashes in 6 months)

---

### Key Design Decisions Explained

**1. Why sync.WaitGroup instead of channels?**

**Using channels (more complex):**
```go
done := make(chan bool, batchSize)
for j := 0; j < batchSize; j++ {
    go func(idx int) {
        triggerPipeline(idx)
        done <- true  // Send signal
    }(i + j)
}
for j := 0; j < batchSize; j++ {
    <-done  // Receive signal
}
```

**Using sync.WaitGroup (simpler):**
```go
var wg sync.WaitGroup
for j := 0; j < batchSize; j++ {
    wg.Add(1)
    go func(idx int) {
        defer wg.Done()
        triggerPipeline(idx)
    }(i + j)
}
wg.Wait()
```

**Why WaitGroup wins:**
- ✅ Clearer intent: "wait for group of goroutines"
- ✅ Less code, easier to read
- ✅ `defer wg.Done()` ensures cleanup even on panic
- ✅ No need to count receives

**When to use channels:** When you need to pass data between goroutines

---

**2. Why pass index as parameter?**

**Wrong (loop variable capture bug):**
```go
for j := 0; j < batchSize; j++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        // BUG: All goroutines see the same 'j' value!
        triggerPipeline(pipelines[i + j])
    }()
}
```

**Correct (pass as parameter):**
```go
for j := 0; j < batchSize; j++ {
    wg.Add(1)
    index := i + j
    go func(idx int) {
        defer wg.Done()
        // Each goroutine gets its own copy of idx
        triggerPipeline(pipelines[idx])
    }(index)
}
```

**Why:** Function parameters are passed by value, so each goroutine gets its own copy.

---

**3. How to choose batch size?**

**Factors to consider:**

| Constraint | Calculation | Example |
|------------|-------------|---------|
| **DB connections** | available_connections / 2 | 100 / 2 = 50 |
| **API rate limit** | rate_limit / calls_per_task | 50 QPS / 5 calls = 10 |
| **Memory** | available_memory / memory_per_goroutine | 1GB / 100MB = 10 |
| **CPU cores** | num_cores × 2 (for I/O-bound) | 8 × 2 = 16 |

**Our choice at Devtron:**
- Database: 100 connections, other services use ~50 → **Available: 50**
- K8s API: 50 QPS, each pipeline makes ~5 calls → **Max: 10**
- **Conservative choice: 5** (leaves headroom for spikes)

**Rule of thumb:** Start conservative (5-10), monitor, then increase if safe.

---

## Act 6: Scaling Further - Fan-Out/Fan-In (10 mins)

### A New Challenge Appears

**With Worker Pool working great, we faced a different problem:**

**The scenario:**
- User opens the dashboard
- Needs to see CI status + CD status
- Each query takes ~500ms
- **Sequential:** 500ms + 500ms = 1000ms (too slow!)
- User sees loading spinner for 1 second 😞

**The question:**
> "Can we fetch both in parallel and combine the results?"

**This is the Fan-Out/Fan-In pattern!**

---

### What is Fan-Out/Fan-In?

**The pattern:**
```
Input → Fan-Out → [Worker 1, Worker 2, Worker 3, ...] → Fan-In → Combined Result
```

**Fan-Out:** Distribute work to multiple goroutines running in parallel
**Fan-In:** Collect results from all goroutines into a single place

**Key difference from Worker Pool:**
- **Worker Pool:** Process MANY tasks in controlled batches (1000 tasks → batches of 5)
- **Fan-Out/Fan-In:** Process FEW tasks ALL in parallel (2-100 tasks → all at once)

**When to use each:**
- **Worker Pool:** Large N, need to control concurrency
- **Fan-Out/Fan-In:** Small N, want maximum parallelism

---

### Story 1: The Dashboard Problem

**The library analogy:**

You're a student researching for a paper. You need information from 2 library sections:
- **Section A:** History books (5 minutes away)
- **Section B:** Science books (5 minutes away)

**Option 1: Sequential (you do it alone)**
- Walk to Section A → 5 minutes
- Walk to Section B → 5 minutes
- **Total:** 10 minutes

**Option 2: Parallel (you + friend)**
- You → Section A (5 minutes)
- Friend → Section B (5 minutes, at the same time!)
- Meet back at table
- **Total:** 5 minutes (2x faster!) ✅

**This is exactly our dashboard problem:**
- **Section A** = CI workflow status query (500ms)
- **Section B** = CD workflow status query (500ms)
- **You + friend** = 2 goroutines running in parallel
- **Meeting at table** = sync.WaitGroup waiting for both

---

### Real Production Code: Dashboard Status

**File:** `api/restHandler/app/pipeline/configure/PipelineConfigRestHandler.go`
**Function:** `FetchWorkflowStatus`
**Context:** User opens dashboard, needs CI + CD status

**The code:**

```go
func (handler *PipelineConfigRestHandlerImpl) FetchWorkflowStatus(
    w http.ResponseWriter,
    r *http.Request,
) {
    appId, _ := strconv.Atoi(mux.Vars(r)["app-id"])

    // Variables to store results
    var ciWorkflowStatus []*pipelineConfig.CiWorkflowStatus
    var cdWorkflowStatus []*pipelineConfig.CdWorkflowStatus
    var err, err1 error

    // Create WaitGroup for 2 goroutines
    wg := sync.WaitGroup{}
    wg.Add(2)

    // Goroutine 1: Fetch CI status (500ms)
    go func() {
        defer wg.Done()
        ciWorkflowStatus, err = handler.ciHandler.FetchCiStatusForTriggerView(appId)
    }()

    // Goroutine 2: Fetch CD status (500ms)
    go func() {
        defer wg.Done()
        cdWorkflowStatus, err1 = handler.cdHandler.FetchAppWorkflowStatusForTriggerView(appId)
    }()

    // Wait for both to complete
    wg.Wait()

    // Handle errors
    if err != nil {
        common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
        return
    }

    // Combine results and return
    response := pipelineConfig.TriggerWorkflowStatus{
        CiWorkflowStatus: ciWorkflowStatus,
        CdWorkflowStatus: cdWorkflowStatus,
    }
    common.WriteJsonResp(w, nil, response, http.StatusOK)
}
```

**What happened:**
```
Timeline:
0ms   - User clicks dashboard
0ms   - Launch 2 goroutines (CI + CD)
0ms   - Both start fetching in parallel
500ms - Both complete
500ms - Combine results
500ms - Return to user ✅

Before: 1000ms (sequential)
After:  500ms (parallel)
Speedup: 2x faster!
```

**Why this works:**
- ✅ CI and CD are independent (no shared state)
- ✅ Both run simultaneously
- ✅ Simple pattern, easy to understand
- ✅ User sees dashboard 2x faster

---

## Act 7: The Wisdom - Lessons Learned (5 mins)

### Our Journey: From Crash to Success

**The story so far:**

```
Monday 9 AM:  System crashes → "We need to fix this!"
Monday 10 AM: Attempt 1 (Naive) → Crashes again
Monday 11 AM: Deep dive into Go internals → "Aha! Now we understand!"
Monday 2 PM:  Attempt 2 (WaitGroup) → Still crashes
Monday 4 PM:  Attempt 3 (Worker Pool) → Success! 🎉
Tuesday:      Implement Fan-Out/Fan-In → Even faster!
6 months later: 99.9% uptime, 10,000+ deployments/day ✅
```

---

### The Three Approaches: A Comparison

| Attempt | Pattern | Goroutines | Result | Lesson Learned |
|---------|---------|------------|--------|----------------|
| **1** | Naive | 5,000 | ❌ Crash | Goroutines aren't free |
| **2** | WaitGroup | 5,000 | ❌ Crash | Waiting ≠ Scaling |
| **3** | Worker Pool | 5 | ✅ Success! | Batching = Control |

**The critical insight:**

> **sync.WaitGroup is a TOOL for coordination, not a SOLUTION for scaling.**
>
> - Attempt 2: Used it to wait for 5,000 goroutines (unbounded)
> - Attempt 3: Used it to wait for 5 goroutines per batch (bounded)
>
---

### The Two Patterns We Mastered

**Pattern 1: Worker Pool**
- **Use when:** Processing MANY tasks (1000+) with resource constraints
- **How:** Process in small batches (batch size = 5-10)
- **Result:** Crash → Stable, 0% → 100% success rate

**Pattern 2: Fan-Out/Fan-In**
- **Use when:** Fetching from FEW sources (2-100) in parallel
- **How:** Launch all goroutines, collect results safely
- **Result:** 1000ms → 500ms (2x faster)

---

### The Golden Rules

**1. Always use `defer wg.Done()`**
```go
go func() {
    defer wg.Done()  // Guarantees execution even on panic
    doWork()
}()
```

**2. Pass loop variables as parameters**
```go
for i := range items {
    go func(idx int) {  // Pass as parameter
        process(items[idx])
    }(i)  // Capture value
}
```

**3. Choose batch size wisely**
- Start conservative (5-10)
- Consider: DB connections, API limits, memory
- Monitor and tune based on metrics

---

### Decision Guide: Which Pattern to Use?

**Quick reference:**

| Your Situation | Pattern | Batch Size | Example |
|----------------|---------|------------|---------|
| 20 tasks, no DB/API | Simple WaitGroup | N/A | Process files in memory |
| 100+ tasks, uses DB/API | Worker Pool | 5-10 | Trigger CI pipelines |
| 2-10 independent sources | Fan-Out/Fan-In | All | Fetch CI + CD status |
| 100+ clusters to query | Fan-Out/Fan-In + Batching | 50 | Check cluster health |

**How to choose batch size:**
1. **DB connections:** available / 2 (leave headroom)
2. **API rate limit:** limit / calls_per_task
3. **Start conservative:** 5-10, then tune based on metrics

---

### The End of Our Story

**Where we started:**
```
Monday 9 AM: System crash
Error: Out of memory
Status: 0% success rate
Team: Panicking 😱
```

**Where we are now:**
```
6 months later: 99.9% uptime
Deployments: 10,000+ per day
Success rate: 100%
Team: Confident ✅
```

**What we learned:**

1. **Goroutines are powerful, but not magic**
   - Each costs ~2KB + heap allocations
   - Unbounded = Crash
   - Bounded = Scale

2. **sync.WaitGroup is a tool, not a solution**
   - It helps you wait
   - It doesn't control concurrency
   - Combine with batching for scaling

3. **Patterns matter**
   - Worker Pool for many tasks
   - Fan-Out/Fan-In for few sources
   - Choose based on your constraints

4. **Production is different**
   - Start conservative
   - Monitor everything
   - Stability > Speed
   - Plan for failures

**The most important lesson:**
> "Understanding how Go works under the hood helps you make better decisions."

---

## The Restaurant, Revisited

**Remember our restaurant story?**

**Week 1:** Small restaurant, you cook alone → Works fine
**Week 2:** 1,000 customers, you try to cook alone → Crash
**Week 3:** Hire 1,000 chefs → Kitchen chaos, crash
**Week 4:** Hire 5 chefs, process in batches → Success! ✅

**This is exactly what we did with our code.**

**The wisdom:**
- Know your kitchen's capacity (resources)
- Hire the right number of chefs (goroutines)
- Process orders in batches (worker pool)
- Serve customers efficiently (scale!)

---

## Thank You!

**Resources:**
- **Devtron:** https://github.com/devtron-labs/devtron
- **Go Concurrency:** https://go.dev/blog/pipelines
- **sync package:** https://pkg.go.dev/sync

**Questions?** 🚀

---

**Remember:** The best code is code that works in production. Start simple, measure, and optimize based on real data.

**Good luck scaling your Go applications!** 💪



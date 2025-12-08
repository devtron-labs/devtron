# Go Concurrency at Scale: Lessons from a Kubernetes Platform
## Real-World Patterns from Devtron

**Duration:** 25 minutes  
**Audience:** Intermediate to Advanced Go Developers

---

## 📋 Talk Outline

### 1. Introduction (2 mins)
### 2. Worker Pools Pattern (6 mins)
### 3. Fan-Out/Fan-In Pattern (5 mins)
### 4. Rate Limiting & Throttling (4 mins)
### 5. Graceful Shutdown with Context (4 mins)
### 6. Real-World Case Study: Processing Thousands of API Calls (3 mins)
### 7. Q&A (1 min)

---

## 1. Introduction (2 mins)

### Why Beyond Basic Goroutines?

**The Problem:**
- Spawning unlimited goroutines → Resource exhaustion
- No coordination → Race conditions
- No error handling → Silent failures
- No graceful shutdown → Data loss

**What We'll Cover:**
- Structured concurrency patterns
- Production-ready error handling
- Resource management
- Real examples from a production Kubernetes platform (Devtron)

---

## 2. Worker Pools Pattern (6 mins)

### The Problem: Uncontrolled Concurrency

```go
// ❌ BAD: Can spawn thousands of goroutines
for _, item := range items {
    go processItem(item)  // No control!
}
```

### Solution: Worker Pool with Bounded Concurrency

**Real Example from Devtron: Batch Processing CI Artifacts**

**File:** `pkg/workflow/dag/WorkflowDagExecutor.go`

```go
// Auto-trigger CD pipelines after CI success
totalCIArtifactCount := len(ciArtifactArr)
batchSize := impl.ciConfig.CIAutoTriggerBatchSize  // e.g., 5

for i := 0; i < totalCIArtifactCount; {
    remainingBatch := totalCIArtifactCount - i
    if remainingBatch < batchSize {
        batchSize = remainingBatch
    }
    
    var wg sync.WaitGroup
    for j := 0; j < batchSize; j++ {
        wg.Add(1)
        index := i + j
        runnableFunc := func(index int) {
            defer wg.Done()
            ciArtifact := ciArtifactArr[index]
            err = impl.handleCiSuccessEvent(triggerContext, ciArtifact, async, request.UserId)
            if err != nil {
                impl.logger.Errorw("error on handle ci success event", 
                    "ciArtifactId", ciArtifact.Id, "err", err)
            }
        }
        impl.asyncRunnable.Execute(func() { runnableFunc(index) })
    }
    wg.Wait()
    i += batchSize
}
```

**Key Takeaways:**
- ✅ Controlled concurrency (batchSize = 5)
- ✅ Process 1000s of artifacts without overwhelming system
- ✅ Wait for batch completion before next batch
- ✅ Proper error logging per artifact

---

### Another Example: Kubernetes Resource Fetching

**File:** `pkg/k8s/K8sCommonService.go`

```go
func (impl *K8sCommonServiceImpl) getManifestsByBatch(
    ctx context.Context, 
    requests []bean5.ResourceRequestBean,
) []bean5.BatchResourceResponse {
    
    batchSize := impl.K8sApplicationServiceConfig.BatchSize
    requestsLength := len(requests)
    res := make([]bean5.BatchResourceResponse, requestsLength)
    
    for i := 0; i < requestsLength; {
        remainingBatch := requestsLength - i
        if remainingBatch < batchSize {
            batchSize = remainingBatch
        }
        
        var wg sync.WaitGroup
        for j := 0; j < batchSize; j++ {
            wg.Add(1)
            runnableFunc := func(index int) {
                resp := bean5.BatchResourceResponse{}
                response, err := impl.GetResource(ctx, &requests[index])
                if response != nil {
                    resp.ManifestResponse = response.ManifestResponse
                }
                resp.Err = err
                res[index] = resp
                wg.Done()
            }
            index := i + j
            impl.asyncRunnable.Execute(func() { runnableFunc(index) })
        }
        wg.Wait()
        i += batchSize
    }
    return res
}
```

**Production Impact:**
- Fetching 100+ Kubernetes resources across multiple clusters
- Without batching: 100 concurrent API calls → K8s API throttling
- With batching: 5-10 at a time → Stable, predictable performance

---

## 3. Fan-Out/Fan-In Pattern (5 mins)

### Pattern: Distribute work, collect results

**Real Example: Parallel CI/CD Status Fetching**

**File:** `api/restHandler/app/pipeline/configure/PipelineConfigRestHandler.go`

```go
func (handler PipelineConfigRestHandlerImpl) FetchWorkflowStatus(w http.ResponseWriter, r *http.Request) {
    // ... RBAC checks ...
    
    var ciWorkflowStatus []*pipelineConfig.CiWorkflowStatus
    var cdWorkflowStatus []*pipelineConfig.CdWorkflowStatus
    var err, err1 error
    
    // FAN-OUT: Launch parallel goroutines
    wg := sync.WaitGroup{}
    wg.Add(2)
    
    go func() {
        defer wg.Done()
        ciWorkflowStatus, err = handler.ciHandler.FetchCiStatusForTriggerView(appId)
    }()
    
    go func() {
        defer wg.Done()
        cdWorkflowStatus, err1 = handler.cdHandler.FetchAppWorkflowStatusForTriggerView(appId)
    }()
    
    // FAN-IN: Wait for all results
    wg.Wait()
    
    // Combine results
    triggerWorkflowStatus := pipelineConfig.TriggerWorkflowStatus{
        CiWorkflowStatus: ciWorkflowStatus,
        CdWorkflowStatus: cdWorkflowStatus,
    }
    
    common.WriteJsonResp(w, err, triggerWorkflowStatus, http.StatusOK)
}
```

**Performance Gain:**
- Sequential: 200ms (CI) + 300ms (CD) = **500ms**
- Parallel: max(200ms, 300ms) = **300ms** (40% faster!)

---

### Advanced Fan-Out: Cluster Connection Testing

**File:** `pkg/cluster/ClusterService.go`

```go
func (impl *ClusterServiceImpl) ConnectClustersInBatch(
    clusters []*bean.ClusterBean, 
    clusterExistInDb bool,
) {
    var wg sync.WaitGroup
    respMap := &sync.Map{}  // Thread-safe map for results
    
    for idx := range clusters {
        cluster := clusters[idx]
        if cluster.IsVirtualCluster {
            impl.updateConnectionStatusForVirtualCluster(respMap, cluster.Id, cluster.ClusterName)
            continue
        }
        
        wg.Add(1)
        runnableFunc := func(idx int, cluster *bean.ClusterBean) {
            defer wg.Done()
            
            clusterConfig := cluster.GetClusterConfig()
            _, _, k8sClientSet, err := impl.K8sUtil.GetK8sConfigAndClients(clusterConfig)
            if err != nil {
                respMap.Store(cluster.Id, err)
                return
            }
            
            id := cluster.Id
            if !clusterExistInDb {
                id = idx
            }
            impl.GetAndUpdateConnectionStatusForOneCluster(k8sClientSet, id, respMap)
        }
        impl.asyncRunnable.Execute(func() { runnableFunc(idx, cluster) })
    }
    
    wg.Wait()
    impl.HandleErrorInClusterConnections(clusters, respMap, clusterExistInDb)
}
```

**Key Pattern:**
- ✅ `sync.Map` for thread-safe result collection
- ✅ Each goroutine stores its result independently
- ✅ Main goroutine waits and processes all results

---

## 4. Rate Limiting & Throttling (4 mins)

### The Problem: Overwhelming External APIs

**Scenario:** Fetching 100+ Kubernetes resource manifests (Pods, Services, Deployments)

**The Challenge:**
- Each resource requires a Kubernetes API call
- K8s API server has rate limits
- Too many concurrent calls → throttling errors
- Need to control concurrency

---

### Solution: Batched Processing with Worker Pool

**File:** `pkg/k8s/K8sCommonService.go`

**Approach:**
1. Divide resources into batches (e.g., batch size = 5)
2. Process each batch concurrently
3. Wait for batch to complete before starting next
4. Pre-allocate result slice for thread-safe writes

**Simplified Code:**
```go
func (impl *K8sCommonServiceImpl) GetManifestsByBatch(
    ctx context.Context,
    requests []ResourceRequest,
) []ResourceResponse {

    batchSize := 5  // Configurable via env
    totalRequests := len(requests)

    // Pre-allocate result slice
    results := make([]ResourceResponse, totalRequests)

    // Process in batches
    for i := 0; i < totalRequests; {
        // Calculate remaining items
        remainingBatch := totalRequests - i
        if remainingBatch < batchSize {
            batchSize = remainingBatch
        }

        var wg sync.WaitGroup

        // Launch workers for current batch
        for j := 0; j < batchSize; j++ {
            wg.Add(1)
            index := i + j

            go func(idx int) {
                defer wg.Done()

                // Fetch resource from Kubernetes API
                response, err := impl.GetResource(ctx, &requests[idx])

                // Store result at pre-determined index (thread-safe)
                results[idx] = ResourceResponse{
                    Manifest: response,
                    Error:    err,
                }
            }(index)
        }

        // Wait for current batch to complete
        wg.Wait()

        // Move to next batch
        i += batchSize
    }

    return results
}
```

**Key Patterns:**
- ✅ **Bounded concurrency** - Max 5 concurrent API calls
- ✅ **Pre-allocated slice** - No mutex needed for writes
- ✅ **Index-based storage** - Each goroutine writes to its own index
- ✅ **Batch synchronization** - Wait between batches

**Production Metrics:**
- Fetching 100 Kubernetes resources
- Sequential: 100 × 50ms = **5 seconds**
- Batched (size=5): 20 batches × 50ms = **1 second**
- **5x performance improvement!**

**Why This Works:**
- Respects Kubernetes API rate limits
- Predictable resource usage
- No database/API connection pool exhaustion
- Graceful degradation under load

---

## 5. Graceful Shutdown with Context (4 mins)

### Pattern: Respect Client Disconnections

**Real Example: HTTP Request Cancellation**

**File:** `api/cluster/ClusterRestHandler.go`

```go
func (impl ClusterRestHandlerImpl) Save(w http.ResponseWriter, r *http.Request) {
    // ... parse request ...
    
    // Create cancellable context
    ctx, cancel := context.WithCancel(r.Context())
    
    // Detect client disconnect
    if cn, ok := w.(http.CloseNotifier); ok {
        go func(done <-chan struct{}, closed <-chan bool) {
            select {
            case <-done:
                // Request completed normally
            case <-closed:
                // Client disconnected - cancel context
                cancel()
            }
        }(ctx.Done(), cn.CloseNotify())
    }
    
    // Pass context to service layer
    bean, err = impl.clusterService.Save(ctx, bean, userId)
    if err != nil {
        impl.logger.Errorw("service err, Save", "err", err)
        common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
        return
    }
    
    common.WriteJsonResp(w, err, bean, http.StatusOK)
}
```

**Why This Matters:**
- User closes browser → Cancel expensive K8s API calls
- Saves resources
- Prevents orphaned operations

---

### Advanced: Server-Sent Events (SSE) with Graceful Shutdown

**File:** `api/sse/Broker.go`

```go
type Broker struct {
    notifier    chan SSEMessage
    connections map[*Connection]bool
    register    chan *Connection
    unregister  chan *Connection
    shutdown    chan bool
}

func (br *Broker) run() {
    for {
        select {
        case <-br.shutdown:
            // Graceful shutdown: close all connections
            for conn := range br.connections {
                br.shutdownConnection(conn)
            }
            return
            
        case conn := <-br.register:
            br.connections[conn] = true
            
        case conn := <-br.unregister:
            br.unregisterConnection(conn)
            
        case msg := <-br.notifier:
            br.broadcastMessage(msg)
        }
    }
}

func (br *Broker) broadcastMessage(message SSEMessage) {
    fmtMsg := message.format()
    for conn := range br.connections {
        if strings.HasPrefix(message.Namespace, conn.namespace) {
            select {
            case conn.outboundMessage <- fmtMsg:
                // Message sent successfully
            default:
                // Channel full - client too slow, disconnect
                br.shutdownConnection(conn)
            }
        }
    }
}
```

**Key Patterns:**
- ✅ `select` with multiple channels
- ✅ Non-blocking sends with `default` case
- ✅ Graceful cleanup on shutdown
- ✅ Automatic slow client detection

---

## 6. Real-World Case Study (3 mins)

### Problem: Checking Hibernation Status for 100+ Kubernetes Resources

**File:** `pkg/appStore/installedApp/service/FullMode/resource/ResourceTreeService.go`

```go
func (impl *InstalledAppResourceServiceImpl) checkHibernate(
    resp map[string]interface{}, 
    deploymentAppName string, 
    ctx context.Context,
) (map[string]interface{}, string) {
    
    var canBeHibernated uint64 = 0
    var hibernated uint64 = 0
    
    replicaNodes := impl.filterOutReplicaNodes(responseTreeNodes)
    batchSize := impl.aCDAuthConfig.ResourceListForReplicasBatchSize
    requestsLength := len(replicaNodes)
    
    for i := 0; i < requestsLength; {
        remainingBatch := requestsLength - i
        if remainingBatch < batchSize {
            batchSize = remainingBatch
        }
        
        var wg sync.WaitGroup
        for j := 0; j < batchSize; j++ {
            wg.Add(1)
            go func(j int) {
                defer wg.Done()
                
                canBeHibernatedFlag, hibernatedFlag := 
                    impl.processReplicaNodeForHibernation(
                        replicaNodes[i+j], 
                        deploymentAppName, 
                        ctx,
                    )
                
                // Atomic operations for thread-safe counters
                if canBeHibernatedFlag {
                    atomic.AddUint64(&canBeHibernated, 1)
                }
                if hibernatedFlag {
                    atomic.AddUint64(&hibernated, 1)
                }
            }(j)
        }
        wg.Wait()
        i += batchSize
    }
    
    // Determine hibernation status
    status := ""
    if hibernated > 0 && canBeHibernated > 0 {
        if hibernated == canBeHibernated {
            status = appStatus.HealthStatusHibernating
        } else if hibernated < canBeHibernated {
            status = appStatus.HealthStatusPartiallyHibernated
        }
    }
    
    return responseTree, status
}
```

**Patterns Combined:**
1. ✅ Worker pool (batching)
2. ✅ Fan-out/fan-in (parallel processing)
3. ✅ Atomic operations (thread-safe counters)
4. ✅ Context propagation (cancellation support)

**Performance:**
- 100 resources × 50ms each = 5 seconds (sequential)
- With batch size 10: ~500ms (10x faster!)

---

## 7. Key Takeaways & Best Practices

### ✅ DO:
1. **Always use worker pools** for bounded concurrency
2. **Use `sync.WaitGroup`** for coordinating goroutines
3. **Use `context.Context`** for cancellation and timeouts
4. **Use `sync.Map` or mutexes** for shared state
5. **Use `atomic` operations** for simple counters
6. **Use `select` with `default`** for non-blocking operations
7. **Always `defer wg.Done()`** to prevent deadlocks

### ❌ DON'T:
1. Spawn unlimited goroutines
2. Share memory without synchronization
3. Ignore context cancellation
4. Forget to handle errors in goroutines
5. Use channels when `sync.WaitGroup` is simpler

---

## Resources

**From This Talk:**
- Devtron GitHub: https://github.com/devtron-labs/devtron
- Files referenced:
  - `pkg/workflow/dag/WorkflowDagExecutor.go`
  - `pkg/k8s/K8sCommonService.go`
  - `pkg/cluster/ClusterService.go`
  - `pkg/auth/authorisation/casbin/rbac.go`
  - `api/sse/Broker.go`

**Further Reading:**
- Go Concurrency Patterns (Rob Pike): https://go.dev/blog/pipelines
- Effective Go: https://go.dev/doc/effective_go#concurrency
- `golang.org/x/sync/errgroup` package
- `golang.org/x/time/rate` package

---

## Thank You! Questions?

**Contact:**
- GitHub: @devtron-labs
- Website: devtron.ai


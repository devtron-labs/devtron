# API Spec Validation Report

Generated: 2025-08-04T19:24:02+05:30

## Summary

- Total Endpoints: 246
- Passed: 33
- Failed: 213
- Warnings: 0
- Success Rate: 13.41%

## Detailed Results

### ❌ GET /orchestrator/app/helm/meta/info/{appId}

- **Status**: FAIL
- **Duration**: 671.611583ms
- **Spec File**: ../../specs/application/labels.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/app/labels/list

- **Status**: PASS
- **Duration**: 129.500834ms
- **Spec File**: ../../specs/application/labels.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/app/meta/info/{appId}

- **Status**: PASS
- **Duration**: 110.831625ms
- **Spec File**: ../../specs/application/labels.yaml
- **Response Code**: 200

---

### ❌ POST /app/workflow/trigger/{pipelineId}

- **Status**: FAIL
- **Duration**: 101.527291ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /app/workflow/{workflowId}

- **Status**: FAIL
- **Duration**: 103.727208ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ GET /app/commit-info/{ciPipelineMaterialId}/{gitHash}

- **Status**: FAIL
- **Duration**: 100.845ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ POST /orchestrator/deployment/template/validate

- **Status**: FAIL
- **Duration**: 106.865ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/deployment/template/fetch

- **Status**: PASS
- **Duration**: 210.260583ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Response Code**: 200

---

### ✅ PUT /orchestrator/deployment/template/upload

- **Status**: PASS
- **Duration**: 109.712875ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/app/deployment-status/timeline/{appId}/{envId}

- **Status**: FAIL
- **Duration**: 109.492708ms
- **Spec File**: ../../specs/deployment/timeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/app/deployment/template/data

- **Status**: FAIL
- **Duration**: 104.927834ms
- **Spec File**: ../../specs/gitops/manifest-generation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/deployments/{app-id}/{env-id}

- **Status**: FAIL
- **Duration**: 104.549ms
- **Spec File**: ../../specs/gitops/manifest-generation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PATCH /orchestrator/app/env/patch

- **Status**: FAIL
- **Duration**: 115.206166ms
- **Spec File**: ../../specs/helm/deployment-chart-type.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /operate

- **Status**: FAIL
- **Duration**: 120.036375ms
- **Spec File**: ../../specs/jobs/batch.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /bulk/v1beta1/application

- **Status**: FAIL
- **Duration**: 122.452333ms
- **Spec File**: ../../specs/jobs/batch.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /bulk/v1beta1/application/dryrun

- **Status**: FAIL
- **Duration**: 185.56575ms
- **Spec File**: ../../specs/jobs/batch.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/cluster

- **Status**: PASS
- **Duration**: 117.79275ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Response Code**: 200

---

### ❌ PUT /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 107.881291ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/cluster/auth-list

- **Status**: PASS
- **Duration**: 105.699208ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/notification/channel/delete

- **Status**: FAIL
- **Duration**: 107.785583ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/team/autocomplete

- **Status**: PASS
- **Duration**: 121.649833ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/git/provider/delete

- **Status**: FAIL
- **Duration**: 108.987875ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/cluster/delete

- **Status**: FAIL
- **Duration**: 102.163666ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/env/delete

- **Status**: FAIL
- **Duration**: 119.801792ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/team

- **Status**: PASS
- **Duration**: 118.564291ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/team

- **Status**: FAIL
- **Duration**: 107.871208ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/team

- **Status**: FAIL
- **Duration**: 120.860333ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/team

- **Status**: FAIL
- **Duration**: 112.5205ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/team/{id}

- **Status**: PASS
- **Duration**: 107.257208ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/app-store/repo/delete

- **Status**: FAIL
- **Duration**: 106.84225ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/chart-group/delete

- **Status**: FAIL
- **Duration**: 119.102834ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/docker/registry/delete

- **Status**: FAIL
- **Duration**: 100.334459ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/app/material/delete

- **Status**: FAIL
- **Duration**: 101.133542ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/config/data

- **Status**: FAIL
- **Duration**: 106.667875ms
- **Spec File**: ../../specs/environment/config-diff.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/config/manifest

- **Status**: FAIL
- **Duration**: 118.924291ms
- **Spec File**: ../../specs/environment/config-diff.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/config/autocomplete

- **Status**: PASS
- **Duration**: 131.7765ms
- **Spec File**: ../../specs/environment/config-diff.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/config/compare/{resource}

- **Status**: FAIL
- **Duration**: 106.661334ms
- **Spec File**: ../../specs/environment/config-diff.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/k8s/pod/exec/session/{identifier}/{namespace}/{pod}/{shell}/{container}

- **Status**: FAIL
- **Duration**: 104.931208ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/k8s/pods/logs/{podName}

- **Status**: FAIL
- **Duration**: 110.894542ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/k8s/resource

- **Status**: FAIL
- **Duration**: 105.546875ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/k8s/resource

- **Status**: FAIL
- **Duration**: 113.0795ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/k8s/resource/create

- **Status**: FAIL
- **Duration**: 109.809208ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/k8s/resource/delete

- **Status**: FAIL
- **Duration**: 118.069083ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/k8s/resource/list

- **Status**: FAIL
- **Duration**: 126.505791ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/k8s/api-resources/{clusterId}

- **Status**: PASS
- **Duration**: 338.601708ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/k8s/events

- **Status**: FAIL
- **Duration**: 117.195ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/k8s/capacity/node/list

- **Status**: FAIL
- **Duration**: 118.981375ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/k8s/capacity/node/taints/edit

- **Status**: FAIL
- **Duration**: 114.664209ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/k8s/capacity/cluster/list

- **Status**: PASS
- **Duration**: 149.1185ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/k8s/capacity/cluster/list/raw

- **Status**: PASS
- **Duration**: 116.316459ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/k8s/capacity/cluster/{clusterId}

- **Status**: PASS
- **Duration**: 323.578291ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Response Code**: 200

---

### ❌ DELETE /orchestrator/k8s/capacity/node

- **Status**: FAIL
- **Duration**: 112.622458ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/k8s/capacity/node

- **Status**: FAIL
- **Duration**: 103.795834ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/k8s/capacity/node

- **Status**: FAIL
- **Duration**: 107.837375ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /orchestrator/k8s/capacity/node/cordon

- **Status**: FAIL
- **Duration**: 109.539709ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /orchestrator/k8s/capacity/node/drain

- **Status**: FAIL
- **Duration**: 120.633041ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 123.415834ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 174.022209ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 106.316417ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/cluster

- **Status**: PASS
- **Duration**: 110.05425ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/cluster/auth-list

- **Status**: PASS
- **Duration**: 107.826875ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/cluster/namespaces

- **Status**: PASS
- **Duration**: 109.248209ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/cluster/namespaces/{clusterId}

- **Status**: PASS
- **Duration**: 116.289458ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/cluster/saveClusters

- **Status**: FAIL
- **Duration**: 118.291875ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/cluster/validate

- **Status**: FAIL
- **Duration**: 119.289709ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PUT /orchestrator/resource/update

- **Status**: FAIL
- **Duration**: 114.535417ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/resource/urls

- **Status**: FAIL
- **Duration**: 114.752584ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/api-resources/{clusterId}

- **Status**: FAIL
- **Duration**: 104.394208ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/resource

- **Status**: FAIL
- **Duration**: 105.038833ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/resource/create

- **Status**: FAIL
- **Duration**: 109.794541ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/resource/delete

- **Status**: FAIL
- **Duration**: 103.944833ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/resource/inception/info

- **Status**: FAIL
- **Duration**: 104.234417ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/application/rollback

- **Status**: FAIL
- **Duration**: 104.425417ms
- **Spec File**: ../../specs/openapiClient/api/openapi.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrator/application/template-chart

- **Status**: FAIL
- **Duration**: 107.101375ms
- **Spec File**: ../../specs/openapiClient/api/openapi.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrator/user

- **Status**: FAIL
- **Duration**: 106.182584ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/user

- **Status**: FAIL
- **Duration**: 109.703333ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/user/bulk

- **Status**: FAIL
- **Duration**: 111.718042ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/user/v2

- **Status**: PASS
- **Duration**: 112.468458ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 200

---

### ❌ DELETE /orchestrator/user/{id}

- **Status**: FAIL
- **Duration**: 105.158875ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/{id}

- **Status**: PASS
- **Duration**: 115.233375ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 200

---

### ❌ GET /app-store/discover

- **Status**: FAIL
- **Duration**: 125.219542ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /app-store/discover/application/chartInfo/{appStoreApplicationVersionId}

- **Status**: FAIL
- **Duration**: 118.16475ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /app-store/discover/application/{appStoreId}/version/autocomplete

- **Status**: FAIL
- **Duration**: 109.422542ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /app-store/discover/application/{id}

- **Status**: FAIL
- **Duration**: 111.071958ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /app-store/discover/search

- **Status**: FAIL
- **Duration**: 106.422791ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/flux-application

- **Status**: FAIL
- **Duration**: 119.970083ms
- **Spec File**: ../../specs/gitops/fluxcd.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/flux-application/app

- **Status**: FAIL
- **Duration**: 110.69025ms
- **Spec File**: ../../specs/gitops/fluxcd.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/chart-repo/sync

- **Status**: FAIL
- **Duration**: 114.135375ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/chart-repo/validate

- **Status**: FAIL
- **Duration**: 110.645ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/chart-repo/{id}

- **Status**: PASS
- **Duration**: 117.595542ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/chart-group

- **Status**: FAIL
- **Duration**: 225.259ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 405

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 405

---

### ❌ PUT /orchestrator/chart-group

- **Status**: FAIL
- **Duration**: 209.57425ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 405

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 405

---

### ❌ DELETE /orchestrator/chart-group/{id}

- **Status**: FAIL
- **Duration**: 106.72075ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/chart-repo

- **Status**: FAIL
- **Duration**: 104.418917ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 405

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 405

---

### ❌ PUT /orchestrator/chart-repo

- **Status**: FAIL
- **Duration**: 99.884125ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 405

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 405

---

### ✅ GET /orchestrator/chart-repo/list

- **Status**: PASS
- **Duration**: 109.974958ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/app-store/chart-provider/sync-chart

- **Status**: FAIL
- **Duration**: 126.325708ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/chart-group/list

- **Status**: PASS
- **Duration**: 112.198084ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/app-store/chart-provider/list

- **Status**: PASS
- **Duration**: 132.501083ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/app-store/chart-provider/update

- **Status**: FAIL
- **Duration**: 105.167792ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/chart-group/entries

- **Status**: FAIL
- **Duration**: 107.774333ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/history/deployed-component/detail/{appId}/{pipelineId}/{id}

- **Status**: FAIL
- **Duration**: 104.396292ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/history/deployed-component/list/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 110.413209ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/history/deployed-configuration/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 137.302375ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Response Code**: 200

---

### ❌ PATCH /orchestrator/app/ci-pipeline/patch-source

- **Status**: FAIL
- **Duration**: 133.864833ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-change-source.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /deployment/pipeline/trigger

- **Status**: FAIL
- **Duration**: 104.541458ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /deployment/pipeline/configure

- **Status**: FAIL
- **Duration**: 99.8685ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /deployment/pipeline/history

- **Status**: FAIL
- **Duration**: 127.164417ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /deployment/pipeline/rollback

- **Status**: FAIL
- **Duration**: 120.067042ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/deployment-configuration/all/latest/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 108.324666ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/history/deployed-configuration/all/latest/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 106.783542ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/history/deployed-configuration/all/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 135.511084ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/app/template/default/{appId}/{chartRefId}

- **Status**: FAIL
- **Duration**: 116.073792ms
- **Spec File**: ../../specs/environment/templates.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/git/host

- **Status**: FAIL
- **Duration**: 118.074042ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/git/host

- **Status**: PASS
- **Duration**: 108.28425ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/git/host/event/{eventId}

- **Status**: FAIL
- **Duration**: 109.012042ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/webhook-meta-config/{gitProviderId}

- **Status**: FAIL
- **Duration**: 108.994333ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/{id}

- **Status**: FAIL
- **Duration**: 115.873042ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/{id}/event

- **Status**: FAIL
- **Duration**: 104.917083ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/app-store/installed-app/notes/{installed-app-id}/{env-id}

- **Status**: FAIL
- **Duration**: 101.898834ms
- **Spec File**: ../../specs/helm/charts.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/app-store/installed-app

- **Status**: PASS
- **Duration**: 129.354ms
- **Spec File**: ../../specs/helm/charts.yaml
- **Response Code**: 200

---

### ❌ POST /create

- **Status**: FAIL
- **Duration**: 104.570917ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /update

- **Status**: FAIL
- **Duration**: 111.226208ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /validate

- **Status**: FAIL
- **Duration**: 124.962125ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/ci-pipeline/{ciPipelineId}/linked-ci/downstream/cd

- **Status**: FAIL
- **Duration**: 107.881ms
- **Spec File**: ../../specs/ci-pipeline/ciPipelineDownstream/downstream-linked-ci-view-spec.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/ci-pipeline/{ciPipelineId}/linked-ci/downstream/env

- **Status**: FAIL
- **Duration**: 104.963459ms
- **Spec File**: ../../specs/ci-pipeline/ciPipelineDownstream/downstream-linked-ci-view-spec.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrtor/batch/v1beta1/cd-pipeline

- **Status**: FAIL
- **Duration**: 104.473ms
- **Spec File**: ../../specs/environment/bulk-delete.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /job

- **Status**: FAIL
- **Duration**: 109.277875ms
- **Spec File**: ../../specs/jobs/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /job/ci-pipeline/list/{jobId}

- **Status**: FAIL
- **Duration**: 107.945167ms
- **Spec File**: ../../specs/jobs/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /job/list

- **Status**: FAIL
- **Duration**: 123.231042ms
- **Spec File**: ../../specs/jobs/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/k8s/resources/ephemeralContainers

- **Status**: FAIL
- **Duration**: 104.018375ms
- **Spec File**: ../../specs/kubernetes/ephemeral-containers.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/k8s/resources/ephemeralContainers

- **Status**: FAIL
- **Duration**: 108.044542ms
- **Spec File**: ../../specs/kubernetes/ephemeral-containers.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ DELETE /orchestrator/notification

- **Status**: PASS
- **Duration**: 119.6485ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/notification

- **Status**: FAIL
- **Duration**: 103.38025ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/notification

- **Status**: FAIL
- **Duration**: 126.6695ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/notification

- **Status**: FAIL
- **Duration**: 113.721125ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/notification/channel

- **Status**: PASS
- **Duration**: 129.965375ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/notification/channel

- **Status**: FAIL
- **Duration**: 115.7915ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200

**Issues:**
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: unexpected end of JSON input

---

### ❌ GET /orchestrator/notification/recipient

- **Status**: FAIL
- **Duration**: 109.863542ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/api-token

- **Status**: PASS
- **Duration**: 108.622834ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/api-token

- **Status**: FAIL
- **Duration**: 107.949333ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/api-token/webhook

- **Status**: PASS
- **Duration**: 106.518042ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 200

---

### ✅ DELETE /orchestrator/api-token/{id}

- **Status**: PASS
- **Duration**: 241.333917ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 200

---

### ❌ PUT /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 103.204917ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/plugin/global/list/detail/v2

- **Status**: FAIL
- **Duration**: 106.745416ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/plugin/global/list/v2

- **Status**: FAIL
- **Duration**: 100.740125ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/plugin/global/list/v2/min

- **Status**: FAIL
- **Duration**: 110.157584ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/plugin/global/migrate

- **Status**: FAIL
- **Duration**: 109.034209ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/plugin/global/create

- **Status**: FAIL
- **Duration**: 110.331ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/plugin/global/list/global-variable

- **Status**: FAIL
- **Duration**: 112.9355ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/plugin/global/list/tags

- **Status**: FAIL
- **Duration**: 112.501917ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/plugin/global/detail/all

- **Status**: FAIL
- **Duration**: 120.042459ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/plugin/global/detail/{pluginId}

- **Status**: FAIL
- **Duration**: 121.06625ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/api-token

- **Status**: FAIL
- **Duration**: 114.417458ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/api-token

- **Status**: FAIL
- **Duration**: 124.86825ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ DELETE /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 119.904625ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 122.40125ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/user

- **Status**: FAIL
- **Duration**: 117.272ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/user

- **Status**: FAIL
- **Duration**: 110.55575ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ DELETE /orchestrator/user/bulk

- **Status**: FAIL
- **Duration**: 109.750875ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 104.523667ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 115.447334ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ DELETE /orchestrator/user/role/group/bulk

- **Status**: FAIL
- **Duration**: 120.013875ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 119.803208ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/user/v2

- **Status**: FAIL
- **Duration**: 114.852291ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/app/history/deployed-component/detail/{appId}/{pipelineId}/{id}

- **Status**: FAIL
- **Duration**: 104.467375ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/app/history/deployed-component/list/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 115.832583ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/app/history/deployed-configuration/{appId}/{pipelineId}/{wfrId}

- **Status**: FAIL
- **Duration**: 130.704458ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /env/namespace/autocomplete

- **Status**: FAIL
- **Duration**: 107.925833ms
- **Spec File**: ../../specs/environment/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 129.886834ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 100.324333ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 109.5405ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 112.028416ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/external-links/tools

- **Status**: FAIL
- **Duration**: 111.441708ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ DELETE /orchestrator/cluster/access/{id}

- **Status**: FAIL
- **Duration**: 109.01725ms
- **Spec File**: ../../specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ GET /orchestrator/cluster/access/{id}

- **Status**: FAIL
- **Duration**: 123.667209ms
- **Spec File**: ../../specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ PUT /orchestrator/cluster/access

- **Status**: FAIL
- **Duration**: 123.642959ms
- **Spec File**: ../../specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/cluster/access

- **Status**: FAIL
- **Duration**: 106.165791ms
- **Spec File**: ../../specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/cluster/access/list

- **Status**: FAIL
- **Duration**: 103.108292ms
- **Spec File**: ../../specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/webhook/notification/{id}

- **Status**: FAIL
- **Duration**: 100.058625ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/ci/workflow

- **Status**: FAIL
- **Duration**: 106.559709ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/webhook/ext-ci/{externalCiId}

- **Status**: FAIL
- **Duration**: 130.431708ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/webhook/git

- **Status**: FAIL
- **Duration**: 109.832125ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/git/{gitHostId}

- **Status**: FAIL
- **Duration**: 103.585ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/git/{gitHostId}/{secret}

- **Status**: FAIL
- **Duration**: 106.140458ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/notification

- **Status**: FAIL
- **Duration**: 114.409875ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/webhook/notification/variables

- **Status**: FAIL
- **Duration**: 119.60625ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/user/role/group/{id}

- **Status**: FAIL
- **Duration**: 107.404792ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/user/role/group/{id}

- **Status**: FAIL
- **Duration**: 114.756375ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 121.351167ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 106.240708ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 109.738959ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ DELETE /orchestrator/user/role/group/bulk

- **Status**: FAIL
- **Duration**: 101.23925ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/user/role/group/detailed/get

- **Status**: FAIL
- **Duration**: 123.109792ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/user/role/group/search

- **Status**: FAIL
- **Duration**: 116.97925ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 124.769417ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 115.334916ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 119.195875ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/version

- **Status**: FAIL
- **Duration**: 105.329208ms
- **Spec File**: ../../specs/common/version.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/app/cd-pipeline/patch/deployment

- **Status**: FAIL
- **Duration**: 120.292167ms
- **Spec File**: ../../specs/deployment/app-type-change.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 123.564458ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 102.352ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 108.899166ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/gitops/config-by-provider

- **Status**: FAIL
- **Duration**: 123.414125ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/gitops/config/{id}

- **Status**: FAIL
- **Duration**: 111.432583ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/gitops/configured

- **Status**: FAIL
- **Duration**: 109.980166ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/gitops/validate

- **Status**: FAIL
- **Duration**: 120.023042ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /gitops/config

- **Status**: FAIL
- **Duration**: 117.7785ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /gitops/config-by-provider

- **Status**: FAIL
- **Duration**: 101.876583ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ GET /gitops/config/{id}

- **Status**: FAIL
- **Duration**: 114.740417ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ GET /gitops/configured

- **Status**: FAIL
- **Duration**: 124.297459ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /validate

- **Status**: FAIL
- **Duration**: 107.6615ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /config

- **Status**: FAIL
- **Duration**: 107.72925ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /config

- **Status**: FAIL
- **Duration**: 110.300541ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/chartref/autocomplete/{appId}

- **Status**: FAIL
- **Duration**: 229.016ms
- **Spec File**: ../../specs/helm/dynamic-charts.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /v1beta1/deploy

- **Status**: FAIL
- **Duration**: 104.450458ms
- **Spec File**: ../../specs/jobs/bulk-actions.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /v1beta1/hibernate

- **Status**: FAIL
- **Duration**: 113.611333ms
- **Spec File**: ../../specs/jobs/bulk-actions.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /v1beta1/unhibernate

- **Status**: FAIL
- **Duration**: 103.27775ms
- **Spec File**: ../../specs/jobs/bulk-actions.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/module

- **Status**: FAIL
- **Duration**: 119.034542ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/module

- **Status**: FAIL
- **Duration**: 99.553166ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/server

- **Status**: FAIL
- **Duration**: 101.2865ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/server

- **Status**: FAIL
- **Duration**: 125.117333ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /app/list

- **Status**: FAIL
- **Duration**: 107.89725ms
- **Spec File**: ../../specs/application/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /core/v1beta1/application

- **Status**: FAIL
- **Duration**: 122.395958ms
- **Spec File**: ../../specs/application/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /app/details/{appId}

- **Status**: FAIL
- **Duration**: 116.3725ms
- **Spec File**: ../../specs/application/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ POST /app/edit

- **Status**: FAIL
- **Duration**: 104.284375ms
- **Spec File**: ../../specs/application/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ POST /app/edit/projects

- **Status**: FAIL
- **Duration**: 112.713792ms
- **Spec File**: ../../specs/application/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ GET /orchestrator/app/ci-pipeline/{appId}

- **Status**: FAIL
- **Duration**: 105.188625ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/app/wf/all/component-names/{appId}

- **Status**: FAIL
- **Duration**: 124.955834ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/app/ci-pipeline/patch

- **Status**: FAIL
- **Duration**: 109.662917ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/configmap/global

- **Status**: FAIL
- **Duration**: 105.732ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/configmap/global/{appId}

- **Status**: FAIL
- **Duration**: 114.274625ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/configmap/environment/{appId}/{envId}/{id}

- **Status**: FAIL
- **Duration**: 100.476959ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/configmap/global/{appId}/{id}

- **Status**: FAIL
- **Duration**: 105.515083ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/configmap/bulk/patch

- **Status**: FAIL
- **Duration**: 104.6695ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/configmap/environment/edit/{appId}/{envId}/{id}

- **Status**: FAIL
- **Duration**: 113.86225ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/configmap/global/edit/{appId}/{id}

- **Status**: FAIL
- **Duration**: 119.01425ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/configmap/environment

- **Status**: FAIL
- **Duration**: 116.116584ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/configmap/environment/{appId}/{envId}

- **Status**: FAIL
- **Duration**: 112.448708ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/user/v2

- **Status**: FAIL
- **Duration**: 108.73525ms
- **Spec File**: ../../specs/security/policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/user/{id}

- **Status**: FAIL
- **Duration**: 114.293458ms
- **Spec File**: ../../specs/security/policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/user

- **Status**: FAIL
- **Duration**: 115.0715ms
- **Spec File**: ../../specs/security/policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/user

- **Status**: FAIL
- **Duration**: 109.480708ms
- **Spec File**: ../../specs/security/policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/user

- **Status**: FAIL
- **Duration**: 105.927458ms
- **Spec File**: ../../specs/security/policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ DELETE /orchestrator/user/bulk

- **Status**: FAIL
- **Duration**: 154.184792ms
- **Spec File**: ../../specs/security/policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/user/resource/options/{kind}/{version}

- **Status**: FAIL
- **Duration**: 107.894542ms
- **Spec File**: ../../specs/userResource/userResource.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---


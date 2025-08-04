# API Spec Validation Report

Generated: 2025-08-04T17:14:31+05:30

## Summary

- Total Endpoints: 241
- Passed: 0
- Failed: 241
- Warnings: 0
- Success Rate: 0.00%

## Detailed Results

### ❌ POST /orchestrtor/batch/v1beta1/cd-pipeline

- **Status**: FAIL
- **Duration**: 9.757959ms
- **Spec File**: specs/environment/bulk-delete.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PATCH /orchestrator/app/env/patch

- **Status**: FAIL
- **Duration**: 3.831ms
- **Spec File**: specs/helm/deployment-chart-type.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/app/ci-pipeline/patch

- **Status**: FAIL
- **Duration**: 260.333µs
- **Spec File**: specs/infrastructure/docker-build.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/app/ci-pipeline/{appId}

- **Status**: FAIL
- **Duration**: 334.333µs
- **Spec File**: specs/infrastructure/docker-build.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/app/wf/all/component-names/{appId}

- **Status**: FAIL
- **Duration**: 370.5µs
- **Spec File**: specs/infrastructure/docker-build.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/cluster/access/{id}

- **Status**: FAIL
- **Duration**: 427.875µs
- **Spec File**: specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ DELETE /orchestrator/cluster/access/{id}

- **Status**: FAIL
- **Duration**: 183.708µs
- **Spec File**: specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ POST /orchestrator/cluster/access

- **Status**: FAIL
- **Duration**: 79.208µs
- **Spec File**: specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /orchestrator/cluster/access

- **Status**: FAIL
- **Duration**: 88.084µs
- **Spec File**: specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/cluster/access/list

- **Status**: FAIL
- **Duration**: 128.5µs
- **Spec File**: specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/k8s/resource/create

- **Status**: FAIL
- **Duration**: 161.416µs
- **Spec File**: specs/kubernetes/apis.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/k8s/resource/delete

- **Status**: FAIL
- **Duration**: 124.916µs
- **Spec File**: specs/kubernetes/apis.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/k8s/resource/list

- **Status**: FAIL
- **Duration**: 317.583µs
- **Spec File**: specs/kubernetes/apis.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/k8s/api-resources/{clusterId}

- **Status**: FAIL
- **Duration**: 259.25µs
- **Spec File**: specs/kubernetes/apis.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/k8s/events

- **Status**: FAIL
- **Duration**: 189.625µs
- **Spec File**: specs/kubernetes/apis.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/k8s/pod/exec/session/{identifier}/{namespace}/{pod}/{shell}/{container}

- **Status**: FAIL
- **Duration**: 177.167µs
- **Spec File**: specs/kubernetes/apis.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/k8s/pods/logs/{podName}

- **Status**: FAIL
- **Duration**: 271.5µs
- **Spec File**: specs/kubernetes/apis.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/k8s/resource

- **Status**: FAIL
- **Duration**: 82.208µs
- **Spec File**: specs/kubernetes/apis.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/k8s/resource

- **Status**: FAIL
- **Duration**: 89.708µs
- **Spec File**: specs/kubernetes/apis.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ DELETE /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 95.208µs
- **Spec File**: specs/kubernetes/cluster-management.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 88.334µs
- **Spec File**: specs/kubernetes/cluster-management.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 127.5µs
- **Spec File**: specs/kubernetes/cluster-management.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 87.875µs
- **Spec File**: specs/kubernetes/cluster-management.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/cluster/auth-list

- **Status**: FAIL
- **Duration**: 95.375µs
- **Spec File**: specs/kubernetes/cluster-management.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/cluster/namespaces

- **Status**: FAIL
- **Duration**: 93.958µs
- **Spec File**: specs/kubernetes/cluster-management.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/cluster/namespaces/{clusterId}

- **Status**: FAIL
- **Duration**: 285.458µs
- **Spec File**: specs/kubernetes/cluster-management.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/cluster/saveClusters

- **Status**: FAIL
- **Duration**: 90.916µs
- **Spec File**: specs/kubernetes/cluster-management.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/cluster/validate

- **Status**: FAIL
- **Duration**: 175.708µs
- **Spec File**: specs/kubernetes/cluster-management.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/webhook/git/{gitHostId}/{secret}

- **Status**: FAIL
- **Duration**: 256.484709ms
- **Spec File**: specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/notification

- **Status**: FAIL
- **Duration**: 212.333µs
- **Spec File**: specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/webhook/notification/variables

- **Status**: FAIL
- **Duration**: 132.375µs
- **Spec File**: specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/webhook/notification/{id}

- **Status**: FAIL
- **Duration**: 113.375µs
- **Spec File**: specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/ci/workflow

- **Status**: FAIL
- **Duration**: 121.375µs
- **Spec File**: specs/notifications/webhooks.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/webhook/ext-ci/{externalCiId}

- **Status**: FAIL
- **Duration**: 120.208µs
- **Spec File**: specs/notifications/webhooks.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/webhook/git

- **Status**: FAIL
- **Duration**: 249.635125ms
- **Spec File**: specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/git/{gitHostId}

- **Status**: FAIL
- **Duration**: 216.608083ms
- **Spec File**: specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/k8s/resources/ephemeralContainers

- **Status**: FAIL
- **Duration**: 1.7845ms
- **Spec File**: specs/kubernetes/ephemeral-containers.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/k8s/resources/ephemeralContainers

- **Status**: FAIL
- **Duration**: 284.208µs
- **Spec File**: specs/kubernetes/ephemeral-containers.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 553.417µs
- **Spec File**: specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 228.458µs
- **Spec File**: specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/api-token

- **Status**: FAIL
- **Duration**: 234.417µs
- **Spec File**: specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/api-token

- **Status**: FAIL
- **Duration**: 241.917µs
- **Spec File**: specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/api-token/webhook

- **Status**: FAIL
- **Duration**: 222.125µs
- **Spec File**: specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/user/resource/options/{kind}/{version}

- **Status**: FAIL
- **Duration**: 1.891917ms
- **Spec File**: specs/userResource/userResource.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/labels/list

- **Status**: FAIL
- **Duration**: 494.666µs
- **Spec File**: specs/application/labels.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/app/meta/info/{appId}

- **Status**: FAIL
- **Duration**: 196.041µs
- **Spec File**: specs/application/labels.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/app/helm/meta/info/{appId}

- **Status**: FAIL
- **Duration**: 185.416µs
- **Spec File**: specs/application/labels.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/version

- **Status**: FAIL
- **Duration**: 185.334µs
- **Spec File**: specs/common/version.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/deployment/template/validate

- **Status**: FAIL
- **Duration**: 200µs
- **Spec File**: specs/deployment/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/deployment/template/fetch

- **Status**: FAIL
- **Duration**: 336.25µs
- **Spec File**: specs/deployment/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/deployment/template/upload

- **Status**: FAIL
- **Duration**: 209.292µs
- **Spec File**: specs/deployment/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/git/host/webhook-meta-config/{gitProviderId}

- **Status**: FAIL
- **Duration**: 220.167µs
- **Spec File**: specs/gitops/submodules.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/git/host/{id}

- **Status**: FAIL
- **Duration**: 149.584µs
- **Spec File**: specs/gitops/submodules.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/git/host/{id}/event

- **Status**: FAIL
- **Duration**: 156.042µs
- **Spec File**: specs/gitops/submodules.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/git/host

- **Status**: FAIL
- **Duration**: 145.083µs
- **Spec File**: specs/gitops/submodules.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/git/host

- **Status**: FAIL
- **Duration**: 285.125µs
- **Spec File**: specs/gitops/submodules.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/git/host/event/{eventId}

- **Status**: FAIL
- **Duration**: 235.375µs
- **Spec File**: specs/gitops/submodules.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /validate

- **Status**: FAIL
- **Duration**: 181µs
- **Spec File**: specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /config

- **Status**: FAIL
- **Duration**: 176.167µs
- **Spec File**: specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /config

- **Status**: FAIL
- **Duration**: 212.084µs
- **Spec File**: specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /gitops/config

- **Status**: FAIL
- **Duration**: 166.167µs
- **Spec File**: specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /gitops/config-by-provider

- **Status**: FAIL
- **Duration**: 175.75µs
- **Spec File**: specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ GET /gitops/config/{id}

- **Status**: FAIL
- **Duration**: 174.125µs
- **Spec File**: specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ GET /gitops/configured

- **Status**: FAIL
- **Duration**: 141.5µs
- **Spec File**: specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /orchestrator/application/rollback

- **Status**: FAIL
- **Duration**: 435.042µs
- **Spec File**: specs/openapiClient/api/openapi.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/application/template-chart

- **Status**: FAIL
- **Duration**: 246.709µs
- **Spec File**: specs/openapiClient/api/openapi.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ DELETE /orchestrator/user/role/group/bulk

- **Status**: FAIL
- **Duration**: 259.708µs
- **Spec File**: specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/user/role/group/detailed/get

- **Status**: FAIL
- **Duration**: 138.25µs
- **Spec File**: specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/user/role/group/search

- **Status**: FAIL
- **Duration**: 164.875µs
- **Spec File**: specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 136.458µs
- **Spec File**: specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 185.5µs
- **Spec File**: specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 178.834µs
- **Spec File**: specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ DELETE /orchestrator/user/role/group/{id}

- **Status**: FAIL
- **Duration**: 205.125µs
- **Spec File**: specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/user/role/group/{id}

- **Status**: FAIL
- **Duration**: 146.875µs
- **Spec File**: specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 169.333µs
- **Spec File**: specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 225.75µs
- **Spec File**: specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 129.459µs
- **Spec File**: specs/security/group-policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/app/history/deployed-component/detail/{appId}/{pipelineId}/{id}

- **Status**: FAIL
- **Duration**: 206.208µs
- **Spec File**: specs/audit/definitions.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/app/history/deployed-component/list/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 142.584µs
- **Spec File**: specs/audit/definitions.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/app/history/deployed-configuration/{appId}/{pipelineId}/{wfrId}

- **Status**: FAIL
- **Duration**: 149.917µs
- **Spec File**: specs/audit/definitions.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /deployment/pipeline/trigger

- **Status**: FAIL
- **Duration**: 117.375µs
- **Spec File**: specs/deployment/pipeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /deployment/pipeline/configure

- **Status**: FAIL
- **Duration**: 108.333µs
- **Spec File**: specs/deployment/pipeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /deployment/pipeline/history

- **Status**: FAIL
- **Duration**: 151.375µs
- **Spec File**: specs/deployment/pipeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /deployment/pipeline/rollback

- **Status**: FAIL
- **Duration**: 107.083µs
- **Spec File**: specs/deployment/pipeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/configmap/bulk/patch

- **Status**: FAIL
- **Duration**: 198.833µs
- **Spec File**: specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/configmap/environment

- **Status**: FAIL
- **Duration**: 113.208µs
- **Spec File**: specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/configmap/environment/edit/{appId}/{envId}/{id}

- **Status**: FAIL
- **Duration**: 142.917µs
- **Spec File**: specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/configmap/environment/{appId}/{envId}

- **Status**: FAIL
- **Duration**: 134.208µs
- **Spec File**: specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/configmap/environment/{appId}/{envId}/{id}

- **Status**: FAIL
- **Duration**: 130.584µs
- **Spec File**: specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/configmap/global/edit/{appId}/{id}

- **Status**: FAIL
- **Duration**: 125.791µs
- **Spec File**: specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/configmap/global

- **Status**: FAIL
- **Duration**: 136.25µs
- **Spec File**: specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/configmap/global/{appId}

- **Status**: FAIL
- **Duration**: 132.25µs
- **Spec File**: specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/configmap/global/{appId}/{id}

- **Status**: FAIL
- **Duration**: 127.417µs
- **Spec File**: specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/user/v2

- **Status**: FAIL
- **Duration**: 133.875µs
- **Spec File**: specs/security/user-management.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ DELETE /orchestrator/user/{id}

- **Status**: FAIL
- **Duration**: 122.417µs
- **Spec File**: specs/security/user-management.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/user/{id}

- **Status**: FAIL
- **Duration**: 115.917µs
- **Spec File**: specs/security/user-management.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/user

- **Status**: FAIL
- **Duration**: 117.583µs
- **Spec File**: specs/security/user-management.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/user

- **Status**: FAIL
- **Duration**: 178.084µs
- **Spec File**: specs/security/user-management.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ DELETE /orchestrator/user/bulk

- **Status**: FAIL
- **Duration**: 153.958µs
- **Spec File**: specs/security/user-management.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/app/history/deployed-configuration/all/{appId}/{pipelineId}/{wfrId}

- **Status**: FAIL
- **Duration**: 177.083µs
- **Spec File**: specs/deployment/rollback.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/app/deployment-configuration/all/latest/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 157.791µs
- **Spec File**: specs/deployment/rollback.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/history/deployed-configuration/all/latest/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 163.375µs
- **Spec File**: specs/deployment/rollback.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /app/details/{appId}

- **Status**: FAIL
- **Duration**: 108.709µs
- **Spec File**: specs/application/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ POST /app/edit

- **Status**: FAIL
- **Duration**: 113.75µs
- **Spec File**: specs/application/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ POST /app/edit/projects

- **Status**: FAIL
- **Duration**: 147µs
- **Spec File**: specs/application/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ POST /app/list

- **Status**: FAIL
- **Duration**: 128.166µs
- **Spec File**: specs/application/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /core/v1beta1/application

- **Status**: FAIL
- **Duration**: 165.833µs
- **Spec File**: specs/application/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PATCH /orchestrator/app/ci-pipeline/patch-source

- **Status**: FAIL
- **Duration**: 490.25µs
- **Spec File**: specs/ci-pipeline/ci-pipeline-change-source.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/app/ci-pipeline/{ciPipelineId}/linked-ci/downstream/env

- **Status**: FAIL
- **Duration**: 133.458µs
- **Spec File**: specs/ci-pipeline/ciPipelineDownstream/downstream-linked-ci-view-spec.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/app/ci-pipeline/{ciPipelineId}/linked-ci/downstream/cd

- **Status**: FAIL
- **Duration**: 123.75µs
- **Spec File**: specs/ci-pipeline/ciPipelineDownstream/downstream-linked-ci-view-spec.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/config/manifest

- **Status**: FAIL
- **Duration**: 120.667µs
- **Spec File**: specs/environment/config-diff.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/config/autocomplete

- **Status**: FAIL
- **Duration**: 115.5µs
- **Spec File**: specs/environment/config-diff.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/config/compare/{resource}

- **Status**: FAIL
- **Duration**: 122.834µs
- **Spec File**: specs/environment/config-diff.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/config/data

- **Status**: FAIL
- **Duration**: 119.333µs
- **Spec File**: specs/environment/config-diff.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ DELETE /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 124.458µs
- **Spec File**: specs/external-links/external-links-specs.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 111.166µs
- **Spec File**: specs/external-links/external-links-specs.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 100.333µs
- **Spec File**: specs/external-links/external-links-specs.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 108.458µs
- **Spec File**: specs/external-links/external-links-specs.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/external-links/tools

- **Status**: FAIL
- **Duration**: 98.458µs
- **Spec File**: specs/external-links/external-links-specs.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/app/deployment/template/data

- **Status**: FAIL
- **Duration**: 148.917µs
- **Spec File**: specs/gitops/manifest-generation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/deployments/{app-id}/{env-id}

- **Status**: FAIL
- **Duration**: 146.541µs
- **Spec File**: specs/gitops/manifest-generation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /create

- **Status**: FAIL
- **Duration**: 162.5µs
- **Spec File**: specs/helm/repo-validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /update

- **Status**: FAIL
- **Duration**: 125.25µs
- **Spec File**: specs/helm/repo-validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /validate

- **Status**: FAIL
- **Duration**: 125.5µs
- **Spec File**: specs/helm/repo-validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /env/namespace/autocomplete

- **Status**: FAIL
- **Duration**: 100.208µs
- **Spec File**: specs/environment/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /operate

- **Status**: FAIL
- **Duration**: 104.5µs
- **Spec File**: specs/jobs/batch.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /bulk/v1beta1/application

- **Status**: FAIL
- **Duration**: 159.541µs
- **Spec File**: specs/jobs/batch.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /bulk/v1beta1/application/dryrun

- **Status**: FAIL
- **Duration**: 110.667µs
- **Spec File**: specs/jobs/batch.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /job

- **Status**: FAIL
- **Duration**: 94.625µs
- **Spec File**: specs/jobs/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /job/ci-pipeline/list/{jobId}

- **Status**: FAIL
- **Duration**: 146.959µs
- **Spec File**: specs/jobs/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /job/list

- **Status**: FAIL
- **Duration**: 106.417µs
- **Spec File**: specs/jobs/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/k8s/capacity/cluster/{clusterId}

- **Status**: FAIL
- **Duration**: 165.208µs
- **Spec File**: specs/kubernetes/capacity.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/k8s/capacity/node

- **Status**: FAIL
- **Duration**: 367.416µs
- **Spec File**: specs/kubernetes/capacity.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/k8s/capacity/node

- **Status**: FAIL
- **Duration**: 110.875µs
- **Spec File**: specs/kubernetes/capacity.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ DELETE /orchestrator/k8s/capacity/node

- **Status**: FAIL
- **Duration**: 121.084µs
- **Spec File**: specs/kubernetes/capacity.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/k8s/capacity/node/cordon

- **Status**: FAIL
- **Duration**: 122µs
- **Spec File**: specs/kubernetes/capacity.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/k8s/capacity/node/drain

- **Status**: FAIL
- **Duration**: 157.208µs
- **Spec File**: specs/kubernetes/capacity.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/k8s/capacity/node/list

- **Status**: FAIL
- **Duration**: 131.417µs
- **Spec File**: specs/kubernetes/capacity.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/k8s/capacity/node/taints/edit

- **Status**: FAIL
- **Duration**: 114.416µs
- **Spec File**: specs/kubernetes/capacity.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/k8s/capacity/cluster/list

- **Status**: FAIL
- **Duration**: 113.292µs
- **Spec File**: specs/kubernetes/capacity.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/k8s/capacity/cluster/list/raw

- **Status**: FAIL
- **Duration**: 114µs
- **Spec File**: specs/kubernetes/capacity.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 110.25µs
- **Spec File**: specs/kubernetes/cluster.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 100.292µs
- **Spec File**: specs/kubernetes/cluster.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/cluster/auth-list

- **Status**: FAIL
- **Duration**: 110.583µs
- **Spec File**: specs/kubernetes/cluster.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/resource

- **Status**: FAIL
- **Duration**: 114.209µs
- **Spec File**: specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/resource/create

- **Status**: FAIL
- **Duration**: 110.167µs
- **Spec File**: specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/resource/delete

- **Status**: FAIL
- **Duration**: 108.792µs
- **Spec File**: specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/resource/inception/info

- **Status**: FAIL
- **Duration**: 115.25µs
- **Spec File**: specs/kubernetes/resources.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/resource/update

- **Status**: FAIL
- **Duration**: 100.125µs
- **Spec File**: specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/resource/urls

- **Status**: FAIL
- **Duration**: 128.708µs
- **Spec File**: specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/api-resources/{clusterId}

- **Status**: FAIL
- **Duration**: 116.125µs
- **Spec File**: specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/module

- **Status**: FAIL
- **Duration**: 85.5µs
- **Spec File**: specs/modularisation/v1.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/module

- **Status**: FAIL
- **Duration**: 83.791µs
- **Spec File**: specs/modularisation/v1.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/server

- **Status**: FAIL
- **Duration**: 99.584µs
- **Spec File**: specs/modularisation/v1.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/server

- **Status**: FAIL
- **Duration**: 106.958µs
- **Spec File**: specs/modularisation/v1.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/plugin/global/list/global-variable

- **Status**: FAIL
- **Duration**: 134.875µs
- **Spec File**: specs/plugins/global.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/plugin/global/list/tags

- **Status**: FAIL
- **Duration**: 138.208µs
- **Spec File**: specs/plugins/global.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/plugin/global/list/detail/v2

- **Status**: FAIL
- **Duration**: 238.125µs
- **Spec File**: specs/plugins/global.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/plugin/global/migrate

- **Status**: FAIL
- **Duration**: 149.875µs
- **Spec File**: specs/plugins/global.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/plugin/global/create

- **Status**: FAIL
- **Duration**: 95.625µs
- **Spec File**: specs/plugins/global.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/plugin/global/detail/{pluginId}

- **Status**: FAIL
- **Duration**: 112.75µs
- **Spec File**: specs/plugins/global.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/plugin/global/list/v2

- **Status**: FAIL
- **Duration**: 123µs
- **Spec File**: specs/plugins/global.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/plugin/global/list/v2/min

- **Status**: FAIL
- **Duration**: 108.667µs
- **Spec File**: specs/plugins/global.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/plugin/global/detail/all

- **Status**: FAIL
- **Duration**: 225.125µs
- **Spec File**: specs/plugins/global.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/chartref/autocomplete/{appId}

- **Status**: FAIL
- **Duration**: 116.083µs
- **Spec File**: specs/helm/dynamic-charts.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/app/cd-pipeline/patch/deployment

- **Status**: FAIL
- **Duration**: 161.167µs
- **Spec File**: specs/deployment/app-type-change.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/app/template/default/{appId}/{chartRefId}

- **Status**: FAIL
- **Duration**: 149.417µs
- **Spec File**: specs/environment/templates.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/chart-group

- **Status**: FAIL
- **Duration**: 122.167µs
- **Spec File**: specs/helm/provider.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/chart-group

- **Status**: FAIL
- **Duration**: 121.875µs
- **Spec File**: specs/helm/provider.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/app-store/chart-provider/sync-chart

- **Status**: FAIL
- **Duration**: 284.625µs
- **Spec File**: specs/helm/provider.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/app-store/chart-provider/update

- **Status**: FAIL
- **Duration**: 129.25µs
- **Spec File**: specs/helm/provider.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/chart-group/entries

- **Status**: FAIL
- **Duration**: 148.584µs
- **Spec File**: specs/helm/provider.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/chart-repo

- **Status**: FAIL
- **Duration**: 90.542µs
- **Spec File**: specs/helm/provider.yaml
- **Response Code**: 405

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 405

---

### ❌ PUT /orchestrator/chart-repo

- **Status**: FAIL
- **Duration**: 84.584µs
- **Spec File**: specs/helm/provider.yaml
- **Response Code**: 405

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 405

---

### ❌ POST /orchestrator/chart-repo/sync

- **Status**: FAIL
- **Duration**: 91.042µs
- **Spec File**: specs/helm/provider.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/chart-repo/validate

- **Status**: FAIL
- **Duration**: 119.208µs
- **Spec File**: specs/helm/provider.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/app-store/chart-provider/list

- **Status**: FAIL
- **Duration**: 141.459µs
- **Spec File**: specs/helm/provider.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ DELETE /orchestrator/chart-group/{id}

- **Status**: FAIL
- **Duration**: 93.667µs
- **Spec File**: specs/helm/provider.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/chart-repo/{id}

- **Status**: FAIL
- **Duration**: 117.125µs
- **Spec File**: specs/helm/provider.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/chart-group/list

- **Status**: FAIL
- **Duration**: 83.959µs
- **Spec File**: specs/helm/provider.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/chart-repo/list

- **Status**: FAIL
- **Duration**: 94.458µs
- **Spec File**: specs/helm/provider.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/notification

- **Status**: FAIL
- **Duration**: 124.5µs
- **Spec File**: specs/notifications/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/notification

- **Status**: FAIL
- **Duration**: 104.125µs
- **Spec File**: specs/notifications/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/notification

- **Status**: FAIL
- **Duration**: 101.75µs
- **Spec File**: specs/notifications/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ DELETE /orchestrator/notification

- **Status**: FAIL
- **Duration**: 97.709µs
- **Spec File**: specs/notifications/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/notification/channel

- **Status**: FAIL
- **Duration**: 110.917µs
- **Spec File**: specs/notifications/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/notification/channel

- **Status**: FAIL
- **Duration**: 91.041µs
- **Spec File**: specs/notifications/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/notification/recipient

- **Status**: FAIL
- **Duration**: 119.625µs
- **Spec File**: specs/notifications/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 114.166µs
- **Spec File**: specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/user/v2

- **Status**: FAIL
- **Duration**: 95.792µs
- **Spec File**: specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/api-token

- **Status**: FAIL
- **Duration**: 132.458µs
- **Spec File**: specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/api-token

- **Status**: FAIL
- **Duration**: 159.292µs
- **Spec File**: specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ DELETE /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 130.208µs
- **Spec File**: specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 133.792µs
- **Spec File**: specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/user

- **Status**: FAIL
- **Duration**: 226.083µs
- **Spec File**: specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/user

- **Status**: FAIL
- **Duration**: 90.292µs
- **Spec File**: specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ DELETE /orchestrator/user/bulk

- **Status**: FAIL
- **Duration**: 120.208µs
- **Spec File**: specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 87.958µs
- **Spec File**: specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 115.584µs
- **Spec File**: specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ DELETE /orchestrator/user/role/group/bulk

- **Status**: FAIL
- **Duration**: 95.292µs
- **Spec File**: specs/security/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/user/{id}

- **Status**: FAIL
- **Duration**: 83.292µs
- **Spec File**: specs/security/policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/user

- **Status**: FAIL
- **Duration**: 90.792µs
- **Spec File**: specs/security/policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/user

- **Status**: FAIL
- **Duration**: 83.541µs
- **Spec File**: specs/security/policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/user

- **Status**: FAIL
- **Duration**: 81.917µs
- **Spec File**: specs/security/policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ DELETE /orchestrator/user/bulk

- **Status**: FAIL
- **Duration**: 100.167µs
- **Spec File**: specs/security/policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/user/v2

- **Status**: FAIL
- **Duration**: 109.375µs
- **Spec File**: specs/security/policy.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /app-store/discover

- **Status**: FAIL
- **Duration**: 87.417µs
- **Spec File**: specs/app-store.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /app-store/discover/application/chartInfo/{appStoreApplicationVersionId}

- **Status**: FAIL
- **Duration**: 96.75µs
- **Spec File**: specs/app-store.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /app-store/discover/application/{appStoreId}/version/autocomplete

- **Status**: FAIL
- **Duration**: 84.458µs
- **Spec File**: specs/app-store.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /app-store/discover/application/{id}

- **Status**: FAIL
- **Duration**: 68.834µs
- **Spec File**: specs/app-store.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /app-store/discover/search

- **Status**: FAIL
- **Duration**: 78.25µs
- **Spec File**: specs/app-store.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/history/deployed-component/detail/{appId}/{pipelineId}/{id}

- **Status**: FAIL
- **Duration**: 132.125µs
- **Spec File**: specs/audit/api-changes.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/app/history/deployed-component/list/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 98µs
- **Spec File**: specs/audit/api-changes.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/app/history/deployed-configuration/{appId}/{pipelineId}/{wfrId}

- **Status**: FAIL
- **Duration**: 94.417µs
- **Spec File**: specs/audit/api-changes.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/app/material/delete

- **Status**: FAIL
- **Duration**: 89.416µs
- **Spec File**: specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/docker/registry/delete

- **Status**: FAIL
- **Duration**: 91.583µs
- **Spec File**: specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/env/delete

- **Status**: FAIL
- **Duration**: 100.833µs
- **Spec File**: specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/git/provider/delete

- **Status**: FAIL
- **Duration**: 88.583µs
- **Spec File**: specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/app-store/repo/delete

- **Status**: FAIL
- **Duration**: 161.916µs
- **Spec File**: specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/cluster/delete

- **Status**: FAIL
- **Duration**: 93.291µs
- **Spec File**: specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/team/delete

- **Status**: FAIL
- **Duration**: 84.25µs
- **Spec File**: specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/chart-group/delete

- **Status**: FAIL
- **Duration**: 73.291µs
- **Spec File**: specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/notification/channel/delete

- **Status**: FAIL
- **Duration**: 93.167µs
- **Spec File**: specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/deployment-status/timeline/{appId}/{envId}

- **Status**: FAIL
- **Duration**: 102.5µs
- **Spec File**: specs/deployment/timeline.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/gitops/validate

- **Status**: FAIL
- **Duration**: 121.417µs
- **Spec File**: specs/gitops/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 430.5µs
- **Spec File**: specs/gitops/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 236.542µs
- **Spec File**: specs/gitops/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 236.833µs
- **Spec File**: specs/gitops/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/gitops/config-by-provider

- **Status**: FAIL
- **Duration**: 206.542µs
- **Spec File**: specs/gitops/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/gitops/config/{id}

- **Status**: FAIL
- **Duration**: 779.292µs
- **Spec File**: specs/gitops/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/gitops/configured

- **Status**: FAIL
- **Duration**: 100.416µs
- **Spec File**: specs/gitops/core.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/flux-application

- **Status**: FAIL
- **Duration**: 91.5µs
- **Spec File**: specs/gitops/fluxcd.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/flux-application/app

- **Status**: FAIL
- **Duration**: 95.625µs
- **Spec File**: specs/gitops/fluxcd.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app-store/installed-app

- **Status**: FAIL
- **Duration**: 125.5µs
- **Spec File**: specs/helm/charts.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ GET /orchestrator/app-store/installed-app/notes/{installed-app-id}/{env-id}

- **Status**: FAIL
- **Duration**: 123.333µs
- **Spec File**: specs/helm/charts.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /v1beta1/hibernate

- **Status**: FAIL
- **Duration**: 86.333µs
- **Spec File**: specs/jobs/bulk-actions.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /v1beta1/unhibernate

- **Status**: FAIL
- **Duration**: 75.625µs
- **Spec File**: specs/jobs/bulk-actions.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /v1beta1/deploy

- **Status**: FAIL
- **Duration**: 81.041µs
- **Spec File**: specs/jobs/bulk-actions.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /app/workflow/trigger/{pipelineId}

- **Status**: FAIL
- **Duration**: 92µs
- **Spec File**: specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /app/workflow/{workflowId}

- **Status**: FAIL
- **Duration**: 86.584µs
- **Spec File**: specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ GET /app/commit-info/{ciPipelineMaterialId}/{gitHash}

- **Status**: FAIL
- **Duration**: 89.666µs
- **Spec File**: specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---


# API Spec Validation Report

Generated: 2025-08-05T03:54:12+05:30

## Summary

- Total Endpoints: 178
- Passed: 47
- Failed: 131
- Warnings: 0
- Success Rate: 26.40%

## Detailed Results

### ❌ POST /orchestrator/app/cd-pipeline/patch/deployment

- **Status**: FAIL
- **Duration**: 592.461708ms
- **Spec File**: ../../specs/deployment/app-type-change.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/deployment/template/validate

- **Status**: FAIL
- **Duration**: 112.239416ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/deployment/template/fetch

- **Status**: PASS
- **Duration**: 125.824791ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Response Code**: 200

---

### ✅ PUT /orchestrator/deployment/template/upload

- **Status**: PASS
- **Duration**: 148.201958ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 113.378458ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 144.546333ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 111.005708ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/external-links

- **Status**: PASS
- **Duration**: 122.927333ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/external-links/tools

- **Status**: PASS
- **Duration**: 115.142458ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/app/ci-pipeline/{appId}

- **Status**: PASS
- **Duration**: 121.844959ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/app/wf/all/component-names/{appId}

- **Status**: PASS
- **Duration**: 132.803333ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/app/ci-pipeline/patch

- **Status**: FAIL
- **Duration**: 123.335584ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/cluster/auth-list

- **Status**: PASS
- **Duration**: 116.018084ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/cluster

- **Status**: PASS
- **Duration**: 122.421666ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Response Code**: 200

---

### ❌ PUT /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 117.259834ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/application/rollback

- **Status**: FAIL
- **Duration**: 114.139042ms
- **Spec File**: ../../specs/openapiClient/api/openapi.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrator/application/template-chart

- **Status**: FAIL
- **Duration**: 112.322625ms
- **Spec File**: ../../specs/openapiClient/api/openapi.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrtor/batch/v1beta1/cd-pipeline

- **Status**: FAIL
- **Duration**: 117.474292ms
- **Spec File**: ../../specs/environment/bulk-delete.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/gitops/validate

- **Status**: FAIL
- **Duration**: 113.63525ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/gitops/config

- **Status**: PASS
- **Duration**: 113.7205ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 115.272291ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 134.292375ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/gitops/config-by-provider

- **Status**: FAIL
- **Duration**: 111.126333ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/gitops/config/{id}

- **Status**: PASS
- **Duration**: 117.179167ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/gitops/configured

- **Status**: PASS
- **Duration**: 114.632292ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 200

---

### ❌ PATCH /orchestrator/app/env/patch

- **Status**: FAIL
- **Duration**: 114.003834ms
- **Spec File**: ../../specs/helm/deployment-chart-type.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/k8s/resources/ephemeralContainers

- **Status**: FAIL
- **Duration**: 112.676667ms
- **Spec File**: ../../specs/kubernetes/ephemeral-containers.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/k8s/resources/ephemeralContainers

- **Status**: FAIL
- **Duration**: 110.972959ms
- **Spec File**: ../../specs/kubernetes/ephemeral-containers.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/git

- **Status**: FAIL
- **Duration**: 111.615459ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/git/{gitHostId}

- **Status**: FAIL
- **Duration**: 113.51275ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/git/{gitHostId}/{secret}

- **Status**: FAIL
- **Duration**: 112.676791ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/notification

- **Status**: FAIL
- **Duration**: 112.179375ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/webhook/notification/variables

- **Status**: FAIL
- **Duration**: 112.968833ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/webhook/notification/{id}

- **Status**: FAIL
- **Duration**: 112.124209ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/ci/workflow

- **Status**: FAIL
- **Duration**: 118.022292ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrator/webhook/ext-ci/{externalCiId}

- **Status**: FAIL
- **Duration**: 113.444875ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/user

- **Status**: FAIL
- **Duration**: 113.827209ms
- **Spec File**: ../../specs/security/policy.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user

- **Status**: PASS
- **Duration**: 115.4865ms
- **Spec File**: ../../specs/security/policy.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/user

- **Status**: FAIL
- **Duration**: 114.079167ms
- **Spec File**: ../../specs/security/policy.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/user/bulk

- **Status**: FAIL
- **Duration**: 112.769875ms
- **Spec File**: ../../specs/security/policy.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/user/v2

- **Status**: PASS
- **Duration**: 117.102166ms
- **Spec File**: ../../specs/security/policy.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/user/{id}

- **Status**: PASS
- **Duration**: 116.190167ms
- **Spec File**: ../../specs/security/policy.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/user/resource/options/{kind}/{version}

- **Status**: FAIL
- **Duration**: 112.233584ms
- **Spec File**: ../../specs/userResource/userResource.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/template/default/{appId}/{chartRefId}

- **Status**: FAIL
- **Duration**: 117.922166ms
- **Spec File**: ../../specs/environment/templates.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app-store/installed-app/notes/{installed-app-id}/{env-id}

- **Status**: FAIL
- **Duration**: 112.32275ms
- **Spec File**: ../../specs/helm/charts.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/app-store/installed-app

- **Status**: PASS
- **Duration**: 116.553584ms
- **Spec File**: ../../specs/helm/charts.yaml
- **Response Code**: 200

---

### ❌ GET /job/ci-pipeline/list/{jobId}

- **Status**: FAIL
- **Duration**: 112.10325ms
- **Spec File**: ../../specs/jobs/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /job/list

- **Status**: FAIL
- **Duration**: 111.273042ms
- **Spec File**: ../../specs/jobs/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/job

- **Status**: FAIL
- **Duration**: 118.348583ms
- **Spec File**: ../../specs/jobs/core.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/cluster/access

- **Status**: FAIL
- **Duration**: 111.00925ms
- **Spec File**: ../../specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /orchestrator/cluster/access

- **Status**: FAIL
- **Duration**: 109.616542ms
- **Spec File**: ../../specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/cluster/access/list

- **Status**: FAIL
- **Duration**: 111.602042ms
- **Spec File**: ../../specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/cluster/access/{id}

- **Status**: FAIL
- **Duration**: 111.861042ms
- **Spec File**: ../../specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ GET /orchestrator/cluster/access/{id}

- **Status**: FAIL
- **Duration**: 110.685667ms
- **Spec File**: ../../specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ GET /app-store/discover/application/{appStoreId}/version/autocomplete

- **Status**: FAIL
- **Duration**: 112.451958ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /app-store/discover/application/{id}

- **Status**: FAIL
- **Duration**: 114.302167ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /app-store/discover/search

- **Status**: FAIL
- **Duration**: 112.556625ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/app-store/discover

- **Status**: PASS
- **Duration**: 806.034625ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 200

---

### ❌ GET /app-store/discover/application/chartInfo/{appStoreApplicationVersionId}

- **Status**: FAIL
- **Duration**: 110.322584ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/history/deployed-component/detail/{appId}/{pipelineId}/{id}

- **Status**: FAIL
- **Duration**: 120.983542ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/history/deployed-component/list/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 113.998667ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/history/deployed-configuration/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 118.987084ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/flux-application

- **Status**: FAIL
- **Duration**: 117.885ms
- **Spec File**: ../../specs/gitops/fluxcd.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/flux-application/app

- **Status**: FAIL
- **Duration**: 110.391042ms
- **Spec File**: ../../specs/gitops/fluxcd.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/app-store/chart-provider/update

- **Status**: FAIL
- **Duration**: 115.451166ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/chart-repo/{id}

- **Status**: PASS
- **Duration**: 113.059625ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/app-store/chart-provider/list

- **Status**: PASS
- **Duration**: 119.446042ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200

---

### ❌ DELETE /orchestrator/chart-group/{id}

- **Status**: FAIL
- **Duration**: 109.725292ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/chart-repo/list

- **Status**: PASS
- **Duration**: 152.787417ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/chart-repo/validate

- **Status**: FAIL
- **Duration**: 113.295125ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/app-store/chart-provider/sync-chart

- **Status**: FAIL
- **Duration**: 116.090292ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/chart-group

- **Status**: FAIL
- **Duration**: 224.847ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 405

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 405

---

### ❌ PUT /orchestrator/chart-group

- **Status**: FAIL
- **Duration**: 224.263125ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 405

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 405

---

### ❌ POST /orchestrator/chart-repo/sync

- **Status**: FAIL
- **Duration**: 111.693167ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/chart-group/entries

- **Status**: FAIL
- **Duration**: 110.718584ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/chart-group/list

- **Status**: PASS
- **Duration**: 114.002708ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/chart-repo

- **Status**: FAIL
- **Duration**: 111.598125ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 405

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 405

---

### ❌ PUT /orchestrator/chart-repo

- **Status**: FAIL
- **Duration**: 111.930208ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 405

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 405

---

### ❌ PUT /orchestrator/resource/update

- **Status**: FAIL
- **Duration**: 111.209375ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/resource/urls

- **Status**: FAIL
- **Duration**: 110.179459ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/api-resources/{clusterId}

- **Status**: FAIL
- **Duration**: 112.701333ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/resource

- **Status**: FAIL
- **Duration**: 112.640708ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/resource/create

- **Status**: FAIL
- **Duration**: 111.785ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/resource/delete

- **Status**: FAIL
- **Duration**: 110.440166ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/resource/inception/info

- **Status**: FAIL
- **Duration**: 112.074583ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/user

- **Status**: FAIL
- **Duration**: 117.504708ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/user

- **Status**: FAIL
- **Duration**: 114.06075ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/user/bulk

- **Status**: FAIL
- **Duration**: 115.207125ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/user/v2

- **Status**: PASS
- **Duration**: 116.514166ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 200

---

### ❌ DELETE /orchestrator/user/{id}

- **Status**: FAIL
- **Duration**: 116.741375ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/{id}

- **Status**: PASS
- **Duration**: 116.933042ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/app/ci-pipeline/{ciPipelineId}/linked-ci/downstream/env

- **Status**: FAIL
- **Duration**: 114.435583ms
- **Spec File**: ../../specs/ci-pipeline/ciPipelineDownstream/downstream-linked-ci-view-spec.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/ci-pipeline/{ciPipelineId}/linked-ci/downstream/cd

- **Status**: FAIL
- **Duration**: 167.021542ms
- **Spec File**: ../../specs/ci-pipeline/ciPipelineDownstream/downstream-linked-ci-view-spec.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/notification/recipient

- **Status**: FAIL
- **Duration**: 110.201167ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ DELETE /orchestrator/notification

- **Status**: PASS
- **Duration**: 115.023292ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/notification

- **Status**: FAIL
- **Duration**: 112.179916ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/notification

- **Status**: FAIL
- **Duration**: 113.874083ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/notification

- **Status**: FAIL
- **Duration**: 113.772084ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/notification/channel

- **Status**: PASS
- **Duration**: 115.087167ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/notification/channel

- **Status**: FAIL
- **Duration**: 113.912083ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200

**Issues:**
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: unexpected end of JSON input

---

### ❌ DELETE /orchestrator/configmap/global/{appId}/{id}

- **Status**: FAIL
- **Duration**: 111.968291ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/configmap/global

- **Status**: FAIL
- **Duration**: 110.30525ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/configmap/global/edit/{appId}/{id}

- **Status**: FAIL
- **Duration**: 111.148625ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/configmap/global/{appId}

- **Status**: FAIL
- **Duration**: 112.163875ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/configmap/bulk/patch

- **Status**: FAIL
- **Duration**: 111.880167ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/configmap/environment

- **Status**: FAIL
- **Duration**: 110.15075ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/configmap/environment/{appId}/{envId}

- **Status**: FAIL
- **Duration**: 111.130875ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/configmap/environment/{appId}/{envId}/{id}

- **Status**: FAIL
- **Duration**: 113.93825ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/configmap/environment/edit/{appId}/{envId}/{id}

- **Status**: FAIL
- **Duration**: 110.965792ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/history/deployed-component/detail/{appId}/{pipelineId}/{id}

- **Status**: FAIL
- **Duration**: 115.659708ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/history/deployed-component/list/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 205.195041ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/history/deployed-configuration/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 116.101875ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/env/namespace/autocomplete

- **Status**: FAIL
- **Duration**: 114.127167ms
- **Spec File**: ../../specs/environment/core.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/event/{eventId}

- **Status**: FAIL
- **Duration**: 115.216375ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/webhook-meta-config/{gitProviderId}

- **Status**: FAIL
- **Duration**: 116.038042ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/{id}

- **Status**: FAIL
- **Duration**: 116.05425ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/{id}/event

- **Status**: FAIL
- **Duration**: 115.660708ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/git/host

- **Status**: PASS
- **Duration**: 304.190292ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/git/host

- **Status**: FAIL
- **Duration**: 114.101208ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/chartref/autocomplete/{appId}

- **Status**: PASS
- **Duration**: 117.620542ms
- **Spec File**: ../../specs/helm/dynamic-charts.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/validate

- **Status**: FAIL
- **Duration**: 111.261833ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /update

- **Status**: FAIL
- **Duration**: 110.614625ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /create

- **Status**: FAIL
- **Duration**: 110.755625ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/module

- **Status**: PASS
- **Duration**: 119.152666ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/module

- **Status**: FAIL
- **Duration**: 115.881833ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/server

- **Status**: PASS
- **Duration**: 191.671541ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/server

- **Status**: FAIL
- **Duration**: 117.230167ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/helm/meta/info/{appId}

- **Status**: FAIL
- **Duration**: 116.306917ms
- **Spec File**: ../../specs/application/labels.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/app/labels/list

- **Status**: PASS
- **Duration**: 112.628625ms
- **Spec File**: ../../specs/application/labels.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/app/meta/info/{appId}

- **Status**: PASS
- **Duration**: 119.800959ms
- **Spec File**: ../../specs/application/labels.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/app/deployment-status/timeline/{appId}/{envId}

- **Status**: FAIL
- **Duration**: 115.772833ms
- **Spec File**: ../../specs/deployment/timeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /gitops/config-by-provider

- **Status**: FAIL
- **Duration**: 112.206042ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ GET /gitops/config/{id}

- **Status**: FAIL
- **Duration**: 112.919375ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ GET /gitops/configured

- **Status**: FAIL
- **Duration**: 113.529416ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/validate

- **Status**: FAIL
- **Duration**: 110.182625ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /config

- **Status**: FAIL
- **Duration**: 111.050583ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /config

- **Status**: FAIL
- **Duration**: 114.023834ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /gitops/config

- **Status**: FAIL
- **Duration**: 111.492666ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ PUT /orchestrator/plugin/global/migrate

- **Status**: PASS
- **Duration**: 118.168208ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/plugin/global/list/v2

- **Status**: PASS
- **Duration**: 279.7585ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/plugin/global/create

- **Status**: FAIL
- **Duration**: 142.18425ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/plugin/global/detail/all

- **Status**: PASS
- **Duration**: 929.797083ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/plugin/global/detail/{pluginId}

- **Status**: PASS
- **Duration**: 164.005791ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/plugin/global/list/detail/v2

- **Status**: FAIL
- **Duration**: 113.46175ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/plugin/global/list/tags

- **Status**: PASS
- **Duration**: 115.619667ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/plugin/global/list/v2/min

- **Status**: PASS
- **Duration**: 115.185292ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/plugin/global/list/global-variable

- **Status**: FAIL
- **Duration**: 116.283334ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/user/role/group/bulk

- **Status**: FAIL
- **Duration**: 115.871ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/user/role/group/detailed/get

- **Status**: PASS
- **Duration**: 114.933708ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/user/role/group/search

- **Status**: FAIL
- **Duration**: 113.84825ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 116.088ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/user/role/group/v2

- **Status**: PASS
- **Duration**: 114.557291ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 115.944209ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ DELETE /orchestrator/user/role/group/{id}

- **Status**: PASS
- **Duration**: 118.599459ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/user/role/group/{id}

- **Status**: FAIL
- **Duration**: 114.106ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/user/role/group

- **Status**: PASS
- **Duration**: 121.693917ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 200

---

### ✅ POST /orchestrator/user/role/group

- **Status**: PASS
- **Duration**: 120.096542ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 200

---

### ❌ PUT /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 115.228417ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PATCH /orchestrator/app/ci-pipeline/patch-source

- **Status**: FAIL
- **Duration**: 160.191875ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-change-source.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrator/deployment/pipeline/rollback

- **Status**: FAIL
- **Duration**: 115.220458ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/deployment/pipeline/trigger

- **Status**: FAIL
- **Duration**: 110.1175ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/deployment/pipeline/configure

- **Status**: FAIL
- **Duration**: 112.404166ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/deployment/pipeline/history

- **Status**: FAIL
- **Duration**: 110.430833ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/deployment-configuration/all/latest/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 112.422708ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/history/deployed-configuration/all/latest/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 121.035042ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/history/deployed-configuration/all/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 124.03875ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/v1beta1/hibernate

- **Status**: FAIL
- **Duration**: 110.511416ms
- **Spec File**: ../../specs/jobs/bulk-actions.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /v1beta1/deploy

- **Status**: FAIL
- **Duration**: 114.206042ms
- **Spec File**: ../../specs/jobs/bulk-actions.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /v1beta1/unhibernate

- **Status**: FAIL
- **Duration**: 112.251417ms
- **Spec File**: ../../specs/jobs/bulk-actions.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/api-token

- **Status**: PASS
- **Duration**: 117.086792ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/api-token

- **Status**: FAIL
- **Duration**: 113.402416ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/api-token/webhook

- **Status**: PASS
- **Duration**: 113.790916ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 200

---

### ❌ DELETE /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 115.381833ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PUT /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 114.127917ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/commit-info/{ciPipelineMaterialId}/{gitHash}

- **Status**: FAIL
- **Duration**: 111.071ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ GET /{appId}/ci-pipeline/{pipelineId}/workflow/{workflowId}

- **Status**: FAIL
- **Duration**: 109.867916ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ POST /ci-pipeline/trigger

- **Status**: FAIL
- **Duration**: 111.912375ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/version

- **Status**: PASS
- **Duration**: 113.113083ms
- **Spec File**: ../../specs/common/version.yaml
- **Response Code**: 200

---


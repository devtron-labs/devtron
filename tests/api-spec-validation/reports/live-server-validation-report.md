# API Spec Validation Report

Generated: 2025-08-05T04:50:26+05:30

## Summary

- Total Endpoints: 172
- Passed: 45
- Failed: 127
- Warnings: 0
- Success Rate: 26.16%

## Detailed Results

### ❌ POST /config

- **Status**: FAIL
- **Duration**: 353.701875ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /config

- **Status**: FAIL
- **Duration**: 107.269959ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /gitops/config

- **Status**: FAIL
- **Duration**: 108.011083ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /gitops/config-by-provider

- **Status**: FAIL
- **Duration**: 106.48525ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ GET /gitops/config/{id}

- **Status**: FAIL
- **Duration**: 108.703708ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ GET /gitops/configured

- **Status**: FAIL
- **Duration**: 108.450333ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/validate

- **Status**: FAIL
- **Duration**: 155.382708ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/cluster/access

- **Status**: FAIL
- **Duration**: 205.509125ms
- **Spec File**: ../../specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /orchestrator/cluster/access

- **Status**: FAIL
- **Duration**: 107.53025ms
- **Spec File**: ../../specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/cluster/access/list

- **Status**: FAIL
- **Duration**: 107.685375ms
- **Spec File**: ../../specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/cluster/access/{id}

- **Status**: FAIL
- **Duration**: 105.775041ms
- **Spec File**: ../../specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ GET /orchestrator/cluster/access/{id}

- **Status**: FAIL
- **Duration**: 108.046917ms
- **Spec File**: ../../specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ✅ GET /orchestrator/cluster/auth-list

- **Status**: PASS
- **Duration**: 110.035625ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/cluster

- **Status**: PASS
- **Duration**: 116.716417ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Response Code**: 200

---

### ❌ PUT /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 108.857875ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/k8s/resources/ephemeralContainers

- **Status**: FAIL
- **Duration**: 106.304958ms
- **Spec File**: ../../specs/kubernetes/ephemeral-containers.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/k8s/resources/ephemeralContainers

- **Status**: FAIL
- **Duration**: 107.189042ms
- **Spec File**: ../../specs/kubernetes/ephemeral-containers.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/helm/meta/info/{appId}

- **Status**: FAIL
- **Duration**: 112.853917ms
- **Spec File**: ../../specs/application/labels.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/app/labels/list

- **Status**: PASS
- **Duration**: 112.008083ms
- **Spec File**: ../../specs/application/labels.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/app/meta/info/{appId}

- **Status**: PASS
- **Duration**: 114.030375ms
- **Spec File**: ../../specs/application/labels.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/app/history/deployed-component/detail/{appId}/{pipelineId}/{id}

- **Status**: FAIL
- **Duration**: 112.222917ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/history/deployed-component/list/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 109.726625ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/history/deployed-configuration/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 113.006375ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/version

- **Status**: PASS
- **Duration**: 110.260167ms
- **Spec File**: ../../specs/common/version.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 111.016583ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 113.349416ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 108.321125ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ✅ GET /orchestrator/external-links

- **Status**: PASS
- **Duration**: 111.410833ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/external-links/tools

- **Status**: PASS
- **Duration**: 111.780125ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/flux-application

- **Status**: FAIL
- **Duration**: 107.910583ms
- **Spec File**: ../../specs/gitops/fluxcd.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/flux-application/app

- **Status**: FAIL
- **Duration**: 110.859291ms
- **Spec File**: ../../specs/gitops/fluxcd.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/app-store/installed-app

- **Status**: PASS
- **Duration**: 111.527542ms
- **Spec File**: ../../specs/helm/charts.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/app-store/installed-app/notes/{installed-app-id}/{env-id}

- **Status**: FAIL
- **Duration**: 107.263458ms
- **Spec File**: ../../specs/helm/charts.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PATCH /orchestrator/app/env/patch

- **Status**: FAIL
- **Duration**: 108.838291ms
- **Spec File**: ../../specs/helm/deployment-chart-type.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /job/ci-pipeline/list/{jobId}

- **Status**: FAIL
- **Duration**: 105.015459ms
- **Spec File**: ../../specs/jobs/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /job/list

- **Status**: FAIL
- **Duration**: 106.130125ms
- **Spec File**: ../../specs/jobs/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/job

- **Status**: FAIL
- **Duration**: 109.302417ms
- **Spec File**: ../../specs/jobs/core.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/template/default/{appId}/{chartRefId}

- **Status**: FAIL
- **Duration**: 112.209375ms
- **Spec File**: ../../specs/environment/templates.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/chartref/autocomplete/{appId}

- **Status**: PASS
- **Duration**: 113.8465ms
- **Spec File**: ../../specs/helm/dynamic-charts.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/chart-group/list

- **Status**: PASS
- **Duration**: 110.537167ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/app-store/chart-provider/sync-chart

- **Status**: FAIL
- **Duration**: 108.65575ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/chart-group

- **Status**: FAIL
- **Duration**: 218.447583ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 405

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 405

---

### ❌ PUT /orchestrator/chart-group

- **Status**: FAIL
- **Duration**: 217.249ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 405

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 405

---

### ❌ DELETE /orchestrator/chart-group/{id}

- **Status**: FAIL
- **Duration**: 106.797458ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/app-store/chart-provider/update

- **Status**: FAIL
- **Duration**: 108.443791ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/chart-repo/sync

- **Status**: FAIL
- **Duration**: 106.316916ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/chart-repo/validate

- **Status**: FAIL
- **Duration**: 109.972375ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/chart-repo/{id}

- **Status**: PASS
- **Duration**: 116.77625ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/app-store/chart-provider/list

- **Status**: PASS
- **Duration**: 111.705084ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/chart-group/entries

- **Status**: FAIL
- **Duration**: 107.436667ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/chart-repo

- **Status**: FAIL
- **Duration**: 105.782083ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 405

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 405

---

### ❌ PUT /orchestrator/chart-repo

- **Status**: FAIL
- **Duration**: 107.05875ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 405

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 405

---

### ✅ GET /orchestrator/chart-repo/list

- **Status**: PASS
- **Duration**: 113.838417ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/server

- **Status**: PASS
- **Duration**: 111.269542ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/server

- **Status**: FAIL
- **Duration**: 110.200208ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/module

- **Status**: PASS
- **Duration**: 115.464708ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/module

- **Status**: FAIL
- **Duration**: 106.159417ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/configmap/environment/edit/{appId}/{envId}/{id}

- **Status**: FAIL
- **Duration**: 105.543333ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/configmap/global/edit/{appId}/{id}

- **Status**: FAIL
- **Duration**: 108.917542ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/configmap/environment

- **Status**: FAIL
- **Duration**: 106.030459ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/configmap/environment/{appId}/{envId}

- **Status**: FAIL
- **Duration**: 107.560542ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/configmap/global

- **Status**: FAIL
- **Duration**: 108.363042ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/configmap/bulk/patch

- **Status**: FAIL
- **Duration**: 109.529042ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/configmap/environment/{appId}/{envId}/{id}

- **Status**: FAIL
- **Duration**: 106.907417ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/configmap/global/{appId}

- **Status**: FAIL
- **Duration**: 106.291625ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/configmap/global/{appId}/{id}

- **Status**: FAIL
- **Duration**: 108.37075ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/plugin/global/create

- **Status**: FAIL
- **Duration**: 110.451792ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/plugin/global/detail/all

- **Status**: PASS
- **Duration**: 715.67825ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/plugin/global/detail/{pluginId}

- **Status**: PASS
- **Duration**: 113.228958ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/plugin/global/list/global-variable

- **Status**: FAIL
- **Duration**: 113.434167ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/plugin/global/list/v2/min

- **Status**: PASS
- **Duration**: 111.267083ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200

---

### ✅ PUT /orchestrator/plugin/global/migrate

- **Status**: PASS
- **Duration**: 109.501291ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/plugin/global/list/detail/v2

- **Status**: FAIL
- **Duration**: 110.023375ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/plugin/global/list/tags

- **Status**: PASS
- **Duration**: 109.743333ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/plugin/global/list/v2

- **Status**: PASS
- **Duration**: 133.480292ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200

---

### ❌ DELETE /orchestrator/user/{id}

- **Status**: FAIL
- **Duration**: 110.997ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/{id}

- **Status**: PASS
- **Duration**: 114.191ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/user

- **Status**: FAIL
- **Duration**: 113.658958ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/user

- **Status**: FAIL
- **Duration**: 108.74775ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/user/bulk

- **Status**: FAIL
- **Duration**: 111.610458ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/user/v2

- **Status**: PASS
- **Duration**: 240.401542ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/app-store/discover

- **Status**: PASS
- **Duration**: 700.572ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/app-store/discover/application/chartInfo/{appStoreApplicationVersionId}

- **Status**: PASS
- **Duration**: 122.80725ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/app-store/discover/application/{appStoreId}/version/autocomplete

- **Status**: PASS
- **Duration**: 110.490333ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/app-store/discover/application/{id}

- **Status**: PASS
- **Duration**: 216.85375ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/app-store/discover/search

- **Status**: FAIL
- **Duration**: 105.242625ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrtor/batch/v1beta1/cd-pipeline

- **Status**: FAIL
- **Duration**: 106.503791ms
- **Spec File**: ../../specs/environment/bulk-delete.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/app/wf/all/component-names/{appId}

- **Status**: PASS
- **Duration**: 114.080333ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/app/ci-pipeline/patch

- **Status**: FAIL
- **Duration**: 112.509375ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/app/ci-pipeline/{appId}

- **Status**: PASS
- **Duration**: 114.319958ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Response Code**: 200

---

### ❌ DELETE /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 111.788125ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PUT /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 112.53125ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/api-token

- **Status**: PASS
- **Duration**: 111.617708ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/api-token

- **Status**: FAIL
- **Duration**: 109.674792ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/api-token/webhook

- **Status**: PASS
- **Duration**: 111.081125ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 200

---

### ❌ PUT /orchestrator/application/rollback

- **Status**: FAIL
- **Duration**: 111.942125ms
- **Spec File**: ../../specs/openapiClient/api/openapi.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrator/application/template-chart

- **Status**: FAIL
- **Duration**: 109.004625ms
- **Spec File**: ../../specs/openapiClient/api/openapi.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/app/history/deployed-component/detail/{appId}/{pipelineId}/{id}

- **Status**: FAIL
- **Duration**: 108.8525ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/history/deployed-component/list/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 109.662125ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/history/deployed-configuration/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 111.565583ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/app/ci-pipeline/{ciPipelineId}/linked-ci/downstream/cd

- **Status**: FAIL
- **Duration**: 123.199958ms
- **Spec File**: ../../specs/ci-pipeline/ciPipelineDownstream/downstream-linked-ci-view-spec.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/ci-pipeline/{ciPipelineId}/linked-ci/downstream/env

- **Status**: FAIL
- **Duration**: 113.205708ms
- **Spec File**: ../../specs/ci-pipeline/ciPipelineDownstream/downstream-linked-ci-view-spec.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/notification

- **Status**: FAIL
- **Duration**: 351.622709ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ DELETE /orchestrator/notification

- **Status**: PASS
- **Duration**: 112.449667ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/notification

- **Status**: FAIL
- **Duration**: 108.9995ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ POST /orchestrator/notification

- **Status**: FAIL
- **Duration**: 108.678917ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/notification/channel

- **Status**: PASS
- **Duration**: 111.810167ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/notification/channel

- **Status**: FAIL
- **Duration**: 109.512708ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200

**Issues:**
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: unexpected end of JSON input

---

### ❌ GET /orchestrator/notification/recipient

- **Status**: FAIL
- **Duration**: 109.236709ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ POST /orchestrator/webhook/ci/workflow

- **Status**: FAIL
- **Duration**: 111.211166ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrator/webhook/ext-ci/{externalCiId}

- **Status**: FAIL
- **Duration**: 117.109667ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/webhook/git

- **Status**: FAIL
- **Duration**: 108.754834ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/git/{gitHostId}

- **Status**: FAIL
- **Duration**: 108.967792ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/git/{gitHostId}/{secret}

- **Status**: FAIL
- **Duration**: 106.062834ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/notification

- **Status**: FAIL
- **Duration**: 106.070375ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/webhook/notification/variables

- **Status**: FAIL
- **Duration**: 106.968708ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/webhook/notification/{id}

- **Status**: FAIL
- **Duration**: 106.813792ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/app/cd-pipeline/patch/deployment

- **Status**: FAIL
- **Duration**: 111.383833ms
- **Spec File**: ../../specs/deployment/app-type-change.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/history/deployed-configuration/all/latest/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 110.711541ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/history/deployed-configuration/all/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 118.646ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/app/deployment-configuration/all/latest/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 106.786666ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /create

- **Status**: FAIL
- **Duration**: 105.261917ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/validate

- **Status**: FAIL
- **Duration**: 107.325709ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /update

- **Status**: FAIL
- **Duration**: 108.351917ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/user/resource/options/{kind}/{version}

- **Status**: FAIL
- **Duration**: 107.601958ms
- **Spec File**: ../../specs/userResource/userResource.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ PUT /orchestrator/deployment/template/upload

- **Status**: PASS
- **Duration**: 109.281875ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/deployment/template/validate

- **Status**: FAIL
- **Duration**: 109.825667ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/deployment/template/fetch

- **Status**: PASS
- **Duration**: 117.576875ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/deployment/pipeline/rollback

- **Status**: FAIL
- **Duration**: 104.515208ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/deployment/pipeline/trigger

- **Status**: FAIL
- **Duration**: 104.593209ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/deployment/pipeline/configure

- **Status**: FAIL
- **Duration**: 107.128833ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/deployment/pipeline/history

- **Status**: FAIL
- **Duration**: 113.479375ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/deployment-status/timeline/{appId}/{envId}

- **Status**: FAIL
- **Duration**: 110.030041ms
- **Spec File**: ../../specs/deployment/timeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/gitops/config/{id}

- **Status**: PASS
- **Duration**: 112.203209ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/gitops/configured

- **Status**: PASS
- **Duration**: 112.755541ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/gitops/validate

- **Status**: FAIL
- **Duration**: 109.076666ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 113.294333ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 109.022666ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/gitops/config

- **Status**: PASS
- **Duration**: 108.868583ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/gitops/config-by-provider

- **Status**: FAIL
- **Duration**: 108.030167ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/git/host

- **Status**: PASS
- **Duration**: 112.078084ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/git/host

- **Status**: FAIL
- **Duration**: 109.338334ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/git/host/event/{eventId}

- **Status**: FAIL
- **Duration**: 109.330667ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/webhook-meta-config/{gitProviderId}

- **Status**: FAIL
- **Duration**: 110.980792ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/{id}

- **Status**: FAIL
- **Duration**: 111.852916ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/{id}/event

- **Status**: FAIL
- **Duration**: 112.090417ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrator/v1beta1/hibernate

- **Status**: FAIL
- **Duration**: 107.047875ms
- **Spec File**: ../../specs/jobs/bulk-actions.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /v1beta1/deploy

- **Status**: FAIL
- **Duration**: 106.763375ms
- **Spec File**: ../../specs/jobs/bulk-actions.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /v1beta1/unhibernate

- **Status**: FAIL
- **Duration**: 105.952917ms
- **Spec File**: ../../specs/jobs/bulk-actions.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/resource/delete

- **Status**: FAIL
- **Duration**: 107.468083ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/resource/inception/info

- **Status**: FAIL
- **Duration**: 108.627167ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/resource/update

- **Status**: FAIL
- **Duration**: 106.858542ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/resource/urls

- **Status**: FAIL
- **Duration**: 118.894125ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/api-resources/{clusterId}

- **Status**: FAIL
- **Duration**: 108.601959ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/resource

- **Status**: FAIL
- **Duration**: 108.28525ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/resource/create

- **Status**: FAIL
- **Duration**: 105.939791ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/user/role/group

- **Status**: PASS
- **Duration**: 112.132416ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 112.686208ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PUT /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 110.8235ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/user/role/group/bulk

- **Status**: FAIL
- **Duration**: 109.1565ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/user/role/group/detailed/get

- **Status**: PASS
- **Duration**: 110.999875ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/user/role/group/search

- **Status**: FAIL
- **Duration**: 332.173292ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/role/group/v2

- **Status**: PASS
- **Duration**: 111.488667ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 111.048125ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PUT /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 111.935959ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/user/role/group/{id}

- **Status**: FAIL
- **Duration**: 109.071375ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/user/role/group/{id}

- **Status**: FAIL
- **Duration**: 109.803708ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /ci-pipeline/trigger

- **Status**: FAIL
- **Duration**: 105.007125ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/commit-info/{ciPipelineMaterialId}/{gitHash}

- **Status**: FAIL
- **Duration**: 130.658292ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ GET /{appId}/ci-pipeline/{pipelineId}/workflow/{workflowId}

- **Status**: FAIL
- **Duration**: 109.001ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ PATCH /orchestrator/app/ci-pipeline/patch-source

- **Status**: FAIL
- **Duration**: 109.047916ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-change-source.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/env/namespace/autocomplete

- **Status**: FAIL
- **Duration**: 144.192875ms
- **Spec File**: ../../specs/environment/core.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---


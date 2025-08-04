# API Spec Validation Report

Generated: 2025-07-31T20:07:52+05:30

## Summary

- Total Endpoints: 242
- Passed: 40
- Failed: 202
- Warnings: 0
- Success Rate: 16.53%

## Detailed Results

### ✅ GET /orchestrator/api-token

- **Status**: PASS
- **Duration**: 479.46475ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/api-token

- **Status**: FAIL
- **Duration**: 118.536ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/api-token/webhook

- **Status**: PASS
- **Duration**: 118.469208ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 200

---

### ❌ DELETE /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 118.510583ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 141.853917ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/user/bulk

- **Status**: FAIL
- **Duration**: 120.998791ms
- **Spec File**: ../../specs/security/policy.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/user/email

- **Status**: FAIL
- **Duration**: 117.038792ms
- **Spec File**: ../../specs/security/policy.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/v2

- **Status**: PASS
- **Duration**: 122.898833ms
- **Spec File**: ../../specs/security/policy.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/user/{id}

- **Status**: FAIL
- **Duration**: 159.432958ms
- **Spec File**: ../../specs/security/policy.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user

- **Status**: PASS
- **Duration**: 122.566833ms
- **Spec File**: ../../specs/security/policy.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/user

- **Status**: FAIL
- **Duration**: 118.69825ms
- **Spec File**: ../../specs/security/policy.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/user

- **Status**: FAIL
- **Duration**: 119.0935ms
- **Spec File**: ../../specs/security/policy.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/deployment/template/validate

- **Status**: FAIL
- **Duration**: 165.913875ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/deployment/template/fetch

- **Status**: PASS
- **Duration**: 232.987958ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Response Code**: 200

---

### ✅ PUT /orchestrator/deployment/template/upload

- **Status**: PASS
- **Duration**: 289.090834ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Response Code**: 200

---

### ❌ POST /repo/create

- **Status**: FAIL
- **Duration**: 115.911583ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /repo/update

- **Status**: FAIL
- **Duration**: 115.300834ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /repo/validate

- **Status**: FAIL
- **Duration**: 115.87325ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /orchestrator/k8s/capacity/node/cordon

- **Status**: FAIL
- **Duration**: 174.45925ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /orchestrator/k8s/capacity/node/drain

- **Status**: FAIL
- **Duration**: 119.258666ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/k8s/capacity/node/list

- **Status**: FAIL
- **Duration**: 121.739125ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/k8s/capacity/node/taints/edit

- **Status**: FAIL
- **Duration**: 120.955708ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/k8s/capacity/cluster/list

- **Status**: PASS
- **Duration**: 166.356625ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/k8s/capacity/cluster/list/raw

- **Status**: PASS
- **Duration**: 128.538709ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/k8s/capacity/cluster/{clusterId}

- **Status**: FAIL
- **Duration**: 120.45525ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/k8s/capacity/node

- **Status**: FAIL
- **Duration**: 119.857167ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/k8s/capacity/node

- **Status**: FAIL
- **Duration**: 189.487125ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/k8s/capacity/node

- **Status**: FAIL
- **Duration**: 118.39675ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/module

- **Status**: PASS
- **Duration**: 122.691167ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/module

- **Status**: FAIL
- **Duration**: 114.615708ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/server

- **Status**: PASS
- **Duration**: 133.673709ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/server

- **Status**: FAIL
- **Duration**: 118.776834ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /api-token

- **Status**: FAIL
- **Duration**: 116.595333ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /api-token

- **Status**: FAIL
- **Duration**: 116.548958ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /api-token/{id}

- **Status**: FAIL
- **Duration**: 173.787167ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /api-token/{id}

- **Status**: FAIL
- **Duration**: 120.727792ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/user

- **Status**: FAIL
- **Duration**: 124.475ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/user

- **Status**: FAIL
- **Duration**: 117.480833ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/user/bulk

- **Status**: FAIL
- **Duration**: 161.0135ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 124.760708ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PUT /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 120.84825ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/user/role/group/bulk

- **Status**: FAIL
- **Duration**: 120.948333ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/user/role/group/v2

- **Status**: PASS
- **Duration**: 154.705458ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/user/v2

- **Status**: PASS
- **Duration**: 123.60975ms
- **Spec File**: ../../specs/security/core.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/version

- **Status**: PASS
- **Duration**: 120.072791ms
- **Spec File**: ../../specs/common/version.yaml
- **Response Code**: 200

---

### ❌ DELETE /orchestrator/k8s/resources/ephemeralContainers

- **Status**: FAIL
- **Duration**: 284.429708ms
- **Spec File**: ../../specs/kubernetes/ephemeral-containers.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/k8s/resources/ephemeralContainers

- **Status**: FAIL
- **Duration**: 134.34725ms
- **Spec File**: ../../specs/kubernetes/ephemeral-containers.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/user/resource/options/{kind}/{version}

- **Status**: FAIL
- **Duration**: 116.374792ms
- **Spec File**: ../../specs/userResource/userResource.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /gitops/config

- **Status**: FAIL
- **Duration**: 115.459ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /gitops/config-by-provider

- **Status**: FAIL
- **Duration**: 870.515959ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ GET /gitops/config/{id}

- **Status**: FAIL
- **Duration**: 115.433208ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ GET /gitops/configured

- **Status**: FAIL
- **Duration**: 410.15675ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /validate

- **Status**: FAIL
- **Duration**: 525.703583ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /config

- **Status**: FAIL
- **Duration**: 119.10325ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /config

- **Status**: FAIL
- **Duration**: 211.424167ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/history/deployed-component/list/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 118.663417ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/history/deployed-configuration/{appId}/{pipelineId}/{wfrId}

- **Status**: FAIL
- **Duration**: 119.026417ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/history/deployed-component/detail/{appId}/{pipelineId}/{id}

- **Status**: FAIL
- **Duration**: 230.447541ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/ci-pipeline/{ciPipelineId}/linked-ci/downstream/env

- **Status**: FAIL
- **Duration**: 120.535416ms
- **Spec File**: ../../specs/ci-pipeline/ciPipelineDownstream/downstream-linked-ci-view-spec.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/ci-pipeline/{ciPipelineId}/linked-ci/downstream/cd

- **Status**: FAIL
- **Duration**: 118.206792ms
- **Spec File**: ../../specs/ci-pipeline/ciPipelineDownstream/downstream-linked-ci-view-spec.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/app/cd-pipeline/patch/deployment

- **Status**: FAIL
- **Duration**: 118.1435ms
- **Spec File**: ../../specs/deployment/app-type-change.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/config/autocomplete

- **Status**: PASS
- **Duration**: 121.241833ms
- **Spec File**: ../../specs/environment/config-diff.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/config/compare/{resource}

- **Status**: FAIL
- **Duration**: 123.593ms
- **Spec File**: ../../specs/environment/config-diff.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/config/data

- **Status**: FAIL
- **Duration**: 123.123791ms
- **Spec File**: ../../specs/environment/config-diff.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/config/manifest

- **Status**: FAIL
- **Duration**: 119.085084ms
- **Spec File**: ../../specs/environment/config-diff.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/flux-application

- **Status**: FAIL
- **Duration**: 122.152042ms
- **Spec File**: ../../specs/gitops/fluxcd.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/flux-application/app

- **Status**: FAIL
- **Duration**: 114.954458ms
- **Spec File**: ../../specs/gitops/fluxcd.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/chart-group/{id}

- **Status**: FAIL
- **Duration**: 115.376209ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/chart-repo/validate

- **Status**: FAIL
- **Duration**: 120.099125ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/app-store/chart-provider/update

- **Status**: FAIL
- **Duration**: 118.931917ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/chart-repo/list

- **Status**: PASS
- **Duration**: 127.737875ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/chart-group/entries

- **Status**: FAIL
- **Duration**: 117.407667ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/chart-repo/{id}

- **Status**: FAIL
- **Duration**: 121.148708ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/app-store/chart-provider/sync-chart

- **Status**: FAIL
- **Duration**: 120.195292ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/chart-group

- **Status**: FAIL
- **Duration**: 235.190333ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 405

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 405

---

### ❌ PUT /orchestrator/chart-group

- **Status**: FAIL
- **Duration**: 236.327ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 405

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 405

---

### ✅ GET /orchestrator/chart-group/list

- **Status**: PASS
- **Duration**: 120.542125ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/chart-repo

- **Status**: FAIL
- **Duration**: 116.564792ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 405

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 405

---

### ❌ PUT /orchestrator/chart-repo

- **Status**: FAIL
- **Duration**: 117.767167ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 405

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 405

---

### ❌ POST /orchestrator/chart-repo/sync

- **Status**: FAIL
- **Duration**: 115.320083ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/app-store/chart-provider/list

- **Status**: PASS
- **Duration**: 122.632917ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/app/ci-pipeline/patch

- **Status**: FAIL
- **Duration**: 120.695833ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/ci-pipeline/{appId}

- **Status**: FAIL
- **Duration**: 116.694584ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/wf/all/component-names/{appId}

- **Status**: FAIL
- **Duration**: 121.882666ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /app/details/{appId}

- **Status**: FAIL
- **Duration**: 116.849792ms
- **Spec File**: ../../specs/application/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ POST /app/edit

- **Status**: FAIL
- **Duration**: 115.570958ms
- **Spec File**: ../../specs/application/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ POST /app/edit/projects

- **Status**: FAIL
- **Duration**: 117.067291ms
- **Spec File**: ../../specs/application/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ POST /app/list

- **Status**: FAIL
- **Duration**: 117.3865ms
- **Spec File**: ../../specs/application/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /core/v1beta1/application

- **Status**: FAIL
- **Duration**: 118.774625ms
- **Spec File**: ../../specs/application/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /env/namespace/autocomplete

- **Status**: FAIL
- **Duration**: 115.381292ms
- **Spec File**: ../../specs/environment/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/template/default/{appId}/{chartRefId}

- **Status**: FAIL
- **Duration**: 119.37625ms
- **Spec File**: ../../specs/environment/templates.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/cluster/access/list

- **Status**: FAIL
- **Duration**: 116.28025ms
- **Spec File**: ../../specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/cluster/access/{id}

- **Status**: FAIL
- **Duration**: 117.6395ms
- **Spec File**: ../../specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ DELETE /orchestrator/cluster/access/{id}

- **Status**: FAIL
- **Duration**: 116.88425ms
- **Spec File**: ../../specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ POST /orchestrator/cluster/access

- **Status**: FAIL
- **Duration**: 304.25775ms
- **Spec File**: ../../specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /orchestrator/cluster/access

- **Status**: FAIL
- **Duration**: 124.565625ms
- **Spec File**: ../../specs/kubernetes/access-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/k8s/resource/list

- **Status**: FAIL
- **Duration**: 120.502417ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/k8s/api-resources/{clusterId}

- **Status**: FAIL
- **Duration**: 120.263958ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/k8s/events

- **Status**: FAIL
- **Duration**: 121.905875ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/k8s/pod/exec/session/{identifier}/{namespace}/{pod}/{shell}/{container}

- **Status**: FAIL
- **Duration**: 118.073959ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/k8s/pods/logs/{podName}

- **Status**: FAIL
- **Duration**: 121.257416ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /orchestrator/k8s/resource

- **Status**: FAIL
- **Duration**: 211.301958ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/k8s/resource

- **Status**: FAIL
- **Duration**: 120.449917ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/k8s/resource/create

- **Status**: FAIL
- **Duration**: 197.586583ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/k8s/resource/delete

- **Status**: FAIL
- **Duration**: 180.214833ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/cluster

- **Status**: PASS
- **Duration**: 163.791958ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Response Code**: 200

---

### ❌ PUT /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 161.645334ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/cluster/auth-list

- **Status**: PASS
- **Duration**: 177.577333ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/webhook/git

- **Status**: FAIL
- **Duration**: 158.385583ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/git/{gitHostId}

- **Status**: FAIL
- **Duration**: 173.31475ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/git/{gitHostId}/{secret}

- **Status**: FAIL
- **Duration**: 177.235458ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/notification

- **Status**: FAIL
- **Duration**: 160.024333ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/webhook/notification/variables

- **Status**: FAIL
- **Duration**: 175.396291ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/webhook/notification/{id}

- **Status**: FAIL
- **Duration**: 174.723709ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/ci/workflow

- **Status**: FAIL
- **Duration**: 162.11725ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrator/webhook/ext-ci/{externalCiId}

- **Status**: FAIL
- **Duration**: 175.914833ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 401

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ PUT /orchestrator/application/rollback

- **Status**: FAIL
- **Duration**: 175.931291ms
- **Spec File**: ../../specs/openapiClient/api/openapi.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrator/application/template-chart

- **Status**: FAIL
- **Duration**: 158.102084ms
- **Spec File**: ../../specs/openapiClient/api/openapi.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/app/deployment-status/timeline/{appId}/{envId}

- **Status**: FAIL
- **Duration**: 178.149459ms
- **Spec File**: ../../specs/deployment/timeline.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/gitops/validate

- **Status**: FAIL
- **Duration**: 184.953875ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/gitops/config

- **Status**: PASS
- **Duration**: 162.137208ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 216.571083ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 172.148ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/gitops/config-by-provider

- **Status**: FAIL
- **Duration**: 172.28ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/gitops/config/{id}

- **Status**: FAIL
- **Duration**: 176.816708ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/gitops/configured

- **Status**: PASS
- **Duration**: 177.287ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/app-store/installed-app

- **Status**: PASS
- **Duration**: 175.094583ms
- **Spec File**: ../../specs/helm/charts.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/app-store/installed-app/notes/{installed-app-id}/{env-id}

- **Status**: FAIL
- **Duration**: 176.657625ms
- **Spec File**: ../../specs/helm/charts.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /bulk/v1beta1/application

- **Status**: FAIL
- **Duration**: 175.17025ms
- **Spec File**: ../../specs/jobs/batch.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /bulk/v1beta1/application/dryrun

- **Status**: FAIL
- **Duration**: 175.674625ms
- **Spec File**: ../../specs/jobs/batch.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /operate

- **Status**: FAIL
- **Duration**: 174.532708ms
- **Spec File**: ../../specs/jobs/batch.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /v1beta1/unhibernate

- **Status**: FAIL
- **Duration**: 178.839791ms
- **Spec File**: ../../specs/jobs/bulk-actions.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /v1beta1/deploy

- **Status**: FAIL
- **Duration**: 174.395833ms
- **Spec File**: ../../specs/jobs/bulk-actions.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /v1beta1/hibernate

- **Status**: FAIL
- **Duration**: 175.969833ms
- **Spec File**: ../../specs/jobs/bulk-actions.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/resource/inception/info

- **Status**: FAIL
- **Duration**: 174.812667ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/resource/update

- **Status**: FAIL
- **Duration**: 174.885917ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/resource/urls

- **Status**: FAIL
- **Duration**: 177.189125ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/api-resources/{clusterId}

- **Status**: FAIL
- **Duration**: 117.6665ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/resource

- **Status**: FAIL
- **Duration**: 115.611292ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/resource/create

- **Status**: FAIL
- **Duration**: 115.685ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/resource/delete

- **Status**: FAIL
- **Duration**: 116.14625ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ DELETE /orchestrator/notification

- **Status**: PASS
- **Duration**: 118.944083ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/notification

- **Status**: FAIL
- **Duration**: 116.952ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/notification

- **Status**: FAIL
- **Duration**: 127.82625ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/notification

- **Status**: FAIL
- **Duration**: 123.050959ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/notification/channel

- **Status**: PASS
- **Duration**: 121.294292ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/notification/channel

- **Status**: FAIL
- **Duration**: 118.747042ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200

**Issues:**
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: unexpected end of JSON input

---

### ❌ GET /orchestrator/notification/recipient

- **Status**: FAIL
- **Duration**: 121.428375ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /orchestrator/user

- **Status**: FAIL
- **Duration**: 120.095375ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/user

- **Status**: FAIL
- **Duration**: 119.352042ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/user/bulk

- **Status**: FAIL
- **Duration**: 119.197042ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/user/v2

- **Status**: PASS
- **Duration**: 120.174125ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 200

---

### ❌ DELETE /orchestrator/user/{id}

- **Status**: FAIL
- **Duration**: 118.866959ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/user/{id}

- **Status**: FAIL
- **Duration**: 119.380667ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/meta/info/{appId}

- **Status**: FAIL
- **Duration**: 118.602292ms
- **Spec File**: ../../specs/application/labels.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/helm/meta/info/{appId}

- **Status**: FAIL
- **Duration**: 120.589334ms
- **Spec File**: ../../specs/application/labels.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/app/labels/list

- **Status**: PASS
- **Duration**: 124.203875ms
- **Spec File**: ../../specs/application/labels.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/app/history/deployed-component/detail/{appId}/{pipelineId}/{id}

- **Status**: FAIL
- **Duration**: 117.41825ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/history/deployed-component/list/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 121.731167ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/history/deployed-configuration/{appId}/{pipelineId}/{wfrId}

- **Status**: FAIL
- **Duration**: 120.252167ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /deployment/pipeline/configure

- **Status**: FAIL
- **Duration**: 117.095875ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /deployment/pipeline/history

- **Status**: FAIL
- **Duration**: 115.47675ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /deployment/pipeline/rollback

- **Status**: FAIL
- **Duration**: 115.507709ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /deployment/pipeline/trigger

- **Status**: FAIL
- **Duration**: 116.048875ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/chartref/autocomplete/{appId}

- **Status**: FAIL
- **Duration**: 117.470542ms
- **Spec File**: ../../specs/helm/dynamic-charts.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /job

- **Status**: FAIL
- **Duration**: 127.352792ms
- **Spec File**: ../../specs/jobs/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /job/ci-pipeline/list/{jobId}

- **Status**: FAIL
- **Duration**: 114.604792ms
- **Spec File**: ../../specs/jobs/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /job/list

- **Status**: FAIL
- **Duration**: 114.410666ms
- **Spec File**: ../../specs/jobs/core.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/configmap/global/{appId}

- **Status**: FAIL
- **Duration**: 301.662ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/configmap/environment/edit/{appId}/{envId}/{id}

- **Status**: FAIL
- **Duration**: 115.159708ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/configmap/global/{appId}/{id}

- **Status**: FAIL
- **Duration**: 113.942792ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/configmap/environment/{appId}/{envId}/{id}

- **Status**: FAIL
- **Duration**: 114.438833ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/configmap/global

- **Status**: FAIL
- **Duration**: 116.601917ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/configmap/global/edit/{appId}/{id}

- **Status**: FAIL
- **Duration**: 117.131083ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/configmap/bulk/patch

- **Status**: FAIL
- **Duration**: 116.508792ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/configmap/environment

- **Status**: FAIL
- **Duration**: 115.701583ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/configmap/environment/{appId}/{envId}

- **Status**: FAIL
- **Duration**: 115.986208ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/user/role/group/bulk

- **Status**: FAIL
- **Duration**: 121.139166ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/user/role/group/detailed/get

- **Status**: PASS
- **Duration**: 119.129042ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/user/role/group/search

- **Status**: FAIL
- **Duration**: 118.836917ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/role/group/v2

- **Status**: PASS
- **Duration**: 131.786209ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 427.608334ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PUT /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 117.282792ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/user/role/group/{id}

- **Status**: FAIL
- **Duration**: 118.62325ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/user/role/group/{id}

- **Status**: FAIL
- **Duration**: 116.704292ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/role/group

- **Status**: PASS
- **Duration**: 120.238959ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 120.792583ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PUT /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 119.761ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/git/host/{id}/event

- **Status**: FAIL
- **Duration**: 116.646083ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/git/host

- **Status**: PASS
- **Duration**: 120.400291ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/git/host

- **Status**: FAIL
- **Duration**: 117.04825ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/git/host/event/{eventId}

- **Status**: FAIL
- **Duration**: 122.459459ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/webhook-meta-config/{gitProviderId}

- **Status**: FAIL
- **Duration**: 120.226584ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/{id}

- **Status**: FAIL
- **Duration**: 118.656292ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /app-store/discover

- **Status**: FAIL
- **Duration**: 114.944458ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /app-store/discover/application/chartInfo/{appStoreApplicationVersionId}

- **Status**: FAIL
- **Duration**: 115.366667ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /app-store/discover/application/{appStoreId}/version/autocomplete

- **Status**: FAIL
- **Duration**: 115.001333ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /app-store/discover/application/{id}

- **Status**: FAIL
- **Duration**: 114.712917ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /app-store/discover/search

- **Status**: FAIL
- **Duration**: 116.11125ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /app/workflow/{workflowId}

- **Status**: FAIL
- **Duration**: 117.421042ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ GET /app/commit-info/{ciPipelineMaterialId}/{gitHash}

- **Status**: FAIL
- **Duration**: 115.206875ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ POST /app/workflow/trigger/{pipelineId}

- **Status**: FAIL
- **Duration**: 115.732542ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PATCH /orchestrator/app/ci-pipeline/patch-source

- **Status**: FAIL
- **Duration**: 118.433792ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-change-source.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/app/deployment-configuration/all/latest/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 117.736666ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/history/deployed-configuration/all/latest/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 122.067458ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/history/deployed-configuration/all/{appId}/{pipelineId}/{wfrId}

- **Status**: FAIL
- **Duration**: 117.698958ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrtor/batch/v1beta1/cd-pipeline

- **Status**: FAIL
- **Duration**: 118.760833ms
- **Spec File**: ../../specs/environment/bulk-delete.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/app/deployment/template/data

- **Status**: FAIL
- **Duration**: 122.021041ms
- **Spec File**: ../../specs/gitops/manifest-generation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/deployments/{app-id}/{env-id}

- **Status**: FAIL
- **Duration**: 115.805916ms
- **Spec File**: ../../specs/gitops/manifest-generation.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/plugin/global/create

- **Status**: FAIL
- **Duration**: 128.198875ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/plugin/global/list/global-variable

- **Status**: FAIL
- **Duration**: 118.194542ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/plugin/global/list/v2

- **Status**: PASS
- **Duration**: 264.343958ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200

---

### ✅ PUT /orchestrator/plugin/global/migrate

- **Status**: PASS
- **Duration**: 120.794959ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/plugin/global/list/tags

- **Status**: PASS
- **Duration**: 118.12825ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/plugin/global/list/v2/min

- **Status**: PASS
- **Duration**: 119.343834ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/plugin/global/detail/{pluginId}

- **Status**: FAIL
- **Duration**: 118.107458ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/plugin/global/list/detail/v2

- **Status**: FAIL
- **Duration**: 120.629542ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/plugin/global/detail/all

- **Status**: PASS
- **Duration**: 649.987333ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200

---

### ❌ PATCH /orchestrator/app/env/patch

- **Status**: FAIL
- **Duration**: 118.278792ms
- **Spec File**: ../../specs/helm/deployment-chart-type.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/git/provider/delete

- **Status**: FAIL
- **Duration**: 116.134791ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/notification/channel/delete

- **Status**: FAIL
- **Duration**: 116.006708ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/docker/registry/delete

- **Status**: FAIL
- **Duration**: 115.072083ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/team/delete

- **Status**: FAIL
- **Duration**: 115.510667ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/app-store/repo/delete

- **Status**: FAIL
- **Duration**: 115.382916ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/app/material/delete

- **Status**: FAIL
- **Duration**: 113.997541ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/env/delete

- **Status**: FAIL
- **Duration**: 117.202834ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/cluster/delete

- **Status**: FAIL
- **Duration**: 113.643ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/chart-group/delete

- **Status**: FAIL
- **Duration**: 116.344833ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/external-links/tools

- **Status**: PASS
- **Duration**: 120.402792ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/external-links

- **Status**: PASS
- **Duration**: 120.738083ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 200

---

### ❌ POST /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 117.397708ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 121.034375ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 115.200625ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 404

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/cluster/saveClusters

- **Status**: FAIL
- **Duration**: 121.105ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/cluster/validate

- **Status**: FAIL
- **Duration**: 121.443291ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Response Code**: 500

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 125.265208ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 117.989833ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 116.531625ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/cluster

- **Status**: PASS
- **Duration**: 157.975417ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/cluster/auth-list

- **Status**: PASS
- **Duration**: 121.821ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Response Code**: 200

---

### ✅ GET /orchestrator/cluster/namespaces

- **Status**: PASS
- **Duration**: 120.859583ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Response Code**: 200

---

### ❌ GET /orchestrator/cluster/namespaces/{clusterId}

- **Status**: FAIL
- **Duration**: 117.962167ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Response Code**: 400

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---


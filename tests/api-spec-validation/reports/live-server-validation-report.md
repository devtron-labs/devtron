# API Spec Validation Report

Generated: 2025-08-08T13:39:39+05:30

## Summary

- Total Endpoints: 213
- Passed: 80
- Failed: 133
- Warnings: 0
- Success Rate: 37.56%

## Detailed Results

### ❌ GET /orchestrator/flux-application/app

- **Status**: FAIL
- **Duration**: 314.137375ms
- **Spec File**: ../../specs/gitops/fluxcd.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/flux-application

- **Status**: FAIL
- **Duration**: 95.3985ms
- **Spec File**: ../../specs/gitops/fluxcd.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{error in getting cluster ids}]","userMessage":"error in getting cluster ids"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/app/ci-pipeline/patch

- **Status**: FAIL
- **Duration**: 81.477458ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "appWorkflowId": 1,
  "ciPipeline": {
    "afterDockerBuildScripts": [
      {
        "index": 1,
        "name": "sample_string",
        "script": "sample_string"
      },
      {
        "index": 1,
        "name": "sample-item-2",
        "script": "sample_string"
      }
    ],
    "ciMaterial": [
      {
        "active": true,
        "type": "sample_string",
        "value": "sample_string"
      },
      {
        "active": true,
        "type": "sample_string",
        "value": "sample_string"
      }
    ],
    "isExternal": true,
    "isManual": true,
    "name": "sample_string",
    "pipelineType": "CI",
    "scanEnabled": true
  },
  "isCloneJob": true,
  "isJob": true
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'CiPatchRequest.CiPipeline.CiMaterial[0].Source' Error:Field validation for 'Source' failed on the 'dive' tag","userMessage":"Key:...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/ci-pipeline/{appId}

- **Status**: PASS
- **Duration**: 106.986875ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"appId":1,"dockerRegistry":"quay","dockerRepository":"devtron/test","ciBuildConfig":{"id":1,"gitMaterialId":1,"buildContextGitMaterialId":1,"useRootBuildCont...

---

### ✅ GET /orchestrator/app/wf/all/component-names/{appId}

- **Status**: PASS
- **Duration**: 84.812417ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"workflows":[{"id":1,"name":"wf-1-lcu5","ciPipelineId":0,"ciPipelineName":"","cdPipelines":null},{"id":2,"name":"wf-1-4rww","ciPipelineId":0,"ciPipelineName":"","cd...

---

### ✅ GET /orchestrator/app-store/discover

- **Status**: PASS
- **Duration**: 760.414042ms
- **Spec File**: ../../specs/app-store.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":65,"appStoreApplicationVersionId":4028,"name":"ai-agent","chart_repo_id":2,"docker_artifact_store_id":"","chart_name":"devtron","icon":"","active":true,"chart...

---

### ✅ GET /orchestrator/app-store/discover/application/chartInfo/{appStoreApplicationVersionId}

- **Status**: PASS
- **Duration**: 97.211833ms
- **Spec File**: ../../specs/app-store.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"appStoreApplicationVersionId":1,"readme":"# cluster-autoscaler\n\nScales Kubernetes worker nodes within autoscaling groups.\n\n## TL;DR\n\n```console\n$ helm repo ...

---

### ✅ GET /orchestrator/app-store/discover/application/{appStoreId}/version/autocomplete

- **Status**: PASS
- **Duration**: 79.915ms
- **Spec File**: ../../specs/app-store.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"version":"9.49.0","id":5507},{"version":"9.48.0","id":1},{"version":"9.47.0","id":2},{"version":"9.46.6","id":3},{"version":"9.46.5","id":4},{"version":"9.46.4","...

---

### ✅ GET /orchestrator/app-store/discover/application/{id}

- **Status**: PASS
- **Duration**: 107.296875ms
- **Spec File**: ../../specs/app-store.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"version":"9.48.0","appVersion":"1.33.0","created":"2025-07-11T21:16:00.149315Z","deprecated":false,"description":"Scales Kubernetes worker nodes within auto...

---

### ❌ GET /orchestrator/app-store/discover/search

- **Status**: FAIL
- **Duration**: 74.959708ms
- **Spec File**: ../../specs/app-store.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/history/deployed-component/detail/{appId}/{pipelineId}/{id}

- **Status**: FAIL
- **Duration**: 77.612333ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","result":"invalid historyComponent"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/history/deployed-component/list/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 78.309583ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","result":"invalid historyComponent"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/history/deployed-configuration/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 96.5035ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":8,"name":"DEPLOYMENT_TEMPLATE"},{"id":3,"name":"PIPELINE_STRATEGY"}]}

---

### ❌ GET /orchestrator/app/ci-pipeline/{ciPipelineId}/linked-ci/downstream/cd

- **Status**: FAIL
- **Duration**: 77.626834ms
- **Spec File**: ../../specs/ci-pipeline/ciPipelineDownstream/downstream-linked-ci-view-spec.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{invalid CiPipelineId 1}]","userMessage":"invalid CiPipelineId 1"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/ci-pipeline/{ciPipelineId}/linked-ci/downstream/env

- **Status**: FAIL
- **Duration**: 77.608291ms
- **Spec File**: ../../specs/ci-pipeline/ciPipelineDownstream/downstream-linked-ci-view-spec.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{invalid CiPipelineId 1}]","userMessage":"invalid CiPipelineId 1"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/version

- **Status**: PASS
- **Duration**: 77.113041ms
- **Spec File**: ../../specs/common/version.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"result":{"gitCommit":"908eae83","buildTime":"2025-08-05T20:18:14Z","serverMode":"FULL"}}

---

### ❌ GET /orchestrator/app/deployment-status/timeline/{appId}/{envId}

- **Status**: FAIL
- **Duration**: 79.638583ms
- **Spec File**: ../../specs/deployment/timeline.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/config/environment/cm

- **Status**: FAIL
- **Duration**: 87.626542ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "configData": [
    {
      "external": true,
      "id": 1,
      "mountPath": "sample_string",
      "name": "sample_string",
      "type": "CONFIGMAP"
    },
    {
      "external": true,
      "id": 2,
      "mountPath": "sample_string",
      "name": "sample-item-2",
      "type": "CONFIGMAP"
    }
  ],
  "environmentId": 1,
  "id": 1,
  "userId": 1
}
```
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{invalid request multiple config found for add or update}]","userMessage":"invalid request multiple config foun...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/config/environment/cm/edit/{appId}/{envId}/{id}

- **Status**: FAIL
- **Duration**: 74.750542ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/config/environment/cm/{appId}/{envId}

- **Status**: PASS
- **Duration**: 80.833ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":0,"appId":1,"environmentId":1,"configData":[],"isDeletable":false,"isExpressEdit":false}}

---

### ❌ DELETE /orchestrator/config/global/cm/{appId}/{id}

- **Status**: FAIL
- **Duration**: 73.862ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/config/global/cm/edit/{appId}/{id}

- **Status**: FAIL
- **Duration**: 74.770166ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/config/bulk/patch

- **Status**: FAIL
- **Duration**: 79.022667ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "configData": [
    {
      "data": {},
      "externalSecret": [
        {
          "isBinary": true,
          "key": "sample_string",
          "name": "sample_string",
          "property": "sample_string"
        },
        {
          "isBinary": true,
          "key": "sample_string",
          "name": "sample-item-2",
          "property": "sample_string"
        }
      ],
      "filePermission": "sample_string",
      "name": "sample_string",
      "type": "CONFIGMAP"
    },
    {
      "data": {},
      "externalSecret": [
        {
          "isBinary": true,
          "key": "sample_string",
          "name": "sample_string",
          "property": "sample_string"
        },
        {
          "isBinary": true,
          "key": "sample_string",
          "name": "sample-item-2",
          "property": "sample_string"
        }
      ],
      "filePermission": "sample_string",
      "name": "sample-item-2",
      "type": "CONFIGMAP"
    }
  ],
  "environmentId": 1,
  "global": true,
  "userId": 1
}
```
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{invalid request no payload found for sync}]","userMessage":"invalid request no payload found for sync"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ DELETE /orchestrator/config/environment/cm/{appId}/{envId}/{id}

- **Status**: FAIL
- **Duration**: 74.486958ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/config/global/cm

- **Status**: FAIL
- **Duration**: 79.251833ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "configData": [
    {
      "filePermission": "sample_string",
      "name": "sample_string",
      "roleARN": "sample_string",
      "subPath": "sample_string",
      "type": "CONFIGMAP"
    },
    {
      "filePermission": "sample_string",
      "name": "sample-item-2",
      "roleARN": "sample_string",
      "subPath": "sample_string",
      "type": "CONFIGMAP"
    }
  ],
  "environmentId": 1,
  "id": 1,
  "userId": 1
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{json: cannot unmarshal string into Go struct field ConfigData.configData.subPath of type bool}]","userMessage":"json: ca...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/config/global/cm/{appId}

- **Status**: PASS
- **Duration**: 78.841208ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":0,"appId":1,"configData":[],"isDeletable":false,"isExpressEdit":false}}

---

### ❌ GET /orchestrator/app/commit-info/{ciPipelineMaterialId}/{gitHash}

- **Status**: FAIL
- **Duration**: 79.107833ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/ci-pipeline/trigger

- **Status**: FAIL
- **Duration**: 75.435583ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Request Payload**:
```json
{
  "ciArtifactLastFetch": "2023-01-01T00:00:00Z",
  "ciPipelineMaterials": [
    {
      "GitTag": "v1.0.0",
      "Id": 1,
      "Type": "GIT"
    },
    {
      "GitTag": "v1.0.0",
      "Id": 1,
      "Type": "GIT"
    }
  ],
  "pipelineId": 123,
  "pipelineType": "CI",
  "triggeredBy": 1
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/{appId}/ci-pipeline/{pipelineId}/workflow/{workflowId}

- **Status**: FAIL
- **Duration**: 84.279709ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ GET /orchestrator/app/history/deployed-configuration/all/latest/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 77.022083ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"latest\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"latest\": invalid syntax"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/history/deployed-configuration/all/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 87.360792ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"deploymentTemplate":{"templateName":"Deployment","templateVersion":"4.21.0","isAppMetricsEnabled":false,"codeEditorValue":{"displayName":"values.yaml","value":"{\"...

---

### ✅ GET /orchestrator/chart-repo/list

- **Status**: PASS
- **Duration**: 85.261958ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"default-chartmuseum","url":"http://devtron-chartmuseum.devtroncd:8080/","authMode":"ANONYMOUS","active":true,"default":true,"allow_insecure_connecti...

---

### ✅ POST /orchestrator/chart-repo/validate

- **Status**: PASS
- **Duration**: 265.691084ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**:
```json
{
  "active": true,
  "allow_insecure_connection": false,
  "authMode": "ANONYMOUS",
  "default": false,
  "name": "sample-chart-repo",
  "url": "https://charts.example.com"
}
```
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"customErrMsg":"Could not validate the repo. Please try again.","actualErrMsg":"Get \"https://charts.example.com/index.yaml\": dial tcp: lookup charts.example.com o...

---

### ✅ POST /orchestrator/app-store/chart-provider/sync-chart

- **Status**: PASS
- **Duration**: 7.172846083s
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**:
```json
{
  "active": true,
  "id": "sample_string",
  "isOCIRegistry": true
}
```
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK"}

---

### ✅ POST /orchestrator/chart-repo/sync-charts

- **Status**: PASS
- **Duration**: 7.17838225s
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"status":"ok"}}

---

### ✅ GET /orchestrator/chart-repo/{id}

- **Status**: PASS
- **Duration**: 92.338084ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"name":"default-chartmuseum","url":"http://devtron-chartmuseum.devtroncd:8080/","authMode":"ANONYMOUS","active":true,"default":true,"allow_insecure_connectio...

---

### ✅ GET /orchestrator/app-store/chart-provider/list

- **Status**: PASS
- **Duration**: 93.780208ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":"prakhar","name":"prakhar","active":true,"isEditable":true,"isOCIRegistry":true,"registryProvider":"docker-hub"},{"id":"1","name":"default-chartmuseum","activ...

---

### ❌ POST /orchestrator/app-store/chart-provider/update

- **Status**: FAIL
- **Duration**: 91.907708ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**:
```json
{
  "active": true,
  "id": "sample_string",
  "isOCIRegistry": true
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/chart-group/list

- **Status**: PASS
- **Duration**: 93.172291ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":null}

---

### ✅ POST /orchestrator/chart-repo/create

- **Status**: PASS
- **Duration**: 320.083792ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**:
```json
{
  "active": true,
  "allow_insecure_connection": false,
  "authMode": "ANONYMOUS",
  "default": false,
  "name": "sample-chart-repo",
  "url": "https://charts.example.com"
}
```
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"customErrMsg":"Could not validate the repo. Please try again.","actualErrMsg":"Get \"https://charts.example.com/index.yaml\": dial tcp: lookup charts.example.com o...

---

### ❌ PUT /orchestrator/chart-group/entries

- **Status**: FAIL
- **Duration**: 93.290875ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**:
```json
{
  "entries": [
    {
      "chartId": 1,
      "chartRepoId": 1,
      "chartRepoName": "sample_string",
      "chartVersion": "sample_string"
    },
    {
      "chartId": 1,
      "chartRepoId": 1,
      "chartRepoName": "sample_string",
      "chartVersion": "sample_string"
    }
  ],
  "groupId": 1
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/chart-repo/update

- **Status**: FAIL
- **Duration**: 98.186166ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**:
```json
{
  "active": true,
  "allow_insecure_connection": false,
  "authMode": "ANONYMOUS",
  "default": false,
  "name": "sample-chart-repo",
  "url": "https://charts.example.com"
}
```
- **Response Code**: 412
- **Error/Msg**: {"code":412,"status":"Precondition Failed","result":{"customErrMsg":"Could not validate the repo. Please try again.","actualErrMsg":"Get \"https://charts.example.com/index.yaml\": dial tcp: lookup cha...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 412

---

### ❌ POST /orchestrator/chart-group/

- **Status**: FAIL
- **Duration**: 100.217ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**:
```json
{
  "entries": [
    {
      "chartId": 1,
      "chartName": "sample_string",
      "chartRepoId": 1,
      "chartRepoName": "sample_string"
    },
    {
      "chartId": 1,
      "chartName": "sample_string",
      "chartRepoId": 1,
      "chartRepoName": "sample_string"
    }
  ],
  "id": 1,
  "name": "sample_string",
  "userId": 1
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'ChartGroupBean.Name' Error:Field validation for 'Name' failed on the 'name-component' tag","userMessage":"Key: 'ChartGroupBean.Na...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/chart-group/

- **Status**: FAIL
- **Duration**: 90.134041ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**:
```json
{
  "description": "sample_string",
  "id": 1,
  "name": "sample_string",
  "userId": 1
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'ChartGroupBean.Name' Error:Field validation for 'Name' failed on the 'name-component' tag","userMessage":"Key: 'ChartGroupBean.Na...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/chart-group/{id}

- **Status**: FAIL
- **Duration**: 87.248125ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/batch/v1beta1/hibernate

- **Status**: FAIL
- **Duration**: 130.197ms
- **Spec File**: ../../specs/jobs/bulk-operations.yaml
- **Request Payload**:
```json
{
  "applications": [
    1
  ]
}
```
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{ERROR #42601 syntax error at or near \")\"}]","userMessage":"ERROR #42601 syntax error at or near \")\""}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrator/batch/v1beta1/unhibernate

- **Status**: FAIL
- **Duration**: 146.544166ms
- **Spec File**: ../../specs/jobs/bulk-operations.yaml
- **Request Payload**:
```json
{
  "applications": [
    1
  ]
}
```
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{ERROR #42601 syntax error at or near \")\"}]","userMessage":"ERROR #42601 syntax error at or near \")\""}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/batch/{apiVersion}/{kind}/readme

- **Status**: FAIL
- **Duration**: 117.374125ms
- **Spec File**: ../../specs/jobs/bulk-operations.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/batch/v1beta1/application

- **Status**: FAIL
- **Duration**: 119.588583ms
- **Spec File**: ../../specs/jobs/bulk-operations.yaml
- **Request Payload**:
```json
{
  "excludes": [
    {},
    {}
  ],
  "filter": {},
  "includes": [
    {},
    {}
  ]
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'BulkUpdateScript.ApiVersion' Error:Field validation for 'ApiVersion' failed on the 'required' tag","userMessage":"Key: 'BulkUpdat...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/batch/v1beta1/application/dryrun

- **Status**: FAIL
- **Duration**: 90.33325ms
- **Spec File**: ../../specs/jobs/bulk-operations.yaml
- **Request Payload**:
```json
{
  "excludes": [
    {},
    {}
  ],
  "filter": {},
  "includes": [
    {},
    {}
  ]
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'BulkUpdateScript.ApiVersion' Error:Field validation for 'ApiVersion' failed on the 'required' tag","userMessage":"Key: 'BulkUpdat...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ POST /orchestrator/batch/v1beta1/build

- **Status**: PASS
- **Duration**: 98.608958ms
- **Spec File**: ../../specs/jobs/bulk-operations.yaml
- **Request Payload**:
```json
{
  "applications": [
    1
  ]
}
```
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"invalidateCache":false,"deployLatestEligibleArtifact":false,"response":{}}}

---

### ❌ POST /orchestrator/batch/v1beta1/deploy

- **Status**: FAIL
- **Duration**: 119.371666ms
- **Spec File**: ../../specs/jobs/bulk-operations.yaml
- **Request Payload**:
```json
{
  "applications": [
    1
  ]
}
```
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{please mention environment id or environment name}]","userMessage":"please mention environment id or environme...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/module

- **Status**: PASS
- **Duration**: 100.332042ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"name":"security.trivy","status":"installed","moduleResourcesStatus":null,"enabled":true,"moduleType":"security"},{"name":"security.clair","status":"installed","mo...

---

### ❌ POST /orchestrator/module

- **Status**: FAIL
- **Duration**: 87.364209ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Request Payload**:
```json
{
  "action": "INSTALL",
  "name": "sample-module"
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/server

- **Status**: FAIL
- **Duration**: 94.680917ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Request Payload**:
```json
{
  "action": "RESTART"
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'ServerActionRequestDto.Action' Error:Field validation for 'Action' failed on the 'oneof' tag","userMessage":"Key: 'ServerActionRe...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/server

- **Status**: PASS
- **Duration**: 91.85725ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"status":"unknown","releaseName":"devtron","installationType":"enterprise"}}

---

### ✅ GET /orchestrator/notification/variables

- **Status**: PASS
- **Duration**: 89.404083ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"devtronAppId":"{{devtronAppId}}","devtronAppName":"{{devtronAppName}}","devtronApprovedByEmail":"{{devtronApprovedByEmail}}","devtronBuildGitCommitHash":"{{devtron...

---

### ✅ GET /orchestrator/notification/channel/smtp/{id}

- **Status**: PASS
- **Duration**: 91.968083ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":0,"port":"","host":"","authType":"","authUser":"","authPassword":"","fromEmail":"","configName":"","description":"","ownerId":0,"default":false,"deleted":false...

---

### ❌ POST /orchestrator/notification/v2

- **Status**: FAIL
- **Duration**: 91.191959ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**:
```json
{
  "notificationConfigRequest": {
    "appId": 1,
    "envId": 1,
    "eventTypeIds": [
      1
    ],
    "pipelineId": 1,
    "pipelineType": "CI",
    "providers": [
      {
        "configId": 1,
        "dest": "test@example.com"
      }
    ],
    "teamId": 1
  }
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{json: cannot unmarshal object into Go struct field NotificationRequest.notificationConfigRequest of type []*beans.Notifi...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/notification/channel

- **Status**: FAIL
- **Duration**: 89.35075ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**:
```json
{
  "notificationConfigRequest": {
    "appId": 1,
    "envId": 1,
    "eventTypeIds": [
      1
    ],
    "pipelineId": 1,
    "pipelineType": "CI",
    "providers": [
      {
        "configId": 1,
        "dest": "test@example.com"
      }
    ],
    "teamId": 1
  }
}
```
- **Response Code**: 200
- **Error/Msg**: {}

**Issues:**
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: unexpected end of JSON input

---

### ❌ DELETE /orchestrator/notification/channel

- **Status**: FAIL
- **Duration**: 89.949ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{EOF}]","userMessage":"EOF"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/notification/channel

- **Status**: PASS
- **Duration**: 95.67275ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"slackConfigs":[],"webhookConfigs":[],"sesConfigs":[{"userId":2,"teamId":0,"region":"ap-south-1","accessKey":"AKIAWEAVHFOOSNAORPHU","secretKey":"**********","fromEm...

---

### ✅ GET /orchestrator/notification/channel/autocomplete/{type}

- **Status**: PASS
- **Duration**: 88.91025ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[]}

---

### ✅ GET /orchestrator/notification/channel/slack/{id}

- **Status**: PASS
- **Duration**: 91.22925ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"userId":0,"teamId":0,"webhookUrl":"","configName":"","description":"","id":0}}

---

### ❌ GET /orchestrator/notification/recipient

- **Status**: FAIL
- **Duration**: 87.604667ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/notification

- **Status**: FAIL
- **Duration**: 86.324458ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/notification

- **Status**: FAIL
- **Duration**: 91.880417ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**:
```json
{
  "notificationConfigRequest": {
    "appId": 1,
    "envId": 1,
    "eventTypeIds": [
      1
    ],
    "pipelineId": 1,
    "pipelineType": "CI",
    "providers": [
      {
        "configId": 1,
        "dest": "test@example.com"
      }
    ],
    "teamId": 1
  }
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{json: cannot unmarshal object into Go struct field NotificationRequest.notificationConfigRequest of type []*beans.Notifi...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/notification

- **Status**: FAIL
- **Duration**: 90.805833ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**:
```json
{
  "notificationConfigRequest": {
    "appId": 1,
    "envId": 1,
    "eventTypeIds": [
      1
    ],
    "pipelineId": 1,
    "pipelineType": "CI",
    "providers": [
      {
        "configId": 1,
        "dest": "test@example.com"
      }
    ],
    "teamId": 1
  }
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{json: cannot unmarshal object into Go struct field NotificationUpdateRequest.notificationConfigRequest of type []*beans....

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/notification

- **Status**: FAIL
- **Duration**: 89.600833ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{EOF}]","userMessage":"EOF"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/notification/channel/ses/{id}

- **Status**: PASS
- **Duration**: 92.012416ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"userId":2,"teamId":0,"region":"ap-south-1","accessKey":"AKIAWEAVHFOOSNAORPHU","secretKey":"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","fromEmail":"watcher@devtron.i...

---

### ✅ GET /orchestrator/notification/channel/webhook/{id}

- **Status**: PASS
- **Duration**: 89.638833ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"userId":0,"webhookUrl":"","configName":"","header":null,"payload":"","description":"","id":0}}

---

### ✅ POST /orchestrator/notification/search

- **Status**: PASS
- **Duration**: 103.030084ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**:
```json
{
  "notificationConfigRequest": {
    "appId": 1,
    "envId": 1,
    "eventTypeIds": [
      1
    ],
    "pipelineId": 1,
    "pipelineType": "CI",
    "providers": [
      {
        "configId": 1,
        "dest": "test@example.com"
      }
    ],
    "teamId": 1
  }
}
```
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"team":null,"app":null,"environment":null,"cluster":null,"pipeline":{"id":1,"name":"cd-3-pgxd","environmentName":"devtron-demo","appName":"pk-test","isVirtualEnvir...

---

### ❌ POST /orchestrator/webhook/ext-ci/{externalCiId}

- **Status**: FAIL
- **Duration**: 90.738583ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Request Payload**:
```json
{
  "message": "sample_string",
  "pipelineId": 1,
  "startedOn": "2023-01-01T00:00:00Z",
  "status": "SUCCESS",
  "triggeredBy": 1
}
```
- **Response Code**: 401
- **Error/Msg**: {"code":401,"status":"Unauthorized","errors":[{"code":"6005","internalMessage":"no token provided"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/webhook/git

- **Status**: FAIL
- **Duration**: 88.789541ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Request Payload**:
```json
{
  "commits": [
    {
      "added": [
        "sample_string"
      ],
      "removed": [
        "sample_string"
      ],
      "timestamp": "2023-01-01T00:00:00Z"
    },
    {
      "added": [
        "sample_string"
      ],
      "removed": [
        "sample_string"
      ],
      "timestamp": "2023-01-01T00:00:00Z"
    }
  ],
  "deleted": true,
  "forced": true
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/git/{gitHostId}

- **Status**: FAIL
- **Duration**: 94.582208ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Request Payload**:
```json
{
  "after": "sample_string",
  "ref": "sample_string",
  "sender": {
    "email": "sample_string",
    "name": "sample_string",
    "username": "sample_string"
  }
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"git host not found: sample_string","userMessage":"git host with ID 'sample_string' not found","userDetailMessage":"The req...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/git/{gitHostId}/{secret}

- **Status**: FAIL
- **Duration**: 89.362959ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Request Payload**:
```json
{
  "before": "sample_string",
  "head_commit": {
    "added": [
      "sample_string"
    ],
    "removed": [
      "sample_string"
    ],
    "timestamp": "2023-01-01T00:00:00Z"
  },
  "repository": {
    "fork": true,
    "full_name": "sample_string",
    "html_url": "sample_string"
  }
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"git host not found: sample_string","userMessage":"git host with ID 'sample_string' not found","userDetailMessage":"The req...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/ci/workflow

- **Status**: FAIL
- **Duration**: 89.049209ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Request Payload**:
```json
{
  "finishedOn": "2023-01-01T00:00:00Z",
  "message": "sample_string",
  "pipelineId": 1,
  "startedOn": "2023-01-01T00:00:00Z",
  "status": "SUCCESS"
}
```
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{invalid wf name}]","userMessage":0}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/refresh

- **Status**: FAIL
- **Duration**: 89.519166ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: internal server error

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/user/role/cache

- **Status**: PASS
- **Duration**: 298.576083ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":"{}"}

---

### ✅ GET /orchestrator/user/role/cache/invalidate

- **Status**: PASS
- **Duration**: 90.146083ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":"Cache Cleaned Successfully"}

---

### ❌ GET /orchestrator/api/dex/{path}

- **Status**: FAIL
- **Duration**: 91.702166ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 502
- **Error/Msg**: {}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 502

---

### ❌ POST /orchestrator/api/v1/session

- **Status**: FAIL
- **Duration**: 87.981ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**:
```json
{
  "password": "sample_string",
  "username": "sample_string"
}
```
- **Response Code**: 403
- **Error/Msg**: {"code":403,"status":"Forbidden","errors":[{"code":"000","internalMessage":"[{invalid username or password}]","userMessage":"invalid username or password"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 403

---

### ❌ GET /orchestrator/auth/callback

- **Status**: FAIL
- **Duration**: 315.689917ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: Failed to query provider "https://devtron.example.com/orchestrator/api/dex": Get "https://devtron.example.com/orchestrator/api/dex/.well-known/openid-configuration": dial tcp: lookup devtron.example.c...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/auth/login

- **Status**: FAIL
- **Duration**: 323.822875ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: Failed to query provider "https://devtron.example.com/orchestrator/api/dex": Get "https://devtron.example.com/orchestrator/api/dex/.well-known/openid-configuration": dial tcp: lookup devtron.example.c...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'F' looking for beginning of value

---

### ❌ GET /orchestrator/login

- **Status**: FAIL
- **Duration**: 326.905875ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: Failed to query provider "https://devtron.example.com/orchestrator/api/dex": Get "https://devtron.example.com/orchestrator/api/dex/.well-known/openid-configuration": dial tcp: lookup devtron.example.c...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'F' looking for beginning of value

---

### ❌ GET /orchestrator/rbac/roles/default

- **Status**: FAIL
- **Duration**: 161.593083ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /devtron/auth/verify/v2

- **Status**: FAIL
- **Duration**: 89.661125ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: <html>
<head><title>404 Not Found</title></head>
<body>
<center><h1>404 Not Found</h1></center>
<hr><center>nginx</center>
</body>
</html>


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/devtron/auth/verify

- **Status**: PASS
- **Duration**: 87.457042ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":true}

---

### ✅ GET /orchestrator/user/check/roles

- **Status**: PASS
- **Duration**: 94.045208ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"roles":["role:super-admin___","role:chart-group_admin"],"superAdmin":true}}

---

### ❌ GET /orchestrator/app/history/deployed-component/list/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 89.502167ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","result":"invalid historyComponent"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/history/deployed-configuration/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 92.770208ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":8,"name":"DEPLOYMENT_TEMPLATE"},{"id":3,"name":"PIPELINE_STRATEGY"}]}

---

### ❌ GET /orchestrator/app/history/deployed-component/detail/{appId}/{pipelineId}/{id}

- **Status**: FAIL
- **Duration**: 90.186875ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","result":"invalid historyComponent"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/batch/v1beta1/cd-pipeline

- **Status**: FAIL
- **Duration**: 229.974125ms
- **Spec File**: ../../specs/environment/bulk-delete.yaml
- **Request Payload**:
```json
{
  "action": 1,
  "envNames": [
    "sample_string"
  ],
  "nonCascadeDelete": true
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","userMessage":"invalid payload, can not get pipelines for this filter"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/gitops/configured

- **Status**: PASS
- **Duration**: 92.171875ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"allowCustomRepository":false,"authMode":"PASSWORD","exists":true,"provider":"GITHUB"}}

---

### ❌ POST /orchestrator/gitops/validate

- **Status**: FAIL
- **Duration**: 136.543333ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Request Payload**:
```json
{
  "host": "sample_string",
  "id": 1,
  "isTLSKeyDataPresent": true
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/gitops/config

- **Status**: PASS
- **Duration**: 95.802209ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"provider":"GITHUB","username":"systemsdt","token":"","gitLabGroupId":"","gitHubOrgId":"stage-gitops","host":"https://github.com/","active":true,"azureProje...

---

### ❌ POST /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 91.1045ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Request Payload**:
```json
{
  "active": true,
  "azureProjectId": "",
  "gitHubOrgId": "sample-org",
  "gitLabGroupId": "",
  "host": "https://github.com",
  "provider": "GITHUB",
  "token": "sample-token",
  "username": "sample-user"
}
```
- **Response Code**: 409
- **Error/Msg**: {"code":409,"status":"Conflict","errors":[{"internalMessage":"gitops provider already exists","userMessage":"gitops provider already exists"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 409

---

### ❌ PUT /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 91.042042ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Request Payload**:
```json
{
  "enableTLSVerification": true,
  "provider": "GITLAB",
  "token": "sample_string"
}
```
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"internalMessage":"gitops config update failed, does not exist","userMessage":"gitops config update failed, does not exist"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/gitops/config-by-provider

- **Status**: FAIL
- **Duration**: 88.602333ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/gitops/config/{id}

- **Status**: PASS
- **Duration**: 90.995375ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"provider":"GITHUB","username":"systemsdt","token":"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","gitLabGro...

---

### ❌ PATCH /orchestrator/app/env/patch

- **Status**: FAIL
- **Duration**: 92.679708ms
- **Spec File**: ../../specs/helm/deployment-chart-type.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "currentViewEditor": "BASIC",
  "envId": 1,
  "isBasicViewLocked": true,
  "targetChartRefId": 1,
  "template": {}
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","result":"env properties not found"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/k8s/resources/ephemeralContainers

- **Status**: FAIL
- **Duration**: 86.8225ms
- **Spec File**: ../../specs/kubernetes/ephemeral-containers.yaml
- **Request Payload**:
```json
{
  "advancedData": {
    "manifest": "sample_string"
  },
  "basicData": {
    "containerName": "sample_string",
    "image": "sample_string",
    "targetContainerName": "sample_string"
  },
  "clusterId": 1,
  "namespace": "sample_string",
  "podName": "sample_string",
  "userId": 1
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/k8s/resources/ephemeralContainers

- **Status**: FAIL
- **Duration**: 86.973084ms
- **Spec File**: ../../specs/kubernetes/ephemeral-containers.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /orchestrator/application/rollback

- **Status**: FAIL
- **Duration**: 554.401792ms
- **Spec File**: ../../specs/openapiClient/api/openapi.yaml
- **Request Payload**:
```json
{
  "hAppId": "1|default|someName",
  "installedAppId": 1,
  "installedAppVersionId": 2,
  "version": 10
}
```
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"500","internalMessage":"release name is invalid: someName","userMessage":"release name is invalid: someName"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrator/application/template-chart

- **Status**: FAIL
- **Duration**: 722.849583ms
- **Spec File**: ../../specs/openapiClient/api/openapi.yaml
- **Request Payload**:
```json
{
  "appStoreApplicationVersionId": 10,
  "clusterId": 1,
  "environmentId": 1,
  "namespace": "1",
  "releaseName": "some name",
  "valuesYaml": "some values yaml"
}
```
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{rpc error: code = Unknown desc = error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string i...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrator/admin/policy/default

- **Status**: FAIL
- **Duration**: 87.321ms
- **Spec File**: ../../specs/security/policy-management.yaml
- **Request Payload**:
```json
{
  "policies": [
    {
      "description": "sample_string",
      "name": "sample_string",
      "rules": [
        {
          "action": "sample_string",
          "effect": "allow",
          "resource": "sample_string"
        },
        {
          "action": "sample_string",
          "effect": "allow",
          "resource": "sample_string"
        }
      ]
    },
    {
      "description": "sample_string",
      "name": "sample-item-2",
      "rules": [
        {
          "action": "sample_string",
          "effect": "allow",
          "resource": "sample_string"
        },
        {
          "action": "sample_string",
          "effect": "allow",
          "resource": "sample_string"
        }
      ]
    }
  ],
  "roles": [
    {
      "description": "sample_string",
      "name": "sample_string",
      "policies": [
        "sample_string"
      ]
    },
    {
      "description": "sample_string",
      "name": "sample-item-2",
      "policies": [
        "sample_string"
      ]
    }
  ]
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/validate

- **Status**: FAIL
- **Duration**: 87.062834ms
- **Spec File**: ../../specs/security/policy-management.yaml
- **Request Payload**:
```json
{
  "dryRun": true,
  "gitOpsConfig": {}
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ PUT /orchestrator/sso/update

- **Status**: PASS
- **Duration**: 110.429375ms
- **Spec File**: ../../specs/sso/configuration.yaml
- **Request Payload**:
```json
{
  "active": true,
  "config": {},
  "id": 1,
  "label": "sample_string",
  "name": "sample_string",
  "url": "https://devtron.example.com/orchestrator"
}
```
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"name":"sample_string","label":"sample_string","url":"https://devtron.example.com/orchestrator","config":{"id":"","type":"","name":"","config":null},"active"...

---

### ✅ GET /orchestrator/sso/{id}

- **Status**: PASS
- **Duration**: 91.067125ms
- **Spec File**: ../../specs/sso/configuration.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"name":"google","label":"sample_string","url":"https://devtron.example.com/orchestrator","config":{"id":"","type":"","name":"","config":null},"active":true,"...

---

### ❌ GET /orchestrator/sso

- **Status**: FAIL
- **Duration**: 87.68575ms
- **Spec File**: ../../specs/sso/configuration.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ✅ POST /orchestrator/sso/create

- **Status**: PASS
- **Duration**: 106.093084ms
- **Spec File**: ../../specs/sso/configuration.yaml
- **Request Payload**:
```json
{
  "active": true,
  "config": {},
  "id": 1,
  "label": "sample_string",
  "name": "sample_string",
  "url": "https://devtron.example.com/orchestrator"
}
```
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":3,"name":"sample_string","label":"sample_string","url":"https://devtron.example.com/orchestrator","config":{},"active":true,"globalAuthConfigType":""}}

---

### ✅ GET /orchestrator/sso/list

- **Status**: PASS
- **Duration**: 88.134875ms
- **Spec File**: ../../specs/sso/configuration.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":3,"name":"sample_string","label":"sample_string","url":"https://devtron.example.com/orchestrator","active":true,"globalAuthConfigType":""},{"id":1,"name":"goo...

---

### ✅ GET /orchestrator/chartref/autocomplete/{appId}

- **Status**: PASS
- **Duration**: 95.621708ms
- **Spec File**: ../../specs/helm/dynamic-charts.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"chartRefs":[{"id":10,"version":"3.9.0","name":"Rollout Deployment","description":"","userUploaded":false,"isAppMetricsSupported":true},{"id":11,"version":"3.10.0",...

---

### ❌ POST /orchestrator/user/resource/options/{kind}/{version}

- **Status**: FAIL
- **Duration**: 87.721833ms
- **Spec File**: ../../specs/userResource/userResource.yaml
- **Request Payload**:
```json
{
  "emailId": "test@example.com",
  "groups": [],
  "roleFilters": [],
  "superAdmin": false
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/git/host/event/{eventId}

- **Status**: FAIL
- **Duration**: 156.442458ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"{eventId}\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"{eventId}\": inv...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/webhook-meta-config/{gitProviderId}

- **Status**: FAIL
- **Duration**: 91.231084ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{ERROR #22P02 invalid input syntax for type integer: \"{gitProviderId}\"}]","userMessage":"ERROR #22P02 invalid...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/{id}

- **Status**: FAIL
- **Duration**: 91.028291ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"{id}\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"{id}\": invalid synta...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/{id}/event

- **Status**: FAIL
- **Duration**: 90.275ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"{id}\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"{id}\": invalid synta...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/git/host

- **Status**: PASS
- **Duration**: 91.709084ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"Github","active":true,"webhookUrl":"","webhookSecret":"","eventTypeHeader":"","secretHeader":"","secretValidator":""},{"id":2,"name":"Bitbucket Clou...

---

### ❌ POST /orchestrator/git/host

- **Status**: FAIL
- **Duration**: 92.208083ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Request Payload**:
```json
{
  "active": true,
  "id": 1,
  "name": "sample_string",
  "url": "sample_string"
}
```
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{sample_string: git host already exists}]","userMessage":"sample_string: git host already exists"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ PUT /orchestrator/deployment/template/upload

- **Status**: PASS
- **Duration**: 90.115125ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Request Payload**:
```json
{
  "action": "sample_string",
  "fileId": "sample_string"
}
```
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":"Processed successfully"}

---

### ❌ POST /orchestrator/deployment/template/validate

- **Status**: FAIL
- **Duration**: 91.244041ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Request Payload**:
```json
{}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{request Content-Type isn't multipart/form-data}]","userMessage":"request Content-Type isn't multipart/form-data"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/deployment/template/fetch

- **Status**: PASS
- **Duration**: 98.368625ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":10,"name":"Rollout Deployment","chartDescription":"This chart deploys an advanced version of deployment that supports Blue/Green and Canary deployments. For f...

---

### ❌ GET /orchestrator/deployment/pipeline/history

- **Status**: FAIL
- **Duration**: 86.679166ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/deployment/pipeline/rollback

- **Status**: FAIL
- **Duration**: 86.742ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Request Payload**:
```json
{
  "pipelineId": 1,
  "triggeredBy": 1,
  "version": "sample_string"
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/deployment/pipeline/trigger

- **Status**: FAIL
- **Duration**: 88.074417ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Request Payload**:
```json
{
  "artifactId": 1,
  "pipelineId": 1,
  "triggeredBy": 1
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/deployment/pipeline/configure

- **Status**: FAIL
- **Duration**: 89.202458ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "envId": 1,
  "postDeploymentScript": "sample_string",
  "preDeploymentScript": "sample_string",
  "triggerType": "AUTOMATIC"
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/external-links/tools

- **Status**: PASS
- **Duration**: 94.653875ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"Grafana","icon":"","category":2},{"id":2,"name":"Kibana","icon":"","category":2},{"id":3,"name":"Newrelic","icon":"","category":2},{"id":4,"name":"C...

---

### ✅ GET /orchestrator/external-links

- **Status**: PASS
- **Duration**: 94.418625ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":2,"name":"sample-external-link","url":"https://grafana.example.com","active":true,"monitoringToolId":1,"type":"appLevel","identifiers":[{"type":"devtron-app",...

---

### ✅ POST /orchestrator/external-links

- **Status**: PASS
- **Duration**: 99.673625ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Request Payload**:
```json
[
  {
    "description": "Sample external link for testing",
    "id": 0,
    "identifiers": [
      {
        "clusterId": 1,
        "identifier": "1",
        "type": "devtron-app"
      }
    ],
    "isEditable": true,
    "monitoringToolId": 1,
    "name": "sample-external-link",
    "type": "appLevel",
    "url": "https://grafana.example.com"
  }
]
```
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"success":true}}

---

### ✅ PUT /orchestrator/external-links

- **Status**: PASS
- **Duration**: 100.78025ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Request Payload**:
```json
{
  "description": "Updated external link for testing",
  "id": 1,
  "identifiers": [
    {
      "clusterId": 1,
      "identifier": "1",
      "type": "devtron-app"
    }
  ],
  "isEditable": true,
  "monitoringToolId": 1,
  "name": "updated-external-link",
  "type": "appLevel",
  "url": "https://grafana-updated.example.com"
}
```
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"success":true}}

---

### ❌ DELETE /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 88.05625ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ✅ GET /orchestrator/cluster

- **Status**: PASS
- **Duration**: 99.320208ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":6,"cluster_name":"abcd","description":"","active":true,"config":{"bearer_token":""},"prometheusAuth":{"isAnonymous":false},"defaultClusterComponent":[],"agent...

---

### ❌ PUT /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 95.172083ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Request Payload**:
```json
{
  "createdBy": "sample_string",
  "prometheusAuth": {
    "tlsClientCert": "sample_string",
    "tlsClientKey": "sample_string",
    "userName": "sample_string"
  },
  "server_url": "sample_string"
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'ClusterBean.ClusterName' Error:Field validation for 'ClusterName' failed on the 'required' tag","userMessage":"Key: 'ClusterBean....

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/cluster/auth-list

- **Status**: PASS
- **Duration**: 91.8025ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":6,"cluster_name":"abcd","description":"","active":false,"defaultClusterComponent":null,"agentInstallationStage":0,"k8sVersion":"","insecureSkipTlsVerify":fals...

---

### ✅ POST /orchestrator/app/edit/projects

- **Status**: PASS
- **Duration**: 176.013375ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "projectIds": [
    1
  ]
}
```
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":null}

---

### ❌ POST /orchestrator/app

- **Status**: FAIL
- **Duration**: 98.401708ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**:
```json
{
  "appName": "sample_string",
  "description": "sample_string",
  "projectIds": [
    1
  ],
  "teamId": 1
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'CreateAppDTO.AppName' Error:Field validation for 'AppName' failed on the 'name-component' tag","userMessage":"Key: 'CreateAppDTO....

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/autocomplete

- **Status**: PASS
- **Duration**: 91.469708ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"helm-sanity-application","createdBy":"admin","description":""},{"id":2,"name":"gitops-sanity-application","createdBy":"admin","description":""},{"id...

---

### ❌ POST /orchestrator/app/edit

- **Status**: FAIL
- **Duration**: 99.236041ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**:
```json
{
  "appName": "sample_string",
  "description": "sample_string",
  "id": 1,
  "labels": [
    {
      "key": "sample_string",
      "propagate": true,
      "value": "sample_string"
    },
    {
      "key": "sample_string",
      "propagate": true,
      "value": "sample_string"
    }
  ]
}
```
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{duplicate key found for app 1, sample_string}]","userMessage":"duplicate key found for app 1, sample_string"}]...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/app/details/{appId}

- **Status**: FAIL
- **Duration**: 87.006833ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ✅ POST /orchestrator/app/workflow/clone

- **Status**: PASS
- **Duration**: 91.68075ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**:
```json
{
  "sourceAppId": 1,
  "targetAppId": 1,
  "workflowName": "sample_string"
}
```
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"status":"FAILED","message":"pg: no rows in result set"}}

---

### ✅ POST /orchestrator/app/list

- **Status**: PASS
- **Duration**: 93.862417ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**:
```json
{
  "environmentIds": [
    1
  ],
  "offset": 1,
  "projectIds": [
    1
  ]
}
```
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"appContainers":null,"appCount":0,"deploymentGroup":{"id":0,"name":"","appCount":0,"noOfApps":"","environmentId":0,"ciPipelineId":0,"ciMaterialDTOs":null,"isVirtual...

---

### ❌ POST /orchestrator/app/workflow

- **Status**: FAIL
- **Duration**: 85.809792ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "cdPipelines": [
    {},
    {}
  ],
  "tree": [
    {},
    {}
  ]
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/app/workflow/{app-wf-id}/app/{app-id}

- **Status**: FAIL
- **Duration**: 87.168875ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ POST /orchestrator/core/v1beta1/application

- **Status**: FAIL
- **Duration**: 92.415542ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**:
```json
{
  "appWorkflows": [
    {},
    {}
  ],
  "environmentOverrides": [
    {},
    {}
  ],
  "metadata": {
    "appName": "sample_string",
    "id": 1,
    "labels": [
      {
        "key": "sample_string",
        "propagate": true,
        "value": "sample_string"
      },
      {
        "key": "sample_string",
        "propagate": true,
        "value": "sample_string"
      }
    ],
    "projectIds": [
      1
    ]
  }
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'AppDetail.Metadata.ProjectName' Error:Field validation for 'ProjectName' failed on the 'required' tag","userMessage":"Key: 'AppDe...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PATCH /orchestrator/app/ci-pipeline/patch-source

- **Status**: FAIL
- **Duration**: 172.338ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-change-source.yaml
- **Request Payload**:
```json
{
  "appId": 16,
  "environmentId": 1,
  "source": {
    "regex": "feature-*",
    "type": "SOURCE_TYPE_BRANCH_FIXED",
    "value": "main"
  }
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/k8s/resources/apply

- **Status**: FAIL
- **Duration**: 89.287917ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Request Payload**:
```json
{
  "clusterId": 1,
  "manifest": "sample_string",
  "namespace": "sample_string"
}
```
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{failed to unmarshal manifest: error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrator/k8s/resources/rotate

- **Status**: FAIL
- **Duration**: 85.833833ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Request Payload**:
```json
{
  "clusterId": 1,
  "deploymentName": "sample_string",
  "namespace": "sample_string"
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/k8s/api-resources/{clusterId}

- **Status**: PASS
- **Duration**: 202.201792ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"apiResources":[{"gvk":{"Group":"","Version":"v1","Kind":"PersistentVolume"},"gvr":{"Group":"","Version":"v1","Resource":"persistentvolumes"},"namespaced":false,"sh...

---

### ❌ POST /orchestrator/k8s/resource

- **Status**: FAIL
- **Duration**: 192.259583ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "k8sRequest": {
    "patch": "sample_string",
    "resourceIdentifier": {
      "groupVersionKind": {
        "group": "sample_string",
        "kind": "sample_string",
        "version": "sample_string"
      },
      "name": "sample_string",
      "namespace": "sample_string"
    }
  }
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{json: cannot unmarshal number into Go struct field ResourceRequestBean.appId of type string}]","userMessage":"json: cann...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/k8s/resource

- **Status**: FAIL
- **Duration**: 89.905375ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "k8sRequest": {
    "patch": "sample_string",
    "resourceIdentifier": {
      "groupVersionKind": {
        "group": "sample_string",
        "kind": "sample_string",
        "version": "sample_string"
      },
      "name": "sample_string",
      "namespace": "sample_string"
    }
  }
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{json: cannot unmarshal number into Go struct field ResourceRequestBean.appId of type string}]","userMessage":"json: cann...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/k8s/resource/create

- **Status**: FAIL
- **Duration**: 89.378708ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "k8sRequest": {
    "patch": "sample_string",
    "resourceIdentifier": {
      "groupVersionKind": {
        "group": "sample_string",
        "kind": "sample_string",
        "version": "sample_string"
      },
      "name": "sample_string",
      "namespace": "sample_string"
    }
  }
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{json: cannot unmarshal number into Go struct field ResourceRequestBean.appId of type string}]","userMessage":"json: cann...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/k8s/resource/delete

- **Status**: FAIL
- **Duration**: 91.222875ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "k8sRequest": {
    "patch": "sample_string",
    "resourceIdentifier": {
      "groupVersionKind": {
        "group": "sample_string",
        "kind": "sample_string",
        "version": "sample_string"
      },
      "name": "sample_string",
      "namespace": "sample_string"
    }
  }
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{json: cannot unmarshal number into Go struct field ResourceRequestBean.appId of type string}]","userMessage":"json: cann...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/k8s/resource/inception/info

- **Status**: FAIL
- **Duration**: 164.077208ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"404","userMessage":"error on getting resource from k8s"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/k8s/resource/urls

- **Status**: FAIL
- **Duration**: 87.45925ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/api-token/webhook

- **Status**: PASS
- **Duration**: 94.145541ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[]}

---

### ❌ PUT /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 91.129208ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Request Payload**:
```json
{
  "description": "Updated API token for testing",
  "id": 1,
  "name": "updated-api-token"
}
```
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{api-token corresponds to apiTokenId '1' is not found}]","userMessage":"api-token corresponds to apiTokenId '1'...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ DELETE /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 91.988417ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{api-token corresponds to apiTokenId '1' is not found}]","userMessage":"api-token corresponds to apiTokenId '1'...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/api-token

- **Status**: PASS
- **Duration**: 91.724667ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"description":"some description","expireAtInMs":12344546,"id":2,"name":"some-name","token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6IkFQSS1UT0tFTjpzb21lLW...

---

### ❌ POST /orchestrator/api-token

- **Status**: FAIL
- **Duration**: 92.457625ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Request Payload**:
```json
{
  "description": "Sample API token for testing",
  "expireAtInMs": 1735689600000,
  "name": "sample-api-token"
}
```
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{name 'sample-api-token' is already used. please use another name}]","userMessage":"name 'sample-api-token' is ...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/user/role/group/detailed/get

- **Status**: PASS
- **Duration**: 92.542125ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":2,"roleFilters":[],"superAdmin":false,"canManageAllAccess":false,"hasAccessManagerPermission":false}]}

---

### ❌ GET /orchestrator/user/role/group/search

- **Status**: FAIL
- **Duration**: 91.0785ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"11001","internalMessage":"invalid path parameter id: search","userMessage":"Invalid path parameter 'id'","userDetailMessage":"Please check the par...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 91.744375ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Request Payload**:
```json
{
  "emailId": "test@example.com",
  "groups": [],
  "roleFilters": [],
  "superAdmin": false
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/user/role/group/v2

- **Status**: PASS
- **Duration**: 171.147542ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"roleGroups":[{"id":2,"roleFilters":[],"superAdmin":false,"canManageAllAccess":false,"hasAccessManagerPermission":false}],"totalCount":1}}

---

### ❌ POST /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 92.886ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Request Payload**:
```json
{
  "emailId": "test@example.com",
  "groups": [],
  "roleFilters": [],
  "superAdmin": false
}
```
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{role group already exist}]","userMessage":"role group already exist"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ DELETE /orchestrator/user/role/group/{id}

- **Status**: FAIL
- **Duration**: 90.119042ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/user/role/group/{id}

- **Status**: FAIL
- **Duration**: 91.496834ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"role group not found: 1","userMessage":"role group with ID '1' not found","userDetailMessage":"The requested role group do...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/user/role/group

- **Status**: PASS
- **Duration**: 89.943916ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":2,"roleFilters":[],"superAdmin":false,"canManageAllAccess":false,"hasAccessManagerPermission":false}]}

---

### ❌ POST /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 93.2175ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Request Payload**:
```json
{
  "emailId": "test@example.com",
  "groups": [],
  "roleFilters": [],
  "superAdmin": false
}
```
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{role group already exist}]","userMessage":"role group already exist"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PUT /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 90.935375ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Request Payload**:
```json
{
  "emailId": "test@example.com",
  "groups": [],
  "roleFilters": [],
  "superAdmin": false
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/user/role/group/bulk

- **Status**: FAIL
- **Duration**: 97.880041ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{EOF}]","userMessage":"EOF"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/v2/{id}

- **Status**: PASS
- **Duration**: 95.575417ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"email_id":"system","roleFilters":[],"groups":[],"superAdmin":false,"userRoleGroups":[],"lastLoginTime":"0001-01-01T00:00:00Z","timeoutWindowExpression":"000...

---

### ❌ DELETE /orchestrator/user/{id}

- **Status**: FAIL
- **Duration**: 91.879ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","userMessage":"cannot delete system or admin user"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/{id}

- **Status**: PASS
- **Duration**: 167.241167ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"email_id":"system","roleFilters":[],"groups":[],"superAdmin":false,"userRoleGroups":[],"lastLoginTime":"0001-01-01T00:00:00Z","timeoutWindowExpression":"000...

---

### ❌ POST /orchestrator/user

- **Status**: FAIL
- **Duration**: 90.412417ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Request Payload**:
```json
{
  "emailId": "test@example.com",
  "groups": [],
  "roleFilters": [],
  "superAdmin": false
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'UserInfo.EmailId' Error:Field validation for 'EmailId' failed on the 'required' tag","userMessage":"Key: 'UserInfo.EmailId' Error...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/user

- **Status**: FAIL
- **Duration**: 90.998125ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Request Payload**:
```json
{
  "emailId": "test@example.com",
  "groups": [],
  "roleFilters": [],
  "superAdmin": false
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'UserInfo.EmailId' Error:Field validation for 'EmailId' failed on the 'required' tag","userMessage":"Key: 'UserInfo.EmailId' Error...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/user/bulk

- **Status**: FAIL
- **Duration**: 89.987ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{EOF}]","userMessage":"EOF"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/detail/get

- **Status**: PASS
- **Duration**: 92.620375ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"email_id":"system","roleFilters":[],"groups":[],"superAdmin":false,"userRoleGroups":null,"lastLoginTime":"0001-01-01T00:00:00Z","timeoutWindowExpression":"...

---

### ✅ GET /orchestrator/user/sync/orchestratortocasbin

- **Status**: PASS
- **Duration**: 92.187875ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":true}

---

### ❌ PUT /orchestrator/user/update/trigger/terminal

- **Status**: FAIL
- **Duration**: 85.957167ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Request Payload**:
```json
{
  "emailId": "test@example.com",
  "groups": [],
  "roleFilters": [],
  "superAdmin": false
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/user/v2

- **Status**: PASS
- **Duration**: 93.485459ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"users":[{"id":2,"email_id":"admin","roleFilters":[],"groups":[],"superAdmin":false,"userRoleGroups":[],"lastLoginTime":"2025-08-08T08:09:36.216299Z","timeoutWindow...

---

### ❌ GET /orchestrator/env/clusters

- **Status**: FAIL
- **Duration**: 91.713166ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/env/namespace/autocomplete

- **Status**: FAIL
- **Duration**: 91.293125ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{ERROR #42601 syntax error at or near \")\"}]","userMessage":"ERROR #42601 syntax error at or near \")\""}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/cluster

- **Status**: PASS
- **Duration**: 95.736584ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":6,"cluster_name":"abcd","description":"","active":true,"config":{"bearer_token":""},"prometheusAuth":{"isAnonymous":false},"defaultClusterComponent":[],"agent...

---

### ❌ PUT /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 91.933875ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**:
```json
{
  "cluster_name": "sample_string",
  "created_by": "sample_string",
  "prometheus_url": "sample_string",
  "server_url": "sample_string",
  "updated_on": "2023-01-01T00:00:00Z"
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'ClusterBean.ServerUrl' Error:Field validation for 'ServerUrl' failed on the 'url' tag","userMessage":"Key: 'ClusterBean.ServerUrl...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/cluster/delete

- **Status**: FAIL
- **Duration**: 88.224375ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**:
```json
{
  "forceDelete": true,
  "id": 1
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/cluster/{cluster_id}/env

- **Status**: FAIL
- **Duration**: 87.629458ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ✅ GET /orchestrator/env

- **Status**: PASS
- **Duration**: 99.820333ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"environment_name":"devtron-demo","cluster_id":1,"cluster_name":"default_cluster","active":true,"default":false,"namespace":"devtron-demo","isClusterCdActiv...

---

### ❌ POST /orchestrator/env

- **Status**: FAIL
- **Duration**: 100.763125ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**:
```json
{
  "cluster_id": 1,
  "created_by": "sample_string",
  "description": "sample_string",
  "environment_name": "sample_string",
  "namespace": "sample_string",
  "updated_on": "2023-01-01T00:00:00Z"
}
```
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"422","internalMessage":"Namespace \"sample_string\" is invalid: metadata.name: Invalid value: \"sample_string\": a lowercase RFC 1123 la...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PUT /orchestrator/env

- **Status**: FAIL
- **Duration**: 91.506541ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**:
```json
{
  "cluster_id": 1,
  "created_by": "sample_string",
  "default": true,
  "environment_name": "sample_string",
  "namespace": "sample_string",
  "updated_on": "2023-01-01T00:00:00Z"
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/template/default/{appId}/{chartRefId}

- **Status**: FAIL
- **Duration**: 92.812667ms
- **Spec File**: ../../specs/environment/templates.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/security/scan/executionDetail

- **Status**: FAIL
- **Duration**: 90.011208ms
- **Spec File**: ../../specs/security/security-dashboard-apis.yml
- **Request Payload**: (none)
- **Response Code**: 403
- **Error/Msg**: {"code":403,"status":"Forbidden","errors":[{"code":"000","internalMessage":"[{unauthorized user}]","userMessage":"Unauthorized User"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 403

---

### ❌ POST /orchestrator/app/cd-pipeline/patch/deployment

- **Status**: FAIL
- **Duration**: 91.518084ms
- **Spec File**: ../../specs/deployment/app-type-change.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "envId": 1,
  "targetChartRefId": 1
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'DeploymentAppTypeChangeRequest.DesiredDeploymentType' Error:Field validation for 'DesiredDeploymentType' failed on the 'required'...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/resource/history/deployment/cd-pipeline/v1

- **Status**: FAIL
- **Duration**: 95.45275ms
- **Spec File**: ../../specs/deployment/history.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","internalMessage":"invalid format filter criteria!","userMessage":"invalid format filter criteria!"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/gitops/configured

- **Status**: PASS
- **Duration**: 152.235417ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"allowCustomRepository":false,"authMode":"PASSWORD","exists":true,"provider":"GITHUB"}}

---

### ❌ POST /orchestrator/validate

- **Status**: FAIL
- **Duration**: 86.320042ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Request Payload**:
```json
{
  "allowCustomRepository": true,
  "host": "sample_string",
  "isCADataPresent": true
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 90.920959ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Request Payload**:
```json
{
  "active": true,
  "azureProjectId": "",
  "gitHubOrgId": "sample-org",
  "gitLabGroupId": "",
  "host": "https://github.com",
  "provider": "GITHUB",
  "token": "sample-token",
  "username": "sample-user"
}
```
- **Response Code**: 409
- **Error/Msg**: {"code":409,"status":"Conflict","errors":[{"internalMessage":"gitops provider already exists","userMessage":"gitops provider already exists"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 409

---

### ❌ PUT /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 90.132333ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Request Payload**:
```json
{
  "gitHubOrgId": "sample_string",
  "gitLabGroupId": "sample_string",
  "provider": "GITLAB"
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/gitops/config

- **Status**: PASS
- **Duration**: 90.611833ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"provider":"GITHUB","username":"systemsdt","token":"","gitLabGroupId":"","gitHubOrgId":"stage-gitops","host":"https://github.com/","active":true,"azureProje...

---

### ❌ GET /orchestrator/gitops/config-by-provider

- **Status**: FAIL
- **Duration**: 86.838042ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ✅ GET /orchestrator/gitops/config/{id}

- **Status**: PASS
- **Duration**: 90.454417ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"provider":"GITHUB","username":"systemsdt","token":"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","gitLabGro...

---

### ✅ POST /orchestrator/chart-repo/validate

- **Status**: PASS
- **Duration**: 99.22275ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
- **Request Payload**:
```json
{
  "active": true,
  "allow_insecure_connection": false,
  "authMode": "ANONYMOUS",
  "default": false,
  "name": "sample-chart-repo",
  "url": "https://charts.example.com"
}
```
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"customErrMsg":"Could not validate the repo. Please try again.","actualErrMsg":"Get \"https://charts.example.com/index.yaml\": dial tcp: lookup charts.example.com o...

---

### ✅ POST /orchestrator/chart-repo/create

- **Status**: PASS
- **Duration**: 100.167667ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
- **Request Payload**:
```json
{
  "active": true,
  "allow_insecure_connection": false,
  "authMode": "ANONYMOUS",
  "default": false,
  "name": "sample-chart-repo",
  "url": "https://charts.example.com"
}
```
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"customErrMsg":"Could not validate the repo. Please try again.","actualErrMsg":"Get \"https://charts.example.com/index.yaml\": dial tcp: lookup charts.example.com o...

---

### ❌ POST /orchestrator/chart-repo/update

- **Status**: FAIL
- **Duration**: 333.796667ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
- **Request Payload**:
```json
{
  "active": true,
  "allow_insecure_connection": false,
  "authMode": "ANONYMOUS",
  "default": false,
  "name": "sample-chart-repo",
  "url": "https://charts.example.com"
}
```
- **Response Code**: 412
- **Error/Msg**: {"code":412,"status":"Precondition Failed","result":{"customErrMsg":"Could not validate the repo. Please try again.","actualErrMsg":"Get \"https://charts.example.com/index.yaml\": dial tcp: lookup cha...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 412

---

### ❌ GET /orchestrator/plugin/global/list/global-variable

- **Status**: FAIL
- **Duration**: 88.818875ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"\": invalid syntax}]","userMessage":"invalid appId"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/plugin/global/list/v2/min

- **Status**: PASS
- **Duration**: 92.993459ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[]}

---

### ✅ PUT /orchestrator/plugin/global/migrate

- **Status**: PASS
- **Duration**: 96.540416ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK"}

---

### ❌ POST /orchestrator/plugin/global/create

- **Status**: FAIL
- **Duration**: 91.922667ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Request Payload**:
```json
{
  "icon": "sample_string",
  "id": 1,
  "name": "sample_string"
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","internalMessage":"no step data provided to save, please provide a plugin step to proceed further","userMessage":"no step data provided to sa...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/plugin/global/list/tags

- **Status**: PASS
- **Duration**: 94.391208ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"tagNames":["GCP","Code quality","Code Review","Image source","CI task","Kubernetes","Github","Security","DevSecOps","cloud","Release","Load testing","gcs","AWS EKS...

---

### ✅ GET /orchestrator/plugin/global/detail/all

- **Status**: PASS
- **Duration**: 488.811417ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"K6 Load testing","description":"K6 is an open-source tool and cloud service that makes load testing easy for developers and QA engineers.","type":"P...

---

### ✅ GET /orchestrator/plugin/global/detail/{pluginId}

- **Status**: PASS
- **Duration**: 95.041917ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"name":"K6 Load testing","description":"K6 is an open-source tool and cloud service that makes load testing easy for developers and QA engineers.","type":"PR...

---

### ❌ POST /orchestrator/plugin/global/list/detail/v2

- **Status**: FAIL
- **Duration**: 93.015542ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "parentPluginIdentifiers": [
    "sample_string"
  ]
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","internalMessage":"Empty values for both pluginVersionIds and parentPluginIds. Please provide at least one of them","userMessage":"Empty valu...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/plugin/global/list/v2

- **Status**: PASS
- **Duration**: 113.461917ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"parentPlugins":[{"id":4,"name":"AWS ECR Retag","pluginIdentifier":"aws-retag","description":"AWS ECR Retag plugin that enables retagging of container images within...

---


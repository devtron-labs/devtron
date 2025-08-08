# API Spec Validation Report

Generated: 2025-08-08T11:34:18+05:30

## Summary

- Total Endpoints: 213
- Passed: 78
- Failed: 135
- Warnings: 0
- Success Rate: 36.62%

## Detailed Results

### ✅ GET /orchestrator/notification/channel/smtp/{id}

- **Status**: PASS
- **Duration**: 421.526542ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":0,"port":"","host":"","authType":"","authUser":"","authPassword":"","fromEmail":"","configName":"","description":"","ownerId":0,"default":false,"deleted":false...

---

### ❌ POST /orchestrator/notification/channel

- **Status**: FAIL
- **Duration**: 98.386917ms
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
- **Duration**: 89.774166ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{EOF}]","userMessage":"EOF"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/notification/channel

- **Status**: PASS
- **Duration**: 87.579917ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"slackConfigs":[],"webhookConfigs":[],"sesConfigs":[{"userId":2,"teamId":0,"region":"ap-south-1","accessKey":"AKIAWEAVHFOOSNAORPHU","secretKey":"**********","fromEm...

---

### ❌ GET /orchestrator/notification/recipient

- **Status**: FAIL
- **Duration**: 83.224459ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ POST /orchestrator/notification/search

- **Status**: PASS
- **Duration**: 105.656083ms
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

### ✅ GET /orchestrator/notification/variables

- **Status**: PASS
- **Duration**: 96.172958ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"devtronAppId":"{{devtronAppId}}","devtronAppName":"{{devtronAppName}}","devtronApprovedByEmail":"{{devtronApprovedByEmail}}","devtronBuildGitCommitHash":"{{devtron...

---

### ✅ GET /orchestrator/notification/channel/ses/{id}

- **Status**: PASS
- **Duration**: 90.879334ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"userId":2,"teamId":0,"region":"ap-south-1","accessKey":"AKIAWEAVHFOOSNAORPHU","secretKey":"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","fromEmail":"watcher@devtron.i...

---

### ✅ GET /orchestrator/notification/channel/slack/{id}

- **Status**: PASS
- **Duration**: 87.124042ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"userId":0,"teamId":0,"webhookUrl":"","configName":"","description":"","id":0}}

---

### ❌ POST /orchestrator/notification/v2

- **Status**: FAIL
- **Duration**: 86.833125ms
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

### ❌ DELETE /orchestrator/notification

- **Status**: FAIL
- **Duration**: 86.450333ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{EOF}]","userMessage":"EOF"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/notification

- **Status**: FAIL
- **Duration**: 83.51725ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/notification

- **Status**: FAIL
- **Duration**: 116.825625ms
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
- **Duration**: 87.8645ms
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

### ✅ GET /orchestrator/notification/channel/webhook/{id}

- **Status**: PASS
- **Duration**: 86.923875ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"userId":0,"webhookUrl":"","configName":"","header":null,"payload":"","description":"","id":0}}

---

### ✅ GET /orchestrator/notification/channel/autocomplete/{type}

- **Status**: PASS
- **Duration**: 89.306417ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[]}

---

### ✅ GET /orchestrator/sso/list

- **Status**: PASS
- **Duration**: 85.0925ms
- **Spec File**: ../../specs/sso/configuration.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"google","label":"sample_string","url":"https://devtron.example.com/orchestrator","active":true,"globalAuthConfigType":""}]}

---

### ✅ PUT /orchestrator/sso/update

- **Status**: PASS
- **Duration**: 103.801084ms
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
- **Duration**: 87.523834ms
- **Spec File**: ../../specs/sso/configuration.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"name":"google","label":"sample_string","url":"https://devtron.example.com/orchestrator","config":{"id":"","type":"","name":"","config":null},"active":true,"...

---

### ❌ GET /orchestrator/sso

- **Status**: FAIL
- **Duration**: 83.630459ms
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
- **Duration**: 107.097958ms
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
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":2,"name":"sample_string","label":"sample_string","url":"https://devtron.example.com/orchestrator","config":{},"active":true,"globalAuthConfigType":""}}

---

### ✅ GET /orchestrator/app-store/discover/application/chartInfo/{appStoreApplicationVersionId}

- **Status**: PASS
- **Duration**: 261.549ms
- **Spec File**: ../../specs/app-store.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"appStoreApplicationVersionId":1,"readme":"# cluster-autoscaler\n\nScales Kubernetes worker nodes within autoscaling groups.\n\n## TL;DR\n\n```console\n$ helm repo ...

---

### ✅ GET /orchestrator/app-store/discover/application/{appStoreId}/version/autocomplete

- **Status**: PASS
- **Duration**: 88.698333ms
- **Spec File**: ../../specs/app-store.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"version":"9.49.0","id":5507},{"version":"9.48.0","id":1},{"version":"9.47.0","id":2},{"version":"9.46.6","id":3},{"version":"9.46.5","id":4},{"version":"9.46.4","...

---

### ✅ GET /orchestrator/app-store/discover/application/{id}

- **Status**: PASS
- **Duration**: 179.676292ms
- **Spec File**: ../../specs/app-store.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"version":"9.48.0","appVersion":"1.33.0","created":"2025-07-11T21:16:00.149315Z","deprecated":false,"description":"Scales Kubernetes worker nodes within auto...

---

### ❌ GET /orchestrator/app-store/discover/search

- **Status**: FAIL
- **Duration**: 82.902625ms
- **Spec File**: ../../specs/app-store.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/app-store/discover

- **Status**: PASS
- **Duration**: 437.428334ms
- **Spec File**: ../../specs/app-store.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":65,"appStoreApplicationVersionId":4028,"name":"ai-agent","chart_repo_id":2,"docker_artifact_store_id":"","chart_name":"devtron","icon":"","active":true,"chart...

---

### ❌ GET /orchestrator/flux-application

- **Status**: FAIL
- **Duration**: 87.528542ms
- **Spec File**: ../../specs/gitops/fluxcd.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{error in getting cluster ids}]","userMessage":"error in getting cluster ids"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/flux-application/app

- **Status**: FAIL
- **Duration**: 126.216291ms
- **Spec File**: ../../specs/gitops/fluxcd.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/chartref/autocomplete/{appId}

- **Status**: PASS
- **Duration**: 88.8315ms
- **Spec File**: ../../specs/helm/dynamic-charts.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"chartRefs":[{"id":10,"version":"3.9.0","name":"Rollout Deployment","description":"","userUploaded":false,"isAppMetricsSupported":true},{"id":11,"version":"3.10.0",...

---

### ❌ POST /orchestrator/app/ci-pipeline/patch

- **Status**: FAIL
- **Duration**: 95.041ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Request Payload**:
```json
{
  "action": 1,
  "appId": 1,
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
        "gitMaterialId": 1,
        "type": "sample_string",
        "value": "sample_string"
      },
      {
        "gitMaterialId": 1,
        "type": "sample_string",
        "value": "sample_string"
      }
    ],
    "isDockerConfigOverridden": true,
    "isExternal": true,
    "isManual": true,
    "name": "sample_string",
    "scanEnabled": true
  },
  "isJob": true,
  "userId": 1
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'CiPatchRequest.CiPipeline.CiMaterial[0].Source' Error:Field validation for 'Source' failed on the 'dive' tag","userMessage":"Key:...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/ci-pipeline/{appId}

- **Status**: PASS
- **Duration**: 90.996208ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"appId":1,"dockerRegistry":"quay","dockerRepository":"devtron/test","ciBuildConfig":{"id":1,"gitMaterialId":1,"buildContextGitMaterialId":1,"useRootBuildCont...

---

### ✅ GET /orchestrator/app/wf/all/component-names/{appId}

- **Status**: PASS
- **Duration**: 90.632542ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"workflows":[{"id":1,"name":"wf-1-lcu5","ciPipelineId":0,"ciPipelineName":"","cdPipelines":null},{"id":2,"name":"wf-1-4rww","ciPipelineId":0,"ciPipelineName":"","cd...

---

### ✅ POST /orchestrator/batch/v1beta1/build

- **Status**: PASS
- **Duration**: 87.289ms
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
- **Duration**: 86.650708ms
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

### ❌ POST /orchestrator/batch/v1beta1/hibernate

- **Status**: FAIL
- **Duration**: 87.663416ms
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
- **Duration**: 88.30225ms
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
- **Duration**: 87.236791ms
- **Spec File**: ../../specs/jobs/bulk-operations.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/batch/v1beta1/application

- **Status**: FAIL
- **Duration**: 86.776208ms
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
- **Duration**: 87.761041ms
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

### ❌ POST /orchestrator/module

- **Status**: FAIL
- **Duration**: 84.56275ms
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

### ✅ GET /orchestrator/module

- **Status**: PASS
- **Duration**: 247.170833ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"name":"security.trivy","status":"installed","moduleResourcesStatus":null,"enabled":true,"moduleType":"security"},{"name":"security.clair","status":"installed","mo...

---

### ✅ GET /orchestrator/server

- **Status**: PASS
- **Duration**: 87.975959ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"status":"unknown","releaseName":"devtron","installationType":"enterprise"}}

---

### ❌ POST /orchestrator/server

- **Status**: FAIL
- **Duration**: 87.877708ms
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

### ❌ GET /orchestrator/config/global/cm/edit/{appId}/{id}

- **Status**: FAIL
- **Duration**: 83.771291ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/config/global/cm/{appId}

- **Status**: PASS
- **Duration**: 88.734625ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":0,"appId":1,"configData":[],"isDeletable":false,"isExpressEdit":false}}

---

### ❌ DELETE /orchestrator/config/global/cm/{appId}/{id}

- **Status**: FAIL
- **Duration**: 84.104083ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/config/environment/cm

- **Status**: FAIL
- **Duration**: 90.800708ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "configData": [
    {
      "external": true,
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
      "name": "sample_string",
      "subPath": "sample_string",
      "type": "CONFIGMAP"
    },
    {
      "external": true,
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
      "name": "sample-item-2",
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

### ❌ GET /orchestrator/config/environment/cm/edit/{appId}/{envId}/{id}

- **Status**: FAIL
- **Duration**: 83.590542ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/config/environment/cm/{appId}/{envId}

- **Status**: PASS
- **Duration**: 88.551833ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":0,"appId":1,"environmentId":1,"configData":[],"isDeletable":false,"isExpressEdit":false}}

---

### ❌ DELETE /orchestrator/config/environment/cm/{appId}/{envId}/{id}

- **Status**: FAIL
- **Duration**: 82.961875ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/config/global/cm

- **Status**: FAIL
- **Duration**: 86.518708ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "configData": [
    {
      "external": true,
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
      "name": "sample_string",
      "subPath": "sample_string",
      "type": "CONFIGMAP"
    },
    {
      "external": true,
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
      "name": "sample-item-2",
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

### ❌ POST /orchestrator/config/bulk/patch

- **Status**: FAIL
- **Duration**: 87.823083ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "configData": [
    {
      "external": true,
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
      "name": "sample_string",
      "roleARN": "sample_string",
      "type": "CONFIGMAP"
    },
    {
      "external": true,
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
      "name": "sample-item-2",
      "roleARN": "sample_string",
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

### ✅ GET /orchestrator/app/history/deployed-configuration/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 88.547ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":8,"name":"DEPLOYMENT_TEMPLATE"},{"id":3,"name":"PIPELINE_STRATEGY"}]}

---

### ❌ GET /orchestrator/app/history/deployed-component/detail/{appId}/{pipelineId}/{id}

- **Status**: FAIL
- **Duration**: 85.874791ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","result":"invalid historyComponent"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/history/deployed-component/list/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 86.816125ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","result":"invalid historyComponent"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PATCH /orchestrator/app/ci-pipeline/patch-source

- **Status**: FAIL
- **Duration**: 127.219334ms
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

### ✅ PUT /orchestrator/deployment/template/upload

- **Status**: PASS
- **Duration**: 86.022084ms
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
- **Duration**: 86.893875ms
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
- **Duration**: 93.066708ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":10,"name":"Rollout Deployment","chartDescription":"This chart deploys an advanced version of deployment that supports Blue/Green and Canary deployments. For f...

---

### ❌ GET /orchestrator/resource/history/deployment/cd-pipeline/v1

- **Status**: FAIL
- **Duration**: 140.09625ms
- **Spec File**: ../../specs/deployment/history.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","internalMessage":"invalid format filter criteria!","userMessage":"invalid format filter criteria!"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 87.743916ms
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
- **Duration**: 89.245792ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Request Payload**:
```json
{
  "bitBucketWorkspaceId": "sample_string",
  "gitHubOrgId": "sample_string",
  "gitLabGroupId": "sample_string"
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/gitops/config

- **Status**: PASS
- **Duration**: 86.907208ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"provider":"GITHUB","username":"systemsdt","token":"","gitLabGroupId":"","gitHubOrgId":"stage-gitops","host":"https://github.com/","active":true,"azureProje...

---

### ❌ GET /orchestrator/gitops/config-by-provider

- **Status**: FAIL
- **Duration**: 83.389083ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/gitops/config/{id}

- **Status**: PASS
- **Duration**: 447.832333ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"provider":"GITHUB","username":"systemsdt","token":"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","gitLabGro...

---

### ✅ GET /orchestrator/gitops/configured

- **Status**: PASS
- **Duration**: 87.377791ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"allowCustomRepository":false,"authMode":"PASSWORD","exists":true,"provider":"GITHUB"}}

---

### ❌ POST /orchestrator/gitops/validate

- **Status**: FAIL
- **Duration**: 118.536334ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Request Payload**:
```json
{
  "allowCustomRepository": true,
  "isTLSCertDataPresent": true,
  "isTLSKeyDataPresent": true
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/git/host/{id}/event

- **Status**: FAIL
- **Duration**: 88.368459ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"{id}\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"{id}\": invalid synta...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/git/host

- **Status**: PASS
- **Duration**: 88.102208ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"Github","active":true,"webhookUrl":"","webhookSecret":"","eventTypeHeader":"","secretHeader":"","secretValidator":""},{"id":2,"name":"Bitbucket Clou...

---

### ❌ POST /orchestrator/git/host

- **Status**: FAIL
- **Duration**: 88.093917ms
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

### ❌ GET /orchestrator/git/host/event/{eventId}

- **Status**: FAIL
- **Duration**: 86.928833ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"{eventId}\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"{eventId}\": inv...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/webhook-meta-config/{gitProviderId}

- **Status**: FAIL
- **Duration**: 87.120209ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{ERROR #22P02 invalid input syntax for type integer: \"{gitProviderId}\"}]","userMessage":"ERROR #22P02 invalid...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/{id}

- **Status**: FAIL
- **Duration**: 86.961833ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"{id}\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"{id}\": invalid synta...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/security/scan/executionDetail

- **Status**: FAIL
- **Duration**: 86.877292ms
- **Spec File**: ../../specs/security/security-dashboard-apis.yml
- **Request Payload**: (none)
- **Response Code**: 403
- **Error/Msg**: {"code":403,"status":"Forbidden","errors":[{"code":"000","internalMessage":"[{unauthorized user}]","userMessage":"Unauthorized User"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 403

---

### ❌ GET /app/details/{appId}

- **Status**: FAIL
- **Duration**: 85.314583ms
- **Spec File**: ../../specs/application/core.yaml
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
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ POST /app/edit/projects

- **Status**: FAIL
- **Duration**: 94.666833ms
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
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ POST /app/edit

- **Status**: FAIL
- **Duration**: 90.07675ms
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
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character '<' looking for beginning of value

---

### ❌ POST /orchestrator/app

- **Status**: FAIL
- **Duration**: 96.49ms
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
- **Duration**: 93.365709ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"helm-sanity-application","createdBy":"admin","description":""},{"id":2,"name":"gitops-sanity-application","createdBy":"admin","description":""},{"id...

---

### ❌ POST /orchestrator/core/v1beta1/application

- **Status**: FAIL
- **Duration**: 87.406791ms
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
    "description": "sample_string",
    "projectIds": [
      1
    ],
    "teamId": 1
  }
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'AppDetail.Metadata.ProjectName' Error:Field validation for 'ProjectName' failed on the 'required' tag","userMessage":"Key: 'AppDe...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /app/list

- **Status**: FAIL
- **Duration**: 83.604ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**:
```json
{
  "appNameSearch": "sample_string",
  "environmentIds": [
    1
  ],
  "offset": 1
}
```
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

### ❌ POST /orchestrator/app/workflow

- **Status**: FAIL
- **Duration**: 84.446916ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "name": "sample_string",
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

### ✅ POST /orchestrator/app/workflow/clone

- **Status**: PASS
- **Duration**: 89.869292ms
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

### ❌ DELETE /orchestrator/app/workflow/{app-wf-id}/app/{app-id}

- **Status**: FAIL
- **Duration**: 83.48375ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ GET /orchestrator/app/commit-info/{ciPipelineMaterialId}/{gitHash}

- **Status**: FAIL
- **Duration**: 87.678709ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/ci-pipeline/trigger

- **Status**: FAIL
- **Duration**: 84.676542ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Request Payload**:
```json
{
  "ciPipelineMaterials": [
    {
      "Active": true,
      "GitCommit": {
        "Commit": "a1b2c3d4e5",
        "Date": "2023-01-15T14:30:22Z",
        "Message": "Update README"
      },
      "Value": "main"
    },
    {
      "Active": true,
      "GitCommit": {
        "Commit": "a1b2c3d4e5",
        "Date": "2023-01-15T14:30:22Z",
        "Message": "Update README"
      },
      "Value": "main"
    }
  ],
  "environmentId": 456,
  "invalidateCache": true,
  "pipelineId": 123,
  "pipelineType": "CI"
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/{appId}/ci-pipeline/{pipelineId}/workflow/{workflowId}

- **Status**: FAIL
- **Duration**: 89.38625ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ GET /orchestrator/app/ci-pipeline/{ciPipelineId}/linked-ci/downstream/env

- **Status**: FAIL
- **Duration**: 114.026042ms
- **Spec File**: ../../specs/ci-pipeline/ciPipelineDownstream/downstream-linked-ci-view-spec.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{invalid CiPipelineId 1}]","userMessage":"invalid CiPipelineId 1"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/ci-pipeline/{ciPipelineId}/linked-ci/downstream/cd

- **Status**: FAIL
- **Duration**: 121.501833ms
- **Spec File**: ../../specs/ci-pipeline/ciPipelineDownstream/downstream-linked-ci-view-spec.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{invalid CiPipelineId 1}]","userMessage":"invalid CiPipelineId 1"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/app/cd-pipeline/patch/deployment

- **Status**: FAIL
- **Duration**: 86.45525ms
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

### ❌ POST /orchestrator/deployment/pipeline/configure

- **Status**: FAIL
- **Duration**: 87.503834ms
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

### ❌ GET /orchestrator/deployment/pipeline/history

- **Status**: FAIL
- **Duration**: 84.066583ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/deployment/pipeline/rollback

- **Status**: FAIL
- **Duration**: 84.724584ms
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
- **Duration**: 85.764917ms
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

### ❌ GET /orchestrator/app/template/default/{appId}/{chartRefId}

- **Status**: FAIL
- **Duration**: 90.029625ms
- **Spec File**: ../../specs/environment/templates.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/git/{gitHostId}/{secret}

- **Status**: FAIL
- **Duration**: 85.973166ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Request Payload**:
```json
{
  "commits": [
    {
      "author": {
        "email": "sample_string",
        "name": "sample_string",
        "username": "sample_string"
      },
      "committer": {
        "email": "sample_string",
        "name": "sample_string",
        "username": "sample_string"
      },
      "id": "sample_string"
    },
    {
      "author": {
        "email": "sample_string",
        "name": "sample_string",
        "username": "sample_string"
      },
      "committer": {
        "email": "sample_string",
        "name": "sample_string",
        "username": "sample_string"
      },
      "id": 2
    }
  ],
  "pusher": {
    "email": "sample_string",
    "name": "sample_string",
    "username": "sample_string"
  },
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

### ❌ POST /orchestrator/webhook/ci/workflow

- **Status**: FAIL
- **Duration**: 85.975459ms
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

### ❌ POST /orchestrator/webhook/ext-ci/{externalCiId}

- **Status**: FAIL
- **Duration**: 86.126583ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Request Payload**:
```json
{
  "finishedOn": "2023-01-01T00:00:00Z",
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
- **Duration**: 84.117209ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Request Payload**:
```json
{
  "compare": "sample_string",
  "head_commit": {
    "author": {
      "email": "sample_string",
      "name": "sample_string",
      "username": "sample_string"
    },
    "committer": {
      "email": "sample_string",
      "name": "sample_string",
      "username": "sample_string"
    },
    "id": "sample_string"
  },
  "ref": "sample_string"
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/git/{gitHostId}

- **Status**: FAIL
- **Duration**: 136.587208ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Request Payload**:
```json
{
  "after": "sample_string",
  "compare": "sample_string",
  "repository": {
    "full_name": "sample_string",
    "id": 1,
    "private": true
  }
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"git host not found: sample_string","userMessage":"git host with ID 'sample_string' not found","userDetailMessage":"The req...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/user/sync/orchestratortocasbin

- **Status**: PASS
- **Duration**: 89.36775ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":true}

---

### ❌ PUT /orchestrator/user/update/trigger/terminal

- **Status**: FAIL
- **Duration**: 83.128959ms
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
- **Duration**: 89.608625ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"users":[{"id":2,"email_id":"admin","roleFilters":[],"groups":[],"superAdmin":false,"userRoleGroups":[],"lastLoginTime":"2025-08-08T06:03:52.406471Z","timeoutWindow...

---

### ✅ GET /orchestrator/user/v2/{id}

- **Status**: PASS
- **Duration**: 88.866583ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"email_id":"system","roleFilters":[],"groups":[],"superAdmin":false,"userRoleGroups":[],"lastLoginTime":"0001-01-01T00:00:00Z","timeoutWindowExpression":"000...

---

### ❌ DELETE /orchestrator/user/{id}

- **Status**: FAIL
- **Duration**: 90.962417ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","userMessage":"cannot delete system or admin user"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/{id}

- **Status**: PASS
- **Duration**: 93.714792ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"email_id":"system","roleFilters":[],"groups":[],"superAdmin":false,"userRoleGroups":[],"lastLoginTime":"0001-01-01T00:00:00Z","timeoutWindowExpression":"000...

---

### ❌ POST /orchestrator/user

- **Status**: FAIL
- **Duration**: 87.874542ms
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
- **Duration**: 87.256916ms
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
- **Duration**: 87.92975ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{EOF}]","userMessage":"EOF"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/detail/get

- **Status**: PASS
- **Duration**: 88.672084ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"email_id":"system","roleFilters":[],"groups":[],"superAdmin":false,"userRoleGroups":null,"lastLoginTime":"0001-01-01T00:00:00Z","timeoutWindowExpression":"...

---

### ✅ GET /orchestrator/version

- **Status**: PASS
- **Duration**: 114.740542ms
- **Spec File**: ../../specs/common/version.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"result":{"gitCommit":"908eae83","buildTime":"2025-08-05T20:18:14Z","serverMode":"FULL"}}

---

### ✅ GET /orchestrator/app/history/deployed-configuration/all/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 96.497041ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"deploymentTemplate":{"templateName":"Deployment","templateVersion":"4.21.0","isAppMetricsEnabled":false,"codeEditorValue":{"displayName":"values.yaml","value":"{\"...

---

### ❌ GET /orchestrator/app/history/deployed-configuration/all/latest/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 86.4675ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"latest\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"latest\": invalid syntax"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/deployment-status/timeline/{appId}/{envId}

- **Status**: FAIL
- **Duration**: 87.824791ms
- **Spec File**: ../../specs/deployment/timeline.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/batch/v1beta1/cd-pipeline

- **Status**: FAIL
- **Duration**: 86.6735ms
- **Spec File**: ../../specs/environment/bulk-delete.yaml
- **Request Payload**:
```json
{
  "deleteWfAndCiPipeline": true,
  "nonCascadeDelete": true,
  "projectNames": [
    "sample_string"
  ]
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","userMessage":"invalid payload, can not get pipelines for this filter"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 83.853625ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ✅ GET /orchestrator/external-links

- **Status**: PASS
- **Duration**: 87.087834ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":3,"name":"sample-external-link","url":"https://grafana.example.com","active":true,"monitoringToolId":1,"type":"appLevel","identifiers":[{"type":"devtron-app",...

---

### ✅ POST /orchestrator/external-links

- **Status**: PASS
- **Duration**: 89.003667ms
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
- **Duration**: 89.343458ms
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

### ✅ GET /orchestrator/external-links/tools

- **Status**: PASS
- **Duration**: 87.357042ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"Grafana","icon":"","category":2},{"id":2,"name":"Kibana","icon":"","category":2},{"id":3,"name":"Newrelic","icon":"","category":2},{"id":4,"name":"C...

---

### ✅ POST /orchestrator/chart-repo/validate

- **Status**: PASS
- **Duration**: 317.052542ms
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
- **Duration**: 94.361334ms
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
- **Duration**: 316.044375ms
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

### ❌ GET /orchestrator/rbac/roles/default

- **Status**: FAIL
- **Duration**: 82.64875ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/user/role/cache

- **Status**: PASS
- **Duration**: 86.520833ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":"{}"}

---

### ❌ GET /orchestrator/api/dex/{path}

- **Status**: FAIL
- **Duration**: 84.895459ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 502
- **Error/Msg**: {}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 502

---

### ❌ GET /orchestrator/auth/callback

- **Status**: FAIL
- **Duration**: 313.387208ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: Failed to query provider "https://devtron.example.com/orchestrator/api/dex": Get "https://devtron.example.com/orchestrator/api/dex/.well-known/openid-configuration": dial tcp: lookup devtron.example.c...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/devtron/auth/verify

- **Status**: PASS
- **Duration**: 106.951291ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":true}

---

### ❌ GET /orchestrator/login

- **Status**: FAIL
- **Duration**: 335.551583ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: Failed to query provider "https://devtron.example.com/orchestrator/api/dex": Get "https://devtron.example.com/orchestrator/api/dex/.well-known/openid-configuration": dial tcp: lookup devtron.example.c...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'F' looking for beginning of value

---

### ✅ GET /orchestrator/user/check/roles

- **Status**: PASS
- **Duration**: 87.59975ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"roles":["role:super-admin___","role:chart-group_admin"],"superAdmin":true}}

---

### ❌ GET /devtron/auth/verify/v2

- **Status**: FAIL
- **Duration**: 83.578709ms
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

### ❌ GET /orchestrator/auth/login

- **Status**: FAIL
- **Duration**: 93.576417ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: Failed to query provider "https://devtron.example.com/orchestrator/api/dex": Get "https://devtron.example.com/orchestrator/api/dex/.well-known/openid-configuration": dial tcp: lookup devtron.example.c...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'F' looking for beginning of value

---

### ❌ GET /orchestrator/refresh

- **Status**: FAIL
- **Duration**: 87.167208ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: internal server error

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/user/role/cache/invalidate

- **Status**: PASS
- **Duration**: 115.66025ms
- **Spec File**: ../../specs/security/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":"Cache Cleaned Successfully"}

---

### ❌ POST /orchestrator/api/v1/session

- **Status**: FAIL
- **Duration**: 83.862583ms
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

### ❌ POST /orchestrator/admin/policy/default

- **Status**: FAIL
- **Duration**: 85.367ms
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
- **Duration**: 83.098041ms
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

### ❌ PATCH /orchestrator/app/env/patch

- **Status**: FAIL
- **Duration**: 87.202667ms
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

### ❌ DELETE /orchestrator/k8s/resources/ephemeralContainers

- **Status**: FAIL
- **Duration**: 85.039625ms
- **Spec File**: ../../specs/kubernetes/ephemeral-containers.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/k8s/resources/ephemeralContainers

- **Status**: FAIL
- **Duration**: 92.218875ms
- **Spec File**: ../../specs/kubernetes/ephemeral-containers.yaml
- **Request Payload**:
```json
{
  "advancedData": {
    "manifest": "sample_string"
  },
  "clusterId": 1,
  "externalArgoApplicationName": "sample_string",
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

### ❌ POST /orchestrator/application/template-chart

- **Status**: FAIL
- **Duration**: 483.436875ms
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

### ❌ PUT /orchestrator/application/rollback

- **Status**: FAIL
- **Duration**: 95.53425ms
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

### ❌ POST /orchestrator/user/resource/options/{kind}/{version}

- **Status**: FAIL
- **Duration**: 82.78ms
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

### ❌ PUT /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 87.056291ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**:
```json
{
  "cluster_name": "sample_string",
  "prometheus_url": "sample_string",
  "server_url": "sample_string",
  "updated_by": "sample_string",
  "updated_on": "2023-01-01T00:00:00Z"
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'ClusterBean.ServerUrl' Error:Field validation for 'ServerUrl' failed on the 'url' tag","userMessage":"Key: 'ClusterBean.ServerUrl...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/cluster

- **Status**: PASS
- **Duration**: 91.447208ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"cluster_name":"default_cluster","description":"","server_url":"https://kubernetes.default.svc","active":true,"config":{"bearer_token":""},"prometheusAuth":...

---

### ❌ POST /orchestrator/cluster/delete

- **Status**: FAIL
- **Duration**: 83.283834ms
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
- **Duration**: 83.606792ms
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
- **Duration**: 92.049417ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"environment_name":"devtron-demo","cluster_id":1,"cluster_name":"default_cluster","active":true,"default":false,"namespace":"devtron-demo","isClusterCdActiv...

---

### ❌ POST /orchestrator/env

- **Status**: FAIL
- **Duration**: 96.061458ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**:
```json
{
  "cluster_id": 1,
  "created_by": "sample_string",
  "default": true,
  "environment_name": "sample_string",
  "namespace": "sample_string",
  "updated_by": "sample_string"
}
```
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"422","internalMessage":"Namespace \"sample_string\" is invalid: metadata.name: Invalid value: \"sample_string\": a lowercase RFC 1123 la...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PUT /orchestrator/env

- **Status**: FAIL
- **Duration**: 87.491666ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**:
```json
{
  "active": true,
  "cluster_id": 1,
  "environment_name": "sample_string",
  "namespace": "sample_string",
  "updated_by": "sample_string",
  "updated_on": "2023-01-01T00:00:00Z"
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/env/clusters

- **Status**: FAIL
- **Duration**: 83.214ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/env/namespace/autocomplete

- **Status**: FAIL
- **Duration**: 116.566458ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{ERROR #42601 syntax error at or near \")\"}]","userMessage":"ERROR #42601 syntax error at or near \")\""}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/gitops/config/{id}

- **Status**: PASS
- **Duration**: 87.253208ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"provider":"GITHUB","username":"systemsdt","token":"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","gitLabGro...

---

### ✅ GET /orchestrator/gitops/configured

- **Status**: PASS
- **Duration**: 87.373875ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"allowCustomRepository":false,"authMode":"PASSWORD","exists":true,"provider":"GITHUB"}}

---

### ❌ POST /orchestrator/validate

- **Status**: FAIL
- **Duration**: 83.364167ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Request Payload**:
```json
{
  "allowCustomRepository": true,
  "id": 1,
  "tlsConfig": {
    "caData": "sample_string",
    "tlsCertData": "sample_string",
    "tlsKeyData": "sample_string"
  }
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/gitops/config

- **Status**: PASS
- **Duration**: 94.319375ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"provider":"GITHUB","username":"systemsdt","token":"","gitLabGroupId":"","gitHubOrgId":"stage-gitops","host":"https://github.com/","active":true,"azureProje...

---

### ❌ POST /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 87.437292ms
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
- **Duration**: 87.083ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Request Payload**:
```json
{
  "bitBucketProjectKey": "sample_string",
  "bitBucketWorkspaceId": "sample_string",
  "enableTLSVerification": true
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/gitops/config-by-provider

- **Status**: FAIL
- **Duration**: 83.559625ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ GET /orchestrator/app/history/deployed-component/detail/{appId}/{pipelineId}/{id}

- **Status**: FAIL
- **Duration**: 86.939291ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","result":"invalid historyComponent"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/history/deployed-component/list/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 87.226166ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","result":"invalid historyComponent"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/history/deployed-configuration/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 89.529125ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":8,"name":"DEPLOYMENT_TEMPLATE"},{"id":3,"name":"PIPELINE_STRATEGY"}]}

---

### ✅ GET /orchestrator/api-token

- **Status**: PASS
- **Duration**: 126.738792ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"description":"some description","expireAtInMs":12344546,"id":2,"name":"some-name","token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6IkFQSS1UT0tFTjpzb21lLW...

---

### ❌ POST /orchestrator/api-token

- **Status**: FAIL
- **Duration**: 87.91625ms
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

### ✅ GET /orchestrator/api-token/webhook

- **Status**: PASS
- **Duration**: 86.658416ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[]}

---

### ❌ DELETE /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 87.606542ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{api-token corresponds to apiTokenId '1' is not found}]","userMessage":"api-token corresponds to apiTokenId '1'...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PUT /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 105.627ms
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

### ❌ POST /orchestrator/plugin/global/create

- **Status**: FAIL
- **Duration**: 88.126791ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Request Payload**:
```json
{
  "description": "sample_string",
  "type": "SHARED",
  "versions": [
    {
      "id": 1,
      "updatedBy": 1,
      "updatedOn": "2023-01-01T00:00:00Z"
    },
    {
      "id": 2,
      "updatedBy": 1,
      "updatedOn": "2023-01-01T00:00:00Z"
    }
  ]
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","internalMessage":"no step data provided to save, please provide a plugin step to proceed further","userMessage":"no step data provided to sa...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/plugin/global/detail/all

- **Status**: PASS
- **Duration**: 465.2625ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"K6 Load testing","description":"K6 is an open-source tool and cloud service that makes load testing easy for developers and QA engineers.","type":"P...

---

### ❌ GET /orchestrator/plugin/global/list/global-variable

- **Status**: FAIL
- **Duration**: 86.342208ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"\": invalid syntax}]","userMessage":"invalid appId"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/plugin/global/list/v2

- **Status**: PASS
- **Duration**: 128.743041ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"parentPlugins":[{"id":4,"name":"AWS ECR Retag","pluginIdentifier":"aws-retag","description":"AWS ECR Retag plugin that enables retagging of container images within...

---

### ✅ PUT /orchestrator/plugin/global/migrate

- **Status**: PASS
- **Duration**: 87.147917ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK"}

---

### ✅ GET /orchestrator/plugin/global/detail/{pluginId}

- **Status**: PASS
- **Duration**: 92.251709ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"name":"K6 Load testing","description":"K6 is an open-source tool and cloud service that makes load testing easy for developers and QA engineers.","type":"PR...

---

### ❌ POST /orchestrator/plugin/global/list/detail/v2

- **Status**: FAIL
- **Duration**: 87.855667ms
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

### ✅ GET /orchestrator/plugin/global/list/tags

- **Status**: PASS
- **Duration**: 86.436958ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"tagNames":["Release","Load testing","Code Review","Kubernetes","Google Kubernetes Engine","cloud","DevSecOps","CI task","AWS EKS","Code quality","Security","Image ...

---

### ✅ GET /orchestrator/plugin/global/list/v2/min

- **Status**: PASS
- **Duration**: 86.967667ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[]}

---

### ❌ DELETE /orchestrator/user/role/group/{id}

- **Status**: FAIL
- **Duration**: 89.722ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/user/role/group/{id}

- **Status**: FAIL
- **Duration**: 87.939667ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"role group not found: 1","userMessage":"role group with ID '1' not found","userDetailMessage":"The requested role group do...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/user/role/group

- **Status**: PASS
- **Duration**: 86.849ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":2,"roleFilters":[],"superAdmin":false,"canManageAllAccess":false,"hasAccessManagerPermission":false}]}

---

### ❌ POST /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 89.109625ms
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
- **Duration**: 114.282583ms
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
- **Duration**: 87.8655ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{EOF}]","userMessage":"EOF"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/role/group/detailed/get

- **Status**: PASS
- **Duration**: 88.898833ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":2,"roleFilters":[],"superAdmin":false,"canManageAllAccess":false,"hasAccessManagerPermission":false}]}

---

### ❌ GET /orchestrator/user/role/group/search

- **Status**: FAIL
- **Duration**: 86.684125ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"11001","internalMessage":"invalid path parameter id: search","userMessage":"Invalid path parameter 'id'","userDetailMessage":"Please check the par...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/role/group/v2

- **Status**: PASS
- **Duration**: 88.452959ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"roleGroups":[{"id":2,"roleFilters":[],"superAdmin":false,"canManageAllAccess":false,"hasAccessManagerPermission":false}],"totalCount":1}}

---

### ❌ POST /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 87.730833ms
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

### ❌ PUT /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 87.541834ms
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

### ❌ POST /orchestrator/chart-repo/update

- **Status**: FAIL
- **Duration**: 414.903625ms
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

### ✅ POST /orchestrator/chart-repo/validate

- **Status**: PASS
- **Duration**: 97.252917ms
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

### ✅ GET /orchestrator/chart-repo/{id}

- **Status**: PASS
- **Duration**: 87.010333ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"name":"default-chartmuseum","url":"http://devtron-chartmuseum.devtroncd:8080/","authMode":"ANONYMOUS","active":true,"default":true,"allow_insecure_connectio...

---

### ❌ POST /orchestrator/app-store/chart-provider/update

- **Status**: FAIL
- **Duration**: 88.107125ms
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

### ❌ POST /orchestrator/chart-group/

- **Status**: FAIL
- **Duration**: 86.827875ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**:
```json
{
  "description": "sample_string",
  "entries": [
    {
      "chartId": 1,
      "chartName": "sample_string",
      "chartRepoId": 1,
      "chartVersion": "sample_string"
    },
    {
      "chartId": 1,
      "chartName": "sample_string",
      "chartRepoId": 1,
      "chartVersion": "sample_string"
    }
  ],
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
- **Duration**: 87.724958ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**:
```json
{
  "entries": [
    {
      "chartId": 1,
      "chartName": "sample_string",
      "chartRepoId": 1,
      "chartVersion": "sample_string"
    },
    {
      "chartId": 1,
      "chartName": "sample_string",
      "chartRepoId": 1,
      "chartVersion": "sample_string"
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

### ✅ POST /orchestrator/chart-repo/sync-charts

- **Status**: PASS
- **Duration**: 7.17830525s
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"status":"ok"}}

---

### ✅ GET /orchestrator/app-store/chart-provider/list

- **Status**: PASS
- **Duration**: 92.130417ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":"prakhar","name":"prakhar","active":true,"isEditable":true,"isOCIRegistry":true,"registryProvider":"docker-hub"},{"id":"1","name":"default-chartmuseum","activ...

---

### ✅ GET /orchestrator/chart-group/list

- **Status**: PASS
- **Duration**: 91.213416ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":null}

---

### ✅ GET /orchestrator/chart-repo/list

- **Status**: PASS
- **Duration**: 91.503166ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"default-chartmuseum","url":"http://devtron-chartmuseum.devtroncd:8080/","authMode":"ANONYMOUS","active":true,"default":true,"allow_insecure_connecti...

---

### ✅ POST /orchestrator/app-store/chart-provider/sync-chart

- **Status**: PASS
- **Duration**: 7.256417917s
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

### ❌ PUT /orchestrator/chart-group/entries

- **Status**: FAIL
- **Duration**: 87.168333ms
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
  "groupId": 1
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/chart-group/{id}

- **Status**: FAIL
- **Duration**: 84.025209ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ POST /orchestrator/chart-repo/create

- **Status**: PASS
- **Duration**: 97.901958ms
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

### ✅ GET /orchestrator/cluster

- **Status**: PASS
- **Duration**: 94.703583ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"cluster_name":"default_cluster","description":"","server_url":"https://kubernetes.default.svc","active":true,"config":{"bearer_token":""},"prometheusAuth":...

---

### ❌ PUT /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 89.364875ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Request Payload**:
```json
{
  "agentInstallationStage": 1,
  "defaultClusterComponent": [
    {
      "appId": 1,
      "envId": 1,
      "status": "sample_string"
    },
    {
      "appId": 1,
      "envId": 1,
      "status": "sample_string"
    }
  ],
  "updatedOn": "2023-01-01T00:00:00Z"
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'ClusterBean.ClusterName' Error:Field validation for 'ClusterName' failed on the 'required' tag","userMessage":"Key: 'ClusterBean....

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/cluster/auth-list

- **Status**: PASS
- **Duration**: 87.606667ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"cluster_name":"default_cluster","description":"","active":false,"defaultClusterComponent":null,"agentInstallationStage":0,"k8sVersion":"","insecureSkipTlsV...

---

### ❌ POST /orchestrator/k8s/resources/apply

- **Status**: FAIL
- **Duration**: 87.714208ms
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
- **Duration**: 83.192542ms
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
- **Duration**: 268.388584ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"apiResources":[{"gvk":{"Group":"","Version":"v1","Kind":"ComponentStatus"},"gvr":{"Group":"","Version":"v1","Resource":"componentstatuses"},"namespaced":false,"sho...

---

### ❌ POST /orchestrator/k8s/resource

- **Status**: FAIL
- **Duration**: 91.849333ms
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
- **Duration**: 86.551417ms
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
- **Duration**: 88.86425ms
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
- **Duration**: 100.202375ms
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
- **Duration**: 115.800417ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"404","userMessage":"error on getting resource from k8s"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/k8s/resource/urls

- **Status**: FAIL
- **Duration**: 111.62125ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---


# API Spec Validation Report

Generated: 2025-08-13T20:37:27+05:30

## Summary

- Total Endpoints: 277
- Passed: 86
- Failed: 191
- Warnings: 0
- Success Rate: 31.05%

## Detailed Results

### ❌ PUT /k8s/capacity/node/drain

- **Status**: FAIL
- **Duration**: 367.39125ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Request Payload**:
```json
{
  "clusterId": 1,
  "kind": "sample_string",
  "manifestPatch": "sample_string",
  "name": "sample_string",
  "version": "sample_string"
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

### ❌ GET /k8s/capacity/node/list

- **Status**: FAIL
- **Duration**: 72.227833ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
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

### ❌ PUT /k8s/capacity/node/taints/edit

- **Status**: FAIL
- **Duration**: 73.914625ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Request Payload**:
```json
{
  "clusterId": 1,
  "manifestPatch": "sample_string",
  "name": "sample_string",
  "nodeCordonOptions": {
    "unschedulableDesired": true
  },
  "nodeDrainOptions": {
    "deleteEmptyDirData": true,
    "disableEviction": true,
    "force": true,
    "gracePeriodSeconds": 1,
    "ignoreAllDaemonSets": true
  }
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

### ✅ GET /orchestrator/k8s/capacity/cluster/list/raw

- **Status**: PASS
- **Duration**: 133.715291ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":2,"name":"isolated","errorInNodeListing":"Get virtual cluster 'isolated' error: connection not setup for isolated clusters","nodeDetails":null,"nodeErrors":nu...

---

### ❌ GET /k8s/capacity/cluster/list

- **Status**: FAIL
- **Duration**: 72.632959ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
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

### ❌ GET /k8s/capacity/cluster/{clusterId}

- **Status**: FAIL
- **Duration**: 73.136417ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
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

### ❌ PUT /k8s/capacity/node

- **Status**: FAIL
- **Duration**: 72.232ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Request Payload**:
```json
{
  "clusterId": 1,
  "manifestPatch": "sample_string",
  "name": "sample_string",
  "nodeCordonOptions": {
    "unschedulableDesired": true
  },
  "nodeDrainOptions": {
    "deleteEmptyDirData": true,
    "disableEviction": true,
    "force": true,
    "gracePeriodSeconds": 1,
    "ignoreAllDaemonSets": true
  }
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

### ❌ DELETE /k8s/capacity/node

- **Status**: FAIL
- **Duration**: 72.539458ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
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

### ❌ GET /k8s/capacity/node

- **Status**: FAIL
- **Duration**: 72.9145ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
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

### ❌ PUT /k8s/capacity/node/cordon

- **Status**: FAIL
- **Duration**: 75.5235ms
- **Spec File**: ../../specs/kubernetes/capacity.yaml
- **Request Payload**:
```json
{
  "clusterId": 1,
  "manifestPatch": "sample_string",
  "name": "sample_string",
  "nodeCordonOptions": {
    "unschedulableDesired": true
  },
  "nodeDrainOptions": {
    "deleteEmptyDirData": true,
    "disableEviction": true,
    "force": true,
    "gracePeriodSeconds": 1,
    "ignoreAllDaemonSets": true
  }
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

### ✅ GET /orchestrator/plugin/global/detail/all

- **Status**: PASS
- **Duration**: 691.71775ms
- **Spec File**: ../../specs/plugin/global.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"K6 Load testing","description":"K6 is an open-source tool and cloud service that makes load testing easy for developers and QA engineers.","type":"P...

---

### ✅ GET /orchestrator/plugin/global/detail/{pluginId}

- **Status**: PASS
- **Duration**: 87.836584ms
- **Spec File**: ../../specs/plugin/global.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"name":"K6 Load testing","description":"K6 is an open-source tool and cloud service that makes load testing easy for developers and QA engineers.","type":"PR...

---

### ✅ GET /orchestrator/plugin/global/list/v2

- **Status**: PASS
- **Duration**: 167.250584ms
- **Spec File**: ../../specs/plugin/global.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"parentPlugins":[{"id":4,"name":"AWS ECR Retag","pluginIdentifier":"aws-retag","description":"AWS ECR Retag plugin that enables retagging of container images within...

---

### ✅ GET /orchestrator/plugin/global/list/v2/min

- **Status**: PASS
- **Duration**: 76.331541ms
- **Spec File**: ../../specs/plugin/global.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[]}

---

### ❌ POST /orchestrator/plugin/global/list/detail/v2

- **Status**: FAIL
- **Duration**: 78.441959ms
- **Spec File**: ../../specs/plugin/global.yaml
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

### ✅ PUT /orchestrator/plugin/global/migrate

- **Status**: PASS
- **Duration**: 77.706125ms
- **Spec File**: ../../specs/plugin/global.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK"}

---

### ❌ GET /orchestrator/plugin/global/list/global-variable

- **Status**: FAIL
- **Duration**: 88.272042ms
- **Spec File**: ../../specs/plugin/global.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"\": invalid syntax}]","userMessage":"invalid appId"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/plugin/global/list/tags

- **Status**: PASS
- **Duration**: 78.236625ms
- **Spec File**: ../../specs/plugin/global.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"tagNames":["Google Kubernetes Engine","AWS EKS","Github","Release","Load testing","Code quality","Security","CI task","GCP","cloud","Kubernetes","gcs","DevSecOps",...

---

### ❌ POST /orchestrator/plugin/global/create

- **Status**: FAIL
- **Duration**: 75.870041ms
- **Spec File**: ../../specs/plugin/global.yaml
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

### ❌ POST /app/material/delete

- **Status**: FAIL
- **Duration**: 73.573667ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "material": {
    "checkoutPath": "sample_string",
    "fetchSubmodules": true,
    "url": "sample_string"
  }
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

### ❌ POST /docker/registry/delete

- **Status**: FAIL
- **Duration**: 73.410667ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Request Payload**:
```json
{
  "connection": "sample_string",
  "id": 1,
  "password": "sample_string"
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

### ❌ POST /orchestrator/cluster/delete

- **Status**: FAIL
- **Duration**: 194.554042ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Request Payload**:
```json
{
  "config": {
    "error": "sample_string",
    "stage": "sample_string"
  },
  "defaultClusterComponent": [
    {
      "installedAppId": 1,
      "name": "sample_string",
      "status": "sample_string"
    },
    {
      "installedAppId": 1,
      "name": "sample-item-2",
      "status": "sample_string"
    }
  ],
  "k8sversion": "sample_string"
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /team

- **Status**: FAIL
- **Duration**: 79.977583ms
- **Spec File**: ../../specs/common/delete-options.yaml
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

### ❌ GET /team

- **Status**: FAIL
- **Duration**: 73.359125ms
- **Spec File**: ../../specs/common/delete-options.yaml
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

### ❌ POST /team

- **Status**: FAIL
- **Duration**: 73.590875ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Request Payload**:
```json
{
  "active": true,
  "id": 1,
  "name": "sample_string"
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

### ❌ PUT /team

- **Status**: FAIL
- **Duration**: 73.013166ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Request Payload**:
```json
{
  "active": true,
  "id": 1,
  "name": "sample_string"
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

### ❌ POST /chart-group/delete

- **Status**: FAIL
- **Duration**: 71.827208ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Request Payload**:
```json
{
  "chartGroupEntries": [
    {
      "appStoreValuesVersionId": 1,
      "appStoreValuesVersionName": "sample_string",
      "chartMetaData": {
        "chartName": "sample_string",
        "chartRepoName": "sample_string",
        "environmentId": 1
      }
    },
    {
      "appStoreValuesVersionId": 1,
      "appStoreValuesVersionName": "sample_string",
      "chartMetaData": {
        "chartName": "sample_string",
        "chartRepoName": "sample_string",
        "environmentId": 1
      }
    }
  ],
  "description": "sample_string",
  "id": 1
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

### ❌ GET /team/{id}

- **Status**: FAIL
- **Duration**: 72.3045ms
- **Spec File**: ../../specs/common/delete-options.yaml
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

### ❌ POST /app-store/repo/delete

- **Status**: FAIL
- **Duration**: 75.459125ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Request Payload**:
```json
{
  "active": true,
  "default": true,
  "id": 1
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

### ❌ POST /notification/channel/delete

- **Status**: FAIL
- **Duration**: 73.1455ms
- **Spec File**: ../../specs/common/delete-options.yaml
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

### ❌ GET /team/autocomplete

- **Status**: FAIL
- **Duration**: 73.926208ms
- **Spec File**: ../../specs/common/delete-options.yaml
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

### ❌ POST /env/delete

- **Status**: FAIL
- **Duration**: 76.458417ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Request Payload**:
```json
{
  "cluster_id": 1,
  "environment_name": "sample_string",
  "id": 1
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

### ❌ POST /git/provider/delete

- **Status**: FAIL
- **Duration**: 71.231542ms
- **Spec File**: ../../specs/common/delete-options.yaml
- **Request Payload**:
```json
{
  "gitHostId": 1,
  "name": "sample_string",
  "userName": "sample_string"
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

### ❌ POST /orchestrator/app/cd-pipeline/patch/deployment

- **Status**: FAIL
- **Duration**: 76.694459ms
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

### ❌ POST /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 77.323708ms
- **Spec File**: ../../specs/gitops/bitbucket.yaml
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
- **Duration**: 102.22475ms
- **Spec File**: ../../specs/gitops/bitbucket.yaml
- **Request Payload**:
```json
{
  "active": true,
  "isTLSKeyDataPresent": true,
  "tlsConfig": {
    "caData": "sample_string",
    "tlsCertData": "sample_string",
    "tlsKeyData": "sample_string"
  }
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/gitops/validate

- **Status**: FAIL
- **Duration**: 76.916667ms
- **Spec File**: ../../specs/gitops/bitbucket.yaml
- **Request Payload**:
```json
{
  "active": true,
  "enableTLSVerification": true,
  "gitHubOrgId": "sample_string"
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /app/deployment/template/data

- **Status**: FAIL
- **Duration**: 76.744167ms
- **Spec File**: ../../specs/gitops/manifest-generation.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "chartRefId": 1,
  "environmentId": 1,
  "getValues": true,
  "type": 1,
  "values": "sample_string"
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

### ❌ GET /orchestrator/app/deployments/{app-id}/{env-id}

- **Status**: FAIL
- **Duration**: 75.844292ms
- **Spec File**: ../../specs/gitops/manifest-generation.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /cluster/auth-list

- **Status**: FAIL
- **Duration**: 72.864375ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
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

### ❌ GET /cluster/namespaces

- **Status**: FAIL
- **Duration**: 74.119125ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
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

### ❌ GET /cluster/namespaces/{clusterId}

- **Status**: FAIL
- **Duration**: 74.985542ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
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

### ❌ POST /cluster/saveClusters

- **Status**: FAIL
- **Duration**: 97.424333ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Request Payload**:
```json
[
  {
    "cluster_name": "sample_string",
    "insecure-skip-tls-verify": true,
    "k8sVersion": "sample_string",
    "server_url": "sample_string",
    "userName": "sample_string"
  },
  {
    "cluster_name": "sample_string",
    "insecure-skip-tls-verify": true,
    "k8sVersion": "sample_string",
    "server_url": "sample_string",
    "userName": "sample_string"
  }
]
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

### ❌ POST /cluster/validate

- **Status**: FAIL
- **Duration**: 72.6095ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Request Payload**:
```json
{
  "kubeconfig": "sample_string"
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

### ❌ DELETE /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 138.399792ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{EOF}]","userMessage":"EOF"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/cluster

- **Status**: PASS
- **Duration**: 79.792542ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":2,"cluster_name":"isolated","description":"","active":true,"config":{"bearer_token":""},"prometheusAuth":{"isAnonymous":false},"defaultClusterComponent":[],"a...

---

### ❌ POST /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 75.326709ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Request Payload**:
```json
{
  "cluster_name": "sample_string",
  "errorInConnecting": "sample_string",
  "prometheusAuth": {
    "basic": {
      "password": "sample_string",
      "username": "sample_string"
    },
    "bearer": {
      "token": "sample_string"
    },
    "type": "basic"
  },
  "server_url": "sample_string",
  "userName": "sample_string"
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'ClusterBean.ServerUrl' Error:Field validation for 'ServerUrl' failed on the 'url' tag","userMessage":"Key: 'ClusterBean.ServerUrl...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 77.690708ms
- **Spec File**: ../../specs/kubernetes/cluster-management.yaml
- **Request Payload**:
```json
{
  "clusterUpdated": true,
  "cluster_name": "sample_string",
  "config": {
    "bearer_token": "sample_string",
    "cert_auth_data": "sample_string",
    "cert_data": "sample_string"
  },
  "defaultClusterComponent": [
    {
      "configuration": {
        "type": "yaml"
      },
      "id": "sample_string",
      "name": "sample_string"
    },
    {
      "configuration": {
        "type": "yaml"
      },
      "id": 2,
      "name": "sample-item-2"
    }
  ],
  "server_url": "sample_string"
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'ClusterBean.ServerUrl' Error:Field validation for 'ServerUrl' failed on the 'url' tag","userMessage":"Key: 'ClusterBean.ServerUrl...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/template/default/{appId}/{chartRefId}

- **Status**: FAIL
- **Duration**: 86.847041ms
- **Spec File**: ../../specs/template/templates.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/admin/policy/default

- **Status**: FAIL
- **Duration**: 74.604292ms
- **Spec File**: ../../specs/user/policy-management.yaml
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
- **Duration**: 72.065334ms
- **Spec File**: ../../specs/user/policy-management.yaml
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

### ❌ PUT /orchestrator/user/update/trigger/terminal

- **Status**: FAIL
- **Duration**: 73.637291ms
- **Spec File**: ../../specs/user/user-management.yaml
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
- **Duration**: 87.247916ms
- **Spec File**: ../../specs/user/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"users":[{"id":2,"email_id":"admin","roleFilters":[],"groups":[],"superAdmin":false,"userRoleGroups":[],"lastLoginTime":"2025-08-13T15:06:52.311391Z","timeoutWindow...

---

### ✅ GET /orchestrator/user/v2/{id}

- **Status**: PASS
- **Duration**: 96.134ms
- **Spec File**: ../../specs/user/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"email_id":"system","roleFilters":[],"groups":[],"superAdmin":false,"userRoleGroups":[],"lastLoginTime":"0001-01-01T00:00:00Z","timeoutWindowExpression":"000...

---

### ❌ DELETE /orchestrator/user/{id}

- **Status**: FAIL
- **Duration**: 105.268292ms
- **Spec File**: ../../specs/user/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","userMessage":"cannot delete system or admin user"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/{id}

- **Status**: PASS
- **Duration**: 79.453792ms
- **Spec File**: ../../specs/user/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"email_id":"system","roleFilters":[],"groups":[],"superAdmin":false,"userRoleGroups":[],"lastLoginTime":"0001-01-01T00:00:00Z","timeoutWindowExpression":"000...

---

### ❌ POST /orchestrator/user

- **Status**: FAIL
- **Duration**: 108.670833ms
- **Spec File**: ../../specs/user/user-management.yaml
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
- **Duration**: 111.819416ms
- **Spec File**: ../../specs/user/user-management.yaml
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
- **Duration**: 83.057833ms
- **Spec File**: ../../specs/user/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{EOF}]","userMessage":"EOF"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/detail/get

- **Status**: PASS
- **Duration**: 98.777375ms
- **Spec File**: ../../specs/user/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"email_id":"system","roleFilters":[],"groups":[],"superAdmin":false,"userRoleGroups":null,"lastLoginTime":"0001-01-01T00:00:00Z","timeoutWindowExpression":"...

---

### ✅ GET /orchestrator/user/sync/orchestratortocasbin

- **Status**: PASS
- **Duration**: 78.165542ms
- **Spec File**: ../../specs/user/user-management.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":true}

---

### ❌ GET /orchestrator/app/deployment-status/timeline/{appId}/{envId}

- **Status**: FAIL
- **Duration**: 78.469875ms
- **Spec File**: ../../specs/deployment/timeline.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/env/namespace/autocomplete

- **Status**: FAIL
- **Duration**: 76.751375ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{ERROR #42601 syntax error at or near \")\"}]","userMessage":"ERROR #42601 syntax error at or near \")\""}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/cluster

- **Status**: PASS
- **Duration**: 80.928458ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":2,"cluster_name":"isolated","description":"","active":true,"config":{"bearer_token":""},"prometheusAuth":{"isAnonymous":false},"defaultClusterComponent":[],"a...

---

### ❌ PUT /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 80.360416ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**:
```json
{
  "cd_argo_setup": true,
  "cluster_name": "sample_string",
  "created_on": "2023-01-01T00:00:00Z",
  "prometheus_url": "sample_string",
  "server_url": "sample_string"
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'ClusterBean.ServerUrl' Error:Field validation for 'ServerUrl' failed on the 'url' tag","userMessage":"Key: 'ClusterBean.ServerUrl...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/cluster/delete

- **Status**: FAIL
- **Duration**: 78.823875ms
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
- **Duration**: 73.2975ms
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
- **Duration**: 93.721041ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"environment_name":"devtron-demo","cluster_id":1,"cluster_name":"default_cluster","active":true,"default":false,"namespace":"devtron-demo","isClusterCdActiv...

---

### ❌ POST /orchestrator/env

- **Status**: FAIL
- **Duration**: 87.918834ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**:
```json
{
  "cluster_id": 1,
  "created_on": "2023-01-01T00:00:00Z",
  "default": true,
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
- **Duration**: 76.569167ms
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
- **Duration**: 73.246583ms
- **Spec File**: ../../specs/environment/core.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/flux-application

- **Status**: FAIL
- **Duration**: 75.566833ms
- **Spec File**: ../../specs/gitops/fluxcd.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{error in getting cluster ids}]","userMessage":"error in getting cluster ids"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/flux-application/app

- **Status**: FAIL
- **Duration**: 73.058125ms
- **Spec File**: ../../specs/gitops/fluxcd.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/gitops/configured

- **Status**: PASS
- **Duration**: 76.201458ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"allowCustomRepository":false,"authMode":"PASSWORD","exists":true,"provider":"GITHUB"}}

---

### ❌ POST /orchestrator/validate

- **Status**: FAIL
- **Duration**: 76.626458ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Request Payload**:
```json
{
  "enableTLSVerification": true,
  "gitLabGroupId": "sample_string",
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
- **Duration**: 76.101125ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"provider":"GITHUB","username":"systemsdt","token":"","gitLabGroupId":"","gitHubOrgId":"stage-gitops","host":"https://github.com/","active":true,"azureProje...

---

### ❌ POST /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 76.995666ms
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

### ✅ PUT /orchestrator/gitops/config

- **Status**: PASS
- **Duration**: 76.353916ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Request Payload**:
```json
{
  "gitHubOrgId": "sample_string",
  "provider": "GITLAB",
  "token": "sample_string"
}
```
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"successfulStages":null,"stageErrorMap":{"error in connecting with GITLAB":"gitlab client error: no gitlab group id found"},"validatedOn":"2025-08-13T15:06:54.41716...

---

### ❌ GET /orchestrator/gitops/config-by-provider

- **Status**: FAIL
- **Duration**: 73.22625ms
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
- **Duration**: 76.417375ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"provider":"GITHUB","username":"systemsdt","token":"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","gitLabGro...

---

### ✅ GET /orchestrator/module

- **Status**: PASS
- **Duration**: 122.302416ms
- **Spec File**: ../../specs/global-config/v1.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"name":"security.trivy","status":"installed","moduleResourcesStatus":null,"enabled":true,"moduleType":"security"},{"name":"security.clair","status":"installed","mo...

---

### ❌ POST /orchestrator/module

- **Status**: FAIL
- **Duration**: 72.038708ms
- **Spec File**: ../../specs/global-config/v1.yaml
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

### ✅ GET /orchestrator/server

- **Status**: PASS
- **Duration**: 76.18575ms
- **Spec File**: ../../specs/global-config/v1.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"status":"unknown","releaseName":"devtron","installationType":"enterprise"}}

---

### ❌ POST /orchestrator/server

- **Status**: FAIL
- **Duration**: 76.65825ms
- **Spec File**: ../../specs/global-config/v1.yaml
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

### ❌ PATCH /orchestrator/app/env/patch

- **Status**: FAIL
- **Duration**: 78.907792ms
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

### ✅ GET /orchestrator/chartref/autocomplete/{appId}

- **Status**: PASS
- **Duration**: 82.022542ms
- **Spec File**: ../../specs/helm/dynamic-charts.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"chartRefs":[{"id":10,"version":"3.9.0","name":"Rollout Deployment","description":"","userUploaded":false,"isAppMetricsSupported":true},{"id":11,"version":"3.10.0",...

---

### ❌ POST /orchestrator/app-store/chart-provider/update

- **Status**: FAIL
- **Duration**: 77.126625ms
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

### ❌ PUT /orchestrator/chart-group/

- **Status**: FAIL
- **Duration**: 87.034791ms
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

### ❌ POST /orchestrator/chart-group/

- **Status**: FAIL
- **Duration**: 76.690042ms
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

### ✅ POST /orchestrator/chart-repo/create

- **Status**: PASS
- **Duration**: 333.146208ms
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

### ✅ POST /orchestrator/chart-repo/sync-charts

- **Status**: PASS
- **Duration**: 7.277046209s
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"status":"ok"}}

---

### ❌ POST /orchestrator/chart-repo/update

- **Status**: FAIL
- **Duration**: 330.696416ms
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

### ✅ GET /orchestrator/chart-repo/list

- **Status**: PASS
- **Duration**: 80.206583ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"default-chartmuseum","url":"http://devtron-chartmuseum.devtroncd:8080/","authMode":"ANONYMOUS","active":true,"default":true,"allow_insecure_connecti...

---

### ✅ POST /orchestrator/chart-repo/validate

- **Status**: PASS
- **Duration**: 88.655666ms
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

### ✅ GET /orchestrator/app-store/chart-provider/list

- **Status**: PASS
- **Duration**: 86.064667ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":"prakhar","name":"prakhar","active":true,"isEditable":true,"isOCIRegistry":true,"registryProvider":"docker-hub"},{"id":"1","name":"default-chartmuseum","activ...

---

### ✅ GET /orchestrator/chart-group/list

- **Status**: PASS
- **Duration**: 77.255959ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":null}

---

### ❌ DELETE /orchestrator/chart-group/{id}

- **Status**: FAIL
- **Duration**: 74.069875ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ POST /orchestrator/app-store/chart-provider/sync-chart

- **Status**: PASS
- **Duration**: 7.18167725s
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
- **Duration**: 82.255ms
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

### ✅ GET /orchestrator/chart-repo/{id}

- **Status**: PASS
- **Duration**: 77.828667ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"name":"default-chartmuseum","url":"http://devtron-chartmuseum.devtroncd:8080/","authMode":"ANONYMOUS","active":true,"default":true,"allow_insecure_connectio...

---

### ❌ GET /orchestrator/app/ci-pipeline/{ciPipelineId}/linked-ci/downstream/cd

- **Status**: FAIL
- **Duration**: 77.866667ms
- **Spec File**: ../../specs/ci-pipeline/downstream-linked-ci-view-spec.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{invalid CiPipelineId 1}]","userMessage":"invalid CiPipelineId 1"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/ci-pipeline/{ciPipelineId}/linked-ci/downstream/env

- **Status**: FAIL
- **Duration**: 77.579916ms
- **Spec File**: ../../specs/ci-pipeline/downstream-linked-ci-view-spec.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{invalid CiPipelineId 1}]","userMessage":"invalid CiPipelineId 1"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/resource/history/deployment/cd-pipeline/v1

- **Status**: FAIL
- **Duration**: 76.127292ms
- **Spec File**: ../../specs/deployment/history.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","internalMessage":"invalid format filter criteria!","userMessage":"invalid format filter criteria!"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 74.208083ms
- **Spec File**: ../../specs/external-app/external-links-specs.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ✅ GET /orchestrator/external-links

- **Status**: PASS
- **Duration**: 79.630667ms
- **Spec File**: ../../specs/external-app/external-links-specs.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":3,"name":"sample-external-link","url":"https://grafana.example.com","active":true,"monitoringToolId":1,"type":"appLevel","identifiers":[{"type":"devtron-app",...

---

### ✅ POST /orchestrator/external-links

- **Status**: PASS
- **Duration**: 81.057709ms
- **Spec File**: ../../specs/external-app/external-links-specs.yaml
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
- **Duration**: 78.777ms
- **Spec File**: ../../specs/external-app/external-links-specs.yaml
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
- **Duration**: 76.769958ms
- **Spec File**: ../../specs/external-app/external-links-specs.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"Grafana","icon":"","category":2},{"id":2,"name":"Kibana","icon":"","category":2},{"id":3,"name":"Newrelic","icon":"","category":2},{"id":4,"name":"C...

---

### ✅ GET /orchestrator/version

- **Status**: PASS
- **Duration**: 77.908458ms
- **Spec File**: ../../specs/global-config/version.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"result":{"gitCommit":"908eae83","buildTime":"2025-08-05T20:18:14Z","serverMode":"FULL"}}

---

### ❌ GET /orchestrator/app-store/discover/search

- **Status**: FAIL
- **Duration**: 77.322291ms
- **Spec File**: ../../specs/helm/app-store.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/app-store/discover

- **Status**: PASS
- **Duration**: 737.663833ms
- **Spec File**: ../../specs/helm/app-store.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":65,"appStoreApplicationVersionId":4028,"name":"ai-agent","chart_repo_id":2,"docker_artifact_store_id":"","chart_name":"devtron","icon":"","active":true,"chart...

---

### ✅ GET /orchestrator/app-store/discover/application/chartInfo/{appStoreApplicationVersionId}

- **Status**: PASS
- **Duration**: 87.552667ms
- **Spec File**: ../../specs/helm/app-store.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"appStoreApplicationVersionId":1,"readme":"# cluster-autoscaler\n\nScales Kubernetes worker nodes within autoscaling groups.\n\n## TL;DR\n\n```console\n$ helm repo ...

---

### ✅ GET /orchestrator/app-store/discover/application/{appStoreId}/version/autocomplete

- **Status**: PASS
- **Duration**: 78.065291ms
- **Spec File**: ../../specs/helm/app-store.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"version":"9.50.0","id":21297},{"version":"9.49.0","id":5507},{"version":"9.48.0","id":1},{"version":"9.47.0","id":2},{"version":"9.46.6","id":3},{"version":"9.46....

---

### ✅ GET /orchestrator/app-store/discover/application/{id}

- **Status**: PASS
- **Duration**: 187.492458ms
- **Spec File**: ../../specs/helm/app-store.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"version":"9.48.0","appVersion":"1.33.0","created":"2025-07-11T21:16:00.149315Z","deprecated":false,"description":"Scales Kubernetes worker nodes within auto...

---

### ❌ POST /orchestrator/webhook/git

- **Status**: FAIL
- **Duration**: 89.730167ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Request Payload**:
```json
{
  "compare": "sample_string",
  "head_commit": {
    "message": "sample_string",
    "modified": [
      "sample_string"
    ],
    "removed": [
      "sample_string"
    ]
  },
  "sender": {
    "email": "sample_string",
    "name": "sample_string",
    "username": "sample_string"
  }
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/git/{gitHostId}

- **Status**: FAIL
- **Duration**: 87.591417ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Request Payload**:
```json
{
  "compare": "sample_string",
  "head_commit": {
    "modified": [
      "sample_string"
    ],
    "removed": [
      "sample_string"
    ],
    "timestamp": "2023-01-01T00:00:00Z"
  },
  "repository": {
    "default_branch": "sample_string",
    "name": "sample_string",
    "url": "sample_string"
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
- **Duration**: 80.496666ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Request Payload**:
```json
{
  "compare": "sample_string",
  "head_commit": {
    "added": [
      "sample_string"
    ],
    "author": {
      "email": "sample_string",
      "name": "sample_string",
      "username": "sample_string"
    },
    "committer": {
      "email": "sample_string",
      "name": "sample_string",
      "username": "sample_string"
    }
  },
  "repository": {
    "default_branch": "sample_string",
    "name": "sample_string",
    "private": true
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
- **Duration**: 92.154667ms
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
- **Duration**: 79.220125ms
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
- **Response Code**: 401
- **Error/Msg**: {"code":401,"status":"Unauthorized","errors":[{"code":"6005","internalMessage":"no token provided"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/config/environment/cm

- **Status**: FAIL
- **Duration**: 76.428416ms
- **Spec File**: ../../specs/template/config-maps.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "configData": [
    {
      "data": {},
      "externalSecretType": "sample_string",
      "name": "sample_string",
      "subPath": "sample_string",
      "type": "CONFIGMAP"
    },
    {
      "data": {},
      "externalSecretType": "sample_string",
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

### ✅ GET /orchestrator/config/global/cm/{appId}

- **Status**: PASS
- **Duration**: 78.210959ms
- **Spec File**: ../../specs/template/config-maps.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":0,"appId":1,"configData":[],"isDeletable":false,"isExpressEdit":false}}

---

### ❌ DELETE /orchestrator/config/global/cm/{appId}/{id}

- **Status**: FAIL
- **Duration**: 105.685917ms
- **Spec File**: ../../specs/template/config-maps.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/config/environment/cm/{appId}/{envId}

- **Status**: PASS
- **Duration**: 77.618416ms
- **Spec File**: ../../specs/template/config-maps.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":0,"appId":1,"environmentId":1,"configData":[],"isDeletable":false,"isExpressEdit":false}}

---

### ❌ GET /orchestrator/config/environment/cm/edit/{appId}/{envId}/{id}

- **Status**: FAIL
- **Duration**: 74.554625ms
- **Spec File**: ../../specs/template/config-maps.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/config/environment/cm/{appId}/{envId}/{id}

- **Status**: FAIL
- **Duration**: 163.744541ms
- **Spec File**: ../../specs/template/config-maps.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/config/global/cm/edit/{appId}/{id}

- **Status**: FAIL
- **Duration**: 100.178ms
- **Spec File**: ../../specs/template/config-maps.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/config/global/cm

- **Status**: FAIL
- **Duration**: 75.916084ms
- **Spec File**: ../../specs/template/config-maps.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "configData": [
    {
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
      "subPath": "sample_string",
      "type": "CONFIGMAP"
    },
    {
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
- **Duration**: 75.736875ms
- **Spec File**: ../../specs/template/config-maps.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "configData": [
    {
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
      "subPath": "sample_string",
      "type": "CONFIGMAP"
    },
    {
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
      "subPath": "sample_string",
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

### ❌ DELETE /orchestrator/user/role/group/bulk

- **Status**: FAIL
- **Duration**: 78.212292ms
- **Spec File**: ../../specs/user/core.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{EOF}]","userMessage":"EOF"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/role/group/v2

- **Status**: PASS
- **Duration**: 78.647625ms
- **Spec File**: ../../specs/user/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"roleGroups":[{"id":2,"roleFilters":[],"superAdmin":false,"canManageAllAccess":false,"hasAccessManagerPermission":false}],"totalCount":1}}

---

### ✅ GET /orchestrator/user/v2

- **Status**: PASS
- **Duration**: 77.687917ms
- **Spec File**: ../../specs/user/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"users":[{"id":2,"email_id":"admin","roleFilters":[],"groups":[],"superAdmin":false,"userRoleGroups":[],"lastLoginTime":"2025-08-13T15:07:14.363963Z","timeoutWindow...

---

### ✅ GET /orchestrator/api-token

- **Status**: PASS
- **Duration**: 80.666083ms
- **Spec File**: ../../specs/user/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"description":"some description","expireAtInMs":12344546,"id":2,"name":"some-name","token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6IkFQSS1UT0tFTjpzb21lLW...

---

### ❌ POST /orchestrator/api-token

- **Status**: FAIL
- **Duration**: 78.315084ms
- **Spec File**: ../../specs/user/core.yaml
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

### ❌ PUT /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 75.688584ms
- **Spec File**: ../../specs/user/core.yaml
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
- **Duration**: 76.060542ms
- **Spec File**: ../../specs/user/core.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{api-token corresponds to apiTokenId '1' is not found}]","userMessage":"api-token corresponds to apiTokenId '1'...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PUT /orchestrator/user

- **Status**: FAIL
- **Duration**: 78.066583ms
- **Spec File**: ../../specs/user/core.yaml
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

### ❌ POST /orchestrator/user

- **Status**: FAIL
- **Duration**: 76.44975ms
- **Spec File**: ../../specs/user/core.yaml
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
- **Duration**: 76.504458ms
- **Spec File**: ../../specs/user/core.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{EOF}]","userMessage":"EOF"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 77.100875ms
- **Spec File**: ../../specs/user/core.yaml
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

### ❌ POST /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 77.719791ms
- **Spec File**: ../../specs/user/core.yaml
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

### ❌ POST /orchestrator/ci-pipeline/trigger

- **Status**: FAIL
- **Duration**: 71.706958ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Request Payload**:
```json
{
  "ciPipelineMaterials": [
    {
      "Active": true,
      "GitCommit": {
        "Author": "John Doe",
        "Date": "2023-01-15T14:30:22Z",
        "Message": "Update README"
      },
      "GitMaterialId": 2
    },
    {
      "Active": true,
      "GitCommit": {
        "Author": "John Doe",
        "Date": "2023-01-15T14:30:22Z",
        "Message": "Update README"
      },
      "GitMaterialId": 2
    }
  ],
  "invalidateCache": true,
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
- **Duration**: 78.519ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ GET /orchestrator/app/commit-info/{ciPipelineMaterialId}/{gitHash}

- **Status**: FAIL
- **Duration**: 118.43075ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/config/autocomplete

- **Status**: PASS
- **Duration**: 79.437417ms
- **Spec File**: ../../specs/environment/config-diff.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"resourceConfig":[{"id":0,"name":"","configState":"Published","type":"Deployment Template","configStage":""}]}}

---

### ❌ GET /config/compare/{resource}

- **Status**: FAIL
- **Duration**: 73.655958ms
- **Spec File**: ../../specs/environment/config-diff.yaml
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

### ❌ GET /config/data

- **Status**: FAIL
- **Duration**: 73.168458ms
- **Spec File**: ../../specs/environment/config-diff.yaml
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

### ❌ POST /config/manifest

- **Status**: FAIL
- **Duration**: 71.725167ms
- **Spec File**: ../../specs/environment/config-diff.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "mergeStrategy": "sample_string",
  "resourceId": 1,
  "resourceName": "sample_string",
  "resourceType": "sample_string",
  "values": {}
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

### ✅ GET /orchestrator/cluster

- **Status**: PASS
- **Duration**: 102.943375ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":2,"cluster_name":"isolated","description":"","active":true,"config":{"bearer_token":""},"prometheusAuth":{"isAnonymous":false},"defaultClusterComponent":[],"a...

---

### ❌ PUT /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 76.903917ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Request Payload**:
```json
{
  "defaultClusterComponent": [
    {
      "appId": 1,
      "envId": 1,
      "envname": "sample_string"
    },
    {
      "appId": 1,
      "envId": 1,
      "envname": "sample_string"
    }
  ],
  "k8sVersion": "sample_string",
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
- **Duration**: 77.36ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":2,"cluster_name":"isolated","description":"","active":false,"defaultClusterComponent":null,"agentInstallationStage":0,"k8sVersion":"","insecureSkipTlsVerify":...

---

### ✅ GET /orchestrator/user/role/group/v2

- **Status**: PASS
- **Duration**: 76.661958ms
- **Spec File**: ../../specs/user/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"roleGroups":[{"id":2,"roleFilters":[],"superAdmin":false,"canManageAllAccess":false,"hasAccessManagerPermission":false}],"totalCount":1}}

---

### ❌ POST /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 77.354167ms
- **Spec File**: ../../specs/user/group-policy.yaml
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
- **Duration**: 76.141458ms
- **Spec File**: ../../specs/user/group-policy.yaml
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

### ❌ DELETE /orchestrator/user/role/group/{id}

- **Status**: FAIL
- **Duration**: 76.687833ms
- **Spec File**: ../../specs/user/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/user/role/group/{id}

- **Status**: FAIL
- **Duration**: 76.740291ms
- **Spec File**: ../../specs/user/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"role group not found: 1","userMessage":"role group with ID '1' not found","userDetailMessage":"The requested role group do...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/user/role/group

- **Status**: PASS
- **Duration**: 79.514458ms
- **Spec File**: ../../specs/user/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":2,"roleFilters":[],"superAdmin":false,"canManageAllAccess":false,"hasAccessManagerPermission":false}]}

---

### ❌ POST /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 79.164791ms
- **Spec File**: ../../specs/user/group-policy.yaml
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
- **Duration**: 76.479917ms
- **Spec File**: ../../specs/user/group-policy.yaml
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
- **Duration**: 75.423459ms
- **Spec File**: ../../specs/user/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{EOF}]","userMessage":"EOF"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/role/group/detailed/get

- **Status**: PASS
- **Duration**: 76.144084ms
- **Spec File**: ../../specs/user/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":2,"roleFilters":[],"superAdmin":false,"canManageAllAccess":false,"hasAccessManagerPermission":false}]}

---

### ❌ GET /orchestrator/user/role/group/search

- **Status**: FAIL
- **Duration**: 76.673791ms
- **Spec File**: ../../specs/user/group-policy.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"11001","internalMessage":"invalid path parameter id: search","userMessage":"Invalid path parameter 'id'","userDetailMessage":"Please check the par...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/k8s/resources/ephemeralContainers

- **Status**: FAIL
- **Duration**: 74.230833ms
- **Spec File**: ../../specs/kubernetes/ephemeral-containers.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/k8s/resources/ephemeralContainers

- **Status**: FAIL
- **Duration**: 73.736167ms
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
  "externalArgoApplicationName": "sample_string",
  "namespace": "sample_string",
  "podName": "sample_string"
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/k8s/resource

- **Status**: FAIL
- **Duration**: 78.562708ms
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
- **Duration**: 79.264542ms
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
- **Duration**: 76.087958ms
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
- **Duration**: 76.276167ms
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
- **Duration**: 82.1785ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"404","userMessage":"error on getting resource from k8s"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/k8s/resource/urls

- **Status**: FAIL
- **Duration**: 74.654375ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/k8s/resources/apply

- **Status**: FAIL
- **Duration**: 75.826292ms
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
- **Duration**: 73.503167ms
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
- **Duration**: 96.008625ms
- **Spec File**: ../../specs/kubernetes/resources.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"apiResources":[{"gvk":{"Group":"","Version":"v1","Kind":"ConfigMap"},"gvr":{"Group":"","Version":"v1","Resource":"configmaps"},"namespaced":true,"shortNames":["cm"...

---

### ❌ GET /orchestrator/security/cve/control/list

- **Status**: FAIL
- **Duration**: 72.888ms
- **Spec File**: ../../specs/security/security-policy.yml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/security/cve/control/list

- **Status**: FAIL
- **Duration**: 72.912291ms
- **Spec File**: ../../specs/security/security-policy.yml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /orchestrator/security/cve/control/list

- **Status**: FAIL
- **Duration**: 73.85825ms
- **Spec File**: ../../specs/security/security-policy.yml
- **Request Payload**:
```json
{
  "action": "block",
  "appId": 1,
  "clusterId": 1
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/security/cve/control/list

- **Status**: FAIL
- **Duration**: 71.448792ms
- **Spec File**: ../../specs/security/security-policy.yml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ PUT /orchestrator/sso/update

- **Status**: PASS
- **Duration**: 102.446708ms
- **Spec File**: ../../specs/user/sso-configuration.yaml
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
- **Duration**: 77.586792ms
- **Spec File**: ../../specs/user/sso-configuration.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"name":"google","label":"sample_string","url":"https://devtron.example.com/orchestrator","config":{"id":"","type":"","name":"","config":null},"active":true,"...

---

### ❌ GET /orchestrator/sso

- **Status**: FAIL
- **Duration**: 73.984625ms
- **Spec File**: ../../specs/user/sso-configuration.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ✅ POST /orchestrator/sso/create

- **Status**: PASS
- **Duration**: 93.673125ms
- **Spec File**: ../../specs/user/sso-configuration.yaml
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
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":4,"name":"sample_string","label":"sample_string","url":"https://devtron.example.com/orchestrator","config":{},"active":true,"globalAuthConfigType":""}}

---

### ✅ GET /orchestrator/sso/list

- **Status**: PASS
- **Duration**: 73.046916ms
- **Spec File**: ../../specs/user/sso-configuration.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":4,"name":"sample_string","label":"sample_string","url":"https://devtron.example.com/orchestrator","active":true,"globalAuthConfigType":""},{"id":1,"name":"goo...

---

### ❌ GET /orchestrator/infra-config/profile

- **Status**: FAIL
- **Duration**: 79.564625ms
- **Spec File**: ../../specs/buildInfraConfig/build-infra-config.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /orchestrator/infra-config/profile

- **Status**: FAIL
- **Duration**: 72.640917ms
- **Spec File**: ../../specs/buildInfraConfig/build-infra-config.yaml
- **Request Payload**:
```json
{
  "appCount": 1,
  "configurations": [
    {
      "active": true,
      "id": 1,
      "key": "cpu_limits",
      "platform": "linux/amd64",
      "value": "0.5"
    },
    {
      "active": true,
      "id": 2,
      "key": "cpu_limits",
      "platform": "linux/amd64",
      "value": "0.5"
    }
  ],
  "createdBy": 1,
  "targetPlatforms": [
    "linux/amd64"
  ],
  "updatedBy": 1
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PATCH /orchestrator/app/ci-pipeline/patch-source

- **Status**: FAIL
- **Duration**: 77.150875ms
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

### ❌ POST /orchestrator/app/ci-pipeline/patch

- **Status**: FAIL
- **Duration**: 77.065209ms
- **Spec File**: ../../specs/ci-pipeline/docker-build.yaml
- **Request Payload**:
```json
{
  "action": 1,
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
        "gitMaterialId": 1,
        "value": "sample_string"
      },
      {
        "active": true,
        "gitMaterialId": 1,
        "value": "sample_string"
      }
    ],
    "id": 1,
    "isExternal": true,
    "isManual": true,
    "name": "sample_string",
    "scanEnabled": true
  },
  "isCloneJob": true
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'CiPatchRequest.CiPipeline.CiMaterial[0].Source' Error:Field validation for 'Source' failed on the 'dive' tag","userMessage":"Key:...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/ci-pipeline/{appId}

- **Status**: PASS
- **Duration**: 87.272541ms
- **Spec File**: ../../specs/ci-pipeline/docker-build.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"appId":1,"dockerRegistry":"quay","dockerRepository":"devtron/test","ciBuildConfig":{"id":1,"gitMaterialId":1,"buildContextGitMaterialId":1,"useRootBuildCont...

---

### ✅ GET /orchestrator/app/wf/all/component-names/{appId}

- **Status**: PASS
- **Duration**: 93.875667ms
- **Spec File**: ../../specs/ci-pipeline/docker-build.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"workflows":[{"id":1,"name":"wf-1-lcu5","ciPipelineId":0,"ciPipelineName":"","cdPipelines":null},{"id":2,"name":"wf-1-4rww","ciPipelineId":0,"ciPipelineName":"","cd...

---

### ✅ GET /orchestrator/deployment/template/fetch

- **Status**: PASS
- **Duration**: 97.766917ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":10,"name":"Rollout Deployment","chartDescription":"This chart deploys an advanced version of deployment that supports Blue/Green and Canary deployments. For f...

---

### ✅ PUT /orchestrator/deployment/template/upload

- **Status**: PASS
- **Duration**: 80.339916ms
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
- **Duration**: 76.661333ms
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

### ❌ GET /orchestrator/app/history/deployed-configuration/all/latest/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 75.934375ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"latest\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"latest\": invalid syntax"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/history/deployed-configuration/all/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 89.628ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"deploymentTemplate":{"templateName":"Deployment","templateVersion":"4.21.0","isAppMetricsEnabled":false,"codeEditorValue":{"displayName":"values.yaml","value":"{\"...

---

### ❌ PUT /orchestrator/application/rollback

- **Status**: FAIL
- **Duration**: 104.720542ms
- **Spec File**: ../../specs/deployment/rollback.yaml
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
- **Duration**: 765.530041ms
- **Spec File**: ../../specs/deployment/rollback.yaml
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

### ✅ POST /orchestrator/chart-repo/create

- **Status**: PASS
- **Duration**: 83.508083ms
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
- **Duration**: 83.639792ms
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

### ✅ POST /orchestrator/chart-repo/validate

- **Status**: PASS
- **Duration**: 85.571709ms
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

### ✅ GET /orchestrator/notification/channel/webhook/{id}

- **Status**: PASS
- **Duration**: 78.796208ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"userId":0,"webhookUrl":"","configName":"","header":null,"payload":"","description":"","id":0}}

---

### ❌ POST /orchestrator/notification/v2

- **Status**: FAIL
- **Duration**: 76.348458ms
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
- **Duration**: 76.70675ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{EOF}]","userMessage":"EOF"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/notification

- **Status**: FAIL
- **Duration**: 72.857833ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/notification

- **Status**: FAIL
- **Duration**: 76.676209ms
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
- **Duration**: 76.540625ms
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

### ✅ GET /orchestrator/notification/channel/smtp/{id}

- **Status**: PASS
- **Duration**: 77.242333ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":0,"port":"","host":"","authType":"","authUser":"","authPassword":"","fromEmail":"","configName":"","description":"","ownerId":0,"default":false,"deleted":false...

---

### ✅ GET /orchestrator/notification/variables

- **Status**: PASS
- **Duration**: 76.823292ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"devtronAppId":"{{devtronAppId}}","devtronAppName":"{{devtronAppName}}","devtronApprovedByEmail":"{{devtronApprovedByEmail}}","devtronBuildGitCommitHash":"{{devtron...

---

### ✅ GET /orchestrator/notification/channel/autocomplete/{type}

- **Status**: PASS
- **Duration**: 77.608083ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[]}

---

### ✅ GET /orchestrator/notification/channel/slack/{id}

- **Status**: PASS
- **Duration**: 77.328667ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"userId":0,"teamId":0,"webhookUrl":"","configName":"","description":"","id":0}}

---

### ❌ GET /orchestrator/notification/recipient

- **Status**: FAIL
- **Duration**: 73.658208ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ POST /orchestrator/notification/search

- **Status**: PASS
- **Duration**: 84.452625ms
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

### ❌ DELETE /orchestrator/notification/channel

- **Status**: FAIL
- **Duration**: 76.476125ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{EOF}]","userMessage":"EOF"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/notification/channel

- **Status**: PASS
- **Duration**: 79.250708ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"slackConfigs":[],"webhookConfigs":[],"sesConfigs":[{"userId":2,"teamId":0,"region":"ap-south-1","accessKey":"AKIAWEAVHFOOSNAORPHU","secretKey":"**********","fromEm...

---

### ❌ POST /orchestrator/notification/channel

- **Status**: FAIL
- **Duration**: 78.982041ms
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

### ✅ GET /orchestrator/notification/channel/ses/{id}

- **Status**: PASS
- **Duration**: 75.748584ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"userId":2,"teamId":0,"region":"ap-south-1","accessKey":"AKIAWEAVHFOOSNAORPHU","secretKey":"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","fromEmail":"watcher@devtron.i...

---

### ❌ POST /orchestrator/user/resource/options/{kind}/{version}

- **Status**: FAIL
- **Duration**: 73.581208ms
- **Spec File**: ../../specs/user/userResource.yaml
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

### ❌ GET /devtron/auth/verify/v2

- **Status**: FAIL
- **Duration**: 76.079666ms
- **Spec File**: ../../specs/authentication/authentication.yaml
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

### ❌ GET /orchestrator/api/dex/{path}

- **Status**: FAIL
- **Duration**: 132.866334ms
- **Spec File**: ../../specs/authentication/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 502
- **Error/Msg**: {}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 502

---

### ❌ GET /orchestrator/login

- **Status**: FAIL
- **Duration**: 311.721291ms
- **Spec File**: ../../specs/authentication/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: Failed to query provider "https://devtron.example.com/orchestrator/api/dex": Get "https://devtron.example.com/orchestrator/api/dex/.well-known/openid-configuration": dial tcp: lookup devtron.example.c...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'F' looking for beginning of value

---

### ❌ GET /orchestrator/rbac/roles/default

- **Status**: FAIL
- **Duration**: 74.710292ms
- **Spec File**: ../../specs/authentication/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/refresh

- **Status**: FAIL
- **Duration**: 75.012375ms
- **Spec File**: ../../specs/authentication/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: internal server error

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/auth/login

- **Status**: FAIL
- **Duration**: 309.770125ms
- **Spec File**: ../../specs/authentication/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: Failed to query provider "https://devtron.example.com/orchestrator/api/dex": Get "https://devtron.example.com/orchestrator/api/dex/.well-known/openid-configuration": dial tcp: lookup devtron.example.c...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'F' looking for beginning of value

---

### ❌ GET /orchestrator/auth/callback

- **Status**: FAIL
- **Duration**: 81.110375ms
- **Spec File**: ../../specs/authentication/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: Failed to query provider "https://devtron.example.com/orchestrator/api/dex": Get "https://devtron.example.com/orchestrator/api/dex/.well-known/openid-configuration": dial tcp: lookup devtron.example.c...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/devtron/auth/verify

- **Status**: PASS
- **Duration**: 145.9765ms
- **Spec File**: ../../specs/authentication/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":true}

---

### ✅ GET /orchestrator/user/check/roles

- **Status**: PASS
- **Duration**: 79.002875ms
- **Spec File**: ../../specs/authentication/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"roles":["role:super-admin___","role:chart-group_admin"],"superAdmin":true}}

---

### ❌ POST /orchestrator/api/v1/session

- **Status**: FAIL
- **Duration**: 74.9305ms
- **Spec File**: ../../specs/authentication/authentication.yaml
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

### ✅ GET /orchestrator/user/role/cache

- **Status**: PASS
- **Duration**: 76.740833ms
- **Spec File**: ../../specs/authentication/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":"{}"}

---

### ✅ GET /orchestrator/user/role/cache/invalidate

- **Status**: PASS
- **Duration**: 76.085917ms
- **Spec File**: ../../specs/authentication/authentication.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":"Cache Cleaned Successfully"}

---

### ✅ POST /orchestrator/batch/v1beta1/cd-pipeline

- **Status**: PASS
- **Duration**: 80.790375ms
- **Spec File**: ../../specs/environment/bulk-delete.yaml
- **Request Payload**:
```json
{
  "envIds": [
    1
  ],
  "forceDelete": true,
  "projectNames": [
    "sample_string"
  ]
}
```
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"cdPipelines":null,"ciPipelines":null,"appWorkflows":null}}

---

### ❌ POST /orchestrator/gitops/validate

- **Status**: FAIL
- **Duration**: 83.665875ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Request Payload**:
```json
{
  "bitBucketProjectKey": "sample_string",
  "isCADataPresent": true,
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
- **Duration**: 158.079625ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"provider":"GITHUB","username":"systemsdt","token":"","gitLabGroupId":"","gitHubOrgId":"stage-gitops","host":"https://github.com/","active":true,"azureProje...

---

### ❌ POST /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 84.162541ms
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
- **Duration**: 78.674417ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Request Payload**:
```json
{
  "azureProjectName": "sample_string",
  "bitBucketProjectKey": "sample_string",
  "provider": "GITLAB"
}
```
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/gitops/config-by-provider

- **Status**: FAIL
- **Duration**: 73.980667ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/gitops/config/{id}

- **Status**: PASS
- **Duration**: 133.323458ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"provider":"GITHUB","username":"systemsdt","token":"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","gitLabGro...

---

### ✅ GET /orchestrator/gitops/configured

- **Status**: PASS
- **Duration**: 78.138375ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"allowCustomRepository":false,"authMode":"PASSWORD","exists":true,"provider":"GITHUB"}}

---

### ❌ GET /orchestrator/git/host/{id}/event

- **Status**: FAIL
- **Duration**: 76.188166ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"{id}\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"{id}\": invalid synta...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/git/host

- **Status**: PASS
- **Duration**: 79.567458ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"Github","active":true,"webhookUrl":"","webhookSecret":"","eventTypeHeader":"","secretHeader":"","secretValidator":""},{"id":2,"name":"Bitbucket Clou...

---

### ❌ POST /orchestrator/git/host

- **Status**: FAIL
- **Duration**: 78.1885ms
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
- **Duration**: 83.210084ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"{eventId}\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"{eventId}\": inv...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/webhook-meta-config/{gitProviderId}

- **Status**: FAIL
- **Duration**: 85.195417ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{ERROR #22P02 invalid input syntax for type integer: \"{gitProviderId}\"}]","userMessage":"ERROR #22P02 invalid...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/{id}

- **Status**: FAIL
- **Duration**: 75.713792ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Request Payload**: (none)
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"{id}\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"{id}\": invalid synta...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/security/scan/executionDetail

- **Status**: FAIL
- **Duration**: 78.255917ms
- **Spec File**: ../../specs/security/security-dashboard-apis.yml
- **Request Payload**: (none)
- **Response Code**: 403
- **Error/Msg**: {"code":403,"status":"Forbidden","errors":[{"code":"000","internalMessage":"[{unauthorized user}]","userMessage":"Unauthorized User"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 403

---

### ✅ POST /orchestrator/app/edit/projects

- **Status**: PASS
- **Duration**: 80.967208ms
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

### ❌ POST /orchestrator/core/v1beta1/application

- **Status**: FAIL
- **Duration**: 79.351709ms
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
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'AppDetail.Metadata.ProjectName' Error:Field validation for 'ProjectName' failed on the 'required' tag","userMessage":"Key: 'AppDe...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/autocomplete

- **Status**: PASS
- **Duration**: 76.985833ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":2,"name":"gitops-sanity-application","createdBy":"admin","description":""},{"id":3,"name":"pk-test","createdBy":"admin","description":""},{"id":8,"name":"pk-t...

---

### ❌ POST /orchestrator/app/workflow

- **Status**: FAIL
- **Duration**: 123.544875ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "cdPipelines": [
    {},
    {}
  ],
  "ciPipeline": {}
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ POST /orchestrator/app/edit

- **Status**: PASS
- **Duration**: 83.810292ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**:
```json
{
  "appName": "sample_string",
  "description": "sample_string",
  "id": 1,
  "teamId": 1
}
```
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"appName":"sample_string","description":"sample_string","material":null,"teamId":1,"templateId":0,"appType":0,"workflowCacheConfig":{"type":"","value":false,...

---

### ✅ POST /orchestrator/app/list

- **Status**: PASS
- **Duration**: 84.161083ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**:
```json
{
  "appNameSearch": "sample_string",
  "statuses": [
    "Healthy"
  ],
  "teamIds": [
    1
  ]
}
```
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"appContainers":null,"appCount":0,"deploymentGroup":{"id":0,"name":"","appCount":0,"noOfApps":"","environmentId":0,"ciPipelineId":0,"ciMaterialDTOs":null,"isVirtual...

---

### ✅ POST /orchestrator/app/workflow/clone

- **Status**: PASS
- **Duration**: 76.663375ms
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

### ❌ POST /orchestrator/app

- **Status**: FAIL
- **Duration**: 75.93925ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**:
```json
{
  "appName": "sample_string",
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
  ],
  "teamId": 1
}
```
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'CreateAppDTO.AppName' Error:Field validation for 'AppName' failed on the 'name-component' tag","userMessage":"Key: 'CreateAppDTO....

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/app/workflow/{app-wf-id}/app/{app-id}

- **Status**: FAIL
- **Duration**: 88.46275ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ GET /orchestrator/app/details/{appId}

- **Status**: FAIL
- **Duration**: 80.61825ms
- **Spec File**: ../../specs/application/core.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ GET /orchestrator/app/history/deployed-component/detail/{appId}/{pipelineId}/{id}

- **Status**: FAIL
- **Duration**: 75.062792ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","result":"invalid historyComponent"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/history/deployed-component/list/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 75.688875ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Request Payload**: (none)
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","result":"invalid historyComponent"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/history/deployed-configuration/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 78.623833ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Request Payload**: (none)
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":8,"name":"DEPLOYMENT_TEMPLATE"},{"id":3,"name":"PIPELINE_STRATEGY"}]}

---

### ❌ POST /orchestrator/deployment/pipeline/configure

- **Status**: FAIL
- **Duration**: 76.2355ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Request Payload**:
```json
{
  "appId": 1,
  "deploymentStrategy": "ROLLING",
  "deploymentType": "HELM",
  "envId": 1,
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
- **Duration**: 72.817333ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/deployment/pipeline/rollback

- **Status**: FAIL
- **Duration**: 131.234375ms
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
- **Duration**: 72.449666ms
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

### ❌ POST /batch/bulk/v1beta1/application

- **Status**: FAIL
- **Duration**: 93.648125ms
- **Spec File**: ../../specs/jobs/batch.yaml
- **Request Payload**:
```json
{
  "apiVersion": "sample_string",
  "kind": "sample_string",
  "spec": {
    "configMap": {
      "patchJson": "sample_string"
    },
    "deploymentTemplate": {
      "patchJson": "sample_string"
    },
    "secret": {
      "patchJson": "sample_string"
    }
  }
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

### ❌ POST /batch/bulk/v1beta1/application/dryrun

- **Status**: FAIL
- **Duration**: 74.013375ms
- **Spec File**: ../../specs/jobs/batch.yaml
- **Request Payload**:
```json
{
  "configMap": {
    "patchJson": "sample_string"
  },
  "includes": [
    "sample_string"
  ],
  "secret": {
    "patchJson": "sample_string"
  }
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

### ❌ POST /orchestrator/batch/operate

- **Status**: FAIL
- **Duration**: 75.7675ms
- **Spec File**: ../../specs/jobs/batch.yaml
- **Request Payload**:
```json
{
  "apiVersion": "sample_string",
  "pipelines": [
    {
      "build": {
        "apiVersion": "sample_string",
        "buildMaterials": [
          {
            "gitMaterialUrl": "sample_string",
            "source": {
              "type": "sample_string",
              "value": "BranchFixed"
            }
          },
          {
            "gitMaterialUrl": "sample_string",
            "source": {
              "type": "sample_string",
              "value": "BranchFixed"
            }
          }
        ],
        "dockerArguments": {}
      },
      "deployment": {
        "configMaps": [
          {
            "external": true,
            "externalType": "sample_string",
            "global": true
          },
          {
            "external": true,
            "externalType": "sample_string",
            "global": true
          }
        ],
        "operation": "create",
        "secrets": [
          {
            "data": {},
            "external": true,
            "externalType": "sample_string"
          },
          {
            "data": {},
            "external": true,
            "externalType": "sample_string"
          }
        ]
      }
    },
    {
      "build": {
        "apiVersion": "sample_string",
        "buildMaterials": [
          {
            "gitMaterialUrl": "sample_string",
            "source": {
              "type": "sample_string",
              "value": "BranchFixed"
            }
          },
          {
            "gitMaterialUrl": "sample_string",
            "source": {
              "type": "sample_string",
              "value": "BranchFixed"
            }
          }
        ],
        "dockerArguments": {}
      },
      "deployment": {
        "configMaps": [
          {
            "external": true,
            "externalType": "sample_string",
            "global": true
          },
          {
            "external": true,
            "externalType": "sample_string",
            "global": true
          }
        ],
        "operation": "create",
        "secrets": [
          {
            "data": {},
            "external": true,
            "externalType": "sample_string"
          },
          {
            "data": {},
            "external": true,
            "externalType": "sample_string"
          }
        ]
      }
    }
  ]
}
```
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/batch/v1beta1/application/dryrun

- **Status**: FAIL
- **Duration**: 75.760417ms
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
- **Duration**: 77.897459ms
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
- **Duration**: 76.460917ms
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
- **Duration**: 80.34775ms
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
- **Duration**: 77.189375ms
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
- **Duration**: 75.904459ms
- **Spec File**: ../../specs/jobs/bulk-operations.yaml
- **Request Payload**: (none)
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/batch/v1beta1/application

- **Status**: FAIL
- **Duration**: 79.046125ms
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

### ❌ GET /k8s/api-resources/{clusterId}

- **Status**: FAIL
- **Duration**: 73.792459ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
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

### ❌ POST /k8s/events

- **Status**: FAIL
- **Duration**: 73.262167ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Request Payload**:
```json
{
  "appId": "sample_string",
  "appType": "sample_string",
  "clusterId": 1,
  "devtronAppIdentifier": {
    "appName": "sample_string",
    "clusterId": 1,
    "namespace": "sample_string"
  },
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

### ❌ GET /k8s/pod/exec/session/{identifier}/{namespace}/{pod}/{shell}/{container}

- **Status**: FAIL
- **Duration**: 73.4225ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
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

### ❌ GET /k8s/pods/logs/{podName}

- **Status**: FAIL
- **Duration**: 75.005208ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
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

### ❌ POST /k8s/resource/create

- **Status**: FAIL
- **Duration**: 77.67275ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Request Payload**:
```json
{
  "appId": "sample_string",
  "appIdentifier": {
    "clusterId": 1,
    "namespace": "sample_string",
    "releaseName": "sample_string"
  },
  "clusterId": 1,
  "devtronAppIdentifier": {
    "appName": "sample_string",
    "clusterId": 1,
    "namespace": "sample_string"
  },
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

### ❌ POST /k8s/resource/delete

- **Status**: FAIL
- **Duration**: 75.025958ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Request Payload**:
```json
{
  "appId": "sample_string",
  "appIdentifier": {
    "clusterId": 1,
    "namespace": "sample_string",
    "releaseName": "sample_string"
  },
  "clusterId": 1,
  "devtronAppIdentifier": {
    "appName": "sample_string",
    "clusterId": 1,
    "namespace": "sample_string"
  },
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

### ❌ POST /k8s/resource/list

- **Status**: FAIL
- **Duration**: 72.321708ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Request Payload**:
```json
{
  "appId": "sample_string",
  "appIdentifier": {
    "clusterId": 1,
    "namespace": "sample_string",
    "releaseName": "sample_string"
  },
  "appType": "sample_string",
  "clusterId": 1,
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

### ❌ POST /orchestrator/k8s/resource

- **Status**: FAIL
- **Duration**: 75.373ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Request Payload**:
```json
{
  "appId": "sample_string",
  "appIdentifier": {
    "clusterId": 1,
    "namespace": "sample_string",
    "releaseName": "sample_string"
  },
  "appType": "sample_string",
  "clusterId": 1,
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
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{json: cannot unmarshal string into Go struct field ResourceRequestBean.appType of type int}]","userMessage":"json: canno...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/k8s/resource

- **Status**: FAIL
- **Duration**: 78.35375ms
- **Spec File**: ../../specs/kubernetes/apis.yaml
- **Request Payload**:
```json
{
  "appId": "sample_string",
  "appType": "sample_string",
  "clusterId": 1,
  "devtronAppIdentifier": {
    "appName": "sample_string",
    "clusterId": 1,
    "namespace": "sample_string"
  },
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
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{json: cannot unmarshal string into Go struct field ResourceRequestBean.appType of type int}]","userMessage":"json: canno...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---


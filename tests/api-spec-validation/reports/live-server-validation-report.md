# API Spec Validation Report

Generated: 2025-08-05T22:10:14+05:30

## Summary

- Total Endpoints: 165
- Passed: 59
- Failed: 106
- Warnings: 0
- Success Rate: 35.76%

## Detailed Results

### ❌ GET /orchestrator/app/helm/meta/info/{appId}

- **Status**: FAIL
- **Duration**: 435.672417ms
- **Spec File**: ../../specs/application/labels.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"sample_string\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"sample_strin...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/app/labels/list

- **Status**: PASS
- **Duration**: 111.297791ms
- **Spec File**: ../../specs/application/labels.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[]}

---

### ✅ GET /orchestrator/app/meta/info/{appId}

- **Status**: PASS
- **Duration**: 117.284875ms
- **Spec File**: ../../specs/application/labels.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"appId":1,"appName":"helm-sanity-application","description":"","projectId":2,"projectName":"shared-devtron-demo1","createdBy":"admin","createdOn":"2025-07-15T07:30:...

---

### ❌ POST /orchestrator/gitops/validate

- **Status**: FAIL
- **Duration**: 110.391083ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"000","internalMessage":"[{pg: no rows in result set}]","userMessage":"pg: no rows in result set"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/gitops/config

- **Status**: PASS
- **Duration**: 112.942375ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"provider":"GITHUB","username":"systemsdt","token":"","gitLabGroupId":"","gitHubOrgId":"stage-gitops","host":"https://github.com/","active":true,"azureProje...

---

### ❌ POST /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 107.511958ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{value  is unsupported  Key: 'GitOpsConfigDto.Provider' Error:Field validation for 'Provider' failed on the 'oneof' tag}]...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 105.382917ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"000","internalMessage":"[{pg: no rows in result set}]","userMessage":"pg: no rows in result set"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/gitops/config-by-provider

- **Status**: FAIL
- **Duration**: 183.867208ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/gitops/config/{id}

- **Status**: PASS
- **Duration**: 109.077583ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"provider":"GITHUB","username":"systemsdt","token":"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","gitLabGro...

---

### ✅ GET /orchestrator/gitops/configured

- **Status**: PASS
- **Duration**: 110.496458ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"allowCustomRepository":false,"authMode":"PASSWORD","exists":true,"provider":"GITHUB"}}

---

### ❌ POST /orchestrator/batch/v1beta1/hibernate

- **Status**: FAIL
- **Duration**: 108.301041ms
- **Spec File**: ../../specs/jobs/bulk-actions.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{ERROR #42601 syntax error at or near \")\"}]","userMessage":"ERROR #42601 syntax error at or near \")\""}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /v1beta1/deploy

- **Status**: FAIL
- **Duration**: 119.661834ms
- **Spec File**: ../../specs/jobs/bulk-actions.yaml
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

### ❌ POST /v1beta1/unhibernate

- **Status**: FAIL
- **Duration**: 104.9255ms
- **Spec File**: ../../specs/jobs/bulk-actions.yaml
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

### ❌ PUT /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 109.297583ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'ClusterBean.ClusterName' Error:Field validation for 'ClusterName' failed on the 'required' tag","userMessage":"Key: 'ClusterBean....

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/cluster

- **Status**: PASS
- **Duration**: 108.879583ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":6,"cluster_name":"abcd","description":"","active":true,"config":{"bearer_token":""},"prometheusAuth":{"isAnonymous":false},"defaultClusterComponent":[],"agent...

---

### ✅ GET /orchestrator/cluster/auth-list

- **Status**: PASS
- **Duration**: 109.895125ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":6,"cluster_name":"abcd","description":"","active":false,"defaultClusterComponent":null,"agentInstallationStage":0,"k8sVersion":"","insecureSkipTlsVerify":fals...

---

### ❌ POST /orchestrator/webhook/git/{gitHostId}

- **Status**: FAIL
- **Duration**: 109.933625ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"000","internalMessage":"[{pg: no rows in result set}]","userMessage":"pg: no rows in result set"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/git/{gitHostId}/{secret}

- **Status**: FAIL
- **Duration**: 106.638042ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"000","internalMessage":"[{pg: no rows in result set}]","userMessage":"pg: no rows in result set"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/ci/workflow

- **Status**: FAIL
- **Duration**: 110.802834ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 500
- **Error/Msg**: internal server error

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrator/webhook/ext-ci/{externalCiId}

- **Status**: FAIL
- **Duration**: 106.369958ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 401
- **Error/Msg**: {"code":401,"status":"Unauthorized","errors":[{"code":"6005","internalMessage":"no token provided"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---

### ❌ POST /orchestrator/webhook/git

- **Status**: FAIL
- **Duration**: 114.307416ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"000","internalMessage":"[{pg: no rows in result set}]","userMessage":"pg: no rows in result set"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/plugin/global/list/detail/v2

- **Status**: FAIL
- **Duration**: 163.790334ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","internalMessage":"Empty values for both pluginVersionIds and parentPluginIds. Please provide at least one of them","userMessage":"Empty valu...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/plugin/global/list/tags

- **Status**: PASS
- **Duration**: 172.735041ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"tagNames":["Load testing","Kubernetes","gcs","Github","Code Review","Code quality","DevSecOps","Image source","cloud","AWS EKS","Security","CI task","Google Kubern...

---

### ❌ POST /orchestrator/plugin/global/create

- **Status**: FAIL
- **Duration**: 143.656625ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","internalMessage":"no step data provided to save, please provide a plugin step to proceed further","userMessage":"no step data provided to sa...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/plugin/global/detail/all

- **Status**: PASS
- **Duration**: 756.330042ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"K6 Load testing","description":"K6 is an open-source tool and cloud service that makes load testing easy for developers and QA engineers.","type":"P...

---

### ❌ GET /orchestrator/plugin/global/list/global-variable

- **Status**: FAIL
- **Duration**: 106.387625ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"\": invalid syntax}]","userMessage":"invalid appId"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/plugin/global/list/v2

- **Status**: PASS
- **Duration**: 122.855208ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"parentPlugins":[{"id":4,"name":"AWS ECR Retag","pluginIdentifier":"aws-retag","description":"AWS ECR Retag plugin that enables retagging of container images within...

---

### ✅ GET /orchestrator/plugin/global/list/v2/min

- **Status**: PASS
- **Duration**: 105.0705ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[]}

---

### ✅ PUT /orchestrator/plugin/global/migrate

- **Status**: PASS
- **Duration**: 108.638875ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK"}

---

### ✅ GET /orchestrator/plugin/global/detail/{pluginId}

- **Status**: PASS
- **Duration**: 116.877292ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"name":"K6 Load testing","description":"K6 is an open-source tool and cloud service that makes load testing easy for developers and QA engineers.","type":"PR...

---

### ❌ POST /orchestrator/user/resource/options/{kind}/{version}

- **Status**: FAIL
- **Duration**: 101.167125ms
- **Spec File**: ../../specs/userResource/userResource.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/app-store/installed-app

- **Status**: PASS
- **Duration**: 108.455542ms
- **Spec File**: ../../specs/helm/charts.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"clusterIds":null,"applicationType":"DEVTRON-CHART-STORE","helmApps":[{"lastDeployedAt":"2025-07-15T10:48:16.504431Z","appName":"pk-chart-gitops","appId":"2","chart...

---

### ❌ GET /orchestrator/app-store/installed-app/notes

- **Status**: FAIL
- **Duration**: 103.081792ms
- **Spec File**: ../../specs/helm/charts.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/chartref/autocomplete/{appId}

- **Status**: PASS
- **Duration**: 110.60575ms
- **Spec File**: ../../specs/helm/dynamic-charts.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"chartRefs":[{"id":10,"version":"3.9.0","name":"Rollout Deployment","description":"","userUploaded":false,"isAppMetricsSupported":true},{"id":11,"version":"3.10.0",...

---

### ✅ GET /orchestrator/app/history/deployed-configuration/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 110.757ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":8,"name":"DEPLOYMENT_TEMPLATE"},{"id":3,"name":"PIPELINE_STRATEGY"}]}

---

### ❌ GET /orchestrator/app/history/deployed-component/detail/{appId}/{pipelineId}/{id}

- **Status**: FAIL
- **Duration**: 107.952708ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","result":"invalid historyComponent"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/history/deployed-component/list/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 107.582625ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","result":"invalid historyComponent"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/version

- **Status**: PASS
- **Duration**: 105.365875ms
- **Spec File**: ../../specs/common/version.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"result":{"gitCommit":"a73ff9ef","buildTime":"2025-07-16T13:44:58Z","serverMode":"FULL"}}

---

### ✅ GET /orchestrator/external-links/tools

- **Status**: PASS
- **Duration**: 108.95ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"Grafana","icon":"","category":2},{"id":2,"name":"Kibana","icon":"","category":2},{"id":3,"name":"Newrelic","icon":"","category":2},{"id":4,"name":"C...

---

### ❌ DELETE /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 105.058209ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ✅ GET /orchestrator/external-links

- **Status**: PASS
- **Duration**: 108.622542ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[]}

---

### ❌ POST /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 108.2415ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{json: cannot unmarshal object into Go value of type []*externalLink.ExternalLinkDto}]","userMessage":"json: cannot unmar...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 107.402833ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"000","internalMessage":"[{pg: no rows in result set}]","userMessage":"pg: no rows in result set"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/git/host/{id}

- **Status**: FAIL
- **Duration**: 104.597875ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"{id}\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"{id}\": invalid synta...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/{id}/event

- **Status**: FAIL
- **Duration**: 106.339792ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"{id}\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"{id}\": invalid synta...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/git/host

- **Status**: PASS
- **Duration**: 107.316ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"Github","active":true,"webhookUrl":"","webhookSecret":"","eventTypeHeader":"","secretHeader":"","secretValidator":""},{"id":2,"name":"Bitbucket Clou...

---

### ❌ POST /orchestrator/git/host

- **Status**: FAIL
- **Duration**: 108.852792ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'GitHostRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag","userMessage":"Key: 'GitHostRequest.Name' Er...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/git/host/event/{eventId}

- **Status**: FAIL
- **Duration**: 115.740541ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"{eventId}\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"{eventId}\": inv...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/webhook-meta-config/{gitProviderId}

- **Status**: FAIL
- **Duration**: 388.766667ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{ERROR #22P02 invalid input syntax for type integer: \"{gitProviderId}\"}]","userMessage":"ERROR #22P02 invalid...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/module

- **Status**: PASS
- **Duration**: 537.444041ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"name":"security.trivy","status":"installed","moduleResourcesStatus":null,"enabled":true,"moduleType":"security"},{"name":"security.clair","status":"installed","mo...

---

### ❌ POST /orchestrator/module

- **Status**: FAIL
- **Duration**: 103.11125ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/server

- **Status**: PASS
- **Duration**: 107.528ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"status":"unknown","releaseName":"devtron","installationType":"enterprise"}}

---

### ❌ POST /orchestrator/server

- **Status**: FAIL
- **Duration**: 168.68225ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'ServerActionRequestDto.Action' Error:Field validation for 'Action' failed on the 'oneof' tag","userMessage":"Key: 'ServerActionRe...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/config/environment/cm

- **Status**: FAIL
- **Duration**: 107.686584ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{invalid request multiple config found for add or update}]","userMessage":"invalid request multiple config foun...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/config/environment/cm/edit/{appId}/{envId}/{id}

- **Status**: FAIL
- **Duration**: 102.67825ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/config/environment/cm/{appId}/{envId}

- **Status**: PASS
- **Duration**: 112.777417ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":0,"appId":1,"environmentId":1,"configData":[],"isDeletable":false,"isExpressEdit":false}}

---

### ❌ DELETE /orchestrator/config/environment/cm/{appId}/{envId}/{id}

- **Status**: FAIL
- **Duration**: 105.252292ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/config/global/cm

- **Status**: FAIL
- **Duration**: 108.859ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{invalid request multiple config found for add or update}]","userMessage":"invalid request multiple config foun...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrator/config/bulk/patch

- **Status**: FAIL
- **Duration**: 109.450916ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{invalid request no payload found for sync}]","userMessage":"invalid request no payload found for sync"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/config/global/cm/edit/{appId}/{id}

- **Status**: FAIL
- **Duration**: 101.235417ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/config/global/cm/{appId}

- **Status**: PASS
- **Duration**: 107.309708ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":0,"appId":1,"configData":[],"isDeletable":false,"isExpressEdit":false}}

---

### ❌ DELETE /orchestrator/config/global/cm/{appId}/{id}

- **Status**: FAIL
- **Duration**: 102.428916ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/app-store/discover/application/chartInfo/{appStoreApplicationVersionId}

- **Status**: PASS
- **Duration**: 520.679166ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"appStoreApplicationVersionId":1,"readme":"# cluster-autoscaler\n\nScales Kubernetes worker nodes within autoscaling groups.\n\n## TL;DR\n\n```console\n$ helm repo ...

---

### ✅ GET /orchestrator/app-store/discover/application/{appStoreId}/version/autocomplete

- **Status**: PASS
- **Duration**: 132.640375ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"version":"9.49.0","id":5507},{"version":"9.48.0","id":1},{"version":"9.47.0","id":2},{"version":"9.46.6","id":3},{"version":"9.46.5","id":4},{"version":"9.46.4","...

---

### ✅ GET /orchestrator/app-store/discover/application/{id}

- **Status**: PASS
- **Duration**: 169.7765ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"version":"9.48.0","appVersion":"1.33.0","created":"2025-07-11T21:16:00.149315Z","deprecated":false,"description":"Scales Kubernetes worker nodes within auto...

---

### ❌ GET /orchestrator/app-store/discover/search

- **Status**: FAIL
- **Duration**: 124.785ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/app-store/discover

- **Status**: PASS
- **Duration**: 1.021184s
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":65,"appStoreApplicationVersionId":4028,"name":"ai-agent","chart_repo_id":2,"docker_artifact_store_id":"","chart_name":"devtron","icon":"","active":true,"chart...

---

### ❌ GET /orchestrator/app/ci-pipeline/{ciPipelineId}/linked-ci/downstream/cd

- **Status**: FAIL
- **Duration**: 107.782667ms
- **Spec File**: ../../specs/ci-pipeline/ciPipelineDownstream/downstream-linked-ci-view-spec.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{invalid CiPipelineId 1}]","userMessage":"invalid CiPipelineId 1"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/ci-pipeline/{ciPipelineId}/linked-ci/downstream/env

- **Status**: FAIL
- **Duration**: 105.740584ms
- **Spec File**: ../../specs/ci-pipeline/ciPipelineDownstream/downstream-linked-ci-view-spec.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{invalid CiPipelineId 1}]","userMessage":"invalid CiPipelineId 1"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/app/cd-pipeline/patch/deployment

- **Status**: FAIL
- **Duration**: 109.075375ms
- **Spec File**: ../../specs/deployment/app-type-change.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'DeploymentAppTypeChangeRequest.EnvId' Error:Field validation for 'EnvId' failed on the 'required' tag","userMessage":"Key: 'Deplo...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrtor/batch/v1beta1/cd-pipeline

- **Status**: FAIL
- **Duration**: 103.406625ms
- **Spec File**: ../../specs/environment/bulk-delete.yaml
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

### ✅ POST /orchestrator/job/list

- **Status**: PASS
- **Duration**: 473.5475ms
- **Spec File**: ../../specs/jobs/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"jobContainers":[],"jobCount":0}}

---

### ❌ POST /orchestrator/job

- **Status**: FAIL
- **Duration**: 108.099375ms
- **Spec File**: ../../specs/jobs/core.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'CreateAppDTO.AppName' Error:Field validation for 'AppName' failed on the 'name-component' tag","userMessage":"Key: 'CreateAppDTO....

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/job/ci-pipeline/list/{jobId}

- **Status**: FAIL
- **Duration**: 110.828375ms
- **Spec File**: ../../specs/jobs/core.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","result":"Job with the given Id does not exist"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ POST /orchestrator/notification/search

- **Status**: PASS
- **Duration**: 169.025166ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"team":null,"app":null,"environment":null,"cluster":null,"pipeline":{"id":1,"name":"cd-3-pgxd","environmentName":"devtron-demo","appName":"pk-test","isVirtualEnvir...

---

### ❌ POST /orchestrator/notification/v2

- **Status**: FAIL
- **Duration**: 143.0195ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'NotificationRequest.NotificationConfigRequest' Error:Field validation for 'NotificationConfigRequest' failed on the 'required' ta...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/notification/channel

- **Status**: FAIL
- **Duration**: 107.426792ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200
- **Error/Msg**: {}

**Issues:**
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: unexpected end of JSON input

---

### ❌ DELETE /orchestrator/notification/channel

- **Status**: FAIL
- **Duration**: 108.504292ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{ The channel you requested is not supported}]","userMessage":" The channel you requested is not supported"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/notification/channel

- **Status**: PASS
- **Duration**: 111.193459ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"slackConfigs":[],"webhookConfigs":[],"sesConfigs":[{"userId":2,"teamId":0,"region":"ap-south-1","accessKey":"AKIAWEAVHFOOSNAORPHU","secretKey":"**********","fromEm...

---

### ✅ GET /orchestrator/notification/channel/autocomplete/{type}

- **Status**: PASS
- **Duration**: 106.45575ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[]}

---

### ✅ GET /orchestrator/notification/channel/webhook/{id}

- **Status**: PASS
- **Duration**: 107.74825ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"userId":0,"webhookUrl":"","configName":"","header":null,"payload":"","description":"","id":0}}

---

### ✅ GET /orchestrator/notification/channel/slack/{id}

- **Status**: PASS
- **Duration**: 125.669083ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"userId":0,"teamId":0,"webhookUrl":"","configName":"","description":"","id":0}}

---

### ✅ GET /orchestrator/notification/channel/ses/{id}

- **Status**: PASS
- **Duration**: 119.674333ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"userId":2,"teamId":0,"region":"ap-south-1","accessKey":"AKIAWEAVHFOOSNAORPHU","secretKey":"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","fromEmail":"watcher@devtron.i...

---

### ❌ GET /orchestrator/notification/recipient

- **Status**: FAIL
- **Duration**: 103.045458ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/notification/variables

- **Status**: PASS
- **Duration**: 129.210916ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"devtronAppId":"{{devtronAppId}}","devtronAppName":"{{devtronAppName}}","devtronApprovedByEmail":"{{devtronApprovedByEmail}}","devtronBuildGitCommitHash":"{{devtron...

---

### ❌ GET /orchestrator/notification

- **Status**: FAIL
- **Duration**: 120.781583ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/notification

- **Status**: FAIL
- **Duration**: 125.769709ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'NotificationRequest.NotificationConfigRequest' Error:Field validation for 'NotificationConfigRequest' failed on the 'required' ta...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/notification

- **Status**: FAIL
- **Duration**: 120.040292ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'NotificationUpdateRequest.NotificationConfigRequest' Error:Field validation for 'NotificationConfigRequest' failed on the 'requir...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ DELETE /orchestrator/notification

- **Status**: PASS
- **Duration**: 177.014667ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK"}

---

### ✅ GET /orchestrator/notification/channel/smtp/{id}

- **Status**: PASS
- **Duration**: 109.910917ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":0,"port":"","host":"","authType":"","authUser":"","authPassword":"","fromEmail":"","configName":"","description":"","ownerId":0,"default":false,"deleted":false...

---

### ❌ PATCH /orchestrator/app/ci-pipeline/patch-source

- **Status**: FAIL
- **Duration**: 107.972084ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-change-source.yaml
- **Response Code**: 500
- **Error/Msg**: internal server error

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PATCH /orchestrator/app/env/patch

- **Status**: FAIL
- **Duration**: 106.570458ms
- **Spec File**: ../../specs/helm/deployment-chart-type.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","result":{"appId":0,"envId":0,"targetChartRefId":0}}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/chart-repo/validate

- **Status**: FAIL
- **Duration**: 108.843208ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","internalMessage":"Key: 'ChartRepoDto.Name' Error:Field validation for 'Name' failed on the 'required' tag\nKey: 'ChartRepoDto.AuthMode' Erro...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/chart-repo/create

- **Status**: FAIL
- **Duration**: 108.111667ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","internalMessage":"Key: 'ChartRepoDto.Name' Error:Field validation for 'Name' failed on the 'required' tag\nKey: 'ChartRepoDto.AuthMode' Erro...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ POST /orchestrator/chart-repo/sync-charts

- **Status**: PASS
- **Duration**: 7.15860475s
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"status":"ok"}}

---

### ❌ POST /orchestrator/chart-repo/update

- **Status**: FAIL
- **Duration**: 108.224708ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","internalMessage":"Key: 'ChartRepoDto.Name' Error:Field validation for 'Name' failed on the 'required' tag\nKey: 'ChartRepoDto.AuthMode' Erro...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app-store/chart-provider/list

- **Status**: PASS
- **Duration**: 109.558958ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":"prakhar","name":"prakhar","active":true,"isEditable":true,"isOCIRegistry":true,"registryProvider":"docker-hub"},{"id":"1","name":"default-chartmuseum","activ...

---

### ❌ POST /orchestrator/app-store/chart-provider/sync-chart

- **Status**: FAIL
- **Duration**: 108.328375ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'ChartProviderRequestDto.Id' Error:Field validation for 'Id' failed on the 'required' tag","userMessage":"Key: 'ChartProviderReque...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/chart-group/

- **Status**: FAIL
- **Duration**: 118.224ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'ChartGroupBean.Name' Error:Field validation for 'Name' failed on the 'name-component' tag","userMessage":"Key: 'ChartGroupBean.Na...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/chart-group/

- **Status**: FAIL
- **Duration**: 108.558833ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'ChartGroupBean.Name' Error:Field validation for 'Name' failed on the 'name-component' tag","userMessage":"Key: 'ChartGroupBean.Na...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/chart-group/list

- **Status**: PASS
- **Duration**: 106.122125ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":null}

---

### ✅ GET /orchestrator/chart-repo/{id}

- **Status**: PASS
- **Duration**: 116.164959ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"name":"default-chartmuseum","url":"http://devtron-chartmuseum.devtroncd:8080/","authMode":"ANONYMOUS","active":true,"default":true,"allow_insecure_connectio...

---

### ❌ POST /orchestrator/app-store/chart-provider/update

- **Status**: FAIL
- **Duration**: 105.508709ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'ChartProviderRequestDto.Id' Error:Field validation for 'Id' failed on the 'required' tag","userMessage":"Key: 'ChartProviderReque...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/chart-group/entries

- **Status**: FAIL
- **Duration**: 105.53ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"000","internalMessage":"[{pg: no rows in result set}]","userMessage":"pg: no rows in result set"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/chart-group/{id}

- **Status**: FAIL
- **Duration**: 103.77225ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/chart-repo/list

- **Status**: PASS
- **Duration**: 111.011625ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"default-chartmuseum","url":"http://devtron-chartmuseum.devtroncd:8080/","authMode":"ANONYMOUS","active":true,"default":true,"allow_insecure_connecti...

---

### ✅ GET /orchestrator/api-token

- **Status**: PASS
- **Duration**: 109.104708ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"description":"some description","expireAtInMs":12344546,"id":2,"name":"some-name","token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6IkFQSS1UT0tFTjpzb21lLW...

---

### ❌ POST /orchestrator/api-token

- **Status**: FAIL
- **Duration**: 109.536917ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'CreateApiTokenRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag","userMessage":"Key: 'CreateApiTokenRe...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/api-token/webhook

- **Status**: PASS
- **Duration**: 109.142292ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[]}

---

### ❌ DELETE /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 107.791666ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{api-token corresponds to apiTokenId '1' is not found}]","userMessage":"api-token corresponds to apiTokenId '1'...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PUT /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 110.273667ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'UpdateApiTokenRequest.Description' Error:Field validation for 'Description' failed on the 'required' tag","userMessage":"Key: 'Up...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/role/group/detailed/get

- **Status**: PASS
- **Duration**: 120.157291ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":2,"roleFilters":[],"superAdmin":false,"canManageAllAccess":false,"hasAccessManagerPermission":false}]}

---

### ❌ GET /orchestrator/user/role/group/search

- **Status**: FAIL
- **Duration**: 111.182083ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"search\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"search\": invalid syntax"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/role/group/v2

- **Status**: PASS
- **Duration**: 127.43525ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"roleGroups":[{"id":2,"roleFilters":[],"superAdmin":false,"canManageAllAccess":false,"hasAccessManagerPermission":false}],"totalCount":1}}

---

### ❌ POST /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 113.402542ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{role group already exist}]","userMessage":"role group already exist"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PUT /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 192.148875ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"000","internalMessage":"[{pg: no rows in result set}]","userMessage":""}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/user/role/group/{id}

- **Status**: FAIL
- **Duration**: 110.706666ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"000","internalMessage":"[{pg: no rows in result set}]","userMessage":""}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/user/role/group/{id}

- **Status**: FAIL
- **Duration**: 112.249625ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"000","internalMessage":"[{pg: no rows in result set}]","userMessage":"Failed to get by id"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 109.076167ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"000","internalMessage":"[{pg: no rows in result set}]","userMessage":""}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/user/role/group

- **Status**: PASS
- **Duration**: 109.09175ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":2,"roleFilters":[],"superAdmin":false,"canManageAllAccess":false,"hasAccessManagerPermission":false}]}

---

### ❌ POST /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 160.916ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{role group already exist}]","userMessage":"role group already exist"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ DELETE /orchestrator/user/role/group/bulk

- **Status**: FAIL
- **Duration**: 110.531667ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{ERROR #42601 syntax error at or near \")\"}]","userMessage":"ERROR #42601 syntax error at or near \")\""}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /{appId}/ci-pipeline/{pipelineId}/workflow/{workflowId}

- **Status**: FAIL
- **Duration**: 103.118625ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
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

### ❌ POST /ci-pipeline/trigger

- **Status**: FAIL
- **Duration**: 102.16925ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
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

### ❌ GET /orchestrator/app/commit-info/{ciPipelineMaterialId}/{gitHash}

- **Status**: FAIL
- **Duration**: 109.109291ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"000","internalMessage":"[{pg: no rows in result set}]","userMessage":"pg: no rows in result set"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/deployment/pipeline/history

- **Status**: FAIL
- **Duration**: 110.7765ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/deployment/pipeline/rollback

- **Status**: FAIL
- **Duration**: 105.640292ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/deployment/pipeline/trigger

- **Status**: FAIL
- **Duration**: 103.362625ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/deployment/pipeline/configure

- **Status**: FAIL
- **Duration**: 106.426125ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/flux-application

- **Status**: FAIL
- **Duration**: 106.788833ms
- **Spec File**: ../../specs/gitops/fluxcd.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{error in getting cluster ids}]","userMessage":"error in getting cluster ids"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/flux-application/app

- **Status**: FAIL
- **Duration**: 106.9545ms
- **Spec File**: ../../specs/gitops/fluxcd.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /create

- **Status**: FAIL
- **Duration**: 344.150333ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
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

### ❌ POST /orchestrator/validate

- **Status**: FAIL
- **Duration**: 105.761834ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /update

- **Status**: FAIL
- **Duration**: 103.281125ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
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

### ✅ GET /orchestrator/app/history/deployed-configuration/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 112.768834ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":8,"name":"DEPLOYMENT_TEMPLATE"},{"id":3,"name":"PIPELINE_STRATEGY"}]}

---

### ❌ GET /orchestrator/app/history/deployed-component/detail/{appId}/{pipelineId}/{id}

- **Status**: FAIL
- **Duration**: 107.579042ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","result":"invalid historyComponent"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/history/deployed-component/list/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 110.888833ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","result":"invalid historyComponent"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/deployment/template/fetch

- **Status**: PASS
- **Duration**: 215.775292ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":10,"name":"Rollout Deployment","chartDescription":"This chart deploys an advanced version of deployment that supports Blue/Green and Canary deployments. For f...

---

### ✅ PUT /orchestrator/deployment/template/upload

- **Status**: PASS
- **Duration**: 141.399542ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":"Processed successfully"}

---

### ❌ POST /orchestrator/deployment/template/validate

- **Status**: FAIL
- **Duration**: 108.184292ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{request Content-Type isn't multipart/form-data}]","userMessage":"request Content-Type isn't multipart/form-data"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/history/deployed-configuration/all/latest/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 107.507875ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"latest\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"latest\": invalid syntax"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/history/deployed-configuration/all/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 172.705959ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"deploymentTemplate":{"templateName":"Deployment","templateVersion":"4.21.0","isAppMetricsEnabled":false,"codeEditorValue":{"displayName":"values.yaml","value":"{\"...

---

### ❌ GET /orchestrator/app/template/default/{appId}/{chartRefId}

- **Status**: FAIL
- **Duration**: 108.519ms
- **Spec File**: ../../specs/environment/templates.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"000","internalMessage":"[{pg: no rows in result set}]","userMessage":"pg: no rows in result set"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/k8s/resources/ephemeralContainers

- **Status**: FAIL
- **Duration**: 104.236917ms
- **Spec File**: ../../specs/kubernetes/ephemeral-containers.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/k8s/resources/ephemeralContainers

- **Status**: FAIL
- **Duration**: 111.923583ms
- **Spec File**: ../../specs/kubernetes/ephemeral-containers.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PUT /orchestrator/application/rollback

- **Status**: FAIL
- **Duration**: 108.570458ms
- **Spec File**: ../../specs/openapiClient/api/openapi.yaml
- **Response Code**: 500
- **Error/Msg**: internal server error

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrator/application/template-chart

- **Status**: FAIL
- **Duration**: 105.56825ms
- **Spec File**: ../../specs/openapiClient/api/openapi.yaml
- **Response Code**: 500
- **Error/Msg**: internal server error

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/app/deployment-status/timeline/{appId}/{envId}

- **Status**: FAIL
- **Duration**: 161.653375ms
- **Spec File**: ../../specs/deployment/timeline.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"000","internalMessage":"[{pg: no rows in result set}]","userMessage":"pg: no rows in result set"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/app/ci-pipeline/patch

- **Status**: FAIL
- **Duration**: 108.342041ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"000","internalMessage":"[{pg: no rows in result set}]","userMessage":"pg: no rows in result set"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/app/ci-pipeline/{appId}

- **Status**: PASS
- **Duration**: 111.224292ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"appId":1,"dockerRegistry":"quay","dockerRepository":"devtron/test","ciBuildConfig":{"id":1,"gitMaterialId":1,"buildContextGitMaterialId":1,"useRootBuildCont...

---

### ✅ GET /orchestrator/app/wf/all/component-names/{appId}

- **Status**: PASS
- **Duration**: 109.274375ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"workflows":[{"id":1,"name":"wf-1-lcu5","ciPipelineId":0,"ciPipelineName":"","cdPipelines":null},{"id":2,"name":"wf-1-4rww","ciPipelineId":0,"ciPipelineName":"","cd...

---

### ❌ POST /orchestrator/user

- **Status**: FAIL
- **Duration**: 109.060167ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'UserInfo.EmailId' Error:Field validation for 'EmailId' failed on the 'required' tag","userMessage":"Key: 'UserInfo.EmailId' Error...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/user

- **Status**: FAIL
- **Duration**: 107.529458ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'UserInfo.EmailId' Error:Field validation for 'EmailId' failed on the 'required' tag","userMessage":"Key: 'UserInfo.EmailId' Error...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/user/bulk

- **Status**: FAIL
- **Duration**: 109.0525ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{ERROR #42601 syntax error at or near \")\"}]","userMessage":"ERROR #42601 syntax error at or near \")\""}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/user/v2

- **Status**: PASS
- **Duration**: 109.714292ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"users":[{"id":2,"email_id":"admin","roleFilters":[],"groups":[],"superAdmin":false,"userRoleGroups":[],"lastLoginTime":"2025-08-05T16:40:13.674161Z","timeoutWindow...

---

### ❌ DELETE /orchestrator/user/{id}

- **Status**: FAIL
- **Duration**: 108.809292ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","userMessage":"cannot delete system or admin user"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/{id}

- **Status**: PASS
- **Duration**: 127.785542ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"email_id":"system","roleFilters":[],"groups":[],"superAdmin":false,"userRoleGroups":[],"lastLoginTime":"0001-01-01T00:00:00Z","timeoutWindowExpression":"000...

---

### ❌ GET /orchestrator/env/namespace/autocomplete

- **Status**: FAIL
- **Duration**: 107.054334ms
- **Spec File**: ../../specs/environment/core.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{ERROR #42601 syntax error at or near \")\"}]","userMessage":"ERROR #42601 syntax error at or near \")\""}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/gitops/configured

- **Status**: PASS
- **Duration**: 108.961292ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"allowCustomRepository":false,"authMode":"PASSWORD","exists":true,"provider":"GITHUB"}}

---

### ❌ POST /orchestrator/validate

- **Status**: FAIL
- **Duration**: 110.406625ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/gitops/config

- **Status**: PASS
- **Duration**: 108.276792ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"provider":"GITHUB","username":"systemsdt","token":"","gitLabGroupId":"","gitHubOrgId":"stage-gitops","host":"https://github.com/","active":true,"azureProje...

---

### ❌ POST /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 109.114625ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{value  is unsupported  Key: 'GitOpsConfigDto.Provider' Error:Field validation for 'Provider' failed on the 'oneof' tag}]...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 107.554459ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"000","internalMessage":"[{pg: no rows in result set}]","userMessage":"pg: no rows in result set"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/gitops/config-by-provider

- **Status**: FAIL
- **Duration**: 104.3385ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ✅ GET /orchestrator/gitops/config/{id}

- **Status**: PASS
- **Duration**: 108.533292ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"provider":"GITHUB","username":"systemsdt","token":"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","gitLabGro...

---


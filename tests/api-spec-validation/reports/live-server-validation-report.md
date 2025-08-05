# API Spec Validation Report

Generated: 2025-08-06T02:26:39+05:30

## Summary

- Total Endpoints: 154
- Passed: 63
- Failed: 91
- Warnings: 0
- Success Rate: 40.91%

## Detailed Results

### ❌ GET /orchestrator/plugin/global/list/global-variable

- **Status**: FAIL
- **Duration**: 280.760625ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"\": invalid syntax}]","userMessage":"invalid appId"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/plugin/global/list/v2

- **Status**: PASS
- **Duration**: 279.143625ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"parentPlugins":[{"id":4,"name":"AWS ECR Retag","pluginIdentifier":"aws-retag","description":"AWS ECR Retag plugin that enables retagging of container images within...

---

### ✅ GET /orchestrator/plugin/global/list/v2/min

- **Status**: PASS
- **Duration**: 111.525667ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[]}

---

### ✅ PUT /orchestrator/plugin/global/migrate

- **Status**: PASS
- **Duration**: 100.328291ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK"}

---

### ✅ GET /orchestrator/plugin/global/detail/{pluginId}

- **Status**: PASS
- **Duration**: 100.77225ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"name":"K6 Load testing","description":"K6 is an open-source tool and cloud service that makes load testing easy for developers and QA engineers.","type":"PR...

---

### ❌ POST /orchestrator/plugin/global/list/detail/v2

- **Status**: FAIL
- **Duration**: 93.231833ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","internalMessage":"Empty values for both pluginVersionIds and parentPluginIds. Please provide at least one of them","userMessage":"Empty valu...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/plugin/global/list/tags

- **Status**: PASS
- **Duration**: 90.020584ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"tagNames":["Load testing","Security","Code Review","CI task","gcs","Google Kubernetes Engine","cloud","Code quality","DevSecOps","Image source","GCP","AWS EKS","Gi...

---

### ❌ POST /orchestrator/plugin/global/create

- **Status**: FAIL
- **Duration**: 91.897917ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","internalMessage":"no step data provided to save, please provide a plugin step to proceed further","userMessage":"no step data provided to sa...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/plugin/global/detail/all

- **Status**: PASS
- **Duration**: 465.936958ms
- **Spec File**: ../../specs/plugins/global.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"K6 Load testing","description":"K6 is an open-source tool and cloud service that makes load testing easy for developers and QA engineers.","type":"P...

---

### ✅ GET /orchestrator/user/role/group

- **Status**: PASS
- **Duration**: 89.666458ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":2,"roleFilters":[],"superAdmin":false,"canManageAllAccess":false,"hasAccessManagerPermission":false}]}

---

### ❌ POST /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 91.655916ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{role group already exist}]","userMessage":"role group already exist"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PUT /orchestrator/user/role/group

- **Status**: FAIL
- **Duration**: 89.798583ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/user/role/group/bulk

- **Status**: FAIL
- **Duration**: 91.006ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{EOF}]","userMessage":"EOF"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/role/group/detailed/get

- **Status**: PASS
- **Duration**: 90.0195ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":2,"roleFilters":[],"superAdmin":false,"canManageAllAccess":false,"hasAccessManagerPermission":false}]}

---

### ❌ GET /orchestrator/user/role/group/search

- **Status**: FAIL
- **Duration**: 90.6895ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"11001","internalMessage":"invalid path parameter id: search","userMessage":"Invalid path parameter 'id'","userDetailMessage":"Please check the par...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/role/group/v2

- **Status**: PASS
- **Duration**: 93.162625ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"roleGroups":[{"id":2,"roleFilters":[],"superAdmin":false,"canManageAllAccess":false,"hasAccessManagerPermission":false}],"totalCount":1}}

---

### ❌ POST /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 91.9575ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{role group already exist}]","userMessage":"role group already exist"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PUT /orchestrator/user/role/group/v2

- **Status**: FAIL
- **Duration**: 91.275625ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/user/role/group/{id}

- **Status**: FAIL
- **Duration**: 91.886667ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/user/role/group/{id}

- **Status**: FAIL
- **Duration**: 119.656167ms
- **Spec File**: ../../specs/security/group-policy.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"role group not found: 1","userMessage":"role group with ID '1' not found","userDetailMessage":"The requested role group do...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/app/cd-pipeline/patch/deployment

- **Status**: FAIL
- **Duration**: 126.893ms
- **Spec File**: ../../specs/deployment/app-type-change.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'DeploymentAppTypeChangeRequest.DesiredDeploymentType' Error:Field validation for 'DesiredDeploymentType' failed on the 'required'...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/gitops/config/{id}

- **Status**: PASS
- **Duration**: 90.992042ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"provider":"GITHUB","username":"systemsdt","token":"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","gitLabGro...

---

### ✅ GET /orchestrator/gitops/configured

- **Status**: PASS
- **Duration**: 92.1365ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"allowCustomRepository":false,"authMode":"PASSWORD","exists":true,"provider":"GITHUB"}}

---

### ❌ POST /orchestrator/validate

- **Status**: FAIL
- **Duration**: 91.444958ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/gitops/config

- **Status**: PASS
- **Duration**: 89.562542ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"provider":"GITHUB","username":"systemsdt","token":"","gitLabGroupId":"","gitHubOrgId":"stage-gitops","host":"https://github.com/","active":true,"azureProje...

---

### ❌ POST /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 91.146292ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 409
- **Error/Msg**: {"code":409,"status":"Conflict","errors":[{"internalMessage":"gitops provider already exists","userMessage":"gitops provider already exists"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 409

---

### ❌ PUT /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 90.018625ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/gitops/config-by-provider

- **Status**: FAIL
- **Duration**: 86.507042ms
- **Spec File**: ../../specs/gitops/validation.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ POST /orchestrator/user

- **Status**: FAIL
- **Duration**: 90.576875ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'UserInfo.EmailId' Error:Field validation for 'EmailId' failed on the 'required' tag","userMessage":"Key: 'UserInfo.EmailId' Error...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/user

- **Status**: FAIL
- **Duration**: 92.875958ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'UserInfo.EmailId' Error:Field validation for 'EmailId' failed on the 'required' tag","userMessage":"Key: 'UserInfo.EmailId' Error...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/user/bulk

- **Status**: FAIL
- **Duration**: 91.008875ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{EOF}]","userMessage":"EOF"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/v2

- **Status**: PASS
- **Duration**: 93.153542ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"users":[{"id":2,"email_id":"admin","roleFilters":[],"groups":[],"superAdmin":false,"userRoleGroups":[],"lastLoginTime":"2025-08-05T20:56:11.278943Z","timeoutWindow...

---

### ❌ DELETE /orchestrator/user/{id}

- **Status**: FAIL
- **Duration**: 91.908417ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","userMessage":"cannot delete system or admin user"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/user/{id}

- **Status**: PASS
- **Duration**: 92.359ms
- **Spec File**: ../../specs/security/user-management.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"email_id":"system","roleFilters":[],"groups":[],"superAdmin":false,"userRoleGroups":[],"lastLoginTime":"0001-01-01T00:00:00Z","timeoutWindowExpression":"000...

---

### ❌ POST /orchestrator/user/resource/options/{kind}/{version}

- **Status**: FAIL
- **Duration**: 86.259083ms
- **Spec File**: ../../specs/userResource/userResource.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/batch/v1beta1/cd-pipeline

- **Status**: FAIL
- **Duration**: 89.393833ms
- **Spec File**: ../../specs/environment/bulk-delete.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"400","userMessage":"invalid payload, can not get pipelines for this filter"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ DELETE /orchestrator/k8s/resources/ephemeralContainers

- **Status**: FAIL
- **Duration**: 88.065625ms
- **Spec File**: ../../specs/kubernetes/ephemeral-containers.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/k8s/resources/ephemeralContainers

- **Status**: FAIL
- **Duration**: 89.016917ms
- **Spec File**: ../../specs/kubernetes/ephemeral-containers.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/app-store/chart-provider/list

- **Status**: PASS
- **Duration**: 91.647458ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":"prakhar","name":"prakhar","active":true,"isEditable":true,"isOCIRegistry":true,"registryProvider":"docker-hub"},{"id":"1","name":"default-chartmuseum","activ...

---

### ❌ POST /orchestrator/chart-group/

- **Status**: FAIL
- **Duration**: 88.286417ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'ChartGroupBean.Name' Error:Field validation for 'Name' failed on the 'name-component' tag","userMessage":"Key: 'ChartGroupBean.Na...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/chart-group/

- **Status**: FAIL
- **Duration**: 88.841ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'ChartGroupBean.Name' Error:Field validation for 'Name' failed on the 'name-component' tag","userMessage":"Key: 'ChartGroupBean.Na...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/chart-group/entries

- **Status**: FAIL
- **Duration**: 94.322125ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/chart-group/{id}

- **Status**: FAIL
- **Duration**: 87.0425ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ POST /orchestrator/chart-repo/create

- **Status**: PASS
- **Duration**: 280.708666ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"customErrMsg":"Could not validate the repo. Please try again.","actualErrMsg":"Get \"https://charts.example.com/index.yaml\": dial tcp: lookup charts.example.com o...

---

### ✅ POST /orchestrator/chart-repo/sync-charts

- **Status**: PASS
- **Duration**: 7.19476625s
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"status":"ok"}}

---

### ✅ POST /orchestrator/app-store/chart-provider/sync-chart

- **Status**: PASS
- **Duration**: 7.135944292s
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK"}

---

### ❌ POST /orchestrator/app-store/chart-provider/update

- **Status**: FAIL
- **Duration**: 92.358292ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/chart-repo/list

- **Status**: PASS
- **Duration**: 96.870208ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"default-chartmuseum","url":"http://devtron-chartmuseum.devtroncd:8080/","authMode":"ANONYMOUS","active":true,"default":true,"allow_insecure_connecti...

---

### ✅ POST /orchestrator/chart-repo/validate

- **Status**: PASS
- **Duration**: 317.718166ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"customErrMsg":"Could not validate the repo. Please try again.","actualErrMsg":"Get \"https://charts.example.com/index.yaml\": dial tcp: lookup charts.example.com o...

---

### ✅ GET /orchestrator/chart-repo/{id}

- **Status**: PASS
- **Duration**: 89.7455ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"name":"default-chartmuseum","url":"http://devtron-chartmuseum.devtroncd:8080/","authMode":"ANONYMOUS","active":true,"default":true,"allow_insecure_connectio...

---

### ✅ GET /orchestrator/chart-group/list

- **Status**: PASS
- **Duration**: 92.584125ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":null}

---

### ❌ POST /orchestrator/chart-repo/update

- **Status**: FAIL
- **Duration**: 276.788208ms
- **Spec File**: ../../specs/helm/provider.yaml
- **Response Code**: 412
- **Error/Msg**: {"code":412,"status":"Precondition Failed","result":{"customErrMsg":"Could not validate the repo. Please try again.","actualErrMsg":"Get \"https://charts.example.com/index.yaml\": dial tcp: lookup cha...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 412

---

### ✅ GET /orchestrator/cluster

- **Status**: PASS
- **Duration**: 98.8505ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"cluster_name":"default_cluster","description":"","server_url":"https://kubernetes.default.svc","active":true,"config":{"bearer_token":""},"prometheusAuth":...

---

### ❌ PUT /orchestrator/cluster

- **Status**: FAIL
- **Duration**: 89.784292ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'ClusterBean.ServerUrl' Error:Field validation for 'ServerUrl' failed on the 'url' tag","userMessage":"Key: 'ClusterBean.ServerUrl...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/cluster/auth-list

- **Status**: PASS
- **Duration**: 90.455833ms
- **Spec File**: ../../specs/kubernetes/cluster.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"cluster_name":"default_cluster","description":"","active":false,"defaultClusterComponent":null,"agentInstallationStage":0,"k8sVersion":"","insecureSkipTlsV...

---

### ❌ GET /orchestrator/app/history/deployed-component/list/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 90.518667ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","result":"invalid historyComponent"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/history/deployed-configuration/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 162.322667ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":8,"name":"DEPLOYMENT_TEMPLATE"},{"id":3,"name":"PIPELINE_STRATEGY"}]}

---

### ❌ GET /orchestrator/app/history/deployed-component/detail/{appId}/{pipelineId}/{id}

- **Status**: FAIL
- **Duration**: 92.102834ms
- **Spec File**: ../../specs/audit/api-changes.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","result":"invalid historyComponent"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/deployment/pipeline/rollback

- **Status**: FAIL
- **Duration**: 88.170125ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/deployment/pipeline/trigger

- **Status**: FAIL
- **Duration**: 93.906125ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/deployment/pipeline/configure

- **Status**: FAIL
- **Duration**: 86.58975ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/deployment/pipeline/history

- **Status**: FAIL
- **Duration**: 87.063166ms
- **Spec File**: ../../specs/deployment/pipeline.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/external-links

- **Status**: FAIL
- **Duration**: 88.707958ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ✅ GET /orchestrator/external-links

- **Status**: PASS
- **Duration**: 92.753208ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[]}

---

### ✅ POST /orchestrator/external-links

- **Status**: PASS
- **Duration**: 94.405542ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"success":true}}

---

### ✅ PUT /orchestrator/external-links

- **Status**: PASS
- **Duration**: 93.690042ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"success":true}}

---

### ✅ GET /orchestrator/external-links/tools

- **Status**: PASS
- **Duration**: 160.059166ms
- **Spec File**: ../../specs/external-links/external-links-specs.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"Grafana","icon":"","category":2},{"id":2,"name":"Kibana","icon":"","category":2},{"id":3,"name":"Newrelic","icon":"","category":2},{"id":4,"name":"C...

---

### ✅ GET /orchestrator/git/host

- **Status**: PASS
- **Duration**: 111.801666ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"name":"Github","active":true,"webhookUrl":"","webhookSecret":"","eventTypeHeader":"","secretHeader":"","secretValidator":""},{"id":2,"name":"Bitbucket Clou...

---

### ✅ POST /orchestrator/git/host

- **Status**: PASS
- **Duration**: 97.95675ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":4}

---

### ❌ GET /orchestrator/git/host/event/{eventId}

- **Status**: FAIL
- **Duration**: 98.820459ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"{eventId}\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"{eventId}\": inv...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/webhook-meta-config/{gitProviderId}

- **Status**: FAIL
- **Duration**: 95.925834ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{ERROR #22P02 invalid input syntax for type integer: \"{gitProviderId}\"}]","userMessage":"ERROR #22P02 invalid...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/{id}

- **Status**: FAIL
- **Duration**: 90.726208ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"{id}\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"{id}\": invalid synta...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ GET /orchestrator/git/host/{id}/event

- **Status**: FAIL
- **Duration**: 119.938958ms
- **Spec File**: ../../specs/gitops/submodules.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"{id}\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"{id}\": invalid synta...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/api-token/webhook

- **Status**: PASS
- **Duration**: 131.131292ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[]}

---

### ❌ DELETE /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 111.104375ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{api-token corresponds to apiTokenId '1' is not found}]","userMessage":"api-token corresponds to apiTokenId '1'...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PUT /orchestrator/api-token/{id}

- **Status**: FAIL
- **Duration**: 141.339625ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{api-token corresponds to apiTokenId '1' is not found}]","userMessage":"api-token corresponds to apiTokenId '1'...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/api-token

- **Status**: PASS
- **Duration**: 95.384959ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"description":"some description","expireAtInMs":12344546,"id":2,"name":"some-name","token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6IkFQSS1UT0tFTjpzb21lLW...

---

### ✅ POST /orchestrator/api-token

- **Status**: PASS
- **Duration**: 98.274584ms
- **Spec File**: ../../specs/openapiClient/api/apiToken_api-openapi.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"hideApiToken":false,"id":3,"success":true,"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6IkFQSS1UT0tFTjpzYW1wbGUtYXBpLXRva2VuIiwidmVyc2lvbiI6IjEiLCJpc3M...

---

### ✅ GET /orchestrator/gitops/config/{id}

- **Status**: PASS
- **Duration**: 90.785666ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"provider":"GITHUB","username":"systemsdt","token":"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","gitLabGro...

---

### ✅ GET /orchestrator/gitops/configured

- **Status**: PASS
- **Duration**: 90.2585ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"allowCustomRepository":false,"authMode":"PASSWORD","exists":true,"provider":"GITHUB"}}

---

### ❌ POST /orchestrator/gitops/validate

- **Status**: FAIL
- **Duration**: 89.963375ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 93.089291ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 409
- **Error/Msg**: {"code":409,"status":"Conflict","errors":[{"internalMessage":"gitops provider already exists","userMessage":"gitops provider already exists"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 409

---

### ❌ PUT /orchestrator/gitops/config

- **Status**: FAIL
- **Duration**: 90.507083ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/gitops/config

- **Status**: PASS
- **Duration**: 92.020584ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":1,"provider":"GITHUB","username":"systemsdt","token":"","gitLabGroupId":"","gitHubOrgId":"stage-gitops","host":"https://github.com/","active":true,"azureProje...

---

### ❌ GET /orchestrator/gitops/config-by-provider

- **Status**: FAIL
- **Duration**: 87.1505ms
- **Spec File**: ../../specs/gitops/core.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ PATCH /orchestrator/app/env/patch

- **Status**: FAIL
- **Duration**: 94.708458ms
- **Spec File**: ../../specs/helm/deployment-chart-type.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","result":"env properties not found"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/app/ci-pipeline/patch

- **Status**: FAIL
- **Duration**: 104.304625ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'CiPatchRequest.CiPipeline.CiMaterial[0].Source' Error:Field validation for 'Source' failed on the 'dive' tag","userMessage":"Key:...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/ci-pipeline/{appId}

- **Status**: PASS
- **Duration**: 98.887584ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"appId":1,"dockerRegistry":"quay","dockerRepository":"devtron/test","ciBuildConfig":{"id":1,"gitMaterialId":1,"buildContextGitMaterialId":1,"useRootBuildCont...

---

### ✅ GET /orchestrator/app/wf/all/component-names/{appId}

- **Status**: PASS
- **Duration**: 94.404208ms
- **Spec File**: ../../specs/infrastructure/docker-build.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"workflows":[{"id":1,"name":"wf-1-lcu5","ciPipelineId":0,"ciPipelineName":"","cdPipelines":null},{"id":2,"name":"wf-1-4rww","ciPipelineId":0,"ciPipelineName":"","cd...

---

### ❌ POST /orchestrator/application/template-chart

- **Status**: FAIL
- **Duration**: 901.217417ms
- **Spec File**: ../../specs/openapiClient/api/openapi.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{rpc error: code = Unknown desc = error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string i...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PUT /orchestrator/application/rollback

- **Status**: FAIL
- **Duration**: 99.004209ms
- **Spec File**: ../../specs/openapiClient/api/openapi.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"500","internalMessage":"release name is invalid: someName","userMessage":"release name is invalid: someName"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/app-store/discover

- **Status**: PASS
- **Duration**: 614.482625ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":65,"appStoreApplicationVersionId":4028,"name":"ai-agent","chart_repo_id":2,"docker_artifact_store_id":"","chart_name":"devtron","icon":"","active":true,"chart...

---

### ✅ GET /orchestrator/app-store/discover/application/chartInfo/{appStoreApplicationVersionId}

- **Status**: PASS
- **Duration**: 99.225417ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"appStoreApplicationVersionId":1,"readme":"# cluster-autoscaler\n\nScales Kubernetes worker nodes within autoscaling groups.\n\n## TL;DR\n\n```console\n$ helm repo ...

---

### ✅ GET /orchestrator/app-store/discover/application/{appStoreId}/version/autocomplete

- **Status**: PASS
- **Duration**: 93.867334ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"version":"9.49.0","id":5507},{"version":"9.48.0","id":1},{"version":"9.47.0","id":2},{"version":"9.46.6","id":3},{"version":"9.46.5","id":4},{"version":"9.46.4","...

---

### ✅ GET /orchestrator/app-store/discover/application/{id}

- **Status**: PASS
- **Duration**: 114.070459ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":1,"version":"9.48.0","appVersion":"1.33.0","created":"2025-07-11T21:16:00.149315Z","deprecated":false,"description":"Scales Kubernetes worker nodes within auto...

---

### ❌ GET /orchestrator/app-store/discover/search

- **Status**: FAIL
- **Duration**: 88.402167ms
- **Spec File**: ../../specs/app-store.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/ci-pipeline/{ciPipelineId}/linked-ci/downstream/cd

- **Status**: FAIL
- **Duration**: 92.017458ms
- **Spec File**: ../../specs/ci-pipeline/ciPipelineDownstream/downstream-linked-ci-view-spec.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{invalid CiPipelineId 1}]","userMessage":"invalid CiPipelineId 1"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/ci-pipeline/{ciPipelineId}/linked-ci/downstream/env

- **Status**: FAIL
- **Duration**: 92.755833ms
- **Spec File**: ../../specs/ci-pipeline/ciPipelineDownstream/downstream-linked-ci-view-spec.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{invalid CiPipelineId 1}]","userMessage":"invalid CiPipelineId 1"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/deployment-status/timeline/{appId}/{envId}

- **Status**: FAIL
- **Duration**: 91.845833ms
- **Spec File**: ../../specs/deployment/timeline.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/flux-application

- **Status**: FAIL
- **Duration**: 89.193875ms
- **Spec File**: ../../specs/gitops/fluxcd.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{error in getting cluster ids}]","userMessage":"error in getting cluster ids"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/flux-application/app

- **Status**: FAIL
- **Duration**: 86.492958ms
- **Spec File**: ../../specs/gitops/fluxcd.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/app/commit-info/{ciPipelineMaterialId}/{gitHash}

- **Status**: FAIL
- **Duration**: 91.786959ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/ci-pipeline/trigger

- **Status**: FAIL
- **Duration**: 89.328625ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/{appId}/ci-pipeline/{pipelineId}/workflow/{workflowId}

- **Status**: FAIL
- **Duration**: 157.807292ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-build-spec.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: invalid character 'p' after top-level value

---

### ❌ GET /orchestrator/app/history/deployed-configuration/all/latest/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 102.396542ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{strconv.Atoi: parsing \"latest\": invalid syntax}]","userMessage":"strconv.Atoi: parsing \"latest\": invalid syntax"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/app/history/deployed-configuration/all/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 101.673625ms
- **Spec File**: ../../specs/deployment/rollback.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"deploymentTemplate":{"templateName":"Deployment","templateVersion":"4.21.0","isAppMetricsEnabled":false,"codeEditorValue":{"displayName":"values.yaml","value":"{\"...

---

### ❌ POST /orchestrator/module

- **Status**: FAIL
- **Duration**: 87.529167ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/module

- **Status**: PASS
- **Duration**: 95.123375ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"name":"security.trivy","status":"installed","moduleResourcesStatus":null,"enabled":true,"moduleType":"security"},{"name":"security.clair","status":"installed","mo...

---

### ❌ POST /orchestrator/server

- **Status**: FAIL
- **Duration**: 92.159666ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"internalMessage":"Key: 'ServerActionRequestDto.Action' Error:Field validation for 'Action' failed on the 'oneof' tag","userMessage":"Key: 'ServerActionRe...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/server

- **Status**: PASS
- **Duration**: 89.233625ms
- **Spec File**: ../../specs/modularisation/v1.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"status":"unknown","releaseName":"devtron","installationType":"enterprise"}}

---

### ❌ GET /orchestrator/env/namespace/autocomplete

- **Status**: FAIL
- **Duration**: 89.794125ms
- **Spec File**: ../../specs/environment/core.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{ERROR #42601 syntax error at or near \")\"}]","userMessage":"ERROR #42601 syntax error at or near \")\""}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ✅ GET /orchestrator/app/history/deployed-configuration/{appId}/{pipelineId}/{wfrId}

- **Status**: PASS
- **Duration**: 92.201667ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":8,"name":"DEPLOYMENT_TEMPLATE"},{"id":3,"name":"PIPELINE_STRATEGY"}]}

---

### ❌ GET /orchestrator/app/history/deployed-component/detail/{appId}/{pipelineId}/{id}

- **Status**: FAIL
- **Duration**: 96.500042ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","result":"invalid historyComponent"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/app/history/deployed-component/list/{appId}/{pipelineId}

- **Status**: FAIL
- **Duration**: 156.065209ms
- **Spec File**: ../../specs/audit/definitions.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","result":"invalid historyComponent"}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/version

- **Status**: PASS
- **Duration**: 93.769333ms
- **Spec File**: ../../specs/common/version.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"result":{"gitCommit":"908eae83","buildTime":"2025-08-05T20:18:14Z","serverMode":"FULL"}}

---

### ❌ GET /orchestrator/app/template/default/{appId}/{chartRefId}

- **Status**: FAIL
- **Duration**: 126.215167ms
- **Spec File**: ../../specs/environment/templates.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ DELETE /orchestrator/config/environment/cm/{appId}/{envId}/{id}

- **Status**: FAIL
- **Duration**: 86.936125ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/config/global/cm/{appId}

- **Status**: PASS
- **Duration**: 92.167417ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":0,"appId":1,"configData":[],"isDeletable":false,"isExpressEdit":false}}

---

### ❌ DELETE /orchestrator/config/global/cm/{appId}/{id}

- **Status**: FAIL
- **Duration**: 88.465583ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/config/environment/cm/edit/{appId}/{envId}/{id}

- **Status**: FAIL
- **Duration**: 88.335583ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ GET /orchestrator/config/global/cm/edit/{appId}/{id}

- **Status**: FAIL
- **Duration**: 88.741625ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/config/environment/cm

- **Status**: FAIL
- **Duration**: 90.59825ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{json: cannot unmarshal string into Go struct field ConfigData.configData.subPath of type bool}]","userMessage":"json: ca...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/config/environment/cm/{appId}/{envId}

- **Status**: PASS
- **Duration**: 90.183917ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":0,"appId":1,"environmentId":1,"configData":[],"isDeletable":false,"isExpressEdit":false}}

---

### ❌ POST /orchestrator/config/global/cm

- **Status**: FAIL
- **Duration**: 154.39875ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{json: cannot unmarshal string into Go struct field ConfigData.configData.subPath of type bool}]","userMessage":"json: ca...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/config/bulk/patch

- **Status**: FAIL
- **Duration**: 94.131834ms
- **Spec File**: ../../specs/plugins/config-maps.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{invalid request no payload found for sync}]","userMessage":"invalid request no payload found for sync"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ PATCH /orchestrator/app/ci-pipeline/patch-source

- **Status**: FAIL
- **Duration**: 90.59375ms
- **Spec File**: ../../specs/ci-pipeline/ci-pipeline-change-source.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/deployment/template/fetch

- **Status**: PASS
- **Duration**: 98.8015ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"id":10,"name":"Rollout Deployment","chartDescription":"This chart deploys an advanced version of deployment that supports Blue/Green and Canary deployments. For f...

---

### ✅ PUT /orchestrator/deployment/template/upload

- **Status**: PASS
- **Duration**: 91.756375ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":"Processed successfully"}

---

### ❌ POST /orchestrator/deployment/template/validate

- **Status**: FAIL
- **Duration**: 91.986084ms
- **Spec File**: ../../specs/deployment/core.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{request Content-Type isn't multipart/form-data}]","userMessage":"request Content-Type isn't multipart/form-data"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/chartref/autocomplete/{appId}

- **Status**: PASS
- **Duration**: 94.592ms
- **Spec File**: ../../specs/helm/dynamic-charts.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"chartRefs":[{"id":10,"version":"3.9.0","name":"Rollout Deployment","description":"","userUploaded":false,"isAppMetricsSupported":true},{"id":11,"version":"3.10.0",...

---

### ❌ POST /orchestrator/chart-repo/update

- **Status**: FAIL
- **Duration**: 98.011083ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
- **Response Code**: 412
- **Error/Msg**: {"code":412,"status":"Precondition Failed","result":{"customErrMsg":"Could not validate the repo. Please try again.","actualErrMsg":"Get \"https://charts.example.com/index.yaml\": dial tcp: lookup cha...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 412

---

### ✅ POST /orchestrator/chart-repo/validate

- **Status**: PASS
- **Duration**: 97.090042ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"customErrMsg":"Could not validate the repo. Please try again.","actualErrMsg":"Get \"https://charts.example.com/index.yaml\": dial tcp: lookup charts.example.com o...

---

### ✅ POST /orchestrator/chart-repo/create

- **Status**: PASS
- **Duration**: 98.074416ms
- **Spec File**: ../../specs/helm/repo-validation.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"customErrMsg":"Could not validate the repo. Please try again.","actualErrMsg":"Get \"https://charts.example.com/index.yaml\": dial tcp: lookup charts.example.com o...

---

### ❌ DELETE /orchestrator/notification/channel

- **Status**: FAIL
- **Duration**: 144.437916ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{EOF}]","userMessage":"EOF"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/notification/channel

- **Status**: PASS
- **Duration**: 94.410292ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"slackConfigs":[],"webhookConfigs":[],"sesConfigs":[{"userId":2,"teamId":0,"region":"ap-south-1","accessKey":"AKIAWEAVHFOOSNAORPHU","secretKey":"**********","fromEm...

---

### ❌ POST /orchestrator/notification/channel

- **Status**: FAIL
- **Duration**: 91.409542ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200
- **Error/Msg**: {}

**Issues:**
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON: unexpected end of JSON input

---

### ✅ GET /orchestrator/notification/channel/autocomplete/{type}

- **Status**: PASS
- **Duration**: 89.415583ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[]}

---

### ✅ GET /orchestrator/notification/channel/webhook/{id}

- **Status**: PASS
- **Duration**: 94.489166ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"userId":0,"webhookUrl":"","configName":"","header":null,"payload":"","description":"","id":0}}

---

### ❌ GET /orchestrator/notification/recipient

- **Status**: FAIL
- **Duration**: 89.333042ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ✅ GET /orchestrator/notification/channel/slack/{id}

- **Status**: PASS
- **Duration**: 91.360417ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"userId":0,"teamId":0,"webhookUrl":"","configName":"","description":"","id":0}}

---

### ✅ GET /orchestrator/notification/channel/smtp/{id}

- **Status**: PASS
- **Duration**: 95.596792ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"id":0,"port":"","host":"","authType":"","authUser":"","authPassword":"","fromEmail":"","configName":"","description":"","ownerId":0,"default":false,"deleted":false...

---

### ✅ POST /orchestrator/notification/search

- **Status**: PASS
- **Duration**: 99.061916ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":[{"team":null,"app":null,"environment":null,"cluster":null,"pipeline":{"id":1,"name":"cd-3-pgxd","environmentName":"devtron-demo","appName":"pk-test","isVirtualEnvir...

---

### ❌ POST /orchestrator/notification/v2

- **Status**: FAIL
- **Duration**: 90.868334ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{json: cannot unmarshal object into Go struct field NotificationRequest.notificationConfigRequest of type []*beans.Notifi...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ✅ GET /orchestrator/notification/variables

- **Status**: PASS
- **Duration**: 138.737917ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"devtronAppId":"{{devtronAppId}}","devtronAppName":"{{devtronAppName}}","devtronApprovedByEmail":"{{devtronApprovedByEmail}}","devtronBuildGitCommitHash":"{{devtron...

---

### ✅ GET /orchestrator/notification/channel/ses/{id}

- **Status**: PASS
- **Duration**: 92.934375ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 200
- **Error/Msg**: {"code":200,"status":"OK","result":{"userId":2,"teamId":0,"region":"ap-south-1","accessKey":"AKIAWEAVHFOOSNAORPHU","secretKey":"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","fromEmail":"watcher@devtron.i...

---

### ❌ DELETE /orchestrator/notification

- **Status**: FAIL
- **Duration**: 89.627458ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{EOF}]","userMessage":"EOF"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ GET /orchestrator/notification

- **Status**: FAIL
- **Duration**: 86.006958ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 404
- **Error/Msg**: 404 page not found


**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/notification

- **Status**: FAIL
- **Duration**: 91.812625ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{json: cannot unmarshal object into Go struct field NotificationRequest.notificationConfigRequest of type []*beans.Notifi...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ PUT /orchestrator/notification

- **Status**: FAIL
- **Duration**: 91.81ms
- **Spec File**: ../../specs/notifications/core.yaml
- **Response Code**: 400
- **Error/Msg**: {"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"[{json: cannot unmarshal object into Go struct field NotificationUpdateRequest.notificationConfigRequest of type []*beans....

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 400

---

### ❌ POST /orchestrator/webhook/git

- **Status**: FAIL
- **Duration**: 88.203708ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"resource not found","userMessage":"Requested resource not found","userDetailMessage":"The requested resource does not exis...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/git/{gitHostId}

- **Status**: FAIL
- **Duration**: 89.022666ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"git host not found: sample_string","userMessage":"git host with ID 'sample_string' not found","userDetailMessage":"The req...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/git/{gitHostId}/{secret}

- **Status**: FAIL
- **Duration**: 88.782709ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 404
- **Error/Msg**: {"code":404,"status":"Not Found","errors":[{"code":"11006","internalMessage":"git host not found: sample_string","userMessage":"git host with ID 'sample_string' not found","userDetailMessage":"The req...

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 404

---

### ❌ POST /orchestrator/webhook/ci/workflow

- **Status**: FAIL
- **Duration**: 90.141709ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 500
- **Error/Msg**: {"code":500,"status":"Internal Server Error","errors":[{"code":"000","internalMessage":"[{invalid wf name}]","userMessage":0}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 500

---

### ❌ POST /orchestrator/webhook/ext-ci/{externalCiId}

- **Status**: FAIL
- **Duration**: 88.939625ms
- **Spec File**: ../../specs/notifications/webhooks.yaml
- **Response Code**: 401
- **Error/Msg**: {"code":401,"status":"Unauthorized","errors":[{"code":"6005","internalMessage":"no token provided"}]}

**Issues:**
- **STATUS_CODE_MISMATCH**: Expected status 200, got 401

---


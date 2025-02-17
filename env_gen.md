

## CD Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | ARGO_APP_MANUAL_SYNC_TIME | int |3 |  |  | false |
 | CD_HELM_PIPELINE_STATUS_CRON_TIME | string |*/2 * * * * |  |  | false |
 | CD_PIPELINE_STATUS_CRON_TIME | string |*/2 * * * * |  |  | false |
 | CD_PIPELINE_STATUS_TIMEOUT_DURATION | string |20 |  |  | false |
 | DEPLOY_STATUS_CRON_GET_PIPELINE_DEPLOYED_WITHIN_HOURS | int |12 |  |  | false |
 | DEVTRON_CHART_ARGO_CD_INSTALL_REQUEST_TIMEOUT | int |1 |  |  | false |
 | DEVTRON_CHART_INSTALL_REQUEST_TIMEOUT | int |6 |  |  | false |
 | EXPOSE_CD_METRICS | bool |false |  |  | false |
 | FEATURE_MIGRATE_ARGOCD_APPLICATION_ENABLE | bool |false | enable migration of external argocd application to devtron pipeline |  | false |
 | HELM_PIPELINE_STATUS_CHECK_ELIGIBLE_TIME | string |120 |  |  | false |
 | IS_INTERNAL_USE | bool |true |  |  | false |
 | MIGRATE_DEPLOYMENT_CONFIG_DATA | bool |false | migrate deployment config data from charts table to deployment_config table |  | false |
 | PIPELINE_DEGRADED_TIME | string |10 |  |  | false |
 | REVISION_HISTORY_LIMIT_DEVTRON_APP | int |1 |  |  | false |
 | REVISION_HISTORY_LIMIT_EXTERNAL_HELM_APP | int |0 |  |  | false |
 | REVISION_HISTORY_LIMIT_HELM_APP | int |1 |  |  | false |
 | RUN_HELM_INSTALL_IN_ASYNC_MODE_HELM_APPS | bool |false |  |  | false |
 | SHOULD_CHECK_NAMESPACE_ON_CLONE | bool |false | should we check if namespace exists or not while cloning app |  | false |
 | USE_DEPLOYMENT_CONFIG_DATA | bool |false | use deployment config data from deployment_config table |  | true |


## CI_RUNNER Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | AZURE_ACCOUNT_KEY | string | |  |  | false |
 | AZURE_ACCOUNT_NAME | string | |  |  | false |
 | AZURE_BLOB_CONTAINER_CI_CACHE | string | |  |  | false |
 | AZURE_BLOB_CONTAINER_CI_LOG | string | |  |  | false |
 | AZURE_GATEWAY_CONNECTION_INSECURE | bool |true |  |  | false |
 | AZURE_GATEWAY_URL | string |http://devtron-minio.devtroncd:9000 |  |  | false |
 | BASE_LOG_LOCATION_PATH | string |/home/devtron/ |  |  | false |
 | BLOB_STORAGE_GCP_CREDENTIALS_JSON | string | |  |  | false |
 | BLOB_STORAGE_PROVIDER |  |S3 |  |  | false |
 | BLOB_STORAGE_S3_ACCESS_KEY | string | |  |  | false |
 | BLOB_STORAGE_S3_BUCKET_VERSIONED | bool |true |  |  | false |
 | BLOB_STORAGE_S3_ENDPOINT | string | |  |  | false |
 | BLOB_STORAGE_S3_ENDPOINT_INSECURE | bool |false |  |  | false |
 | BLOB_STORAGE_S3_SECRET_KEY | string | |  |  | false |
 | BUILDX_CACHE_PATH | string |/var/lib/devtron/buildx |  |  | false |
 | BUILDX_K8S_DRIVER_OPTIONS | string | |  |  | false |
 | BUILDX_PROVENANCE_MODE | string | |  |  | false |
 | BUILD_LOG_TTL_VALUE_IN_SECS | int |3600 |  |  | false |
 | CACHE_LIMIT | int64 |5000000000 |  |  | false |
 | CD_DEFAULT_ADDRESS_POOL_BASE_CIDR | string | |  |  | false |
 | CD_DEFAULT_ADDRESS_POOL_SIZE | int | |  |  | false |
 | CD_LIMIT_CI_CPU | string |0.5 |  |  | false |
 | CD_LIMIT_CI_MEM | string |3G |  |  | false |
 | CD_NODE_LABEL_SELECTOR |  | |  |  | false |
 | CD_NODE_TAINTS_KEY | string |dedicated |  |  | false |
 | CD_NODE_TAINTS_VALUE | string |ci |  |  | false |
 | CD_REQ_CI_CPU | string |0.5 |  |  | false |
 | CD_REQ_CI_MEM | string |3G |  |  | false |
 | CD_WORKFLOW_EXECUTOR_TYPE |  |AWF |  |  | false |
 | CD_WORKFLOW_SERVICE_ACCOUNT | string |cd-runner |  |  | false |
 | CI_DEFAULT_ADDRESS_POOL_BASE_CIDR | string | |  |  | false |
 | CI_DEFAULT_ADDRESS_POOL_SIZE | int | |  |  | false |
 | CI_IGNORE_DOCKER_CACHE | bool | |  |  | false |
 | CI_LOGS_KEY_PREFIX | string | |  |  | false |
 | CI_NODE_LABEL_SELECTOR |  | |  |  | false |
 | CI_NODE_TAINTS_KEY | string | |  |  | false |
 | CI_NODE_TAINTS_VALUE | string | |  |  | false |
 | CI_RUNNER_DOCKER_MTU_VALUE | int |-1 |  |  | false |
 | CI_SUCCESS_AUTO_TRIGGER_BATCH_SIZE | int |1 |  |  | false |
 | CI_VOLUME_MOUNTS_JSON | string | |  |  | false |
 | CI_WORKFLOW_EXECUTOR_TYPE |  |AWF |  |  | false |
 | DEFAULT_ARTIFACT_KEY_LOCATION | string |arsenal-v1/ci-artifacts |  |  | false |
 | DEFAULT_BUILD_LOGS_BUCKET | string |devtron-pro-ci-logs |  |  | false |
 | DEFAULT_BUILD_LOGS_KEY_PREFIX | string |arsenal-v1 |  |  | false |
 | DEFAULT_CACHE_BUCKET | string |ci-caching |  |  | false |
 | DEFAULT_CACHE_BUCKET_REGION | string |us-east-2 |  |  | false |
 | DEFAULT_CD_ARTIFACT_KEY_LOCATION | string | |  |  | false |
 | DEFAULT_CD_LOGS_BUCKET_REGION | string |us-east-2 |  |  | false |
 | DEFAULT_CD_NAMESPACE | string | |  |  | false |
 | DEFAULT_CD_TIMEOUT | int64 |3600 |  |  | false |
 | DEFAULT_CI_IMAGE | string |686244538589.dkr.ecr.us-east-2.amazonaws.com/cirunner:47 |  |  | false |
 | DEFAULT_NAMESPACE | string |devtron-ci |  |  | false |
 | DEFAULT_TARGET_PLATFORM | string | |  |  | false |
 | DOCKER_BUILD_CACHE_PATH | string |/var/lib/docker |  |  | false |
 | ENABLE_BUILD_CONTEXT | bool |false |  |  | false |
 | ENABLE_WORKFLOW_EXECUTION_STAGE | bool |true | if enabled then we will display build stages separately for CI/Job/Pre-Post CD | true | false |
 | EXTERNAL_BLOB_STORAGE_CM_NAME | string |blob-storage-cm |  |  | false |
 | EXTERNAL_BLOB_STORAGE_SECRET_NAME | string |blob-storage-secret |  |  | false |
 | EXTERNAL_CD_NODE_LABEL_SELECTOR |  | |  |  | false |
 | EXTERNAL_CD_NODE_TAINTS_KEY | string |dedicated |  |  | false |
 | EXTERNAL_CD_NODE_TAINTS_VALUE | string |ci |  |  | false |
 | EXTERNAL_CI_API_SECRET | string |devtroncd-secret |  |  | false |
 | EXTERNAL_CI_PAYLOAD | string |{"ciProjectDetails":[{"gitRepository":"https://github.com/vikram1601/getting-started-nodejs.git","checkoutPath":"./abc","commitHash":"239077135f8cdeeccb7857e2851348f558cb53d3","commitTime":"2022-10-30T20:00:00","branch":"master","message":"Update README.md","author":"User Name "}],"dockerImage":"445808685819.dkr.ecr.us-east-2.amazonaws.com/orch:23907713-2"} |  |  | false |
 | EXTERNAL_CI_WEB_HOOK_URL | string | |  |  | false |
 | IGNORE_CM_CS_IN_CI_JOB | bool |false |  |  | false |
 | IMAGE_RETRY_COUNT | int |0 |  |  | false |
 | IMAGE_RETRY_INTERVAL | int |5 |  |  | false |
 | IMAGE_SCANNER_ENDPOINT | string |http://image-scanner-new-demo-devtroncd-service.devtroncd:80 |  |  | false |
 | IMAGE_SCAN_MAX_RETRIES | int |3 |  |  | false |
 | IMAGE_SCAN_RETRY_DELAY | int |5 |  |  | false |
 | IN_APP_LOGGING_ENABLED | bool |false |  |  | false |
 | MAX_CD_WORKFLOW_RUNNER_RETRIES | int |0 |  |  | false |
 | MAX_CI_WORKFLOW_RETRIES | int |0 |  |  | false |
 | MODE | string |DEV |  |  | false |
 | NATS_SERVER_HOST | string |localhost:4222 |  |  | false |
 | ORCH_HOST | string |http://devtroncd-orchestrator-service-prod.devtroncd/webhook/msg/nats |  |  | false |
 | ORCH_TOKEN | string | |  |  | false |
 | PRE_CI_CACHE_PATH | string |/devtroncd-cache |  |  | false |
 | SHOW_DOCKER_BUILD_ARGS | bool |true |  |  | false |
 | SKIP_CI_JOB_BUILD_CACHE_PUSH_PULL | bool |false |  |  | false |
 | SKIP_CREATING_ECR_REPO | bool |false |  |  | false |
 | TERMINATION_GRACE_PERIOD_SECS | int |180 |  |  | false |
 | USE_ARTIFACT_LISTING_QUERY_V2 | bool |true |  |  | false |
 | USE_BLOB_STORAGE_CONFIG_IN_CD_WORKFLOW | bool |true |  |  | false |
 | USE_BLOB_STORAGE_CONFIG_IN_CI_WORKFLOW | bool |true |  |  | false |
 | USE_BUILDX | bool |false |  |  | false |
 | USE_DOCKER_API_TO_GET_DIGEST | bool |false |  |  | false |
 | USE_EXTERNAL_NODE | bool |false |  |  | false |
 | USE_IMAGE_TAG_FROM_GIT_PROVIDER_FOR_TAG_BASED_BUILD | bool |false |  |  | false |
 | WF_CONTROLLER_INSTANCE_ID | string |devtron-runner |  |  | false |
 | WORKFLOW_CACHE_CONFIG | string |{} |  |  | false |
 | WORKFLOW_SERVICE_ACCOUNT | string |ci-runner |  |  | false |


## DEVTRON Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | - |  | |  |  | false |
 | APP_SYNC_IMAGE | string |quay.io/devtron/chart-sync:1227622d-132-3775 |  |  | false |
 | APP_SYNC_JOB_RESOURCES_OBJ | string | |  |  | false |
 | APP_SYNC_SERVICE_ACCOUNT | string |chart-sync |  |  | false |
 | ARGO_AUTO_SYNC_ENABLED | bool |true |  |  | false |
 | ARGO_GIT_COMMIT_RETRY_COUNT_ON_CONFLICT | int |3 |  |  | false |
 | ARGO_GIT_COMMIT_RETRY_DELAY_ON_CONFLICT | int |1 |  |  | false |
 | ARGO_REPO_REGISTER_RETRY_COUNT | int |3 |  |  | false |
 | ARGO_REPO_REGISTER_RETRY_DELAY | int |10 |  |  | false |
 | ASYNC_BUILDX_CACHE_EXPORT | bool |false |  |  | false |
 | BATCH_SIZE | int |5 |  |  | false |
 | BLOB_STORAGE_ENABLED | bool |false |  |  | false |
 | BUILDX_CACHE_MODE_MIN | bool |false |  |  | false |
 | CD_HOST | string |localhost |  |  | false |
 | CD_NAMESPACE | string |devtroncd |  |  | false |
 | CD_PORT | string |8000 |  |  | false |
 | CExpirationTime | int |600 |  |  | false |
 | CI_TRIGGER_CRON_TIME | int |2 |  |  | false |
 | CI_WORKFLOW_STATUS_UPDATE_CRON | string |*/5 * * * * |  |  | false |
 | CLI_CMD_TIMEOUT_GLOBAL_SECONDS | int |0 |  |  | false |
 | CLUSTER_STATUS_CRON_TIME | int |15 |  |  | false |
 | CONSUMER_CONFIG_JSON | string | |  |  | false |
 | DEFAULT_LOG_TIME_LIMIT | int64 |1 |  |  | false |
 | DEFAULT_TIMEOUT | float64 |3600 |  |  | false |
 | DEVTRON_BOM_URL | string |https://raw.githubusercontent.com/devtron-labs/devtron/%s/charts/devtron/devtron-bom.yaml |  |  | false |
 | DEVTRON_DEFAULT_NAMESPACE | string |devtroncd |  |  | false |
 | DEVTRON_DEX_SECRET_NAMESPACE | string |devtroncd |  |  | false |
 | DEVTRON_HELM_RELEASE_CHART_NAME | string |devtron-operator |  |  | false |
 | DEVTRON_HELM_RELEASE_NAME | string |devtron |  |  | false |
 | DEVTRON_HELM_RELEASE_NAMESPACE | string |devtroncd |  |  | false |
 | DEVTRON_HELM_REPO_NAME | string |devtron |  |  | false |
 | DEVTRON_HELM_REPO_URL | string |https://helm.devtron.ai |  |  | false |
 | DEVTRON_INSTALLATION_TYPE | string | |  |  | false |
 | DEVTRON_MODULES_IDENTIFIER_IN_HELM_VALUES | string |installer.modules |  |  | false |
 | DEVTRON_SECRET_NAME | string |devtron-secret |  |  | false |
 | DEVTRON_VERSION_IDENTIFIER_IN_HELM_VALUES | string |installer.release |  |  | false |
 | DEX_CID | string |example-app |  |  | false |
 | DEX_CLIENT_ID | string |argo-cd |  |  | false |
 | DEX_CSTOREKEY | string | |  |  | false |
 | DEX_JWTKEY | string | |  |  | false |
 | DEX_RURL | string |http://127.0.0.1:8080/callback |  |  | false |
 | DEX_SECRET | string | |  |  | false |
 | DEX_URL | string | |  |  | false |
 | ECR_REPO_NAME_PREFIX | string |test/ |  |  | false |
 | ENABLE_ASYNC_ARGO_CD_INSTALL_DEVTRON_CHART | bool |false |  |  | false |
 | ENABLE_ASYNC_INSTALL_DEVTRON_CHART | bool |false |  |  | false |
 | EPHEMERAL_SERVER_VERSION_REGEX | string |v[1-9]\.\b(2[3-9]\|[3-9][0-9])\b.* |  |  | false |
 | EVENT_URL | string |http://localhost:3000/notify |  |  | false |
 | EXECUTE_WIRE_NIL_CHECKER | bool |false |  |  | false |
 | EXPOSE_CI_METRICS | bool |false |  |  | false |
 | FEATURE_RESTART_WORKLOAD_BATCH_SIZE | int |1 |  |  | false |
 | FEATURE_RESTART_WORKLOAD_WORKER_POOL_SIZE | int |5 |  |  | false |
 | FORCE_SECURITY_SCANNING | bool |false |  |  | false |
 | GITOPS_REPO_PREFIX | string | |  |  | false |
 | GO_RUNTIME_ENV | string |production |  |  | false |
 | GRAFANA_HOST | string |localhost |  |  | false |
 | GRAFANA_NAMESPACE | string |devtroncd |  |  | false |
 | GRAFANA_ORG_ID | int |2 |  |  | false |
 | GRAFANA_PASSWORD | string |prom-operator |  |  | false |
 | GRAFANA_PORT | string |8090 |  |  | false |
 | GRAFANA_URL | string | |  |  | false |
 | GRAFANA_USERNAME | string |admin |  |  | false |
 | HIDE_IMAGE_TAGGING_HARD_DELETE | bool |false |  |  | false |
 | IGNORE_AUTOCOMPLETE_AUTH_CHECK | bool |false |  |  | false |
 | INSTALLER_CRD_NAMESPACE | string |devtroncd |  |  | false |
 | INSTALLER_CRD_OBJECT_GROUP_NAME | string |installer.devtron.ai |  |  | false |
 | INSTALLER_CRD_OBJECT_RESOURCE | string |installers |  |  | false |
 | INSTALLER_CRD_OBJECT_VERSION | string |v1alpha1 |  |  | false |
 | JwtExpirationTime | int |120 |  |  | false |
 | K8s_CLIENT_MAX_IDLE_CONNS_PER_HOST | int |25 |  |  | false |
 | K8s_TCP_IDLE_CONN_TIMEOUT | int |300 |  |  | false |
 | K8s_TCP_KEEPALIVE | int |30 |  |  | false |
 | K8s_TCP_TIMEOUT | int |30 |  |  | false |
 | K8s_TLS_HANDSHAKE_TIMEOUT | int |10 |  |  | false |
 | KUBELINK_GRPC_MAX_RECEIVE_MSG_SIZE | int |20 |  |  | false |
 | KUBELINK_GRPC_MAX_SEND_MSG_SIZE | int |4 |  |  | false |
 | LENS_TIMEOUT | int |0 |  |  | false |
 | LENS_URL | string |http://lens-milandevtron-service:80 |  |  | false |
 | LIMIT_CI_CPU | string |0.5 |  |  | false |
 | LIMIT_CI_MEM | string |3G |  |  | false |
 | LOGGER_DEV_MODE | bool |false |  |  | false |
 | LOG_LEVEL | int |-1 |  |  | false |
 | MAX_SESSION_PER_USER | int |5 |  |  | false |
 | MODULE_METADATA_API_URL | string |https://api.devtron.ai/module?name=%s |  |  | false |
 | MODULE_STATUS_HANDLING_CRON_DURATION_MIN | int |3 |  |  | false |
 | NATS_MSG_ACK_WAIT_IN_SECS | int |120 |  |  | false |
 | NATS_MSG_BUFFER_SIZE | int |-1 |  |  | false |
 | NATS_MSG_MAX_AGE | int |86400 |  |  | false |
 | NATS_MSG_PROCESSING_BATCH_SIZE | int |1 |  |  | false |
 | NATS_MSG_REPLICAS | int |0 |  |  | false |
 | NOTIFICATION_MEDIUM | NotificationMedium |rest |  |  | false |
 | OTEL_COLLECTOR_URL | string | |  |  | false |
 | PARALLELISM_LIMIT_FOR_TAG_PROCESSING | int | |  |  | false |
 | PG_EXPORT_PROM_METRICS | bool |true |  |  | false |
 | PG_LOG_ALL_FAILURE_QUERIES | bool |true |  |  | false |
 | PG_LOG_ALL_QUERY | bool |false |  |  | false |
 | PG_LOG_SLOW_QUERY | bool |true |  |  | false |
 | PG_QUERY_DUR_THRESHOLD | int64 |5000 |  |  | false |
 | PLUGIN_NAME | string |Pull images from container repository |  |  | false |
 | PROPAGATE_EXTRA_LABELS | bool |false |  |  | false |
 | PROXY_SERVICE_CONFIG | string |{} |  |  | false |
 | REQ_CI_CPU | string |0.5 |  |  | false |
 | REQ_CI_MEM | string |3G |  |  | false |
 | RESTRICT_TERMINAL_ACCESS_FOR_NON_SUPER_USER | bool |false |  |  | false |
 | RUNTIME_CONFIG_LOCAL_DEV | LocalDevMode |true |  |  | false |
 | SCOPED_VARIABLE_ENABLED | bool |false |  |  | false |
 | SCOPED_VARIABLE_FORMAT | string |@{{%s}} |  |  | false |
 | SCOPED_VARIABLE_HANDLE_PRIMITIVES | bool |false |  |  | false |
 | SCOPED_VARIABLE_NAME_REGEX | string |^[a-zA-Z][a-zA-Z0-9_-]{0,62}[a-zA-Z0-9]$ |  |  | false |
 | SOCKET_DISCONNECT_DELAY_SECONDS | int |5 |  |  | false |
 | SOCKET_HEARTBEAT_SECONDS | int |25 |  |  | false |
 | STREAM_CONFIG_JSON | string | |  |  | false |
 | SYSTEM_VAR_PREFIX | string |DEVTRON_ |  |  | false |
 | TERMINAL_POD_DEFAULT_NAMESPACE | string |default |  |  | false |
 | TERMINAL_POD_INACTIVE_DURATION_IN_MINS | int |10 |  |  | false |
 | TERMINAL_POD_STATUS_SYNC_In_SECS | int |600 |  |  | false |
 | TEST_APP | string |orchestrator |  |  | false |
 | TEST_PG_ADDR | string |127.0.0.1 |  |  | false |
 | TEST_PG_DATABASE | string |orchestrator |  |  | false |
 | TEST_PG_LOG_QUERY | bool |true |  |  | false |
 | TEST_PG_PASSWORD | string |postgrespw |  |  | false |
 | TEST_PG_PORT | string |55000 |  |  | false |
 | TEST_PG_USER | string |postgres |  |  | false |
 | TIMEOUT_FOR_FAILED_CI_BUILD | string |15 |  |  | false |
 | TIMEOUT_IN_SECONDS | int |5 |  |  | false |
 | USER_SESSION_DURATION_SECONDS | int |86400 |  |  | false |
 | USE_ARTIFACT_LISTING_API_V2 | bool |true |  |  | false |
 | USE_CUSTOM_HTTP_TRANSPORT | bool |false |  |  | false |
 | USE_GIT_CLI | bool |false |  |  | false |
 | USE_RBAC_CREATION_V2 | bool |true |  |  | false |
 | VARIABLE_CACHE_ENABLED | bool |true |  |  | false |
 | VARIABLE_EXPRESSION_REGEX | string |@{{([^}]+)}} |  |  | false |
 | WEBHOOK_TOKEN | string | |  |  | false |


## GITOPS Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | ACD_CM | string |argocd-cm |  |  | false |
 | ACD_NAMESPACE | string |devtroncd |  |  | false |
 | ACD_PASSWORD | string | |  |  | false |
 | ACD_USERNAME | string |admin |  |  | false |
 | GITOPS_SECRET_NAME | string |devtron-gitops-secret |  |  | false |
 | RESOURCE_LIST_FOR_REPLICAS | string |Deployment,Rollout,StatefulSet,ReplicaSet |  |  | false |
 | RESOURCE_LIST_FOR_REPLICAS_BATCH_SIZE | int |5 |  |  | false |


## INFRA_SETUP Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | DASHBOARD_HOST | string |localhost |  |  | false |
 | DASHBOARD_NAMESPACE | string |devtroncd |  |  | false |
 | DASHBOARD_PORT | string |3000 |  |  | false |
 | DEX_HOST | string |http://localhost |  |  | false |
 | DEX_PORT | string |5556 |  |  | false |
 | GIT_SENSOR_PROTOCOL | string |REST |  |  | false |
 | GIT_SENSOR_TIMEOUT | int |0 |  |  | false |
 | GIT_SENSOR_URL | string |127.0.0.1:7070 |  |  | false |
 | HELM_CLIENT_URL | string |127.0.0.1:50051 |  |  | false |


## POSTGRES Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | APP | string |orchestrator | Application name |  | false |
 | CASBIN_DATABASE | string |casbin |  |  | false |
 | PG_ADDR | string |127.0.0.1 | address of postgres service | postgresql-postgresql.devtroncd | false |
 | PG_DATABASE | string |orchestrator | postgres database to be made connection with | orchestrator, casbin, git_sensor, lens | false |
 | PG_PASSWORD | string |{password} | password for postgres, associated with PG_USER | confidential ;) | false |
 | PG_PORT | string |5432 | port of postgresql service | 5432 | false |
 | PG_READ_TIMEOUT | int64 |30 |  |  | false |
 | PG_USER | string |postgres | user for postgres | postgres | false |
 | PG_WRITE_TIMEOUT | int64 |30 |  |  | false |


## RBAC Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | ENFORCER_CACHE | bool |false |  |  | false |
 | ENFORCER_CACHE_EXPIRATION_IN_SEC | int |86400 |  |  | false |
 | ENFORCER_MAX_BATCH_SIZE | int |1 |  |  | false |
 | USE_CASBIN_V2 | bool |true |  |  | false |


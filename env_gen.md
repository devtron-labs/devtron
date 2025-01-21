

## CD Related Environment Variables
| Key   | Value        | Description       | Possible values       | Deprecated       |
|-------|--------------|-------------------|-----------------------|------------------|
 | ARGO_APP_MANUAL_SYNC_TIME | 3 |  |  | false |
 | CD_HELM_PIPELINE_STATUS_CRON_TIME | */2 * * * * |  |  | false |
 | CD_PIPELINE_STATUS_CRON_TIME | */2 * * * * |  |  | false |
 | CD_PIPELINE_STATUS_TIMEOUT_DURATION | 20 |  |  | false |
 | DEPLOY_STATUS_CRON_GET_PIPELINE_DEPLOYED_WITHIN_HOURS | 12 |  |  | false |
 | DEVTRON_CHART_ARGO_CD_INSTALL_REQUEST_TIMEOUT | 1 |  |  | false |
 | DEVTRON_CHART_INSTALL_REQUEST_TIMEOUT | 6 |  |  | false |
 | EXPOSE_CD_METRICS | false |  |  | false |
 | HELM_PIPELINE_STATUS_CHECK_ELIGIBLE_TIME | 120 |  |  | false |
 | PIPELINE_DEGRADED_TIME | 10 |  |  | false |
 | REVISION_HISTORY_LIMIT_DEVTRON_APP | 1 |  |  | false |
 | REVISION_HISTORY_LIMIT_EXTERNAL_HELM_APP | 0 |  |  | false |
 | REVISION_HISTORY_LIMIT_HELM_APP | 1 |  |  | false |


## CI_RUNNER Related Environment Variables
| Key   | Value        | Description       | Possible values       | Deprecated       |
|-------|--------------|-------------------|-----------------------|------------------|
 | AZURE_ACCOUNT_KEY |  |  |  | false |
 | AZURE_ACCOUNT_NAME |  |  |  | false |
 | AZURE_BLOB_CONTAINER_CI_CACHE |  |  |  | false |
 | AZURE_BLOB_CONTAINER_CI_LOG |  |  |  | false |
 | AZURE_GATEWAY_CONNECTION_INSECURE | true |  |  | false |
 | AZURE_GATEWAY_URL | http://devtron-minio.devtroncd:9000 |  |  | false |
 | BASE_LOG_LOCATION_PATH | /home/devtron/ |  |  | false |
 | BLOB_STORAGE_GCP_CREDENTIALS_JSON |  |  |  | false |
 | BLOB_STORAGE_PROVIDER | S3 |  |  | false |
 | BLOB_STORAGE_S3_ACCESS_KEY |  |  |  | false |
 | BLOB_STORAGE_S3_BUCKET_VERSIONED | true |  |  | false |
 | BLOB_STORAGE_S3_ENDPOINT |  |  |  | false |
 | BLOB_STORAGE_S3_ENDPOINT_INSECURE | false |  |  | false |
 | BLOB_STORAGE_S3_SECRET_KEY |  |  |  | false |
 | BUILDX_CACHE_PATH | /var/lib/devtron/buildx |  |  | false |
 | BUILDX_K8S_DRIVER_OPTIONS |  |  |  | false |
 | BUILDX_PROVENANCE_MODE |  |  |  | false |
 | BUILD_LOG_TTL_VALUE_IN_SECS | 3600 |  |  | false |
 | CACHE_LIMIT | 5000000000 |  |  | false |
 | CD_DEFAULT_ADDRESS_POOL_BASE_CIDR |  |  |  | false |
 | CD_DEFAULT_ADDRESS_POOL_SIZE |  |  |  | false |
 | CD_LIMIT_CI_CPU | 0.5 |  |  | false |
 | CD_LIMIT_CI_MEM | 3G |  |  | false |
 | CD_NODE_LABEL_SELECTOR |  |  |  | false |
 | CD_NODE_TAINTS_KEY | dedicated |  |  | false |
 | CD_NODE_TAINTS_VALUE | ci |  |  | false |
 | CD_REQ_CI_CPU | 0.5 |  |  | false |
 | CD_REQ_CI_MEM | 3G |  |  | false |
 | CD_WORKFLOW_EXECUTOR_TYPE | AWF |  |  | false |
 | CD_WORKFLOW_SERVICE_ACCOUNT | cd-runner |  |  | false |
 | CI_DEFAULT_ADDRESS_POOL_BASE_CIDR |  |  |  | false |
 | CI_DEFAULT_ADDRESS_POOL_SIZE |  |  |  | false |
 | CI_IGNORE_DOCKER_CACHE |  |  |  | false |
 | CI_LOGS_KEY_PREFIX |  |  |  | false |
 | CI_NODE_LABEL_SELECTOR |  |  |  | false |
 | CI_NODE_TAINTS_KEY |  |  |  | false |
 | CI_NODE_TAINTS_VALUE |  |  |  | false |
 | CI_RUNNER_DOCKER_MTU_VALUE | -1 |  |  | false |
 | CI_SUCCESS_AUTO_TRIGGER_BATCH_SIZE | 1 |  |  | false |
 | CI_VOLUME_MOUNTS_JSON |  |  |  | false |
 | CI_WORKFLOW_EXECUTOR_TYPE | AWF |  |  | false |
 | DEFAULT_ARTIFACT_KEY_LOCATION | arsenal-v1/ci-artifacts |  |  | false |
 | DEFAULT_BUILD_LOGS_BUCKET | devtron-pro-ci-logs |  |  | false |
 | DEFAULT_BUILD_LOGS_KEY_PREFIX | arsenal-v1 |  |  | false |
 | DEFAULT_CACHE_BUCKET | ci-caching |  |  | false |
 | DEFAULT_CACHE_BUCKET_REGION | us-east-2 |  |  | false |
 | DEFAULT_CD_ARTIFACT_KEY_LOCATION |  |  |  | false |
 | DEFAULT_CD_LOGS_BUCKET_REGION | us-east-2 |  |  | false |
 | DEFAULT_CD_NAMESPACE |  |  |  | false |
 | DEFAULT_CD_TIMEOUT | 3600 |  |  | false |
 | DEFAULT_CI_IMAGE | 686244538589.dkr.ecr.us-east-2.amazonaws.com/cirunner:47 |  |  | false |
 | DEFAULT_NAMESPACE | devtron-ci |  |  | false |
 | DEFAULT_TARGET_PLATFORM |  |  |  | false |
 | DOCKER_BUILD_CACHE_PATH | /var/lib/docker |  |  | false |
 | ENABLE_BUILD_CONTEXT | false |  |  | false |
 | EXTERNAL_BLOB_STORAGE_CM_NAME | blob-storage-cm |  |  | false |
 | EXTERNAL_BLOB_STORAGE_SECRET_NAME | blob-storage-secret |  |  | false |
 | EXTERNAL_CD_NODE_LABEL_SELECTOR |  |  |  | false |
 | EXTERNAL_CD_NODE_TAINTS_KEY | dedicated |  |  | false |
 | EXTERNAL_CD_NODE_TAINTS_VALUE | ci |  |  | false |
 | EXTERNAL_CI_API_SECRET | devtroncd-secret |  |  | false |
 | EXTERNAL_CI_PAYLOAD | {"ciProjectDetails":[{"gitRepository":"https://github.com/vikram1601/getting-started-nodejs.git","checkoutPath":"./abc","commitHash":"239077135f8cdeeccb7857e2851348f558cb53d3","commitTime":"2022-10-30T20:00:00","branch":"master","message":"Update README.md","author":"User Name "}],"dockerImage":"445808685819.dkr.ecr.us-east-2.amazonaws.com/orch:23907713-2"} |  |  | false |
 | EXTERNAL_CI_WEB_HOOK_URL |  |  |  | false |
 | IGNORE_CM_CS_IN_CI_JOB | false |  |  | false |
 | IMAGE_RETRY_COUNT | 0 |  |  | false |
 | IMAGE_RETRY_INTERVAL | 5 |  |  | false |
 | IMAGE_SCANNER_ENDPOINT | http://image-scanner-new-demo-devtroncd-service.devtroncd:80 |  |  | false |
 | IMAGE_SCAN_MAX_RETRIES | 3 |  |  | false |
 | IMAGE_SCAN_RETRY_DELAY | 5 |  |  | false |
 | IN_APP_LOGGING_ENABLED | false |  |  | false |
 | MAX_CD_WORKFLOW_RUNNER_RETRIES | 0 |  |  | false |
 | MAX_CI_WORKFLOW_RETRIES | 0 |  |  | false |
 | MODE | DEV |  |  | false |
 | NATS_SERVER_HOST | nats://devtron-nats.devtroncd:4222 |  |  | false |
 | ORCH_HOST | http://devtroncd-orchestrator-service-prod.devtroncd/webhook/msg/nats |  |  | false |
 | ORCH_TOKEN |  |  |  | false |
 | PRE_CI_CACHE_PATH | /devtroncd-cache |  |  | false |
 | SHOW_DOCKER_BUILD_ARGS | true |  |  | false |
 | SKIP_CI_JOB_BUILD_CACHE_PUSH_PULL | false |  |  | false |
 | SKIP_CREATING_ECR_REPO | false |  |  | false |
 | TERMINATION_GRACE_PERIOD_SECS | 180 |  |  | false |
 | USE_ARTIFACT_LISTING_QUERY_V2 | true |  |  | false |
 | USE_BLOB_STORAGE_CONFIG_IN_CD_WORKFLOW | true |  |  | false |
 | USE_BLOB_STORAGE_CONFIG_IN_CI_WORKFLOW | true |  |  | false |
 | USE_BUILDX | false |  |  | false |
 | USE_DOCKER_API_TO_GET_DIGEST | false |  |  | false |
 | USE_EXTERNAL_NODE | false |  |  | false |
 | USE_IMAGE_TAG_FROM_GIT_PROVIDER_FOR_TAG_BASED_BUILD | false |  |  | false |
 | WF_CONTROLLER_INSTANCE_ID | devtron-runner |  |  | false |
 | WORKFLOW_CACHE_CONFIG | {} |  |  | false |
 | WORKFLOW_SERVICE_ACCOUNT | ci-runner |  |  | false |


## DEVTRON Related Environment Variables
| Key   | Value        | Description       | Possible values       | Deprecated       |
|-------|--------------|-------------------|-----------------------|------------------|
 | APP_SYNC_IMAGE | quay.io/devtron/chart-sync:1227622d-132-3775 |  |  | false |
 | APP_SYNC_JOB_RESOURCES_OBJ |  |  |  | false |
 | APP_SYNC_SERVICE_ACCOUNT | chart-sync |  |  | false |
 | ARGO_AUTO_SYNC_ENABLED | true |  |  | false |
 | ARGO_GIT_COMMIT_RETRY_COUNT_ON_CONFLICT | 3 |  |  | false |
 | ARGO_GIT_COMMIT_RETRY_DELAY_ON_CONFLICT | 1 |  |  | false |
 | ARGO_REPO_REGISTER_RETRY_COUNT | 3 |  |  | false |
 | ARGO_REPO_REGISTER_RETRY_DELAY | 10 |  |  | false |
 | ASYNC_BUILDX_CACHE_EXPORT | false |  |  | false |
 | BATCH_SIZE | 5 |  |  | false |
 | BLOB_STORAGE_ENABLED | false |  |  | false |
 | BUILDX_CACHE_MODE_MIN | false |  |  | false |
 | CD_HOST | localhost |  |  | false |
 | CD_NAMESPACE | devtroncd |  |  | false |
 | CD_PORT | 8000 |  |  | false |
 | CExpirationTime | 600 |  |  | false |
 | CI_TRIGGER_CRON_TIME | 2 |  |  | false |
 | CI_WORKFLOW_STATUS_UPDATE_CRON | */5 * * * * |  |  | false |
 | CLI_CMD_TIMEOUT_GLOBAL_SECONDS | 0 |  |  | false |
 | CLUSTER_STATUS_CRON_TIME | 15 |  |  | false |
 | CONSUMER_CONFIG_JSON |  |  |  | false |
 | DEFAULT_LOG_TIME_LIMIT | 1 |  |  | false |
 | DEFAULT_TIMEOUT | 3600 |  |  | false |
 | DEVTRON_BOM_URL | https://raw.githubusercontent.com/devtron-labs/devtron/%s/charts/devtron/devtron-bom.yaml |  |  | false |
 | DEVTRON_DEFAULT_NAMESPACE | devtroncd |  |  | false |
 | DEVTRON_DEX_SECRET_NAMESPACE | devtroncd |  |  | false |
 | DEVTRON_HELM_RELEASE_CHART_NAME | devtron-operator |  |  | false |
 | DEVTRON_HELM_RELEASE_NAME | devtron |  |  | false |
 | DEVTRON_HELM_RELEASE_NAMESPACE | devtroncd |  |  | false |
 | DEVTRON_HELM_REPO_NAME | devtron |  |  | false |
 | DEVTRON_HELM_REPO_URL | https://helm.devtron.ai |  |  | false |
 | DEVTRON_INSTALLATION_TYPE |  |  |  | false |
 | DEVTRON_MODULES_IDENTIFIER_IN_HELM_VALUES | installer.modules |  |  | false |
 | DEVTRON_SECRET_NAME | devtron-secret |  |  | false |
 | DEVTRON_VERSION_IDENTIFIER_IN_HELM_VALUES | installer.release |  |  | false |
 | DEX_CID | example-app |  |  | false |
 | DEX_CLIENT_ID | argo-cd |  |  | false |
 | DEX_CSTOREKEY |  |  |  | false |
 | DEX_JWTKEY |  |  |  | false |
 | DEX_RURL | http://127.0.0.1:8080/callback |  |  | false |
 | DEX_SECRET |  |  |  | false |
 | DEX_URL |  |  |  | false |
 | ECR_REPO_NAME_PREFIX | test/ |  |  | false |
 | ENABLE_ASYNC_ARGO_CD_INSTALL_DEVTRON_CHART | false |  |  | false |
 | ENABLE_ASYNC_INSTALL_DEVTRON_CHART | false |  |  | false |
 | EPHEMERAL_SERVER_VERSION_REGEX | v[1-9]\.\b(2[3-9]\|[3-9][0-9])\b.* |  |  | false |
 | EVENT_URL | http://localhost:3000/notify |  |  | false |
 | EXECUTE_WIRE_NIL_CHECKER | false |  |  | false |
 | EXPOSE_CI_METRICS | false |  |  | false |
 | FEATURE_RESTART_WORKLOAD_BATCH_SIZE | 1 |  |  | false |
 | FEATURE_RESTART_WORKLOAD_WORKER_POOL_SIZE | 5 |  |  | false |
 | FORCE_SECURITY_SCANNING | false |  |  | false |
 | GITOPS_REPO_PREFIX |  |  |  | false |
 | GRAFANA_HOST | localhost |  |  | false |
 | GRAFANA_NAMESPACE | devtroncd |  |  | false |
 | GRAFANA_ORG_ID | 2 |  |  | false |
 | GRAFANA_PASSWORD | prom-operator |  |  | false |
 | GRAFANA_PORT | 8090 |  |  | false |
 | GRAFANA_URL |  |  |  | false |
 | GRAFANA_USERNAME | admin |  |  | false |
 | HIDE_IMAGE_TAGGING_HARD_DELETE | false |  |  | false |
 | IGNORE_AUTOCOMPLETE_AUTH_CHECK | false |  |  | false |
 | INSTALLER_CRD_NAMESPACE | devtroncd |  |  | false |
 | INSTALLER_CRD_OBJECT_GROUP_NAME | installer.devtron.ai |  |  | false |
 | INSTALLER_CRD_OBJECT_RESOURCE | installers |  |  | false |
 | INSTALLER_CRD_OBJECT_VERSION | v1alpha1 |  |  | false |
 | IS_INTERNAL_USE | false |  |  | false |
 | JwtExpirationTime | 120 |  |  | false |
 | K8s_CLIENT_MAX_IDLE_CONNS_PER_HOST | 25 |  |  | false |
 | K8s_TCP_IDLE_CONN_TIMEOUT | 300 |  |  | false |
 | K8s_TCP_KEEPALIVE | 30 |  |  | false |
 | K8s_TCP_TIMEOUT | 30 |  |  | false |
 | K8s_TLS_HANDSHAKE_TIMEOUT | 10 |  |  | false |
 | KUBELINK_GRPC_MAX_RECEIVE_MSG_SIZE | 20 |  |  | false |
 | KUBELINK_GRPC_MAX_SEND_MSG_SIZE | 4 |  |  | false |
 | LENS_TIMEOUT | 0 |  |  | false |
 | LENS_URL | http://lens-milandevtron-service:80 |  |  | false |
 | LIMIT_CI_CPU | 0.5 |  |  | false |
 | LIMIT_CI_MEM | 3G |  |  | false |
 | LOGGER_DEV_MODE | false |  |  | false |
 | LOG_LEVEL | 0 |  |  | false |
 | MAX_SESSION_PER_USER | 5 |  |  | false |
 | MODULE_METADATA_API_URL | https://api.devtron.ai/module?name=%s |  |  | false |
 | MODULE_STATUS_HANDLING_CRON_DURATION_MIN | 3 |  |  | false |
 | NATS_MSG_ACK_WAIT_IN_SECS | 120 |  |  | false |
 | NATS_MSG_BUFFER_SIZE | -1 |  |  | false |
 | NATS_MSG_MAX_AGE | 86400 |  |  | false |
 | NATS_MSG_PROCESSING_BATCH_SIZE | 1 |  |  | false |
 | NATS_MSG_REPLICAS | 0 |  |  | false |
 | NOTIFICATION_MEDIUM | rest |  |  | false |
 | OTEL_COLLECTOR_URL |  |  |  | false |
 | PARALLELISM_LIMIT_FOR_TAG_PROCESSING |  |  |  | false |
 | PG_EXPORT_PROM_METRICS | true |  |  | false |
 | PG_LOG_ALL_FAILURE_QUERIES | true |  |  | false |
 | PG_LOG_ALL_QUERY | false |  |  | false |
 | PG_LOG_SLOW_QUERY | true |  |  | false |
 | PG_QUERY_DUR_THRESHOLD | 5000 |  |  | false |
 | PLUGIN_NAME | Pull images from container repository |  |  | false |
 | PROPAGATE_EXTRA_LABELS | false |  |  | false |
 | PROXY_SERVICE_CONFIG | {} |  |  | false |
 | REQ_CI_CPU | 0.5 |  |  | false |
 | REQ_CI_MEM | 3G |  |  | false |
 | RESTRICT_TERMINAL_ACCESS_FOR_NON_SUPER_USER | false |  |  | false |
 | RUNTIME_CONFIG_LOCAL_DEV | false |  |  | false |
 | RUN_HELM_INSTALL_IN_ASYNC_MODE_HELM_APPS | false |  |  | false |
 | SCOPED_VARIABLE_ENABLED | false |  |  | false |
 | SCOPED_VARIABLE_FORMAT | @{{%s}} |  |  | false |
 | SCOPED_VARIABLE_HANDLE_PRIMITIVES | false |  |  | false |
 | SCOPED_VARIABLE_NAME_REGEX | ^[a-zA-Z][a-zA-Z0-9_-]{0,62}[a-zA-Z0-9]$ |  |  | false |
 | SOCKET_DISCONNECT_DELAY_SECONDS | 5 |  |  | false |
 | SOCKET_HEARTBEAT_SECONDS | 25 |  |  | false |
 | STREAM_CONFIG_JSON |  |  |  | false |
 | SYSTEM_VAR_PREFIX | DEVTRON_ |  |  | false |
 | TERMINAL_POD_DEFAULT_NAMESPACE | default |  |  | false |
 | TERMINAL_POD_INACTIVE_DURATION_IN_MINS | 10 |  |  | false |
 | TERMINAL_POD_STATUS_SYNC_In_SECS | 600 |  |  | false |
 | TEST_APP | orchestrator |  |  | false |
 | TEST_PG_ADDR | 127.0.0.1 |  |  | false |
 | TEST_PG_DATABASE | orchestrator |  |  | false |
 | TEST_PG_LOG_QUERY | true |  |  | false |
 | TEST_PG_PASSWORD | postgrespw |  |  | false |
 | TEST_PG_PORT | 55000 |  |  | false |
 | TEST_PG_USER | postgres |  |  | false |
 | TIMEOUT_FOR_FAILED_CI_BUILD | 15 |  |  | false |
 | TIMEOUT_IN_SECONDS | 5 |  |  | false |
 | USER_SESSION_DURATION_SECONDS | 86400 |  |  | false |
 | USE_ARTIFACT_LISTING_API_V2 | true |  |  | false |
 | USE_CUSTOM_HTTP_TRANSPORT | false |  |  | false |
 | USE_DEPLOYMENT_CONFIG_DATA | false |  |  | false |
 | USE_GIT_CLI | false |  |  | false |
 | USE_RBAC_CREATION_V2 | true |  |  | false |
 | VARIABLE_CACHE_ENABLED | true |  |  | false |
 | VARIABLE_EXPRESSION_REGEX | @{{([^}]+)}} |  |  | false |
 | WEBHOOK_TOKEN |  |  |  | false |


## GITOPS Related Environment Variables
| Key   | Value        | Description       | Possible values       | Deprecated       |
|-------|--------------|-------------------|-----------------------|------------------|
 | ACD_CM | argocd-cm |  |  | false |
 | ACD_NAMESPACE | devtroncd |  |  | false |
 | ACD_PASSWORD |  |  |  | false |
 | ACD_USERNAME | admin |  |  | false |
 | GITOPS_SECRET_NAME | devtron-gitops-secret |  |  | false |
 | RESOURCE_LIST_FOR_REPLICAS | Deployment,Rollout,StatefulSet,ReplicaSet |  |  | false |
 | RESOURCE_LIST_FOR_REPLICAS_BATCH_SIZE | 5 |  |  | false |


## INFRA_SETUP Related Environment Variables
| Key   | Value        | Description       | Possible values       | Deprecated       |
|-------|--------------|-------------------|-----------------------|------------------|
 | DASHBOARD_HOST | localhost |  |  | false |
 | DASHBOARD_NAMESPACE | devtroncd |  |  | false |
 | DASHBOARD_PORT | 3000 |  |  | false |
 | DEX_HOST | http://localhost |  |  | false |
 | DEX_PORT | 5556 |  |  | false |
 | GIT_SENSOR_PROTOCOL | REST |  |  | false |
 | GIT_SENSOR_TIMEOUT | 0 |  |  | false |
 | GIT_SENSOR_URL | 127.0.0.1:7070 |  |  | false |
 | HELM_CLIENT_URL | 127.0.0.1:50051 |  |  | false |


## POSTGRES Related Environment Variables
| Key   | Value        | Description       | Possible values       | Deprecated       |
|-------|--------------|-------------------|-----------------------|------------------|
 | APP | orchestrator | Application name |  | false |
 | CASBIN_DATABASE | casbin |  |  | false |
 | PG_ADDR | 127.0.0.1 | address of postgres service | postgresql-postgresql.devtroncd | false |
 | PG_DATABASE | orchestrator | postgres database to be made connection with | orchestrator, casbin, git_sensor, lens | false |
 | PG_PASSWORD |  | password for postgres, associated with PG_USER | confidential ;) | false |
 | PG_PORT | 5432 | port of postgresql service | 5432 | false |
 | PG_READ_TIMEOUT | 30 |  |  | false |
 | PG_USER |  | user for postgres | postgres | false |
 | PG_WRITE_TIMEOUT | 30 |  |  | false |


## RBAC Related Environment Variables
| Key   | Value        | Description       | Possible values       | Deprecated       |
|-------|--------------|-------------------|-----------------------|------------------|
 | ENFORCER_CACHE | false |  |  | false |
 | ENFORCER_CACHE_EXPIRATION_IN_SEC | 86400 |  |  | false |
 | ENFORCER_MAX_BATCH_SIZE | 1 |  |  | false |
 | USE_CASBIN_V2 | false |  |  | false |


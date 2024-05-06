
## Devtron Environment Variables
| Key   | Value        | Description       |
|-------|--------------|-------------------|
 | ACD_CM | argocd-cm |  | 
 | ACD_NAMESPACE | devtroncd |  | 
 | ACD_PASSWORD |  |  | 
 | ACD_USERNAME | admin |  | 
 | APP | orchestrator |  | 
 | APP_SYNC_IMAGE | quay.io/devtron/chart-sync:1227622d-132-3775 |  | 
 | APP_SYNC_JOB_RESOURCES_OBJ |  |  | 
 | ARGO_APP_MANUAL_SYNC_TIME | 3 |  | 
 | ARGO_AUTO_SYNC_ENABLED | true |  | 
 | AZURE_ACCOUNT_KEY |  |  | 
 | AZURE_ACCOUNT_NAME |  |  | 
 | AZURE_BLOB_CONTAINER_CI_CACHE |  |  | 
 | AZURE_BLOB_CONTAINER_CI_LOG |  |  | 
 | AZURE_GATEWAY_CONNECTION_INSECURE | true |  | 
 | AZURE_GATEWAY_URL | http://devtron-minio.devtroncd:9000 |  | 
 | BASE_LOG_LOCATION_PATH | /home/devtron/ |  | 
 | BATCH_SIZE | 5 |  | 
 | BLOB_STORAGE_ENABLED | false |  | 
 | BLOB_STORAGE_GCP_CREDENTIALS_JSON |  |  | 
 | BLOB_STORAGE_PROVIDER | S3 |  | 
 | BLOB_STORAGE_S3_ACCESS_KEY |  |  | 
 | BLOB_STORAGE_S3_BUCKET_VERSIONED | true |  | 
 | BLOB_STORAGE_S3_ENDPOINT |  |  | 
 | BLOB_STORAGE_S3_ENDPOINT_INSECURE | false |  | 
 | BLOB_STORAGE_S3_SECRET_KEY |  |  | 
 | BUILDX_CACHE_PATH | /var/lib/devtron/buildx |  | 
 | BUILDX_K8S_DRIVER_OPTIONS |  |  | 
 | BUILDX_PROVENANCE_MODE |  |  | 
 | BUILD_LOG_TTL_VALUE_IN_SECS | 3600 |  | 
 | CACHED_GVKs | [] |  | 
 | CACHED_NAMESPACES | gireesh-ns |  | 
 | CACHE_LIMIT | 5000000000 |  | 
 | CAN_APPROVER_DEPLOY | false |  | 
 | CASBIN_CLIENT_URL | 127.0.0.1:9000 |  | 
 | CASBIN_DATABASE | casbin |  | 
 | CASBIN_GRPC_DATA_TRANSFER_MAX_SIZE | 30 |  | 
 | CD_ARTIFACT_LOCATION_FORMAT | %d/%d.zip |  | 
 | CD_DEFAULT_ADDRESS_POOL_BASE_CIDR |  |  | 
 | CD_DEFAULT_ADDRESS_POOL_SIZE |  |  | 
 | CD_HELM_PIPELINE_STATUS_CRON_TIME | */2 * * * * |  | 
 | CD_HOST | localhost |  | 
 | CD_LIMIT_CI_CPU | 0.5 |  | 
 | CD_LIMIT_CI_MEM | 3G |  | 
 | CD_NAMESPACE | devtroncd |  | 
 | CD_NODE_LABEL_SELECTOR |  |  | 
 | CD_NODE_TAINTS_KEY | dedicated |  | 
 | CD_NODE_TAINTS_VALUE | ci |  | 
 | CD_PIPELINE_STATUS_CRON_TIME | */2 * * * * |  | 
 | CD_PIPELINE_STATUS_TIMEOUT_DURATION | 20 |  | 
 | CD_PORT | 8000 |  | 
 | CD_REQ_CI_CPU | 0.5 |  | 
 | CD_REQ_CI_MEM | 3G |  | 
 | CD_WORKFLOW_EXECUTOR_TYPE | AWF |  | 
 | CD_WORKFLOW_SERVICE_ACCOUNT | cd-runner |  | 
 | CExpirationTime | 600 |  | 
 | CI_ARTIFACT_LOCATION_FORMAT | %d/%d.zip |  | 
 | CI_DEFAULT_ADDRESS_POOL_BASE_CIDR |  |  | 
 | CI_DEFAULT_ADDRESS_POOL_SIZE |  |  | 
 | CI_IGNORE_DOCKER_CACHE |  |  | 
 | CI_LOGS_KEY_PREFIX |  |  | 
 | CI_NODE_LABEL_SELECTOR |  |  | 
 | CI_NODE_TAINTS_KEY |  |  | 
 | CI_NODE_TAINTS_VALUE |  |  | 
 | CI_RUNNER_DOCKER_MTU_VALUE | -1 |  | 
 | CI_SUCCESS_AUTO_TRIGGER_BATCH_SIZE | 1 |  | 
 | CI_TRIGGER_CRON_TIME | 2 |  | 
 | CI_VOLUME_MOUNTS_JSON |  |  | 
 | CI_WORKFLOW_EXECUTOR_TYPE | AWF |  | 
 | CI_WORKFLOW_STATUS_UPDATE_CRON | */5 * * * * |  | 
 | CLEAN_UP_RBAC_POLICIES | false |  | 
 | CLEAN_UP_RBAC_POLICIES_CRON_TIME | 0 0 * * * |  | 
 | CLI_CMD_TIMEOUT_GLOBAL_SECONDS | 0 |  | 
 | CLONING_MODE | SHALLOW |  | 
 | CLUSTER_CACHE_ATTEMPT_LIMIT | 1 |  | 
 | CLUSTER_CACHE_LIST_PAGE_BUFFER_SIZE | 10 |  | 
 | CLUSTER_CACHE_LIST_PAGE_SIZE | 500 |  | 
 | CLUSTER_CACHE_LIST_SEMAPHORE_SIZE | 5 |  | 
 | CLUSTER_CACHE_RESYNC_DURATION | 12h |  | 
 | CLUSTER_CACHE_RETRY_USE_BACKOFF |  |  | 
 | CLUSTER_CACHE_WATCH_RESYNC_DURATION | 10m |  | 
 | CLUSTER_STATUS_CRON_TIME | 15 |  | 
 | CLUSTER_SYNC_RETRY_TIMEOUT_DURATION | 10s |  | 
 | CONSUMER_CONFIG_JSON |  |  | 
 | CUSTOM_ROLE_CACHE_ALLOWED | false |  | 
 | DASHBOARD_HOST | localhost |  | 
 | DASHBOARD_NAMESPACE | devtroncd |  | 
 | DASHBOARD_PORT | 3000 |  | 
 | DEFAULT_ARTIFACT_KEY_LOCATION | arsenal-v1/ci-artifacts |  | 
 | DEFAULT_BUILD_LOGS_BUCKET | devtron-pro-ci-logs |  | 
 | DEFAULT_BUILD_LOGS_KEY_PREFIX | arsenal-v1 |  | 
 | DEFAULT_CACHE_BUCKET | ci-caching |  | 
 | DEFAULT_CACHE_BUCKET_REGION | us-east-2 |  | 
 | DEFAULT_CD_ARTIFACT_KEY_LOCATION |  |  | 
 | DEFAULT_CD_LOGS_BUCKET_REGION | us-east-2 |  | 
 | DEFAULT_CD_NAMESPACE |  |  | 
 | DEFAULT_CD_TIMEOUT | 3600 |  | 
 | DEFAULT_CI_IMAGE | 686244538589.dkr.ecr.us-east-2.amazonaws.com/cirunner:47 |  | 
 | DEFAULT_LOG_TIME_LIMIT | 1 |  | 
 | DEFAULT_NAMESPACE | devtron-ci |  | 
 | DEFAULT_TARGET_PLATFORM |  |  | 
 | DEFAULT_TIMEOUT | 3600 |  | 
 | DEPLOYMENT_WINDOW_FETCH_DAYS_BLACKOUT | 90 |  | 
 | DEPLOYMENT_WINDOW_FETCH_DAYS_MAINTENANCE | 90 |  | 
 | DEPLOY_STATUS_CRON_GET_PIPELINE_DEPLOYED_WITHIN_HOURS | 12 |  | 
 | DEVTRON_BOM_URL | https://raw.githubusercontent.com/devtron-labs/devtron/%s/charts/devtron/devtron-bom.yaml |  | 
 | DEVTRON_CHART_INSTALL_REQUEST_TIMEOUT | 6 |  | 
 | DEVTRON_DEFAULT_NAMESPACE | devtroncd |  | 
 | DEVTRON_DEX_SECRET_NAMESPACE | devtroncd |  | 
 | DEVTRON_HELM_RELEASE_CHART_NAME | devtron-operator |  | 
 | DEVTRON_HELM_RELEASE_NAME | devtron |  | 
 | DEVTRON_HELM_RELEASE_NAMESPACE | devtroncd |  | 
 | DEVTRON_HELM_REPO_NAME | devtron |  | 
 | DEVTRON_HELM_REPO_URL | https://helm.devtron.ai |  | 
 | DEVTRON_INSTALLATION_TYPE |  |  | 
 | DEVTRON_MODULES_IDENTIFIER_IN_HELM_VALUES | installer.modules |  | 
 | DEVTRON_SECRET_NAME | devtron-secret |  | 
 | DEVTRON_VERSION_IDENTIFIER_IN_HELM_VALUES | installer.release |  | 
 | DEX_CID | example-app |  | 
 | DEX_CLIENT_ID | argo-cd |  | 
 | DEX_CSTOREKEY |  |  | 
 | DEX_HOST | http://localhost |  | 
 | DEX_JWTKEY |  |  | 
 | DEX_PORT | 5556 |  | 
 | DEX_RURL | http://127.0.0.1:8080/callback |  | 
 | DEX_SECRET |  |  | 
 | DEX_URL |  |  | 
 | DOCKER_BUILD_CACHE_PATH | /var/lib/docker |  | 
 | ECR_REPO_NAME_PREFIX | test/ |  | 
 | ENABLE_ASYNC_INSTALL_DEVTRON_CHART | false |  | 
 | ENABLE_BUILD_CONTEXT | false |  | 
 | ENFORCER_CACHE | false |  | 
 | ENFORCER_CACHE_EXPIRATION_IN_SEC | 86400 |  | 
 | ENFORCER_MAX_BATCH_SIZE | 1 |  | 
 | ENTERPRISE_ENFORCER_ENABLED | true |  | 
 | EPHEMERAL_SERVER_VERSION_REGEX | v[1-9]\.\b(2[3-9]|[3-9][0-9])\b.* |  | 
 | EVENT_URL | http://localhost:3000/notify |  | 
 | EXPOSE_CD_METRICS | false |  | 
 | EXPOSE_CI_METRICS | false |  | 
 | EXTERNAL_BLOB_STORAGE_CM_NAME | blob-storage-cm |  | 
 | EXTERNAL_BLOB_STORAGE_SECRET_NAME | blob-storage-secret |  | 
 | EXTERNAL_CD_NODE_LABEL_SELECTOR |  |  | 
 | EXTERNAL_CD_NODE_TAINTS_KEY | dedicated |  | 
 | EXTERNAL_CD_NODE_TAINTS_VALUE | ci |  | 
 | EXTERNAL_CI_API_SECRET | devtroncd-secret |  | 
 | EXTERNAL_CI_PAYLOAD | {"ciProjectDetails":[{"gitRepository":"https://github.com/vikram1601/getting-started-nodejs.git","checkoutPath":"./abc","commitHash":"239077135f8cdeeccb7857e2851348f558cb53d3","commitTime":"2022-10-30T20:00:00","branch":"master","message":"Update README.md","author":"User Name "}],"dockerImage":"445808685819.dkr.ecr.us-east-2.amazonaws.com/orch:23907713-2"} |  | 
 | EXTERNAL_CI_WEB_HOOK_URL |  |  | 
 | FORCE_SECURITY_SCANNING | false |  | 
 | GITOPS_REPO_PREFIX |  |  | 
 | GITOPS_SECRET_NAME | devtron-gitops-secret |  | 
 | GIT_PROVIDERS | github,gitlab |  | 
 | GIT_SENSOR_GRPC_DATA_TRANSFER_MAX_SIZE | 4 |  | 
 | GIT_SENSOR_PROTOCOL | REST |  | 
 | GIT_SENSOR_TIMEOUT | 0 |  | 
 | GIT_SENSOR_URL | 127.0.0.1:7070 |  | 
 | GRAFANA_HOST | localhost |  | 
 | GRAFANA_NAMESPACE | devtroncd |  | 
 | GRAFANA_ORG_ID | 2 |  | 
 | GRAFANA_PASSWORD | prom-operator |  | 
 | GRAFANA_PORT | 8090 |  | 
 | GRAFANA_URL |  |  | 
 | GRAFANA_USERNAME | admin |  | 
 | HELM_CLIENT_URL | 127.0.0.1:50051 |  | 
 | HELM_PIPELINE_STATUS_CHECK_ELIGIBLE_TIME | 120 |  | 
 | HIDE_IMAGE_TAGGING_HARD_DELETE | false |  | 
 | IGNORE_AUTOCOMPLETE_AUTH_CHECK | false |  | 
 | IMAGE_RETRY_COUNT | 0 |  | 
 | IMAGE_RETRY_INTERVAL | 5 |  | 
 | IMAGE_SCANNER_ENDPOINT | http://image-scanner-new-demo-devtroncd-service.devtroncd:80 |  | 
 | IMAGE_SCAN_MAX_RETRIES | 3 |  | 
 | IMAGE_SCAN_RETRY_DELAY | 5 |  | 
 | INSTALLER_CRD_NAMESPACE | devtroncd |  | 
 | INSTALLER_CRD_OBJECT_GROUP_NAME | installer.devtron.ai |  | 
 | INSTALLER_CRD_OBJECT_RESOURCE | installers |  | 
 | INSTALLER_CRD_OBJECT_VERSION | v1alpha1 |  | 
 | IN_APP_LOGGING_ENABLED | false |  | 
 | IS_AIR_GAP_ENVIRONMENT | false |  | 
 | IS_INTERNAL_USE | false |  | 
 | JwtExpirationTime | 120 |  | 
 | LENS_TIMEOUT | 0 |  | 
 | LENS_URL | http://lens-milandevtron-service:80 |  | 
 | LIMIT_CI_CPU | 0.5 |  | 
 | LIMIT_CI_MEM | 3G |  | 
 | LOGGER_DEV_MODE | false |  | 
 | LOG_LEVEL | 0 |  | 
 | MAX_CD_WORKFLOW_RUNNER_RETRIES | 0 |  | 
 | MAX_CI_WORKFLOW_RETRIES | 0 |  | 
 | MAX_SESSION_PER_USER | 5 |  | 
 | MODE | DEV |  | 
 | MODULE_METADATA_API_URL | https://api.devtron.ai/module?name=%s |  | 
 | MODULE_STATUS_HANDLING_CRON_DURATION_MIN | 3 |  | 
 | NATS_MSG_ACK_WAIT_IN_SECS | 120 |  | 
 | NATS_MSG_BUFFER_SIZE | -1 |  | 
 | NATS_MSG_MAX_AGE | 86400 |  | 
 | NATS_MSG_PROCESSING_BATCH_SIZE | 1 |  | 
 | NATS_SERVER_HOST | nats://devtron-nats.devtroncd:4222 |  | 
 | NOTIFICATION_TOKEN_EXPIRY_TIME_HOURS | 720 |  | 
 | ORCH_HOST | http://devtroncd-orchestrator-service-prod.devtroncd/webhook/msg/nats |  | 
 | ORCH_TOKEN |  |  | 
 | OTEL_COLLECTOR_URL |  |  | 
 | PG_ADDR | 127.0.0.1 |  | 
 | PG_DATABASE | orchestrator |  | 
 | PG_EXPORT_PROM_METRICS | false |  | 
 | PG_LOG_ALL_QUERY | false |  | 
 | PG_LOG_QUERY | true |  | 
 | PG_PASSWORD |  |  | 
 | PG_PORT | 5432 |  | 
 | PG_QUERY_DUR_THRESHOLD | 5000 |  | 
 | PG_READ_TIMEOUT | 30 |  | 
 | PG_USER |  |  | 
 | PG_WRITE_TIMEOUT | 30 |  | 
 | PIPELINE_DEGRADED_TIME | 10 |  | 
 | PLUGIN_NAME | Pull images from container repository |  | 
 | PRE_CI_CACHE_PATH | /devtroncd-cache |  | 
 | PROXY_SERVICE_CONFIG | {} |  | 
 | REQ_CI_CPU | 0.5 |  | 
 | REQ_CI_MEM | 3G |  | 
 | RESOURCE_LIST_FOR_REPLICAS | Deployment,Rollout,StatefulSet,ReplicaSet |  | 
 | RESOURCE_LIST_FOR_REPLICAS_BATCH_SIZE | 5 |  | 
 | REVISION_HISTORY_LIMIT_DEVTRON_APP | 1 |  | 
 | REVISION_HISTORY_LIMIT_EXTERNAL_HELM_APP | 0 |  | 
 | REVISION_HISTORY_LIMIT_HELM_APP | 1 |  | 
 | RUNTIME_CONFIG_LOCAL_DEV | false |  | 
 | RUN_HELM_INSTALL_IN_ASYNC_MODE_HELM_APPS | false |  | 
 | SCAN_V2_ENABLED | false |  | 
 | SCOOP_CLUSTER_CONFIG | {} |  | 
 | SCOPED_VARIABLE_ENABLED | false |  | 
 | SCOPED_VARIABLE_FORMAT | @{{%s}} |  | 
 | SCOPED_VARIABLE_HANDLE_PRIMITIVES | false |  | 
 | SCOPED_VARIABLE_NAME_REGEX | ^[a-zA-Z][a-zA-Z0-9_-]{0,62}[a-zA-Z0-9]$ |  | 
 | SKIP_CREATING_ECR_REPO | false |  | 
 | SOCKET_DISCONNECT_DELAY_SECONDS | 5 |  | 
 | SOCKET_HEARTBEAT_SECONDS | 25 |  | 
 | STREAM_CONFIG_JSON |  |  | 
 | SYSTEM_VAR_PREFIX | DEVTRON_ |  | 
 | TERMINAL_POD_DEFAULT_NAMESPACE | default |  | 
 | TERMINAL_POD_INACTIVE_DURATION_IN_MINS | 10 |  | 
 | TERMINAL_POD_STATUS_SYNC_In_SECS | 600 |  | 
 | TERMINATION_GRACE_PERIOD_SECS | 180 |  | 
 | TEST_APP | orchestrator |  | 
 | TEST_PG_ADDR | 127.0.0.1 |  | 
 | TEST_PG_DATABASE | orchestrator |  | 
 | TEST_PG_LOG_QUERY | true |  | 
 | TEST_PG_PASSWORD | postgrespw |  | 
 | TEST_PG_PORT | 55000 |  | 
 | TEST_PG_USER | postgres |  | 
 | TIMEOUT_FOR_FAILED_CI_BUILD | 15 |  | 
 | TIMEOUT_IN_SECONDS | 5 |  | 
 | USER_SESSION_DURATION_SECONDS | 86400 |  | 
 | USE_ARTIFACT_LISTING_API_V2 | true |  | 
 | USE_ARTIFACT_LISTING_QUERY_V2 | true |  | 
 | USE_BLOB_STORAGE_CONFIG_IN_CD_WORKFLOW | true |  | 
 | USE_BLOB_STORAGE_CONFIG_IN_CI_WORKFLOW | true |  | 
 | USE_BUILDX | false |  | 
 | USE_CASBIN_V2 | false |  | 
 | USE_CUSTOM_ENFORCER | true |  | 
 | USE_EXTERNAL_NODE | false |  | 
 | USE_GIT_CLI | false |  | 
 | USE_IMAGE_TAG_FROM_GIT_PROVIDER_FOR_TAG_BASED_BUILD | false |  | 
 | USE_RBAC_CREATION_V2 | true |  | 
 | USE_RESOURCE_LIST_V2 |  |  | 
 | VARIABLE_CACHE_ENABLED | true |  | 
 | VARIABLE_EXPRESSION_REGEX | @{{([^}]+)}} |  | 
 | WEBHOOK_TOKEN |  |  | 
 | WF_CONTROLLER_INSTANCE_ID | devtron-runner |  | 
 | WORKFLOW_SERVICE_ACCOUNT | ci-runner |  | 

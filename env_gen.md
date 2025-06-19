

## CD Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | ARGO_APP_MANUAL_SYNC_TIME | int |3 | retry argocd app manual sync if the timeline is stuck in ARGOCD_SYNC_INITIATED state for more than this defined time (in mins) |  | false |
 | CD_FLUX_PIPELINE_STATUS_CRON_TIME | string |*/2 * * * * | Cron time to check the pipeline status for flux cd pipeline |  | false |
 | CD_HELM_PIPELINE_STATUS_CRON_TIME | string |*/2 * * * * | Cron time to check the pipeline status  |  | false |
 | CD_PIPELINE_STATUS_CRON_TIME | string |*/2 * * * * | Cron time for CD pipeline status |  | false |
 | CD_PIPELINE_STATUS_TIMEOUT_DURATION | string |20 | Timeout for CD pipeline to get healthy |  | false |
 | DEPLOY_STATUS_CRON_GET_PIPELINE_DEPLOYED_WITHIN_HOURS | int |12 | This flag is used to fetch the deployment status of the application. It retrieves the status of deployments that occurred between 12 hours and 10 minutes prior to the current time. It fetches non-terminal statuses. |  | false |
 | DEVTRON_CHART_ARGO_CD_INSTALL_REQUEST_TIMEOUT | int |1 | Context timeout for gitops concurrent async deployments |  | false |
 | DEVTRON_CHART_INSTALL_REQUEST_TIMEOUT | int |6 | Context timeout for no gitops concurrent async deployments |  | false |
 | EXPOSE_CD_METRICS | bool |false |  |  | false |
 | FEATURE_MIGRATE_ARGOCD_APPLICATION_ENABLE | bool |false | enable migration of external argocd application to devtron pipeline |  | false |
 | FLUX_CD_PIPELINE_STATUS_CHECK_ELIGIBLE_TIME | string |120 | eligible time for checking flux app status periodically and update in db, value is in seconds., default is 120, if wfr is updated within configured time i.e. FLUX_CD_PIPELINE_STATUS_CHECK_ELIGIBLE_TIME then do not include for this cron cycle. |  | false |
 | HELM_PIPELINE_STATUS_CHECK_ELIGIBLE_TIME | string |120 | eligible time for checking helm app status periodically and update in db, value is in seconds., default is 120, if wfr is updated within configured time i.e. HELM_PIPELINE_STATUS_CHECK_ELIGIBLE_TIME then do not include for this cron cycle. |  | false |
 | IS_INTERNAL_USE | bool |true | If enabled then cd pipeline and helm apps will not need the deployment app type mandatorily. Couple this flag with HIDE_GITOPS_OR_HELM_OPTION (in Dashborad) and if gitops is configured and allowed for the env, pipeline/ helm app will gitops else no-gitops. |  | false |
 | MIGRATE_DEPLOYMENT_CONFIG_DATA | bool |false | migrate deployment config data from charts table to deployment_config table |  | false |
 | PIPELINE_DEGRADED_TIME | string |10 | Time to mark a pipeline degraded if not healthy in defined time |  | false |
 | REVISION_HISTORY_LIMIT_DEVTRON_APP | int |1 | Count for devtron application rivision history |  | false |
 | REVISION_HISTORY_LIMIT_EXTERNAL_HELM_APP | int |0 | Count for external helm application rivision history |  | false |
 | REVISION_HISTORY_LIMIT_HELM_APP | int |1 | To set the history limit for the helm app being deployed through devtron |  | false |
 | REVISION_HISTORY_LIMIT_LINKED_HELM_APP | int |15 |  |  | false |
 | RUN_HELM_INSTALL_IN_ASYNC_MODE_HELM_APPS | bool |false |  |  | false |
 | SHOULD_CHECK_NAMESPACE_ON_CLONE | bool |false | should we check if namespace exists or not while cloning app |  | false |
 | USE_DEPLOYMENT_CONFIG_DATA | bool |false | use deployment config data from deployment_config table |  | true |
 | VALIDATE_EXT_APP_CHART_TYPE | bool |false | validate external flux app chart |  | false |


## CI_RUNNER Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | AZURE_ACCOUNT_KEY | string | | If blob storage is being used of azure then pass the secret key to access the bucket |  | false |
 | AZURE_ACCOUNT_NAME | string | | Account name for azure blob storage |  | false |
 | AZURE_BLOB_CONTAINER_CI_CACHE | string | | Cache bucket name for azure blob storage |  | false |
 | AZURE_BLOB_CONTAINER_CI_LOG | string | | Log bucket for azure blob storage |  | false |
 | AZURE_GATEWAY_CONNECTION_INSECURE | bool |true | Azure gateway connection allows insecure if true |  | false |
 | AZURE_GATEWAY_URL | string |http://devtron-minio.devtroncd:9000 | Sent to CI runner for blob |  | false |
 | BASE_LOG_LOCATION_PATH | string |/home/devtron/ | Used to store, download logs of ci workflow, artifact |  | false |
 | BLOB_STORAGE_GCP_CREDENTIALS_JSON | string | | GCP cred json for GCS blob storage |  | false |
 | BLOB_STORAGE_PROVIDER |  |S3 | Blob storage provider name(AWS/GCP/Azure) |  | false |
 | BLOB_STORAGE_S3_ACCESS_KEY | string | | S3 access key for s3 blob storage |  | false |
 | BLOB_STORAGE_S3_BUCKET_VERSIONED | bool |true | To enable buctet versioning for blob storage |  | false |
 | BLOB_STORAGE_S3_ENDPOINT | string | | S3 endpoint URL for s3 blob storage |  | false |
 | BLOB_STORAGE_S3_ENDPOINT_INSECURE | bool |false | To use insecure s3 endpoint |  | false |
 | BLOB_STORAGE_S3_SECRET_KEY | string | | Secret key for s3 blob storage |  | false |
 | BUILDX_CACHE_PATH | string |/var/lib/devtron/buildx | Path for the buildx cache |  | false |
 | BUILDX_K8S_DRIVER_OPTIONS | string | | To enable the k8s driver and pass args for k8s driver in buildx |  | false |
 | BUILDX_PROVENANCE_MODE | string | | provinance is set to true by default by docker. this will add some build related data in generated build manifest.it also adds some unknown:unknown key:value pair which may not be compatible by some container registries. with buildx k8s driver , provinenance=true is causing issue when push manifest to quay registry, so setting it to false |  | false |
 | BUILD_LOG_TTL_VALUE_IN_SECS | int |3600 | This is the time that the pods of ci/pre-cd/post-cd live after completion state. |  | false |
 | CACHE_LIMIT | int64 |5000000000 | Cache limit. |  | false |
 | CD_DEFAULT_ADDRESS_POOL_BASE_CIDR | string | | To pass the IP cidr for Pre/Post cd  |  | false |
 | CD_DEFAULT_ADDRESS_POOL_SIZE | int | | The subnet size to allocate from the base pool for CD |  | false |
 | CD_LIMIT_CI_CPU | string |0.5 | CPU Resource Limit Pre/Post CD |  | false |
 | CD_LIMIT_CI_MEM | string |3G | Memory Resource Limit Pre/Post CD |  | false |
 | CD_NODE_LABEL_SELECTOR |  | | Node label selector for  Pre/Post CD |  | false |
 | CD_NODE_TAINTS_KEY | string |dedicated | Toleration key for Pre/Post CD |  | false |
 | CD_NODE_TAINTS_VALUE | string |ci | Toleration value for Pre/Post CD |  | false |
 | CD_REQ_CI_CPU | string |0.5 | CPU Resource Rquest Pre/Post CD |  | false |
 | CD_REQ_CI_MEM | string |3G | Memory Resource Rquest Pre/Post CD |  | false |
 | CD_WORKFLOW_EXECUTOR_TYPE |  |AWF | Executor type for Pre/Post CD(AWF,System) |  | false |
 | CD_WORKFLOW_SERVICE_ACCOUNT | string |cd-runner | Service account to be used in Pre/Post CD pod |  | false |
 | CI_DEFAULT_ADDRESS_POOL_BASE_CIDR | string | | To pass the IP cidr for CI |  | false |
 | CI_DEFAULT_ADDRESS_POOL_SIZE | int | | The subnet size to allocate from the base pool for CI |  | false |
 | CI_IGNORE_DOCKER_CACHE | bool | | Ignoring docker cache  |  | false |
 | CI_LOGS_KEY_PREFIX | string | | Prefix for build logs |  | false |
 | CI_NODE_LABEL_SELECTOR |  | | Node label selector for  CI |  | false |
 | CI_NODE_TAINTS_KEY | string | | Toleration key for CI |  | false |
 | CI_NODE_TAINTS_VALUE | string | | Toleration value for CI |  | false |
 | CI_RUNNER_DOCKER_MTU_VALUE | int |-1 | this is to control the bytes of inofrmation passed in a network packet in ci-runner.  default is -1 (defaults to the underlying node mtu value) |  | false |
 | CI_SUCCESS_AUTO_TRIGGER_BATCH_SIZE | int |1 | this is to control the no of linked pipelines should be hanled in one go when a ci-success event of an parent ci is received |  | false |
 | CI_VOLUME_MOUNTS_JSON | string | | additional volume mount data for CI and JOB |  | false |
 | CI_WORKFLOW_EXECUTOR_TYPE |  |AWF | Executor type for CI(AWF,System) |  | false |
 | DEFAULT_ARTIFACT_KEY_LOCATION | string |arsenal-v1/ci-artifacts | Key location for artifacts being created |  | false |
 | DEFAULT_BUILD_LOGS_BUCKET | string |devtron-pro-ci-logs |  |  | false |
 | DEFAULT_BUILD_LOGS_KEY_PREFIX | string |arsenal-v1 | Bucket prefix for build logs |  | false |
 | DEFAULT_CACHE_BUCKET | string |ci-caching | Bucket name for build cache |  | false |
 | DEFAULT_CACHE_BUCKET_REGION | string |us-east-2 | Build Cache bucket region |  | false |
 | DEFAULT_CD_ARTIFACT_KEY_LOCATION | string | | Bucket prefix for build cache |  | false |
 | DEFAULT_CD_LOGS_BUCKET_REGION | string |us-east-2 |  |  | false |
 | DEFAULT_CD_NAMESPACE | string | | Namespace for devtron stack |  | false |
 | DEFAULT_CD_TIMEOUT | int64 |3600 | Timeout for Pre/Post-Cd to be completed |  | false |
 | DEFAULT_CI_IMAGE | string |686244538589.dkr.ecr.us-east-2.amazonaws.com/cirunner:47 | To pass the ci-runner image |  | false |
 | DEFAULT_NAMESPACE | string |devtron-ci | Timeout for CI to be completed |  | false |
 | DEFAULT_TARGET_PLATFORM | string | | Default architecture for buildx |  | false |
 | DOCKER_BUILD_CACHE_PATH | string |/var/lib/docker | Path to store cache of docker build  (/var/lib/docker-> for legacy docker build, /var/lib/devtron-> for buildx) |  | false |
 | ENABLE_BUILD_CONTEXT | bool |false | To Enable build context in Devtron. |  | false |
 | ENABLE_WORKFLOW_EXECUTION_STAGE | bool |true | if enabled then we will display build stages separately for CI/Job/Pre-Post CD | true | false |
 | EXTERNAL_BLOB_STORAGE_CM_NAME | string |blob-storage-cm | name of the config map(contains bucket name, etc.) in external cluster when there is some operation related to external cluster, for example:-downloading cd artifact pushed in external cluster's env and we need to download from there, downloads ci logs pushed in external cluster's blob |  | false |
 | EXTERNAL_BLOB_STORAGE_SECRET_NAME | string |blob-storage-secret | name of the secret(contains password, accessId,passKeys, etc.) in external cluster when there is some operation related to external cluster, for example:-downloading cd artifact pushed in external cluster's env and we need to download from there, downloads ci logs pushed in external cluster's blob |  | false |
 | EXTERNAL_CD_NODE_LABEL_SELECTOR |  | | This is an array of strings used when submitting a workflow for pre or post-CD execution. If the  |  | false |
 | EXTERNAL_CD_NODE_TAINTS_KEY | string |dedicated |  |  | false |
 | EXTERNAL_CD_NODE_TAINTS_VALUE | string |ci |  |  | false |
 | EXTERNAL_CI_API_SECRET | string |devtroncd-secret | External CI API secret. |  | false |
 | EXTERNAL_CI_PAYLOAD | string |{"ciProjectDetails":[{"gitRepository":"https://github.com/vikram1601/getting-started-nodejs.git","checkoutPath":"./abc","commitHash":"239077135f8cdeeccb7857e2851348f558cb53d3","commitTime":"2022-10-30T20:00:00","branch":"master","message":"Update README.md","author":"User Name "}],"dockerImage":"445808685819.dkr.ecr.us-east-2.amazonaws.com/orch:23907713-2"} | External CI payload with project details. |  | false |
 | EXTERNAL_CI_WEB_HOOK_URL | string | | default is {{HOST_URL}}/orchestrator/webhook/ext-ci. It is used for external ci. |  | false |
 | IGNORE_CM_CS_IN_CI_JOB | bool |false | Ignore CM/CS in CI-pipeline as Job |  | false |
 | IMAGE_RETRY_COUNT | int |0 | push artifact(image) in ci retry count  |  | false |
 | IMAGE_RETRY_INTERVAL | int |5 | image retry interval takes value in seconds |  | false |
 | IMAGE_SCANNER_ENDPOINT | string |http://image-scanner-new-demo-devtroncd-service.devtroncd:80 | Image-scanner micro-service URL |  | false |
 | IMAGE_SCAN_MAX_RETRIES | int |3 | Max retry count for image-scanning |  | false |
 | IMAGE_SCAN_RETRY_DELAY | int |5 | Delay for the image-scaning to start |  | false |
 | IN_APP_LOGGING_ENABLED | bool |false | Used in case of argo workflow is enabled. If enabled logs push will be managed by us, else will be managed by argo workflow. |  | false |
 | MAX_CD_WORKFLOW_RUNNER_RETRIES | int |0 | Maximum time pre/post-cd-workflow create pod if it fails to complete |  | false |
 | MAX_CI_WORKFLOW_RETRIES | int |0 | Maximum time CI-workflow create pod if it fails to complete |  | false |
 | MODE | string |DEV |  |  | false |
 | NATS_SERVER_HOST | string |localhost:4222 |  |  | false |
 | ORCH_HOST | string |http://devtroncd-orchestrator-service-prod.devtroncd/webhook/msg/nats | Orchestrator micro-service URL  |  | false |
 | ORCH_TOKEN | string | | Orchestrator token |  | false |
 | PRE_CI_CACHE_PATH | string |/devtroncd-cache | Cache path for Pre CI tasks |  | false |
 | SHOW_DOCKER_BUILD_ARGS | bool |true | To enable showing the args passed for CI in build logs |  | false |
 | SKIP_CI_JOB_BUILD_CACHE_PUSH_PULL | bool |false | To skip cache Push/Pull for ci job |  | false |
 | SKIP_CREATING_ECR_REPO | bool |false | By disabling this ECR repo won't get created if it's not available on ECR from build configuration |  | false |
 | TERMINATION_GRACE_PERIOD_SECS | int |180 | this is the time given to workflow pods to shutdown. (grace full termination time) |  | false |
 | USE_ARTIFACT_LISTING_QUERY_V2 | bool |true | To use the V2 query for listing artifacts |  | false |
 | USE_BLOB_STORAGE_CONFIG_IN_CD_WORKFLOW | bool |true | To enable blob storage in pre and post cd |  | false |
 | USE_BLOB_STORAGE_CONFIG_IN_CI_WORKFLOW | bool |true | To enable blob storage in pre and post ci |  | false |
 | USE_BUILDX | bool |false | To enable buildx feature globally |  | false |
 | USE_DOCKER_API_TO_GET_DIGEST | bool |false | when user do not pass the digest  then this flag controls , finding the image digest using docker API or not. if set to true we get the digest from docker API call else use docker pull command. [logic in ci-runner] |  | false |
 | USE_EXTERNAL_NODE | bool |false | It is used in case of Pre/ Post Cd with run in application mode. If enabled the node lebels are read from EXTERNAL_CD_NODE_LABEL_SELECTOR else from CD_NODE_LABEL_SELECTOR MODE: if the vale is DEV, it will read the local kube config file or else from the cluser location. |  | false |
 | USE_IMAGE_TAG_FROM_GIT_PROVIDER_FOR_TAG_BASED_BUILD | bool |false | To use the same tag in container image as that of git tag |  | false |
 | WF_CONTROLLER_INSTANCE_ID | string |devtron-runner | Workflow controller instance ID. |  | false |
 | WORKFLOW_CACHE_CONFIG | string |{} | flag is used to configure how Docker caches are handled during a CI/CD  |  | false |
 | WORKFLOW_SERVICE_ACCOUNT | string |ci-runner |  |  | false |


## DEVTRON Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | - |  | |  |  | false |
 | ADDITIONAL_NODE_GROUP_LABELS |  | | Add comma separated list of additional node group labels to default labels | karpenter.sh/nodepool,cloud.google.com/gke-nodepool | false |
 | APP_SYNC_IMAGE | string |quay.io/devtron/chart-sync:1227622d-132-3775 | For the app sync image, this image will be used in app-manual sync job |  | false |
 | APP_SYNC_JOB_RESOURCES_OBJ | string | | To pass the resource of app sync |  | false |
 | APP_SYNC_SERVICE_ACCOUNT | string |chart-sync | Service account to be used in app sync Job |  | false |
 | APP_SYNC_SHUTDOWN_WAIT_DURATION | int |120 |  |  | false |
 | ARGO_AUTO_SYNC_ENABLED | bool |true | If enabled all argocd application will have auto sync enabled |  | false |
 | ARGO_GIT_COMMIT_RETRY_COUNT_ON_CONFLICT | int |3 | retry argocd app manual sync if the timeline is stuck in ARGOCD_SYNC_INITIATED state for more than this defined time (in mins) |  | false |
 | ARGO_GIT_COMMIT_RETRY_DELAY_ON_CONFLICT | int |1 | Delay on retrying the maifest commit the on gitops |  | false |
 | ARGO_REPO_REGISTER_RETRY_COUNT | int |3 | Argo app registration in argo retries on deployment |  | false |
 | ARGO_REPO_REGISTER_RETRY_DELAY | int |10 | Argo app registration in argo cd on deployment delay between retry |  | false |
 | ASYNC_BUILDX_CACHE_EXPORT | bool |false | To enable async container image cache export |  | false |
 | BATCH_SIZE | int |5 | there is feature to get URL's of services/ingresses. so to extract those, we need to parse all the servcie and ingress objects of the application. this BATCH_SIZE flag controls the no of these objects get parsed in one go. |  | false |
 | BLOB_STORAGE_ENABLED | bool |false |  |  | false |
 | BUILDX_CACHE_MODE_MIN | bool |false | To set build cache mode to minimum in buildx |  | false |
 | CD_HOST | string |localhost | Host for the devtron stack |  | false |
 | CD_NAMESPACE | string |devtroncd |  |  | false |
 | CD_PORT | string |8000 | Port for pre/post-cd |  | false |
 | CExpirationTime | int |600 | Caching expiration time. |  | false |
 | CI_TRIGGER_CRON_TIME | int |2 | For image poll plugin |  | false |
 | CI_WORKFLOW_STATUS_UPDATE_CRON | string |*/5 * * * * | Cron schedule for CI pipeline status |  | false |
 | CLI_CMD_TIMEOUT_GLOBAL_SECONDS | int |0 | Used in git cli opeartion timeout |  | false |
 | CLUSTER_STATUS_CRON_TIME | int |15 | Cron schedule for cluster status on resource browser |  | false |
 | CONSUMER_CONFIG_JSON | string | |  |  | false |
 | DEFAULT_LOG_TIME_LIMIT | int64 |1 |  |  | false |
 | DEFAULT_TIMEOUT | float64 |3600 | Timeout for CI to be completed |  | false |
 | DEVTRON_BOM_URL | string |https://raw.githubusercontent.com/devtron-labs/devtron/%s/charts/devtron/devtron-bom.yaml | Path to devtron-bom.yaml of devtron charts, used for module installation and devtron upgrade |  | false |
 | DEVTRON_DEFAULT_NAMESPACE | string |devtroncd |  |  | false |
 | DEVTRON_DEX_SECRET_NAMESPACE | string |devtroncd | Namespace of dex secret |  | false |
 | DEVTRON_HELM_RELEASE_CHART_NAME | string |devtron-operator |  |  | false |
 | DEVTRON_HELM_RELEASE_NAME | string |devtron | Name of the Devtron Helm release.  |  | false |
 | DEVTRON_HELM_RELEASE_NAMESPACE | string |devtroncd | Namespace of the Devtron Helm release |  | false |
 | DEVTRON_HELM_REPO_NAME | string |devtron | Is used to install modules (stack manager) |  | false |
 | DEVTRON_HELM_REPO_URL | string |https://helm.devtron.ai | Is used to install modules (stack manager) |  | false |
 | DEVTRON_INSTALLATION_TYPE | string | | Devtron Installation type(EA/Full) |  | false |
 | DEVTRON_INSTALLER_MODULES_PATH | string |installer.modules | Path to devtron installer modules, used to find the helm charts and values files |  | false |
 | DEVTRON_INSTALLER_RELEASE_PATH | string |installer.release | Path to devtron installer release, used to find the helm charts and values files |  | false |
 | DEVTRON_MODULES_IDENTIFIER_IN_HELM_VALUES | string |installer.modules |  |  | false |
 | DEVTRON_OPERATOR_BASE_PATH | string | | Base path for devtron operator, used to find the helm charts and values files |  | false |
 | DEVTRON_SECRET_NAME | string |devtron-secret |  |  | false |
 | DEVTRON_VERSION_IDENTIFIER_IN_HELM_VALUES | string |installer.release | devtron operator version identifier in helm values yaml |  | false |
 | DEX_CID | string |example-app | dex client id  |  | false |
 | DEX_CLIENT_ID | string |argo-cd |  |  | false |
 | DEX_CSTOREKEY | string | | DEX CSTOREKEY. |  | false |
 | DEX_JWTKEY | string | | DEX JWT key.   |  | false |
 | DEX_RURL | string |http://127.0.0.1:8080/callback | Dex redirect URL(http://argocd-dex-server.devtroncd:8080/callback) |  | false |
 | DEX_SCOPES |  | |  |  | false |
 | DEX_SECRET | string | | Dex secret |  | false |
 | DEX_URL | string | | Dex service endpoint with dex path(http://argocd-dex-server.devtroncd:5556/dex) |  | false |
 | ECR_REPO_NAME_PREFIX | string |test/ | Prefix for ECR repo to be created in does not exist |  | false |
 | ENABLE_ASYNC_ARGO_CD_INSTALL_DEVTRON_CHART | bool |false | To enable async installation of gitops application |  | false |
 | ENABLE_ASYNC_INSTALL_DEVTRON_CHART | bool |false | To enable async installation of no-gitops application |  | false |
 | ENABLE_NOTIFIER_V2 | bool |false | enable notifier v2 |  | false |
 | EPHEMERAL_SERVER_VERSION_REGEX | string |v[1-9]\.\b(2[3-9]\|[3-9][0-9])\b.* | ephemeral containers support version regex that is compared with k8sServerVersion |  | false |
 | EVENT_URL | string |http://localhost:3000/notify | Notifier service url |  | false |
 | EXECUTE_WIRE_NIL_CHECKER | bool |false | checks for any nil pointer in wire.go |  | false |
 | EXPOSE_CI_METRICS | bool |false | To expose CI metrics |  | false |
 | FEATURE_RESTART_WORKLOAD_BATCH_SIZE | int |1 | restart workload retrieval batch size  |  | false |
 | FEATURE_RESTART_WORKLOAD_WORKER_POOL_SIZE | int |5 | restart workload retrieval pool size |  | false |
 | FORCE_SECURITY_SCANNING | bool |false | By enabling this no one can disable image scaning on ci-pipeline from UI |  | false |
 | GITOPS_REPO_PREFIX | string | | Prefix for Gitops repo being creation for argocd application |  | false |
 | GO_RUNTIME_ENV | string |production |  |  | false |
 | GRAFANA_HOST | string |localhost | Host URL for the grafana dashboard |  | false |
 | GRAFANA_NAMESPACE | string |devtroncd | Namespace for grafana |  | false |
 | GRAFANA_ORG_ID | int |2 | Org ID for grafana for application metrics |  | false |
 | GRAFANA_PASSWORD | string |prom-operator | Password for grafana dashboard |  | false |
 | GRAFANA_PORT | string |8090 | Port for grafana micro-service |  | false |
 | GRAFANA_URL | string | | Host URL for the grafana dashboard |  | false |
 | GRAFANA_USERNAME | string |admin | Username for grafana  |  | false |
 | HIDE_IMAGE_TAGGING_HARD_DELETE | bool |false | Flag to hide the hard delete option in the image tagging service |  | false |
 | IGNORE_AUTOCOMPLETE_AUTH_CHECK | bool |false | flag for ignoring auth check in autocomplete apis. |  | false |
 | INSTALLED_MODULES |  | | List of installed modules given in helm values/yaml are written in cm and used by devtron to know which modules are given | security.trivy,security.clair | false |
 | INSTALLER_CRD_NAMESPACE | string |devtroncd | namespace where Custom Resource Definitions get installed |  | false |
 | INSTALLER_CRD_OBJECT_GROUP_NAME | string |installer.devtron.ai | Devtron installer CRD group name, partially deprecated. |  | false |
 | INSTALLER_CRD_OBJECT_RESOURCE | string |installers | Devtron installer CRD resource name, partially deprecated |  | false |
 | INSTALLER_CRD_OBJECT_VERSION | string |v1alpha1 | version of the CRDs. default is v1alpha1 |  | false |
 | IS_AIR_GAP_ENVIRONMENT | bool |false |  |  | false |
 | JwtExpirationTime | int |120 | JWT expiration time. |  | false |
 | K8s_CLIENT_MAX_IDLE_CONNS_PER_HOST | int |25 |  |  | false |
 | K8s_TCP_IDLE_CONN_TIMEOUT | int |300 |  |  | false |
 | K8s_TCP_KEEPALIVE | int |30 |  |  | false |
 | K8s_TCP_TIMEOUT | int |30 |  |  | false |
 | K8s_TLS_HANDSHAKE_TIMEOUT | int |10 |  |  | false |
 | LENS_TIMEOUT | int |0 | Lens microservice timeout. |  | false |
 | LENS_URL | string |http://lens-milandevtron-service:80 | Lens micro-service URL |  | false |
 | LIMIT_CI_CPU | string |0.5 |  |  | false |
 | LIMIT_CI_MEM | string |3G |  |  | false |
 | LOGGER_DEV_MODE | bool |false | Enables a different logger theme. |  | false |
 | LOG_LEVEL | int |-1 |  |  | false |
 | MAX_SESSION_PER_USER | int |5 | max no of cluster terminal pods can be created by an user |  | false |
 | MODULE_METADATA_API_URL | string |https://api.devtron.ai/module?name=%s | Modules list and meta info will be fetched from this server, that is central api server of devtron. |  | false |
 | MODULE_STATUS_HANDLING_CRON_DURATION_MIN | int |3 |  |  | false |
 | NATS_MSG_ACK_WAIT_IN_SECS | int |120 |  |  | false |
 | NATS_MSG_BUFFER_SIZE | int |-1 |  |  | false |
 | NATS_MSG_MAX_AGE | int |86400 |  |  | false |
 | NATS_MSG_PROCESSING_BATCH_SIZE | int |1 |  |  | false |
 | NATS_MSG_REPLICAS | int |0 |  |  | false |
 | NOTIFICATION_MEDIUM | NotificationMedium |rest | notification medium |  | false |
 | OTEL_COLLECTOR_URL | string | | Opentelemetry URL  |  | false |
 | PARALLELISM_LIMIT_FOR_TAG_PROCESSING | int | | App manual sync job parallel tag processing count. |  | false |
 | PG_EXPORT_PROM_METRICS | bool |true |  |  | false |
 | PG_LOG_ALL_FAILURE_QUERIES | bool |true |  |  | false |
 | PG_LOG_ALL_QUERY | bool |false |  |  | false |
 | PG_LOG_SLOW_QUERY | bool |true |  |  | false |
 | PG_QUERY_DUR_THRESHOLD | int64 |5000 |  |  | false |
 | PLUGIN_NAME | string |Pull images from container repository | Handles image retrieval from a container repository and triggers subsequent CI processes upon detecting new images.Current default plugin name: Pull Images from Container Repository. |  | false |
 | PROPAGATE_EXTRA_LABELS | bool |false | Add additional propagate labels like api.devtron.ai/appName, api.devtron.ai/envName, api.devtron.ai/project along with the user defined ones. |  | false |
 | PROXY_SERVICE_CONFIG | string |{} | Proxy configuration for micro-service to be accessible on orhcestrator ingress |  | false |
 | REQ_CI_CPU | string |0.5 |  |  | false |
 | REQ_CI_MEM | string |3G |  |  | false |
 | RESTRICT_TERMINAL_ACCESS_FOR_NON_SUPER_USER | bool |false | To restrict the cluster terminal from user having non-super admin acceess |  | false |
 | RUNTIME_CONFIG_LOCAL_DEV | LocalDevMode |true |  |  | false |
 | SCOPED_VARIABLE_ENABLED | bool |false | To enable scoped variable option |  | false |
 | SCOPED_VARIABLE_FORMAT | string |@{{%s}} | Its a scope format for varialbe name. |  | false |
 | SCOPED_VARIABLE_HANDLE_PRIMITIVES | bool |false | This describe should we handle primitives or not in scoped variable template parsing. |  | false |
 | SCOPED_VARIABLE_NAME_REGEX | string |^[a-zA-Z][a-zA-Z0-9_-]{0,62}[a-zA-Z0-9]$ | Regex for scoped variable name that must passed this regex. |  | false |
 | SOCKET_DISCONNECT_DELAY_SECONDS | int |5 | The server closes a session when a client receiving connection have not been seen for a while.This delay is configured by this setting. By default the session is closed when a receiving connection wasn't seen for 5 seconds. |  | false |
 | SOCKET_HEARTBEAT_SECONDS | int |25 | In order to keep proxies and load balancers from closing long running http requests we need to pretend that the connection is active and send a heartbeat packet once in a while. This setting controls how often this is done. By default a heartbeat packet is sent every 25 seconds. |  | false |
 | STREAM_CONFIG_JSON | string | |  |  | false |
 | SYSTEM_VAR_PREFIX | string |DEVTRON_ | Scoped variable prefix, variable name must have this prefix. |  | false |
 | TERMINAL_POD_DEFAULT_NAMESPACE | string |default | Cluster terminal default namespace |  | false |
 | TERMINAL_POD_INACTIVE_DURATION_IN_MINS | int |10 | Timeout for cluster terminal to be inactive |  | false |
 | TERMINAL_POD_STATUS_SYNC_In_SECS | int |600 | this is the time interval at which the status of the cluster terminal pod |  | false |
 | TEST_APP | string |orchestrator |  |  | false |
 | TEST_PG_ADDR | string |127.0.0.1 |  |  | false |
 | TEST_PG_DATABASE | string |orchestrator |  |  | false |
 | TEST_PG_LOG_QUERY | bool |true |  |  | false |
 | TEST_PG_PASSWORD | string |postgrespw |  |  | false |
 | TEST_PG_PORT | string |55000 |  |  | false |
 | TEST_PG_USER | string |postgres |  |  | false |
 | TIMEOUT_FOR_FAILED_CI_BUILD | string |15 | Timeout for Failed CI build  |  | false |
 | TIMEOUT_IN_SECONDS | int |5 | timeout to compute the urls from services and ingress objects of an application |  | false |
 | USER_SESSION_DURATION_SECONDS | int |86400 |  |  | false |
 | USE_ARTIFACT_LISTING_API_V2 | bool |true | To use the V2 API for listing artifacts in Listing the images in pipeline |  | false |
 | USE_CUSTOM_HTTP_TRANSPORT | bool |false |  |  | false |
 | USE_GIT_CLI | bool |false | To enable git cli |  | false |
 | USE_RBAC_CREATION_V2 | bool |true | To use the V2 for RBAC creation |  | false |
 | VARIABLE_CACHE_ENABLED | bool |true | This is used to  control caching of all the scope variables defined in the system. |  | false |
 | VARIABLE_EXPRESSION_REGEX | string |@{{([^}]+)}} | Scoped variable expression regex |  | false |
 | WEBHOOK_TOKEN | string | | If you want to continue using jenkins for CI then please provide this for authentication of requests |  | false |


## GITOPS Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | ACD_CM | string |argocd-cm | Name of the argocd CM |  | false |
 | ACD_NAMESPACE | string |devtroncd | To pass the argocd namespace |  | false |
 | ACD_PASSWORD | string | | Password for the Argocd (deprecated) |  | false |
 | ACD_USERNAME | string |admin | User name for argocd |  | false |
 | GITOPS_SECRET_NAME | string |devtron-gitops-secret | devtron-gitops-secret |  | false |
 | RESOURCE_LIST_FOR_REPLICAS | string |Deployment,Rollout,StatefulSet,ReplicaSet | this holds the list of k8s resource names which support replicas key. this list used in hibernate/un hibernate process |  | false |
 | RESOURCE_LIST_FOR_REPLICAS_BATCH_SIZE | int |5 | this the batch size to control no of above resources can be parsed in one go to determine hibernate status |  | false |


## INFRA_SETUP Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | DASHBOARD_HOST | string |localhost | Dashboard micro-service URL |  | false |
 | DASHBOARD_NAMESPACE | string |devtroncd | Dashboard micro-service namespace |  | false |
 | DASHBOARD_PORT | string |3000 | Port for dashboard micro-service |  | false |
 | DEX_HOST | string |http://localhost |  |  | false |
 | DEX_PORT | string |5556 |  |  | false |
 | GIT_SENSOR_PROTOCOL | string |REST | Protocol to connect with git-sensor micro-service |  | false |
 | GIT_SENSOR_SERVICE_CONFIG | string |{"loadBalancingPolicy":"pick_first"} | git-sensor grpc service config |  | false |
 | GIT_SENSOR_TIMEOUT | int |0 | Timeout for getting response from the git-sensor |  | false |
 | GIT_SENSOR_URL | string |127.0.0.1:7070 | git-sensor micro-service url  |  | false |
 | HELM_CLIENT_URL | string |127.0.0.1:50051 | Kubelink micro-service url  |  | false |
 | KUBELINK_GRPC_MAX_RECEIVE_MSG_SIZE | int |20 |  |  | false |
 | KUBELINK_GRPC_MAX_SEND_MSG_SIZE | int |4 |  |  | false |
 | KUBELINK_GRPC_SERVICE_CONFIG | string |{"loadBalancingPolicy":"round_robin"} | kubelink grpc service config |  | false |


## POSTGRES Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | APP | string |orchestrator | Application name |  | false |
 | CASBIN_DATABASE | string |casbin | Database for casbin |  | false |
 | PG_ADDR | string |127.0.0.1 | address of postgres service | postgresql-postgresql.devtroncd | false |
 | PG_DATABASE | string |orchestrator | postgres database to be made connection with | orchestrator, casbin, git_sensor, lens | false |
 | PG_PASSWORD | string |{password} | password for postgres, associated with PG_USER | confidential ;) | false |
 | PG_PORT | string |5432 | port of postgresql service | 5432 | false |
 | PG_READ_TIMEOUT | int64 |30 | Time out for read operation in postgres |  | false |
 | PG_USER | string |postgres | user for postgres | postgres | false |
 | PG_WRITE_TIMEOUT | int64 |30 | Time out for write operation in postgres |  | false |


## RBAC Related Environment Variables
| Key   | Type     | Default Value     | Description       | Example       | Deprecated       |
|-------|----------|-------------------|-------------------|-----------------------|------------------|
 | ENFORCER_CACHE | bool |false | To Enable enforcer cache. |  | false |
 | ENFORCER_CACHE_EXPIRATION_IN_SEC | int |86400 | Expiration time (in seconds) for enforcer cache.  |  | false |
 | ENFORCER_MAX_BATCH_SIZE | int |1 | Maximum batch size for the enforcer. |  | false |
 | USE_CASBIN_V2 | bool |true | To enable casbin V2 API |  | false |


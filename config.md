# DEVTRON CONFIGMAP PARAMETER

| Key                                      | Value                                                     | Description |
|------------------------------------------|-----------------------------------------------------------|-------------|
| APP                                      | orchestrator                                              | Application name. |
| MODE                                     | PROD                                                      | Operating mode. |
| DASHBOARD_PORT                           | 80                                                        | Port number of the dashboard service. |
| DASHBOARD_HOST                           | dashboard-service.devtroncd                               | URL of the dashboard service. |
| CD_HOST                                  | argocd-server.devtroncd                                   | Service name of the ArgoCD service. |
| CD_PORT                                  | 80                                                        | Port of the ArgoCD service. |
| CD_NAMESPACE                             | devtroncd                                                 | Namespace for ArgoCD. |
| GITOPS_REPO_PREFIX                       | devtron                                                   | GitOps repository prefix. |
| EVENT_URL                               | http://notifier-service.devtroncd:80/notify                | URL of the notifier microservice. |
| LENS_URL                                | http://lens-service.devtroncd:80                          | URL of the lens microservice. |
| LENS_TIMEOUT                            | 300                                                       | Lens microservice timeout. |
| HELM_CLIENT_URL                         | kubelink-service:50051                                    | URL of the Helm client microservice. |
| NATS_SERVER_HOST                        | nats://devtron-nats.devtroncd:4222                        | URL of the NATS microservice. |
| PG_ADDR                                 | postgresql-postgresql.devtroncd                          | URL of the PostgreSQL microservice. |
| PG_PORT                                 | 5432                                                     | PostgreSQL port. |
| PG_USER                                 | postgres                                                 | PostgreSQL user. |
| PG_DATABASE                             | orchestrator                                             | PostgreSQL database. |
| GIT_SENSOR_TIMEOUT                      | 300                                                      | Timeout for the Git sensor microservice. |
| GIT_SENSOR_PROTOCOL                      | GRPC                                                     | Git sensor protocol. |
| GIT_SENSOR_URL                          | git-sensor-service.devtroncd:90                          | URL of the Git sensor microservice. |
| APP_SYNC_IMAGE                          | quay.io/devtron/chart-sync:0e8c785e-373-16172            | Image for syncing Devtron chart store. |
| DEX_HOST                               | http://argocd-dex-server.devtroncd                        | URL of Argocd Dex server. |
| DEX_PORT                               | 5556                                                     | Port of Argocd Dex server. |
| DEX_RURL                               | http://argocd-dex-server.devtroncd:8080/callback          | Argocd Dex server redirect URL. |
| DEX_URL                                | http://argocd-dex-server.devtroncd:5556/dex              | URL for Argocd Dex server. |
| CExpirationTime                        | 600                                                      | Caching expiration time. |
| JwtExpirationTime                      | 120                                                      | JWT expiration time. |
| IMAGE_SCANNER_ENDPOINT                 | http://image-scanner-service.devtroncd:80                | URL of the image scanner microservice. |
| LOG_LEVEL                              | 0                                                        | Log level (-2, -1, 0). |
| PG_LOG_QUERY                           | true                                                     | Log PostgreSQL queries. |
| GRAFANA_URL                            | http://%s:%s@devtron-grafana.devtroncd/grafana            | URL for Grafana. |
| GRAFANA_HOST                           | devtron-grafana.devtroncd                                | Grafana host. |
| GRAFANA_PORT                           | 80                                                       | Grafana port. |
| GRAFANA_NAMESPACE                      | devtroncd                                                | Grafana namespace. |
| GRAFANA_ORG_ID                         | 2                                                        | Grafana organization ID. |
| ACD_URL                                | argocd-server.devtroncd                                  | URL for Argocd server. |
| ACD_USERNAME                           | admin                                                    | Argocd username. |
| ACD_USER                               | admin                                                    | Argocd user. |
| ACD_CM                                 | argocd-cm                                                | Argocd config map. |
| ACD_NAMESPACE                          | devtroncd                                                | Namespace for Argocd. |
| ACD_TIMEOUT                            | 300                                                      | Argocd timeout. |
| ACD_SKIP_VERIFY                        | true                                                     | Skip Argocd verification. |
| GIT_WORKING_DIRECTORY                   | /tmp/gitops/                                             | Git working directory. |
| CD_LIMIT_CI_CPU                        | 0.5                                                     | CI CPU limit for post-build workflow. |
| CD_LIMIT_CI_MEM                        | 3G                                                      | CI memory limit for post-build workflow. |
| CD_REQ_CI_CPU                          | 0.5                                                     | CI CPU request for post-build workflow. |
| CD_REQ_CI_MEM                          | 1G                                                      | CI memory request for post-build workflow. |
| CD_NODE_TAINTS_KEY                     | dedicated                                               | CI node taints key. |
| CD_NODE_LABEL_SELECTOR                 | kubernetes.io/os=linux                                   | CI node label selector. |
| CD_WORKFLOW_SERVICE_ACCOUNT             | cd-runner                                                | CI workflow service account. |
| CD_NODE_TAINTS_VALUE                   | ci                                                      | CI node taints value. |
| DEFAULT_CD_ARTIFACT_KEY_LOCATION       | devtron/cd-artifacts                                      | Default location for CI artifacts. |
| CD_ARTIFACT_LOCATION_FORMAT            | %d/%d.zip                                                | Format for CI artifact locations. |
| DEFAULT_CD_NAMESPACE                   | devtron-cd                                               | Default namespace for CI. |
| DEFAULT_CD_TIMEOUT                     | 3600                                                    | Default timeout for CI. |
| ENABLE_BUILD_CONTEXT                   | true                                                    | Enable build context in Devtron. |
| DEFAULT_CI_IMAGE                       | quay.io/devtron/ci-runner:d8d774c3-138-16238             | Default image for CI pods. |
| WF_CONTROLLER_INSTANCE_ID               | devtron-runner                                           | Workflow controller instance ID. |
| CI_LOGS_KEY_PREFIX                     | ci-artifacts                                             | Key prefix for CI artifacts. |
| DEFAULT_NAMESPACE                       | devtron-ci                                              | Default namespace for CI. |
| DEFAULT_TIMEOUT                         | 3600                                                    | Default timeout for CI. |
| LIMIT_CI_CPU                            | 0.5                                                    | CI CPU limit. |
| LIMIT_CI_MEM                            | 3G                                                     | CI memory limit. |
| REQ_CI_CPU                              | 0.5                                                    | CI CPU request. |
| REQ_CI_MEM                              | 1G                                                     | CI memory request. |
| CI_NODE_TAINTS_KEY                      |                                                        | CI node taints key. |
| CI_NODE_TAINTS_VALUE                    |                                                        | CI node taints value. |
| CI_NODE_LABEL_SELECTOR                  |                                                        | CI node label selector. |
| CACHE_LIMIT                             | 5000000000                                            | Cache limit. |
| DEFAULT_ARTIFACT_KEY_LOCATION           | devtron/ci-artifacts                                  | Default location for CI artifacts. |
| WORKFLOW_SERVICE_ACCOUNT                | ci-runner                                            | Workflow service account for CI. |
| CI_ARTIFACT_LOCATION_FORMAT             | %d/%d.zip                                            | Format for CI artifact locations. |
| DEFAULT_BUILD_LOGS_KEY_PREFIX           | devtron                                              | Default key prefix for build logs. |
| MINIO_ENDPOINT                           | http://devtron-minio:9000                           | If Minio is enabled, the Minio endpoint URL.        |
| BLOB_STORAGE_ENABLED                     | "true"                                              | Enables blob storage configuration.                  |
| BLOB_STORAGE_PROVIDER                    | "S3"                                                | The provider for blob storage.                      |
| BLOB_STORAGE_S3_ENDPOINT                 | "http://devtron-minio.devtroncd:9000"              | S3 endpoint URL for blob storage.                   |
| BLOB_STORAGE_S3_ENDPOINT_INSECURE        | "true"                                              | Indicates if S3 endpoint is insecure.               |
| DEFAULT_BUILD_LOGS_BUCKET                | "devtron-ci-log"                                    | Default bucket for build logs.                      |
| DEFAULT_CACHE_BUCKET                     | "devtron-ci-cache"                                  | Default bucket for caching.                         |
| BLOB_STORAGE_S3_BUCKET_VERSIONED         | "false"                                             | If S3 buckets are versioned (false or true).        |
| DEFAULT_CACHE_BUCKET_REGION               | "us-west-2"                                         | Default region for the cache bucket.                |
| DEFAULT_CD_LOGS_BUCKET_REGION             | "us-west-2"                                         | Default region for CD logs bucket.                  |
| BLOB_STORAGE_S3_ENDPOINT                 | ""                                                  | S3 endpoint (empty, may need to be removed).        |
| BLOB_STORAGE_S3_BUCKET_VERSIONED         | "true"                                              | If S3 buckets are versioned (true or false).        |
| ECR_REPO_NAME_PREFIX                     | "devtron/"                                          | Prefix for ECR repository name.                     |
| EXTERNAL_CI_PAYLOAD                      | JSON Payload                                       | External CI payload with project details.           |
| ENFORCER_CACHE                           | "true"                                              | Enable enforcer cache.                              |
| ENFORCER_CACHE_EXPIRATION_IN_SEC          | "345600"                                            | Expiration time (in seconds) for enforcer cache.    |
| ENFORCER_MAX_BATCH_SIZE                  | "1"                                                 | Maximum batch size for the enforcer.                |
| DEVTRON_SECRET_NAME                      | "devtron-secret"                                    | Name of the Devtron secret.                         |
| BLOB_STORAGE_PROVIDER                    | ""                                                  | Provider for blob storage (may need to be removed). |
| DEVTRON_HELM_RELEASE_NAME                | devtron                                             | Name of the Devtron Helm release.                   |
| ENABLE_LEGACY_API                        | "false"                                             | Enable the legacy API.                              |
| INSTALLATION_THROUGH_HELM                | "True"                                              | Installation through Helm (True or False).          |


# DASHBOARD PARAMETER

| Key                               | Value     | Description                                     |
|-----------------------------------|-----------|-------------------------------------------------|
| APPLICATION_METRICS_ENABLED        | "true"    | Show application metrics button                |
| CLUSTER_NAME                       | demo      | Unknown                                         |
| HIDE_APPLICATION_GROUPS            | "false"   | Hide application group from Devtron UI         |
| HIDE_DISCORD                       | "true"    | Hide Discord button from UI                    |
| HIDE_GITOPS_OR_HELM_OPTION         | "false"   | Enable GitOps and Helm option                 |
| HOTJAR_ENABLED                     | "false"   | Hotjar integration status                      |
| POSTHOG_ENABLED                    | "true"    | PostHog integration status                     |
| POSTHOG_TOKEN                      | XXXXXXXX  | PostHog API token                        |
| SENTRY_ENABLED                     | "false"   | Sentry integration status                      |
| SENTRY_ENV                         | stage     | Sentry environment                              |
| USE_V2                             | "true"    | Use the v2 APIs                                 |
| ENABLE_RESTART_WORKLOAD            | "true"    | Show restart pods option in app details page   |
| ENABLE_BUILD_CONTEXT               | "true"    | Enable build context in Devtron UI             |
| FORCE_SECURITY_SCANNING            | "false"   | Force security scanning                         |
| GA_ENABLED                         | "true"    | Enable Google Analytics (GA)                   |
| GA_TRACKING_ID                     | G-XXXXXXXX | Google Analytics tracking ID                 |


#  KUBELINK PARAMAETER


| Key                    | Value                                | Description                                         |
|-----------------------------------|--------------------------------------|-----------------------------------------------------|
| ENABLE_HELM_RELEASE_CACHE         | "true"                               | Enable Helm release cache                          |
| NATS_MSG_PROCESSING_BATCH_SIZE    | "1"                                  | NATS message processing batch size                 |
| NATS_SERVER_HOST                  | nats://devtron-nats.devtroncd:4222   | NATS server host address                            |
| RUN_HELM_INSTALL_IN_ASYNC_MODE    | "true"                               | Run Helm install in async mode                     |
| PG_LOG_QUERY                      | "true"                               | Enable PostgreSQL query logging                     |
| PG_ADDR                           | postgresql-postgresql.devtroncd      | PostgreSQL server address                           |
| PG_DATABASE                       | orchestrator                         | PostgreSQL database name                            |
| PG_PORT                           | "5432"                               | PostgreSQL server port                              |
| PG_USER                           | postgres                             | PostgreSQL database user                            |



# KUBEWATCH PARAMETER


| Variable Name          | Value        | Description                 |
|------------------------|--------------|-----------------------------|
| DEFAULT_NAMESPACE      | devtron-ci   | The default namespace for CI |
| CI_INFORMER            | true         | Enable CI informer           |
| ACD_INFORMER           | true         | Enable ACD informer          |
| NATS_STREAM_MAX_AGE    | 10800        | Maximum age for NATS stream  |
| ACD_NAMESPACE          | devtroncd    | The namespace for ACD        |
| LOG_LEVEL              | 2            | Logging level                |



# IMAGESCANER


| Variable Name       | Value                                  | Description                   |
|---------------------|----------------------------------------|-------------------------------|
| CLAIR_ADDR          | clair-dcd.devtroncd:6060               | For connecting to Clair if it's enabled |
| CLIENT_ID           | client-2                               | Client ID                        |
| NATS_SERVER_HOST    | nats://devtron-nats.devtroncd:4222    | For connecting to NATS         |
| PG_LOG_QUERY        | "false"                                | PostgreSQL Query Logging (false to disable) |
| PG_ADDR             | postgresql-postgresql.devtroncd        | PostgreSQL Server Address       |
| PG_DATABASE         | orchestrator                           | PostgreSQL Database Name       |
| PG_PORT             | "5432"                                 | PostgreSQL Port Number         |
| PG_USER             | postgres                               | PostgreSQL User Name           |




# GITSENSOR

| Variable Name           | Value                    | Description                                 |
|------------------------|--------------------------|---------------------------------------------|
| PG_ADDR                | postgresql-postgresql.devtroncd | Connect to the Postgres database for git-sensor. |
| PG_USER                | postgres                 | Postgres database username.                 |
| PG_DATABASE            | git_sensor               | Name of the Postgres database.              |
| POLL_DURATION          | 1                        | Polling duration in minutes.                |
| POLL_WORKER            | 2                        | Number of workers for fetching commits.     |
| PG_LOG_QUERY           | false                    | Enable or disable Postgres query logging.   |
| COMMIT_STATS_TIMEOUT_IN_SEC | 2                 | Timeout for fetching commit .     |
| ENABLE_FILE_STATS      | false                    | Enable or disable file stats.               |
| GIT_HISTORY_COUNT      | 15                       | Number of commits to display in the UI.     |
| CLONING_MODE           | FULL (or SHALLOW)         | Cloning mode for git repositories.          |
| PG_PASSWORD            | **************            | Password for connecting to the Postgres database. |

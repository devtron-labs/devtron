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


# DEVTRON SECRET PARAMETER
| Key                  | Value                                                     | Description                  |
|----------------------|-----------------------------------------------------------|------------------------------|
| ADMIN_PASSWORD       | RUowOFVQczZvdmJpempxVQ==                                 | Admin password.              |
| DEX_CSTOREKEY        | YTNkNFdHTnViRGM1TkcxVWNrUnNVRVpoVGxWT1ZXTnlURmhKUFFv      | DEX CSTOREKEY.              |
| DEX_JWTKEY           | UnpZMmVYbHZVMlJyYXk5emVtRk9kR0YyTkFwc2RtFjBiMDlyUFFv      | DEX JWT key.                |
| DEX_SECRET           | UVdoaldYTTRkR05LU1ZGWU1FSlpZMXBpWm1jNVprbE1ZbkZCUFFv      | DEX secret.                 |
| EXTERNAL_CI_API_SECRET | VGxOeGJVWk9jMFpMVDBJNWR5OVlaREUwYkdaR0wyMVhjM2RqUFFv      | External CI API secret.     |
| ORCH_TOKEN           | WTBjMlFuVkdSWGxtT1ROc2RsUk9lUzl2WkVNNWFXMWxSMkpGUFFv      | Orchestrator token.         |
| PG_PASSWORD          | RVRVa0lZMHlwN3VSNUJTZGgwbHBFcHRvbTdPeWhvSGE=            | PostgreSQL password.        |
| WEBHOOK_TOKEN        | WVZFMFZYcDJObXRpWXpsVFdXNVRMMk4xTm1WWFRXaFJjbTFyUFFv      | Webhook token.              |
| admin.password       | JDJhJDEwJFZwUWhOc08zSVlVNEdnZzlCZHVrRk9vZHl2TE52MjBxdnhJMkxYeDlqSHA0THpqQTZpdkJl  | Admin password.              |
| admin.passwordMtime  | MjAyMy0xMC0zMFQwNzoxNjozN1o=                          | Admin password modification time. |
| dex.config           | ""                                                        | DEX configuration.           |
| server.secretkey     | RFVtUkFiVnhTQ1QvS2lydGJ0a0N3NjFEOVNtVFNFOEERaTXdIRlA3cmpZRT0= | Server secret key.           |
| url                  | ""                                                        | URL (Uniform Resource Locator). |
| APP                  | orchestrator                                              | Application name.           |
| MODE                 | PROD                                                      | Operating mode.            |

# devtron-custom-cm
| Key      | Value           | Description           |
|----------|-----------------|-----------------------|
| DEFAULT_CI_IMAGE | quay.io/devtron/ci-runner:ad3af321-138-18662    | Default image for CI pods.     |



# devtron-nats-config
| Key                          | Value                                | Description                                 |
|------------------------------|--------------------------------------|---------------------------------------------|
| pid_file                     | "/var/run/nats/nats.pid"             | PID file shared with configuration reloader.|
| http                         | 8222                                 | Monitoring HTTP port.                       |
| server_name                  | $POD_NAME                            | Name of the NATS server.                    |
| jetstream.max_mem            | 1Gi                                  | Maximum memory for NATS JetStream.          |
| jetstream.domain             | devtron-jet                          | Domain for NATS JetStream.                  |
| lame_duck_duration           | 120s                                 | Lame duck duration in seconds.             |

# devtron-cluster-components
  rollout.yaml: |-
    rollout:
      resources:
        limits:
          cpu: 250m
          memory: 200Mi
        requests:
          cpu: 50m
          memory: 100Mi

# postgresql-postgresql-init-scripts
  db_create.sql: |
    create database casbin;
    create database git_sensor;
    create database lens;
    create database clairv4
# argocd-cmd-params-cm
| Key                                    | Value                           | Description                                     |
|----------------------------------------|---------------------------------|-------------------------------------------------|
| controller.log.format                 | text                            | Log format for the controller                  |
| controller.log.level                  | info                            | Log level for the controller                   |
| controller.operation.processors        | 10                              | Number of processors for controller operations |
| controller.repo.server.timeout.seconds | 60                              | Timeout in seconds for the repository server   |
| controller.self.heal.timeout.seconds   | 5                               | Timeout in seconds for self-healing            |
| controller.status.processors           | 20                              | Number of processors for status updates       |
| otlp.address                          | ""                              | Address for OTLP (OpenTelemetry Protocol)      |
| redis.server                          | argocd-redis:6379                | Address for the Redis server                   |
| repo.server                           | argocd-repo-server:8081          | Address for the repository server               |
| reposerver.log.format                 | text                            | Log format for the repository server           |
| reposerver.log.level                  | info                            | Log level for the repository server            |
| reposerver.parallelism.limit          | 0                               | Parallelism limit for the repository server    |
| server.basehref                        | /                               | Base URL for the server                         |
| server.disable.auth                   | "false"                         | Disable authentication for the server          |
| server.enable.gzip                    | "false"                         | Enable GZIP compression for the server         |
| server.insecure                       | "false"                         | Enable insecure mode for the server            |
| server.log.format                     | text                            | Log format for the server                      |
| server.log.level                      | info                            | Log level for the server                       |
| server.rootpath                       | ""                              | Root path for the server                       |
| server.staticassets                   | /shared/app                     | Directory for static assets                    |
| server.x.frame.options                | sameorigin                      | X-Frame-Options header value                   |


# argocd-rbac-cm
| Key                   | Value      | Description                 |
|-----------------------|------------|-----------------------------|
| policy.default        | role:admin | Default role configuration |


# argocd-ssh-known-hosts-cm
| Key                                      | Value                                                     | Description |
|------------------------------------------|-----------------------------------------------------------|-------------|
| bitbucket.org                            | ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAubiN81eDcafrgMeLzaFPsw2kNvEcqTKl/VqLat/MaB33pZy0y3rJZtnqwR2qOOvbwKZYKiEO1O6VqNEBxKvJJelCq0dTXWT5pbO2gDXC6h6QDXCaHo6pOHGPUy+YBaGQRGuSusMEASYiWunYN0vCAI8QaXnWMXNMdFP3jHAJH0eDsoiGnLPBlBp4TNm6rYI74nMzgz3B9IikW4WVK+dc8KZJZWYjAuORU3jc1c/NPskD2ASinf8v3xnfXeukU0sJ5N6m5E8VLjObPEO+mN2t/FZTMZLiFqPWc/ALSqnMnnhwrNi2rbfg/rd/IpL8Le3pSBne8+seeFVBoGqzHM9yXw?== | SSH Key for bitbucket.org |
| github.com                               | ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg= | ECDSA SSH Key for github.com |
| github.com                               | ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl | SSH-Ed25519 Key for github.com |
| github.com                               | ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84Kezm... | SSH-RSA Key for github.com |
| gitlab.com                               | ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFSMqzJeV9rUzU4kWitGjeR4PWSa29SPqJ1fVkhtj3Hw9xjLVXVYrU9QlYWrOLXBpQ6KWjbjTDTdDkoohFzgbEY= | ECDSA SSH Key for gitlab.com |
| gitlab.com                               | ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAfuCHKVTjquxvt6CM6tdG4SLp1Btn/nOeHHE5UOzRdf | SSH-Ed25519 Key for gitlab.com |
| gitlab.com                               | ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCsj2bNKTBSpIYDEGk9KxsGh3mySTRgMtXL583qmBpzeQ+jqCMRgBqB98u3z++J1sKlXHWfM9dyhSevkMwSbhoR8XIq/U0tCNyokEi/ueaBMCvbcTHhO7FcwzY92WK4Yt0aGROY5qX2UKSeOvuP4D6TPqKF1onrSzH9b... | SSH-RSA Key for gitlab.com |
| ssh.dev.azure.com                        | ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7Hr1oTWqNqOlzGJOfGJ4NakVyIzf1rXYd4d7wo6jBlkLvCA4odBlL0mDUyZ0/QUfTTqeu+tm22gOsv+VrVTMk6vwRU75gY/y9ut5Mb3bR5BV58dKXyq9A9UeB5Cakehn5Zgm6x1mKoVyf+FFn26iYqXJRgzIZZcZ5V6hrE0Qg3... | SSH-RSA Key for ssh.dev.azure.com |
| vs-ssh.visualstudio.com                 | ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7Hr1oTWqNqOlzGJOfGJ4NakVyIzf1rXYd4d7wo6jBlkLvCA4odBlL0mDUyZ0/QUfTTqeu+tm22gOsv+VrVTMk6vwRU75gY/y9ut5Mb3bR5BV58dKXyq9A9UeB5Cakehn5Zgm6x1mKoVyf+FFn26iYqXJRgzIZZcZ5V6hrE0Qg3... | SSH-RSA Key for vs-ssh.visualstudio.com |




# devtron-operator-cm
| Key                              | Value      | Description              |
|----------------------------------|------------|--------------------------|
| BLOB_STORAGE_PROVIDER            | ""         | Blob storage provider.   |
| DEVTRON_HELM_RELEASE_NAME        | devtron    | Helm release name for Devtron. |
| ENABLE_LEGACY_API                | "false"    | Enable legacy API.       |
| INSTALLATION_THROUGH_HELM        | "True"     | Installation through Helm. |
| APP                              | orchestrator | Application name.      |
| MODE                             | PROD      | Operating mode.         |

# migrator-override-cm
| Key      | Value           | Description             |
|----------|-----------------|-------------------------|
| override      | ""    |        |

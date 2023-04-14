
[![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)](https://discord.gg/72JDKy4)

# Devtron Installation

Devtron is an open source software delivery workflow for Kubernetes written in Go language. It is designed as a self-serve platform for operationalizing and maintaining applications (AppOps) on Kubernetes in a developer friendly way.

## Introduction

This chart bootstraps deployment of all required components for installation of [Devtron Platform](https://github.com/devtron-labs) on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

It packages third party components like 

 - [Grafana](https://github.com/grafana/grafana) for displaying application metrics 
 - [Argocd](https://github.com/argoproj/argo-cd/) for gitops 
 - [Argo workflows](https://github.com/argoproj/argo) for CI
 - [Clair](https://github.com/quay/clair) & [Guard](https://github.com/guard/guard) for image scanning
 - [External Secret Operator](https://github.com/external-secrets/external-secrets) and [Kubernetes External Secrets (Depricated) ](https://github.com/godaddy/kubernetes-external-secrets) for integrating with external secret management systems like [AWS Secrets Manager](https://aws.amazon.com/secrets-manager/) or [HashiCorp Vault](https://www.vaultproject.io/)
 - [Nats](https://github.com/nats-io) for event streaming
 - [Postgres](https://github.com/postgres/postgres) as datastore
 - Fork of [Argo Rollout](https://github.com/argoproj/argo-rollouts) 

## How to use it

### Install with Helm

Run the following command to install the latest version of Devtron along with the CI/CD integration:

```bash
helm repo add devtron https://helm.devtron.ai 

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd}
```

[Link to](https://docs.devtron.ai/install) more installation options for Devtron.

For more details on configuration, see the [Helm chart configuration](#configuration)

### Access Devtron dashboard

Devtron dashboard in now available at the `BASE_URL/dashboard`, where `BASE_URL` is same as provided in `values.yaml` in case of installation via Helm chart.

#### Check your current devtron version

If you want to check the current version of Devtron you are using, please use the following command.

```
kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.data}' | grep "^LTAG=" | cut -d"=" -f2-
```

#### For Devtron version v0.6.0 and higher

Use username:`admin` and for password run command mentioned below.
```bash
kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ADMIN_PASSWORD}' | base64 -d
```

#### For Devtron version less than v0.6.0

Use username:`admin` and for password run command mentioned below.
```bash
kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
```

### Configuration

All parameters mentioned in the `values.yaml` are mandatory.

First section is ***secrets.env*** and it has following properties
| Parameter | Description | Default |
|----------:|:------------|:--------|
| **POSTGRESQL_PASSWORD*** | password for postgres database (required) | change-me |
| **GIT_TOKEN** | git token for the gitops work flow, please note this is not for source code of repo and this token should have full access to create, delete, update repository (required) |  |
| **WEBHOOK_TOKEN** | If you want to continue using jenkins for CI then please provide this for authentication of requests (optional)  |  |

Second section is ***configs*** and has following properties
| Parameter | Description | Default |
|----------:|:------------|:--------|
| **BASE_URL_SCHEME** | either of http or https | http |
| **BASE_URL** | url without scheme and trailing slash (required) `eg. devtron.ai` | `change-me` |
| **DEX_CONFIG** | dex config if you want to integrate login with SSO (optional) for more information check [Argocd documentation](https://argoproj.github.io/argo-cd/operator-manual/user-management/) | 
| **GIT_PROVIDER** | git provider for storing config files for gitops, currently only GITHUB and GITLAB are supported | `GITHUB` | |
| **GITLAB_NAMESPACE_ID** | if GIT_PROVIDER is GITLAB | | 
| **GITLAB_NAMESPACE_NAME** | if GIT_PROVIDER is GITLAB | |
| **GIT_USERNAME** | git username for the GIT_PROVIDER | |
| **GITHUB_ORGANIZATION** | if GIT_PROVIDER is GITHUB | |
| **DEFAULT_CD_LOGS_BUCKET_REGION** | AWS region of bucket to store CD logs | |
| **DEFAULT_CACHE_BUCKET** | AWS bucket to store docker cache |  |
| **DEFAULT_CACHE_BUCKET_REGION** | AWS region of cache bucket defined in previous step | |
| **DEFAULT_BUILD_LOGS_BUCKET** | AWS bucket to store build logs | |
| **CHARTMUSEUM_STORAGE_AMAZON_BUCKET** | AWS bucket to store charts |  |
| **CHARTMUSEUM_STORAGE_AMAZON_REGION** | AWS region for bucket defined in previous step to store charts | |
| **EXTERNAL_SECRET_AMAZON_REGION** | AWS region for secret manager to pick |  |
| **PROMETHEUS_URL** | url of prometheus where all cluster data is stored, if this is wrong, you will not be able to see application metrics like cpu, ram, http status code, latency and throughput |  |

example of DEX_CONFIG is

    DEX_CONFIG: |-
      connectors:
        - type: oidc
          id: google
          name: Google
          config:
            issuer: https://accounts.google.com
            clientID: xxxxxxxx-qwwdsdsqsxxxxxxxxx.apps.googleusercontent.com
            clientSecret: fssdsdw121wwxssd
            redirectURI: <BASE_URL_SCHEME>://<BASE_URL>/api/dex/callback
            hostedDomains:
            - abc.com

**Please Note:**
Ensure that the cluster has access to the DEFAULT_CACHE_BUCKET, DEFAULT_BUILD_LOGS_BUCKET, CHARTMUSEUM_STORAGE_AMAZON_BUCKET and AWS secrets backends (SSM & secrets manager)


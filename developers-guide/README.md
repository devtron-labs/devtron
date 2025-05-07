# Development Guide for Devtron

This document provides instructions for setting up a development environment to help you get started with contributing to [Devtron](github.com/devtron-labs/devtron). Make sure to review the [contributing.md](https://github.com/devtron-labs/devtron/blob/main/CONTRIBUTING.md) file for more details on contributing. This document will help guide you through understanding the build process.

## Requirements
- [Go Programming language](https://go.dev/)
- [Docker](https://www.docker.com/) 
- [Kubernetes Cluster](https://kubernetes.io/)
- [Helm v3](https://helm.sh/)

## Setup Guide

### Install Devtron in a Kubernetes cluster

1. [Create a Kubernetes Cluster](https://docs.devtron.ai/getting-started#create-a-kubernetes-cluster)
   - 2 vCPUs
   - 4GB+ of free memory
   - 20GB+ free disk space

2. [Install Devtron with CI/CD along with GitOps (Argo CD) - Full mode](https://docs.devtron.ai/install/install-devtron-with-cicd-with-gitops)

    ```bash
    helm repo add devtron https://helm.devtron.ai
    helm repo update devtron
    helm install devtron devtron/devtron-operator \
      --create-namespace --namespace devtroncd \
      --set installer.modules={cicd} \
      --set argo-cd.enabled=true
    ```

    Note: If you want to install Devtron on a Managed Kubernetes cluster, please [refer to the guide](https://docs.devtron.ai/install/demo-tutorials).

3. Download the kubeconfig file to access your cluster. It’s a good idea to set it up locally and configure the context to point to the specific cluster you’re working with.

4. Forward the necessary service ports (e.g., PostgreSQL, NATS) for local access
    
   ```
       #Required
       kubectl -n devtroncd port-forward svc/devtron-nats 4222:4222  # Allows event driven communictation between microservices.   
       kubectl -n devtroncd port-forward svc/postgresql-postgresql 5432:5432 #Allows your local code to connect to the database running in the cluster.
   
       #Optional
       kubectl -n devtroncd port-forward svc/argocd-server 8000:80 #Required to interact with ArgoCD's API (GitOps)
       kubectl -n devtroncd port-forward svc/argocd-dex-server 5556:5556 #Required for SSO Logins.
       kubectl -n devtroncd port-forward svc/kubelink-service 50051:50051 #Required for no-gitops (helm) and fetches the status of resources.
       kubectl -n devtroncd port-forward svc/git-sensor-service 7070:80 # Monitors Git repositories for changes and triggers appropriate workflows in Devtron.
   ```

5. Use this command to extract the password of PostgreSQL (Required to put in ENV):

    ```bash
    kubectl exec -it -n devtroncd postgresql-postgresql-0 -- printenv | grep POSTGRES_PASSWORD
    ```

### Prepare Your Local Workspace

Navigate to the `github.com` directory on your local machine:

- `cd go/src/github.com`
- `git clone http://github.com/devtron-labs/devtron`
- Configure [Environment Variables](https://github.com/devtron-labs/devtron/blob/main/scripts/dev-conf/envfile.env)
- Install [wire](ttps://github.com/google/wire) (Required)
- Run the server: 

    ```bash
    make run
    ```

This will start your server on `localhost:8080`.

## Cleanup

   ```
   1. Uninstall Devtron and delete the 'devtroncd' namespace
         helm uninstall devtron -n devtroncd
         kubectl delete namespace devtroncd 
         
   2. Remove the local Devtron codebase
         rm -rf ~/go/src/github.com/devtron-labs/devtron
   ```
## Contributing
Please check [contributing guidelines](https://github.com/devtron-labs/devtron/blob/main/CONTRIBUTING.md)

## Need help?

If you’re looking for the fastest response to your questions, we encourage you to visit our Discord community. Specifically, you can post your inquiries in the [#setup-and-installation](https://discord.com/channels/769482988882493450/801441246849007667) channel, where our team and community members are actively available to assist you on discord channel.




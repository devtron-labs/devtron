# Glossary

### Artifacts

An immutable blob of data generated as an output after the execution of a job, build, or deployment process, e.g., container image, helm chart. In Devtron, you can view the artifacts in the `Build History` and `Deployment History` of your application. Whereas, job artifacts are visible in the `Run history` of your job.

* Once a build is complete, you can view the build artifacts by going to Applications (choose your app) → Build History (tab) → (choose a pipeline and date of triggering the build) → Artifacts (tab).

* Once a deployment is complete, you can view the deployment artifacts by going to Applications (choose your app) → Deployment History (tab) → (choose an environment and date of deployment) → Artifacts (tab).

* Once a job is complete, you can view the job artifacts by going to Jobs → Run history (tab) → (choose a pipeline and date of triggering the build) → Artifacts (tab).

### Base Deployment Template

A deployment template is a manifest of the application defining its runtime behavior. You can select one of the default deployment charts or custom deployment charts created by super-admin.

It’s a single entry point for you to enter the values, so that when the application is deployed your filled values go to the respective template files (YAML), and accordingly the resources would be created. 

In Devtron, you get the option to select a base deployment template in the `App Configuration` tab at the time of creating an application. [Read More...](https://docs.devtron.ai/usage/applications/creating-application/deployment-template)

### Build Context

For building a docker image we require a [Dockerfile](#dockerfile) and a build context. The Dockerfile contains the instructions to build. Context is the path where the build process may refer for getting the files required for build. 

To build files from the root, use (.) as the build context. Or to refer a subdirectory, enter the path in the format `/myfolder` or `/myfolder/mysubfolder`. If the path is not set, the default path will be the root directory of selected git repository.

Go to Applications (choose your app) → App Configuration (tab) → Build Configuration → (choose 'I have a Dockerfile') → Set Build Context.

### Build Pipeline

A series of automated steps that transform source code into a deployable container image. In Devtron, you can create a build pipeline by going to Applications (choose your app) → App Configuration (tab) → Workflow Editor → New Workflow. [Read More...](https://docs.devtron.ai/usage/applications/creating-application/ci-pipeline)

### Chart Store

A place where all Helm charts are centrally listed for users to deploy applications on Kubernetes. In Devtron, the chart store is available in the left sidebar. You can view, configure, and deploy the existing charts or add new chart repositories too. [Read More...](https://docs.devtron.ai/global-configurations/chart-repo)

### Cluster

A cluster in Kubernetes refers to a set of connected computers (nodes) that collectively manage containerized applications using Kubernetes. It provides resources and services to run, manage, and scale applications. 

In Devtron, you can view the list of clusters in ‘Global Configurations’ available in the left sidebar. [Read More...](https://docs.devtron.ai/usage/clusters)

### Commit Hash

A unique identifier representing a specific version of source code in a Git [repository](#repo). In Devtron, you can view the commit hash of the top 15 commits you pushed to your branch while selecting the git material under the `Build & Deploy` tab of your application.

### ConfigMaps

Kubernetes objects used to store configuration data as key-value pairs. They allow separation of configuration from application code, making it easier to manage and update settings. 

You can use different ConfigMaps for respective environments too. [Read More...](https://docs.devtron.ai/usage/applications/creating-application/config-maps)

### Container/OCI Registry

It is a collection of repositories that store container images. It allows developers to store, share, and manage images used to deploy containers. In Devtron, you can add a container registry by going to Global Configurations → Container / OCI Registry. Your CI images are pushed to the container registry you configure. [Read More...](https://docs.devtron.ai/global-configurations/container-registries). 

An OCI-compliant registry can also store artifacts (such as helm charts). Here, OCI stands for Open Container Initiative. It is an open industry standard for container formats and registries.

### Cordoning

Temporarily marking a node as unschedulable, preventing new pods from being assigned to it. In Devtron, you can cordon a node by going to Resource Browser → (choose a cluster) → Nodes → (click on a node) → Cordon (available in blue). [Read More...](https://docs.devtron.ai/usage/clusters#cordon-a-node)

### CronJob

CronJob is used to create Jobs on a repeating schedule. It is commonly used for running periodic tasks with no manual intervention. In Devtron, you can view a list of cronjobs by going to Resource Browser → (choose a cluster) → Workloads → CronJob. [Read More...](https://docs.devtron.ai/usage/applications/creating-application/deployment-template/job-and-cronjob#2.-cronjob)

### Custom Charts

Devtron offers a variety of ready-made Helm charts for common tasks and functions. If you have a specific need that isn't met by these preconfigured charts, super-admins have the permission to upload their own custom charts. Once uploaded, these custom charts become accessible for use by all users on the Devtron platform. [Read More...](https://docs.devtron.ai/global-configurations/custom-charts)

### DaemonSet

A Kubernetes object that ensures a specific pod runs on all or certain nodes within a cluster, often used for tasks such as logging or monitoring. 

In Devtron, you can view a list of DaemonSets by going to Resource Browser → (choose a cluster) → Workloads → DaemonSet.

### Deployment Strategy

A defined approach for deploying updates or changes to applications. Devtron supports rolling updates, blue-green deployments, canary releases, and recreate strategy. 

In Devtron, you can choose a deployment strategy by going to Applications (choose your app) → App Configuration (tab) → Workflow Editor → (edit deployment pipeline) → Deployment Strategy. [Read More...](https://docs.devtron.ai/usage/applications/creating-application/cd-pipeline#deployment-strategies)

### Devtron Agent

Your Kubernetes cluster gets mapped with Devtron when you save the cluster configurations. Now, the Devtron agent (rollout controller) must be installed from the chart store on the added cluster so that you can deploy your applications on that cluster. [Read More...](https://docs.devtron.ai/global-configurations/cluster-and-environments#installing-devtron-agent)

### Devtron Apps

Devtron Apps are the micro-services deployed using Kubernetes-native CI/CD with Devtron. To create one, go to Applications → Create (button) → Custom App.

### Dockerfile

A script that defines how to build a Docker [container image](#image). It includes instructions to assemble the image's base, dependencies, and application code. It's recommended that you include a Dockerfile with your source code. 

However, in case you don't have a Dockerfile, Devtron helps you create one. Go to Applications (choose your app) → App Configuration (tab) → Build Configuration. [Read More...](https://docs.devtron.ai/usage/applications/creating-application/docker-build-configuration#build-docker-image-by-creating-dockerfile)

### Draining

Evacuating pods from a node before cordoning it, ensuring that running pods are gracefully rescheduled on other nodes. 

In Devtron, you can drain a node by going to Resource Browser → (choose a cluster) → Nodes → (click on a node) → Drain (available in blue). [Read More...](https://docs.devtron.ai/usage/clusters#drain-a-node)

### Environment

You can deploy your application to one or more environments (e.g., development, testing, production). In Devtron, Environment = [Cluster](#cluster) + [Namespace](#namespace). For a given application, you cannot have multiple CD pipelines for an environment. For e.g., if an application named 'test-app' is deployed on an environment named 'test-environment', you cannot create another deployment (CD) pipeline for the same app and environment.
 
Your application can have different deployment configurations for respective environments. For e.g., the number of [ReplicaSet](#replicaset) could be 2 for staging environment, whereas it could be 5 for production.

Similarly, the CPU and memory resources can be different for each environment. This is possible through Environment Overrides. [Read More...](https://docs.devtron.ai/usage/applications/creating-application/environment-overrides)

### External Links

You can add external links related to the application. For e.g., you can add Prometheus, Grafana, and many more to your application by going to Global Configurations → External Links. [Read More...](https://docs.devtron.ai/global-configurations/external-links)

### GitOps

A methodology for managing and automating Kubernetes deployments using Git repositories as the source of truth. Changes to the desired state of the cluster are driven by Git commits. [Read More...](https://docs.devtron.ai/global-configurations/gitops)

### Helm Apps

Apps deployed using Helm Chart from the `Chart Store` section of Devtron. In Devtron, you can view such apps under a tab named `Helm Apps` in the Applications section. To create one, go to Applications → Create (button) → From Chart store.

### Helm Charts/Packages

Packages that contain pre-configured Kubernetes resources and configurations. Helm charts are used to define, install, and upgrade applications on Kubernetes clusters. Refer [chart store](#chart-store) to know more.

### Image

A packaged and standalone software that contains the code and dependencies needed to run a containerized application. Using Devtron, you can build a container image of your application, push it to a container registry, and deploy it on your Kubernetes cluster. 

Since images are platform-agnostic, you don't have to worry about compiling your application to work on different systems. With Devtron, you can enable automatic image builds and vulnerability scanning whenever you make edits to your source code. [Read More...](https://docs.devtron.ai/usage/applications/creating-application/ci-pipeline)

You can also view the list of image builds while preparing your deployment in the `Build & Deploy` tab of your application (provided the CI stage is successful).

### Job

In Devtron, there is a job that is very similar to Kubernetes job. A Kubernetes job is an object used to create one or more pods to complete a specific task or job and then terminate. 

If you are a super-admin in Devtron, you can view Jobs in the sidebar.

### Load Balancer

Distributes incoming network traffic across multiple instances or nodes to ensure efficient resource utilization and improved performance. In Kubernetes, Load Balancer is a service type. Behind the scenes, the managed Kubernetes service connects to the load balancer service of the respective cloud service provider and creates a load balancer, mapping it to the Kubernetes service. 

GKE and AKE provide the public IP of the Load Balancer as the service endpoint, while in the case of EKS, it provides a non-customizable DNS name.

### Manifest

A manifest is a YAML file that describes each component or resource of your Kubernetes object and the state you want your cluster to be in once applied. A manifest specifies the desired state of an object that Kubernetes will maintain when you apply the manifest. 

In Devtron, you can view the manifest of K8s resources under `App Details` and also under `Resource Browser`.

### Material

In Git Repo, the source code of your application in a given commit is referred as material. The option to choose a Git material will be available in the CI stage under the `Build & Deploy` tab of your application. [Read More...](https://docs.devtron.ai/usage/jobs/triggering-job#triggering-job-pipeline)

### Namespace

A namespace is a way to organize and isolate resources within a Kubernetes cluster. It provides a logical separation between different applications or environments within a cluster. 

In Devtron, you can view a list of namespaces by going to Resource Browser → (choose a cluster) → Namespaces.

### Node Taint

A setting applied to a node that influences the scheduling of pods. Taints can restrict which pods are allowed to run on the node. 

In Devtron, you can edit the taints of a node by going to Resource Browser → (choose a cluster) → Nodes → (click on a node) → Edit taints (available in blue). [Read More...](https://docs.devtron.ai/usage/clusters#taint-a-node)

### NodePort

A Kubernetes service type that exposes a port on each node in the cluster, making a service accessible externally.

### Nodes

The physical or virtual machines that make up a Kubernetes cluster, where containers are scheduled to run. 

In Devtron, you can view nodes by going to Resource Browser → (choose a cluster) → Nodes. [Read More...](https://docs.devtron.ai/usage/clusters#nodes)

### Pod

The smallest deployable unit in Kubernetes, consisting of one or more containers that share storage and network resources within the same context. 

In Devtron, you can view a list of Pods by going to Resource Browser → (choose a cluster) → Workloads → Pod. In Devtron, you can create a pod by going to Resource Browser → Create Resource (button).

### Pre-build

Actions or processes performed before the actual image-building process in a containerized application's deployment pipeline, e.g., Jira Issue Validator. 

In Devtron, you can configure pre-build actions by going to Applications (choose your app) → App Configuration (tab) → Workflow Editor → (edit build pipeline) → Pre-build stage (tab) → Add task (button). [Read More...](https://docs.devtron.ai/usage/applications/creating-application/ci-pipeline/ci-build-pre-post-plugins#configuring-pre-post-build-tasks)

### Post-build

Actions or processes performed after the [image](#image) building process in a containerized application's deployment pipeline, e.g., email notification about build status. 

In Devtron, you can configure post-build actions by going to Applications (choose your app) → App Configuration (tab) → Workflow Editor → (edit build pipeline) → Post-build stage (tab) → Add task (button). [Read More...](https://docs.devtron.ai/usage/applications/creating-application/ci-pipeline/ci-build-pre-post-plugins#configuring-pre-post-build-tasks)

### Pre-deployment

Steps, scripts, or configurations executed before deploying a new version of an application to a Kubernetes cluster. 

In Devtron, you can configure pre-deployment actions by going to Applications (choose your app) → App Configuration (tab) → Workflow Editor → (edit deployment pipeline) → Pre-deployment stage (tab) → Add task (button). [Read More...](https://docs.devtron.ai/usage/applications/creating-application/cd-pipeline#3.-pre-deployment-stage)

### Post-deployment

Actions, checks, or processes carried out after a new version of an application is successfully deployed to a Kubernetes cluster, e.g., Jira Issue Updater. 

In Devtron, you can configure post-deployment actions by going to Applications (choose your app) → App Configuration (tab) → Workflow Editor → (edit deployment pipeline) → Post-deployment stage (tab) → Add task (button). [Read More...](https://docs.devtron.ai/usage/applications/creating-application/cd-pipeline#6.-post-deployment-stage)

### ReplicaSet

A Kubernetes object responsible for maintaining a specified number of replica pods, ensuring high availability and desired scaling. 

In Devtron, you can view the deployed ReplicaSet by going to Applications (choose your app) → App Details (tab) → K8s Resources (under Application Metrics section).

### Repo

Abbreviation for "repository". It could either signify a Git repo, container repo, or helm repo.

**Git repo** - A version control system (like Git) that stores and manages source code and other project assets. Once you create a git repo, you can add it in Applications (choose your app) → App Configuration (tab) → Git Repository → Add Git Repository.

**Container repo** - A collection of [container images](#image), e.g., Docker repository.

**Helm repo** - Also known as chart repo. You can add it in Global Configurations.

### Rollback

The process of reverting a deployment to a previously known working version in case of errors or issues with the current version. 

In Devtron, you can rollback a deployment by going to Applications (choose your app) → Build & Deploy (tab) → (click the rollback icon in the deployment pipeline). [Read More...](https://docs.devtron.ai/usage/applications/deploying-application/rollback-deployment)

### Secrets

Kubernetes objects used to store sensitive information, such as passwords and API keys. Secrets are encoded and can be mounted as files or environment variables in pods. 

In Devtron, you get the option to add secrets in the `App Configuration` tab of your application. You can use different secrets for respective environments too. [Read More...](https://docs.devtron.ai/usage/applications/creating-application/secrets)

### Security Context

A Kubernetes resource configuration that defines security settings and permissions for pods and containers. A security context defines privilege and access control settings for a pod or container. [Read More...](https://docs.devtron.ai/usage/applications/creating-application/deployment-template/deployment#security-context)

### StatefulSet

A Kubernetes object designed for managing stateful applications, maintaining stable network identities and storage across pod rescheduling. 

In Devtron, view the list of StatefulSets by going to Resource Browser → (choose a cluster) → Workloads → StatefulSet. [Read More...](https://docs.devtron.ai/usage/applications/creating-application/deployment-template/statefulset)

### Target Platform

The operating system and architecture for which the [container image](#image) will be built, e.g., ubuntu/arm64, linux/amd64. The image will only be compatible to run only on the target platform chosen in the build configuration. 

In Devtron, you can choose the target platform by going to Applications (choose your app) → App Configuration (tab) → Build Configuration → (create build pipeline) → (click `Allow Override` button) → Target platform for the build (section). [Read More...](https://docs.devtron.ai/usage/applications/creating-application/docker-build-configuration)





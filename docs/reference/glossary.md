# Glossary

* **Artifacts**

    Any files or resources generated during the software development lifecycle, including source code, binary files, and compiled packages. In Devtron, you can view the artifacts in the `Build History` tab of your application.

* **Base Deployment Template**

    It’s a single entry point (an all-in-one YAML file) for you to enter the values, so that when the application is deployed the filled values go to the respective template files (YAML), and accordingly the resources would be created. In Devtron, you get the option to select a base deployment template in the `App Configuration` tab at the time of creating an application.

* **Build Context**

    The set of files and directories used as input when building a Docker container image using a Dockerfile. Specify the set of files to be built by referring to a specific subdirectory, relative to the root of your repository. To build all files from the root, use (.) as the build context, or set build context by referring a subdirectory path such as /myfolder or /myfolder/buildhere if path not set, default path will be root dir of selected git repository.

* **Build Pipeline**

    A series of automated steps that transform source code into a deployable container image, often integrated with testing, packaging, and artifact creation. In Devtron, you can create a build pipeline by going to Applications (choose your app) → App Configuration (tab) → Workflow Editor → New Workflow

* **Chart Store**

    A repository of Helm charts that can be used to deploy applications on Kubernetes. Chart Store provides a centralized location for sharing and distributing Helm charts. In Devtron, the chart store is available in the left sidebar.

* **Cluster**

    A cluster in Kubernetes refers to a set of connected computers (nodes) that collectively manage containerized applications using Kubernetes. It provides resources and services to run, manage, and scale applications. In Devtron, you can view the list of clusters in ‘Global Configurations’ available in the left sidebar. 

* **Commit Hash**

    A unique identifier representing a specific version of source code in a Git repository. In Devtron, you can view the commit hash for all the commits you have pushed to your branch while selecting the git material under the `Build & Deploy` tab of your application.

* **ConfigMaps**

    Kubernetes objects used to store configuration data as key-value pairs. They allow separation of configuration from application code, making it easier to manage and update settings. In Devtron, you get the option to add ConfigMaps in the `App Configuration` tab of your application.

* **Container Image**

    A standalone executable software package that includes the application's code, runtime, libraries, and dependencies needed to run the application. In Devtron, you can view the list of image builds while preparing your deployment in the `Build & Deploy` tab of your application (provided the CI stage is successful).

* **Container Registry**

    A repository for storing container images. It allows developers to store, share, and manage images used to deploy containers. In Devtron, you can add a container registry by going to Global Configurations → Container / OCI Registry. Your CI images are pushed to the container registry you configure.

* **Containerization**

    The practice of packaging and isolating applications and their dependencies into containers for consistent and portable deployment.

* **Cordoning**

    Temporarily marking a node as unschedulable, preventing new pods from being assigned to it. In Devtron, you can cordon a node by going to Resource Browser → (choose a cluster) → Nodes → (click on a node) → Cordon (available in blue)

* **CronJob**

    A Kubernetes object that creates pods at specified intervals based on a cron schedule, commonly used for running periodic tasks. In Devtron, you can view a list of CronJob by going to Resource Browser → (choose a cluster) → Workloads → CronJob


* **DaemonSet**

    A Kubernetes object that ensures a specific pod runs on all or certain nodes within a cluster, often used for tasks such as logging or monitoring. In Devtron, you can view a list of DaemonSet by going to Resource Browser → (choose a cluster) → Workloads → DaemonSet

* **Deployment Strategy**

    A defined approach for deploying updates or changes to applications, which may include rolling updates, blue-green deployments, canary releases, and recreate strategy. In Devtron, you can choose a deployment strategy by going to Applications (choose your app) → App Configuration (tab) → Workflow Editor → (edit deployment pipeline) → Deployment Strategy

* **Devtron Agent**

    Your Kubernetes cluster gets mapped with Devtron when you save the cluster configurations. Now, the Devtron agent (rollout controller) must be installed from the chart store on the added cluster so that you can deploy your applications on that cluster.

* **Devtron Apps**

    Apps deployed on Kubernetes cluster using the CI-CD feature of Devtron

* **Dockerfile**

    A script that defines how to build a Docker container image. It includes instructions to assemble the image's base, dependencies, and application code.

* **Draining**

    Evacuating pods from a node before cordoning it, ensuring that running pods are gracefully rescheduled on other nodes. In Devtron, you can drain a node by going to Resource Browser → (choose a cluster) → Nodes → (click on a node) → Drain (available in blue)

* **Environment**

    In Devtron, Environment = Cluster + Namespace. A deployment context for an application that defines variables, settings, and configurations that are specific to a certain stage (e.g., development, testing, production).

* **External Link**

    You can add external links like Prometheus, Grafana, and many more to your application by going to Global Configurations → External Links.

* **GitOps**

    A methodology for managing and automating Kubernetes deployments using Git repositories as the source of truth. Changes to the desired state of the cluster are driven by Git commits.

* **Helm Apps**

    Apps deployed using Helm Chart from the `Chart Store` section of Devtron. In Devtron, you can view such apps under a tab named `Helm Apps` in the Applications section.

* **Helm Charts/Packages**

    Packages that contain pre-configured Kubernetes resources and configurations. Helm charts are used to define, install, and upgrade applications on Kubernetes clusters.

* **Horizontal Scaling**

    Increasing or decreasing the number of instances or replicas of an application to handle varying levels of traffic and demand.

* **Image**

    A packaged and standalone software that contains the code and dependencies needed to run a containerized application.

* **Job**

    A Kubernetes object used to create one or more pods to complete a specific task or job and then terminate. In Devtron, you can view a list of Job by going to Resource Browser → (choose a cluster) → Nodes → (click on a node) → Drain (available in blue)

* **Load Balancing**

    Distributing incoming network traffic across multiple instances or nodes to ensure efficient resource utilization and improved performance.

* **Manifest**

    A human-readable YAML or JSON file that defines the desired state of Kubernetes resources, such as pods, services, and deployments.

* **Material**

    In Git Repo, the source code of your application in a given commit is referred is referred as material. The option to choose a material will be available in the CI stage under the `Build & Deploy` tab of your application.

* **Namespace**

    A namespace is a way to organize and isolate resources within a Kubernetes cluster. It provides a scope for names and helps avoid naming conflicts between different resources. Namespaces can be used to group related resources, manage access control, and provide logical separation between different applications or environments within a cluster. In Devtron, you can a list of namespaces by going to Resource Browser → (choose a cluster) → Namespaces

* **Node Taint**

    A setting applied to a node that influences the scheduling of pods. Taints can restrict which pods are allowed to run on the node. In Devtron, you can edit the taints of a node by going to Resource Browser → (choose a cluster) → Nodes → (click on a node) → Edit taints (available in blue)

* **NodePort**

    A Kubernetes service type that exposes a port on each node in the cluster, making a service accessible externally.

* **Nodes**

    The physical or virtual machines that make up a Kubernetes cluster, where containers are scheduled to run. In Devtron, you can view nodes by going to Resource Browser → (choose a cluster) → Nodes

* **Pod**

    The smallest deployable unit in Kubernetes, consisting of one or more containers that share storage and network resources within the same context. In Devtron, you can a list of Pods by going to Resource Browser → (choose a cluster) → Workloads → Pod. In Devtron, you can create a pod by going to Resource Browser → Create Resource (button)

* **Post-build**

    Actions or processes performed after the image building process in a containerized application's deployment pipeline, e.g. email notification about build status. In Devtron, you can configure post-build actions by going to Applications (choose your app) → App Configuration (tab) → Workflow Editor → (edit build pipeline) → Post-build stage (tab) → Add task (button)

* **Post-deployment**

    Actions, checks, or processes carried out after a new version of an application is successfully deployed to a Kubernetes cluster, e.g. Jira Issue Updater. In Devtron, you can configure post-deployment actions by going to Applications (choose your app) → App Configuration (tab) → Workflow Editor → (edit deployment pipeline) → Post-deployment stage (tab) → Add task (button)

* **Pre-build**

    Actions or processes performed before the actual image-building process in a containerized application's deployment pipeline, e.g. Jira Issue Validator. In Devtron, you can configure pre-build actions by going to Applications (choose your app) → App Configuration (tab) → Workflow Editor → (edit build pipeline) → Pre-build stage (tab) → Add task (button)

* **Pre-deployment**

    Steps, scripts, or configurations executed before deploying a new version of an application to a Kubernetes cluster. In Devtron, you can configure pre-deployment actions by going to Applications (choose your app) → App Configuration (tab) → Workflow Editor → (edit deployment pipeline) → Pre-deployment stage (tab) → Add task (button)

* **ReplicaSet**

    A Kubernetes object responsible for maintaining a specified number of replica pods, ensuring high availability and desired scaling.

* **Repo**

    Abbreviation for "repository" a version control system (like Git) that stores and manages source code and other project assets. Once you create a repo, you can add the repo in Applications (choose your app) → App Configuration (tab) → Git Repository → Add Git Repository

* **Rollback**

    The process of reverting a deployment to a previously known working version in case of errors or issues with the current version. In Devtron, you can rollback a deployment by going to Applications (choose your app) → Build & Deploy (tab) → (click the rollback icon in the deployment pipeline)

* **Secrets**

    Kubernetes objects used to store sensitive information, such as passwords and API keys. Secrets are encoded and can be mounted as files or environment variables in pods. In Devtron, you get the option to add secrets in the `App Configuration` tab of your application.

* **Security Context**

    A Kubernetes resource configuration that defines security settings and permissions for pods and containers. A security context defines privilege and access control settings for a pod or container.

* **StatefulSet**

    A Kubernetes object designed for managing stateful applications, maintaining stable network identities and storage across pod rescheduling. In Devtron, you can edit the taints of a node by going to Resource Browser → (choose a cluster) → Workloads → StatefulSet

* **Target Platform**

    The specific environment and namespace where an application is intended to be deployed. In Devtron, you can choose the target platform by by going to Applications (choose your app) → App Configuration (tab) → Workflow Editor → (edit deployment pipeline) → Deployment Stage (tab) → (choose an environment and add a namespace)

* **Vertical Scaling**

    Adjusting the capacity of an individual instance by changing its resources, such as CPU and memory.





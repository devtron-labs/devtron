# Glossary

* **Artifacts**: Any files or resources generated during the software development lifecycle, including source code, binary files, and compiled packages.

* **Base Deployment Template**: (Need Review) A predefined configuration for deploying an application, used as a starting point to ensure consistent deployment across environments.

* **Build Context**: The set of files and directories used as input when building a Docker container image using a Dockerfile.

* **Build Pipeline**: A series of automated steps that transform source code into a deployable container image, often integrated with testing, packaging, and artifact creation.

* **Chart Store**: A repository of Helm charts that can be used to deploy applications on Kubernetes. Chart Stores provide a centralized location for sharing and distributing Helm charts.

* **Cluster**: A cluster in Kubernetes refers to a set of connected computers (nodes) that collectively manage containerized applications using the Kubernetes platform. It provides resources and services to run, manage, and scale applications.

* **Commit Hash**: A unique identifier representing a specific version of source code in a Git repository.

* **ConfigMaps**: Kubernetes objects used to store configuration data as key-value pairs. They allow separation of configuration from application code, making it easier to manage and update settings.

* **Container Image**: A standalone executable software package that includes the application's code, runtime, libraries, and dependencies needed to run the application.

* **Container Registry**: A repository for storing container images. It allows developers to store, share, and manage images used to deploy containers.

* **Containerization**: The practice of packaging and isolating applications and their dependencies into containers for consistent and portable deployment.

* **Cordoning**: Temporarily marking a node as unschedulable, preventing new pods from being assigned to it.

* **Cron Job**: A Kubernetes object that creates pods at specified intervals based on a cron schedule, commonly used for running periodic tasks.

* **DaemonSet**: A Kubernetes object that ensures a specific pod runs on all or certain nodes within a cluster, often used for tasks such as logging or monitoring.

* **Deployment Strategy**: A defined approach for deploying updates or changes to applications, which may include rolling updates, blue-green deployments, canary releases, etc.

* **Devtron Agent**: (Need Review) A component of the Devtron platform responsible for managing and orchestrating applications, resources, and tools within a Kubernetes cluster.

* **Devtron Apps**: (Need Review) Applications managed and orchestrated using the Devtron platform within a Kubernetes cluster. Devtron Apps streamline the deployment and management of applications, tools, and processes.

* **Dockerfile**: A script that defines how to build a Docker container image. It includes instructions to assemble the image's base, dependencies, and application code.

* **Draining**: Evacuating pods from a node before cordoning it, ensuring that running pods are gracefully rescheduled on other nodes.

* **Environment**: A deployment context for an application that defines variables, settings, and configurations that are specific to a certain stage (e.g., development, testing, production).

* **External Link**:(Need Review) A reference to a resource, service, or URL outside the Kubernetes cluster.

* **GitOps**: A methodology for managing and automating Kubernetes deployments using Git repositories as the source of truth. Changes to the desired state of the cluster are driven by Git commits.

* **Helm Apps**: Applications managed and deployed using Helm, a package manager for Kubernetes. Helm Apps consist of Helm charts that describe the application's structure, dependencies, and configuration.

* **Helm Charts**: Packages that contain pre-configured Kubernetes resources and configurations. Helm charts are used to define, install, and upgrade applications on Kubernetes clusters.

* **Helm Packages**: Synonymous with Helm charts, these are bundles of Kubernetes resources and configurations designed for streamlined deployment and management.

* **Horizontal Scaling**: Increasing or decreasing the number of instances or replicas of an application to handle varying levels of traffic and demand.

* **Image**: A packaged and standalone software that contains the code and dependencies needed to run a containerized application.

-***Image Descriptor**: Metadata associated with a container image, providing information about the image, its version, and dependencies.

* **Job**: A Kubernetes object used to create one or more pods to complete a specific task or job and then terminate.

* **Load Balancing**: Distributing incoming network traffic across multiple instances or nodes to ensure efficient resource utilization and improved performance.

* **Manifest**: A human-readable YAML or JSON file that defines the desired state of Kubernetes resources, such as pods, services, and deployments.

* **Material**: (Need Review) The source code, configuration files, and other resources that constitute an application or service within a Kubernetes environment.

* **Namespace**: (Need Review) A logical partition within a Kubernetes cluster that allows multiple virtual clusters to coexist within the same physical cluster. Namespaces help organize and isolate resources, such as pods, services, and config maps.

* **Node Taint**: A setting applied to a node that influences the scheduling of pods. Taints can restrict which pods are allowed to run on the node.

* **NodePort**: A Kubernetes service type that exposes a port on each node in the cluster, making a service accessible externally.

* **Nodes**: The physical or virtual machines that make up a Kubernetes cluster, where containers are scheduled to run.

* **Pod**: The smallest deployable unit in Kubernetes, consisting of one or more containers that share storage and network resources within the same context.

* **Post-Build**: (Need Review) Actions or processes performed after the image building process in a containerized application's deployment pipeline.

* **Post-deployment**: (Need Review) Actions, checks, or processes carried out after a new version of an application is successfully deployed to a Kubernetes cluster.

* **Pre-Build**: (Need Review)Actions or processes performed before the actual image building process in a containerized application's deployment pipeline.

* **Pre-deployment**: (Need Review) Steps, scripts, or configurations executed before deploying a new version of an application to a Kubernetes cluster.

* **ReplicaSet**: A Kubernetes object responsible for maintaining a specified number of replica pods, ensuring high availability and desired scaling.

* **Repo**: Abbreviation for "repository," a version control system (like Git) that stores and manages source code and other project assets.

* **Rollback**: The process of reverting a deployment to a previous known working version in case of errors or issues with the current version.

* **Secrets**: Kubernetes objects used to store sensitive information, such as passwords and API keys. Secrets are encoded and can be mounted as files or environment variables in pods.

* **Security Context**: A Kubernetes resource configuration that defines security settings and permissions for pods and containers.

* **StatefulSet**: A Kubernetes object designed for managing stateful applications, maintaining stable network identities and storage across pod rescheduling.

* **Target Platform**: (Need Review) The specific Kubernetes environment, cluster, or namespace where an application or workload is intended to be deployed.

* **Vertical Scaling**: Adjusting the capacity of an individual instance by changing its resources, such as CPU and memory.

* **Virtual Cluster**: (Need Review) A logically separated subset of resources and configurations within a Kubernetes cluster, often created using namespaces to isolate applications and workloads.

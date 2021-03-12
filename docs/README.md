# Introduction

Devtron is an open source AppOps platform for kubernetes that simplifies application deployment and management for devops and developers. It offers to bring a layer of simplification for Kubernetes deployments.

[Website](https://devtron.ai/) · [Blog](https://devtron.ai/blog/) · [Join Discord](https://discord.gg/jsRG5qx2gp) · [Twitter](https://twitter.com/DevtronL)

## Why Devtron?

It is designed as a self-serve platform for operationalizing and maintaining applications \(AppOps\) on kubernetes in a developer friendly way.
Devtron enables developers and devops to work collaboratively on one single platform and solve deployment specific problems in real-time.

This is the **AppOps approach** - an applications and tools first way of looking at Kubernetes. Devtrom makes it possible to have deep integrations with all your existing open source as well as commercial software tools and thus, bring about a complete transformation in the way applications are deployed and managed on Kubernetes.

Devtron works on the full lifecycle of applications, from delving on the debuggability aspects with deeper integrations within Kubernetes environment.

![](.gitbook/assets/preview%20%281%29%20%282%29.gif)

## Some of the benefits  provided by devtron are:


### Zero code software delivery workflow

* Workflow which understands the domain of **kubernetes, testing, CD, SecOps** so that you dont have to write scripts
* Reusable and composable components so that workflows are easy to contruct and reason through

### Multi cloud deployment

* deploy to multiple kubernetes cluster
* test on aws cloud

  > coming soon: support for GCP and microsoft azure

### Easy dev-sec-ops integration

* Multi level security policy at global, cluster, environment and application for efficient hierarchical policy management
* Behavior driven security policy
* Define policies and exception for kubernetes resources
* Define policies for events for faster resolution

### Application debugging dashboard

* One place for all historical kubernetes events 
* Access all manifests securely for eg secret obfuscation 
* _**Application metrics**_ for cpu, ram, http status code and latency with comparison between new and old 
* _**Advanced logging**_ with grep and json search 
* Intelligent _**correlation between events, logs**_ for faster triangulation of issue 
* Auto issue identification 

### Enterprise Grade security and compliances

* Fine grained access control; control who can edit configuration and who can deploy.
* Audit log to know who did what and when
* History of all CI and CD events
* Kubernetes events impacting application
* Relevant cloud events and their impact on applications
* Advanced workflow policies like blackout window, branch environment relationship to secure build and deployment pipelines

### Gitops aware

* Gitops exposed through API and UI so that you dont have to interact with git cli
* Gitops backed by postgres for easier analysis
* Enforce finer access control than git

### Operational insights

* Deployment metrics to measure success of agile process. It captures mttr, change failure rate, deployment frequency, deployment size out of the box.
* Audit log to understand the failure causes
* Monitor changes across deployments and revert easily

## Key Features
### Self-service for Developers
* A level of abstraction that requires users to have minimal level of K8s understanding
* In-built blue-green and canary deployment
* Automated microservices management and integrated workflows for CI/CD
* Easy GitOps based deployments with automatic git sync using ArgoCD (Flux in roadmap) integration
* Detailed insights of code and deployment metrics
* One-click rollback
* Easy management of templates and manifests

### App Management and Stability
* Integrations with Prometheus for regular event monitoring and alerting and automated actions as specified
* A single pane view for alerts, metrics, events, and app metrics
* Quick and easy debugging features
* Detailed logs with pattern based filtering for error identification
* Push changes across all environments at once
* A single pane dashboard for aerial view of all aspects of all deployments
* Environment management with features like drift management, cloning, differential view and reconciliation support

### Cost-saving
* Enables easy Instance selection, instance type change & movement between instances
* Detailed cost visibility across namespaces, clusters, and nodes
* Time-based hibernation for cost saving
* Spot management and spot termination handling

### Security
* Define pipeline policies
* Fine grained access control beyond kubernetes RBAC with OIDC integration
* Audit logs for monitoring third-party integrations
* Scanning for vulnerabilities during the deployment stage and detailed reports of the same
* Source integration scanning
* Comprehensive hierarchical policy management system

## Compatibility notes

* Only AWS kubernetes cluster is supported as of now
* It uses modified version of [argo rollout](https://argoproj.github.io/argo-rollouts/).
* application metrics only works for k8s 1.16+

## Community

Get updates on Devtron's development and chat with the project maintainers, contributors and community members.

* Join the [Discord Community](https://discord.gg/jsRG5qx2gp) 
* Follow [@DevtronL on Twitter](https://twitter.com/DevtronL)
* Raise feature requests, suggest enhancements, report bugs at [GitHub issues](https://github.com/devtron-labs/devtron/issues)
* Read the [Devtron blog](https://devtron.ai/blog/)

## Contribute

Check out our [contributing guidelines](https://github.com/devtron-labs/devtron-documentation/tree/1c2b95254995286ac0c3e8379117eb82a7ed8407/CONTRIBUTING.md). Included are directions for opening issues, coding standards, and notes on our development processes.

## Vulnerability Reporting

We at Devtron take security and our users' trust very seriously. If you believe you have found a security issue in Devtron, please responsibly disclose by contacting us at security@devtron.ai.

## License

Devtron is available under the [Apache License, Version 2.0](https://github.com/devtron-labs/devtron-documentation/tree/1c2b95254995286ac0c3e8379117eb82a7ed8407/LICENSE/README.md)


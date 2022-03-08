# Overview

## Devtron 🚀
Devtron leverages popular open source tools to provide a No-Code SaaS like experience for creating Software Delivery workflows for Kubernetes. Devtron integrates seamlessly with multiple open-source tools to provide you an ecosystem for Kubernetes software delivery workflow, debugging, monitoring and holistic access management for the entire team.

> Do check the [Devtron Installation Guide ⎈](setup/install/README.md)


![](.gitbook/assets/preview%20%281%29%20%282%29.gif)


## Why Devtron?

We have seen various tools that are used to greatly increase the ease of using Kubernetes. However, using these tools simultaneously is painful, and hard to use. This is due to the fact these tools don't talk to each other for managing different aspects of application lifecycle; including CI, CD, security, cost, observability, stabilization.

Thus, we built Devtron to solve this problem!

<p align="center"><img src="../assets/readme-comic.png"></p>

Devtron is an open source modular product providing 'seamless', 'implementation agnostic uniform interface', integrated  with open source, and commercial tools across the entire life cycle. This is all achieved while focusing on a slick User Experience, including a self-serve model.
<br>
You can efficiently handle Security, Stability, Cost, and more in a unified experience.




### Devtron Features:

#### Zero code software delivery workflow for Kubernetes

* Workflow which understands the domain of **Kubernetes, testing, CD, SecOps** so that you dont have to write scripts
* Reusable and composable components so that workflows are easy to contruct and reason through

#### Multi cloud deployment

* Deploy to multiple Kubernetes clusters on multiple cloud/on-prem from one Devtron setup.
* Works for all cloud providers and on-premise Kubernetes clusters.


#### Easy DevSecOps integration

* Multi level security policy at global, cluster, environment and application for efficient hierarchical policy management
* Behavior driven security policy
* Define policies and exception for kubernetes resources
* Define policies for events for faster resolution

#### Application debugging dashboard

* One place for all historical kubernetes events
* Access all manifests securely for eg secret obfuscation
* _**Application metrics**_ for cpu, ram, http status code and latency with comparison between new and old
* _**Advanced logging**_ with grep and json search
* Intelligent _**correlation between events, logs**_ for faster triangulation of issue
* Auto issue identification

#### Enterprise Grade security and compliances

* Fine grained access control; control who can edit configuration and who can deploy.
* Audit log to know who did what and when
* History of all CI and CD events
* Kubernetes events impacting application
* Relevant cloud events and their impact on applications
* Advanced workflow policies like blackout window, branch environment relationship to secure build and deployment pipelines

#### GitOps aware

* GitOps exposed through API and UI so that you dont have to interact with git CLI
* GitOps backed by postgres for easier analysis
* Enforce finer access control than git

#### Operational insights

* Deployment metrics to measure success of agile process. It captures mttr, change failure rate, deployment frequency, deployment size out of the box.
* Audit log to understand the failure causes
* Monitor changes across deployments and revert easily

## Compatibility notes

* It uses modified version of [argo rollout](https://argoproj.github.io/argo-rollouts/).
* Application metrics only works for k8s 1.16+

---

## Hyperion 🦹

### Why Hyperion?
Hyperion is a lightweight Dashboard for Kubernetes deployments. Packed with full-fledged debugging features enabled with resource grouping for easier debugging for Development and Infra team.
You can also upgrade to Devtron from Hyperion to enjoy full stack features of Devtron.

> Do check the [Hyperion Installation Guide ⎈](hyperion/setup/install.md)

### Hyperion Features

#### Application-level resource grouping for easier Debugging
- Hyperion groups your deployed microservices and displays them in a slick UI for easier monitoring or debugging. Access pod logs and resource manifests right from the Hyperion UI and even edit them!

#### Centralized Access Management
- Give access to users on Project, Environment and App level and control the level of access with customizable View only and Edit access.

#### Manage and observe Multiple Clusters
- Manage access of all the Kubernetes clusters (hosted on multiple cloud/on-prem) right from one Hyperion setup.

#### View and Edit Kubernetes Manifests
- View and Edit all Kubernetes resources right from the Hyperion dashboard.


---

## Contribute

Check out our [contributing guidelines](https://github.com/devtron-labs/devtron/blob/main/CONTRIBUTING.md). Included are directions for opening issues, coding standards, and notes on our development processes.


## Community

Get updates on Devtron's development and chat with the project maintainers, contributors and community members.

* Join the [Discord Community](https://discord.gg/jsRG5qx2gp)
* Follow [@DevtronL on Twitter](https://twitter.com/DevtronL)
* Raise feature requests, suggest enhancements, report bugs at [GitHub issues](https://github.com/devtron-labs/devtron/issues)
* Read the [Devtron blog](https://devtron.ai/blog/)

## Vulnerability Reporting

We at Devtron take security and our users' trust very seriously. If you believe you have found a security issue in Devtron, please responsibly disclose by contacting us at security@devtron.ai.

## License

Devtron is available under the [Apache License, Version 2.0](https://github.com/devtron-labs/devtron/blob/main/LICENSE)
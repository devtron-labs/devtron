# Software Distribution Hub

## Introduction [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

Software Distribution Hub is a platform that simplifies the packaging, versioning, and delivery of your software products. By using it, you can manage your software release across multiple clients ([tenants](#release-versions)).

![Figure: Software Distribution Hub](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/sdh/sdh-eagle-eye.gif)


### When and Why to Use

Devtron's Software Distribution Hub Devtron is designed to be used in scenarios where:

* **Multi-Tenant Deployments**: You build software solutions for multiple clients (tenants) who require updates deployed to their distinct environments. For one client you may have to deploy a separate instance of your application, which has separate application layer and data layer on their infrastructure.

* **Complex Release Management**: You need a centralized solution to manage the deployment of software updates across various environments and ensure smooth collaboration between different teams involved in the release process.

* **Enhanced Visibility and Monitoring**: You need improved visibility into the release process for stakeholders such as Release Managers, Developers, DevOps teams, System Reliability Engineers (SREs), Tenant Points of Contact (POCs), and other stakeholders. They need to monitor releases, debug issues, and have access to current operational statuses.

* **Standardized Deployment Processes**: You want to standardize and automate deployment processes to ensure error-free releases and streamline operations across multiple tenant environments.

* **Comprehensive Release Documentation**: You require detailed documentation and checklists for each release, specifying versioning, dependencies, configuration changes, and deployment prerequisites to ensure consistent and successful deployments.

---

## Advantages

Devtron's Software Distribution Hub goes beyond basic deployment by providing end-to-end release management. Deployments involving manual processes might be prone to human error. However, Software Distribution Hub automates the [rollout](#rollout) process by enforcing [requirements](#requirements) for each release, and not just for one environment but multiple tenant environments.

### Normal Deployment vs SDH

| Aspect                                     | Normal Deployment                            | Software Distribution Hub                                      |
|--------------------------------------------|----------------------------------------------|----------------------------------------------------------------|
| **Release Management**                     | No versioned deployments                     | Centralizes versioning and deployment into a unified platform  |
| **Visibility**                             | Limited visibility (Apps/App Groups)         | Comprehensive visibility                                       |
| **Automated and Standardized Deployments** | Relies on manual scripts or basic automation (jobs) | Automates and standardizes deployments for consistency  |
| **End-to-End Release Management**          | Focuses on pushing changes quickly           | Manages entire release lifecycle from planning to post-release |

---

## Concepts

Devtron's Software Distribution Hub has 2 sections:

* [Tenants](./tenants.md)
* [Release Hub](./release-hub.md)


Feel free to familiarize yourself with the following concepts (terms) before you proceed to Software Distribution Hub.

### Tenants

Tenants are organizations or clients that use your software. You can think of each tenant as a separate customer. For example, if Microsoft and Google both use your software, they are considered separate tenants. Each tenant has its own environment to ensure their data and operations are kept separate from others.

### Installations

One installation is equivalent to one deployment of your software for a specific [tenant](#tenants). Each installation is a different instance of the software tailored to different stages of use. For example, Microsoft might have three installations: 

* **Production Installation**: Where the live environment of the software runs for everyday use.
* **Development Installation**: Used for testing new features and changes before they go live.
* **QA Installation**: Dedicated to quality assurance testing to ensure the software works correctly before it reaches users.

### Release Tracks

A Release Track in Devtron is where you organize software releases, similar to a project or application. Each release within a Release Track is a unique version of your software. For example, think of Kubernetes as a release track, with "v1.28.8" as one of its releases. This helps in managing different versions and updates of a software project in a structured manner, ensuring that all versions are tracked and organized within their respective tracks.

### Requirements

Requirements refer to specific steps that should be taken before deployment of applications. This includes selecting specific images for each application, specifying the [release order](#release-orderstage) of applications, adding release instructions, and locking these requirements to ensure readiness before a [rollout](#rollout). 

### Release Order/Stage

This is a part of [requirements](#requirements) where you decide the sequence in which applications are deployed. This ensures that all dependencies are met and the software is implemented correctly. For example, you might need to deploy a database update before deploying a new application feature that relies on that update. The release order and stages ensures that everything happens in the correct sequence.

### Rollout

The process of deploying a release to the tenant's environment involves all necessary steps and stages to ensure successful deployment.



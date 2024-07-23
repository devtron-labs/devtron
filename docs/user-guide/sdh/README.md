# Software Distribution Hub

## Introduction [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

Software Distribution Hub is a platform that simplifies the packaging, versioning, and delivery of your software products. By using it, you can manage your software release across multiple clients ([tenants](#release-versions)).

![Figure: Software Distribution Hub](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/sdh/sdh-eagle-eye.gif)


### When and Why to Use

Devtron's Software Distribution Hub is designed to be used in scenarios where:

* **Tenanted deployment**: You build software solutions for clients (tenants) who require updates deployed to their distinct environments. For every tenant, you may have to deploy a separate instance of your application, which has separate application layer and data layer on their infrastructure.

* **Complex Release Management**: You need a centralized solution to manage the deployment of software updates across various environments and ensure smooth collaboration between different teams involved in the release process.

* **Poor Visibility and Monitoring**: You need improved visibility into the release process for stakeholders such as Release Managers, Developers, DevOps teams, System Reliability Engineers (SREs), Tenant Points of Contact (POCs), and other stakeholders. They need to monitor releases, debug issues, and have access to current operational statuses.

* **Inconsistent Deployment Processes**: You need standardization because you face operational challenges and deviation in your release outcomes, while carrying out deployment processes across multiple tenant environments.

* **Insufficient Release Documentation**: You intend to eliminate confusion because your releases lack detailed documentation, including versioning, dependencies, configuration changes, and deployment prerequisites, causing inconsistencies and deployment failures.

---

## Advantages

Devtron's Software Distribution Hub goes beyond basic deployment by providing end-to-end release management. Deployments involving manual processes might be prone to human error. However, Software Distribution Hub streamlines the [rollout](#rollout) process by enforcing [requirements](#requirements) for each release, and not just for one environment but multiple tenant environments.

### Normal Deployment vs SDH

| Aspect                                     | Normal Deployment                            | Software Distribution Hub                                      |
|--------------------------------------------|----------------------------------------------|----------------------------------------------------------------|
| **Release Management**                     | No versioned deployments                     | Centralizes versioning and deployment into a unified platform  |
| **Visibility**                             | Limited visibility                           | Comprehensive visibility                                       |
| **Automated and Standardized Deployments** | Relies on manual scripts or basic automation (jobs) | Automates and standardizes deployments for consistency  |
| **End-to-End Release Management**          | Focuses on pushing changes quickly           | Manages entire release lifecycle from planning to post-release |
| **Collaboration**                          | Gap or siloed communication among teams	| Facilitates collaboration among developers, release managers, and other stakeholders |

---

## Concepts

Devtron's Software Distribution Hub has 2 sections:

* [Tenants](./tenants.md)
* [Release Hub](./release-hub.md)

Feel free to familiarize yourself with the following concepts (terms) before you proceed to Software Distribution Hub.

### Tenants

Tenants are organizations or clients that use your software. You can think of each tenant as a separate customer. For example, if Microsoft and Google both use your software, they are considered separate tenants. Each tenant has its own environment to ensure their data and operations are kept separate from others.

### Installations

One installation represents one deployment of your software for a specific [tenant](#tenants). Each installation serves as a separate instance of the software, customized for different stages of use or separate teams within your organization. For example, your organization might have three installations: 

* **Production Installation**: Where the live environment of the software runs for all your end-users.
* **Development Installation**: Used by your team of developers for testing new features and changes before they go live.
* **QA Installation**: Dedicated to quality assurance (team of testers) to ensure the software works correctly before it reaches users.

### Release Tracks

A Release Track in Devtron is where you organize software releases. Each release within a Release Track is a unique version of your software. For example, think of Kubernetes as a release track, with "v1.28.8" as one of its releases. This helps in managing different versions and updates of a software project in a structured manner, ensuring that all versions are tracked and organized within their respective tracks.

### Requirements

Requirements refer to specific steps that should be taken before deployment of applications. This includes selecting specific images for each application, specifying the [release order](#release-orderstage) of applications, adding release instructions, and locking these requirements to ensure readiness before a [rollout](#rollout). 

### Release Order/Stage

This is a part of [requirements](#requirements) where you decide the stages in which applications are deployed to ensure all dependencies are met. For example, you might need to deploy a backend service before a frontend feature that depends on it. In such a case, release order ensures that backend applications are deployed in the first stage, followed by frontend applications, ensuring a smooth and coordinated rollout.

### Rollout

It is a process of delivering a new release to the tenant's environment. In Software Distribution Hub, this comes right after you lock the basic requirements of a release (i.e., application selection, release order, image selection, and release instructions).


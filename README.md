<p align="center">
<picture>
  <source media="(prefers-color-scheme: dark)"  srcset="./assets/devtron-darkmode-logo.png">
  <source media="(prefers-color-scheme: light)"  srcset="./assets/devtron-lightmode-logo.png">
  <img width="333.333" height="260" src="./assets/devtron-logo-dark-light.png">
</picture>
<h1 align= "center">Kubernetes Dashboard for a Centralized DevOps Hub</h1>
</p>
 
<p align="center">
<br>
<a href="https://docs.dashboard.devtron.ai/" rel="nofollow"><strong>«Explore Documentation»</strong></a> <strong>||</strong>
<a href="https://preview.devtron.ai/dashboard/" rel="nofollow"><strong>«Try Devtron Demo»</strong></a>
<br>
<a href="https://devtron.ai/">Website</a>
·
<a href="https://devtron.ai/blog/">Blogs</a>
·
<a href="https://discord.gg/jsRG5qx2gp">Join Discord channel</a>
·
<a href="https://twitter.com/DevtronL">Twitter</a>
.
<a href="https://www.youtube.com/channel/UCAHRp9qp0z1y9MMtQlcFtcw">YouTube</a>
 

</p>
<p align="center">
<a href="https://discord.gg/jsRG5qx2gp"><img src="https://img.shields.io/discord/687207715902193673?logo=discord&label=Discord&color=5865F2&logoColor=white" alt="Join Discord"></a>
<a href="https://goreportcard.com/badge/github.com/devtron-labs/devtron"><img src="https://goreportcard.com/badge/github.com/devtron-labs/devtron" alt="Go Report Card"></a>
<a href="./LICENSE"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License"></a>
<a href="./CONTRIBUTING.md"><img src="https://img.shields.io/github/contributors/devtron-labs/devtron" alt="contributors"></a>
<a href="https://github.com/devtron-labs/devtron"><img src="https://img.shields.io/github/commit-activity/m/devtron-labs/devtron" alt="Commit Activity"></a>
<a href="https://bestpractices.coreinfrastructure.org/projects/4411"><img src="https://bestpractices.coreinfrastructure.org/projects/4411/badge" alt="CII Best Practices"></a>
<a href="http://golang.org"><img src="https://img.shields.io/badge/Made%20with-Go-1f425f.svg" alt="made-with-Go"></a>
<a href="http://devtron.ai/"><img src="https://img.shields.io/website-up-down-green-red/http/shields.io.svg" alt="Website devtron.ai"></a>
<a href="https://github.com/devtron-labs/devtron/releases"><img src="https://img.shields.io/github/v/release/devtron-labs/devtron"></a>
<a href="https://twitter.com/intent/tweet?text=Devtron%20helps%20in%20simplifying%20software delivery%20workflow%20for%20Kubernetes,%20check%20it%20out!!%20&hashtags=OpenSource,Kubernetes,DevOps,CICD,go&url=https://github.com/devtron-labs/devtron%0a"><img src="https://img.shields.io/twitter/url/http/shields.io.svg?style=social" alt="Tweet"></a>
<h1></h1>

Devtron's **extensible Kubernetes Dashboard** provides clear visibility into your Kubernetes clusters and streamlines Helm app management through a single, intuitive interface.<br> 
With built-in RBAC, it ensures secure access while offering integrated insights into workloads deployed via GitOps tools like **ArgoCD** and **FluxCD** across multiple clusters. Devtron creates a centralized DevOps hub, accelerating operations by up to 20x :rocket:

Check out the below video to experince the full power of the **Kubernetes Dashboard**.

<a href="https://youtu.be/oqCAB9b-SGQ?si=YoUJfHL43VXRU5wx">
<br>
<p align="center"><img src="./assets/dashboard.png"></p>
</a>

Out of the box, Devtron's Kubernetes Dashboard includes:
- [Helm Application Management](https://docs.dashboard.devtron.ai/user-guide/helm-apps) to streamline deploying, configuration, and management of Helm apps 
- [Resource Browser](https://docs.dashboard.devtron.ai/user-guide/resource-browser) to visualize and manage different cluster resources like Nodes, Pods, ConfigMaps, Custom Resource Definations (CRDs), etc
- [Single Sign On (SSO)](https://docs.dashboard.devtron.ai/operator-guide/authorization/sso-login-services) to simplify onboarding and authenticating team members.
- [Fine Grained RBAC](https://docs.dashboard.devtron.ai/operator-guide/authorization) to control the level of access users have to different Dashboard and Cluster resources.

[Devtron](#install-devtron) helps you deploy, observe, manage & debug existing Helm apps in all your clusters.

## Devtron Dashboard 

Devtron's Kubernetes Dashboard offers a variety of different features, giving you flexibility to manage your clusters as you desire.

### Features 🤩

- Simplified configuration and management for Helm applications with easy rollback
- Compare and reconcile configuration drifts across enviornments
- Manage Helm deployments to multiple Kubernetes enviornments
- A single pane of glass for Helm, ArgoCD and FluxCD applications across multiple clusters
- Centralized Dashboard to view and manage K8s resources across multiple clusters
- Comprehensive SSO and RBAC to ensure secure dashboard access for team collaboration

To get a more comprehensive understanding of Devtron's Kubernetes Dashboard, please visit the [documentation](https://docs.dashboard.devtron.ai/)

### Installation

Before you begin, you must create a [Kubernetes cluster](https://kubernetes.io/docs/tutorials/kubernetes-basics/create-cluster/) (preferably K8s 1.16 or higher) and install [Helm](https://helm.sh/docs/intro/install/).

### Install Devtron's Kubernetes Dashboard

Run the following command to install the latest version of Devtron's Kubernetes Dashboard:

```bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd
```

### Access Devtron

**URL**: Use the following command to get the dashboard URL:

```bash
kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}'
```

**Credentials**:

**UserName**:  `admin` <br>
**Password**:   Run the following command to get the admin password for Devtron version v0.6.0 and higher

```bash
kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ADMIN_PASSWORD}' | base64 -d
```

For Devtron version less than v0.6.0, run the following command to get the admin password:

```bash
kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
```


Please refer to the document for more information on how to [access the Devtron Dashboard](https://docs.dashboard.devtron.ai/getting-started/install-devtron-dashboard).

### Install Multi-Architecture Nodes (ARM and AMD)

To install Devtron on clusters with the multi-architecture nodes (ARM and AMD), append the Devtron installation command with ```--set installer.arch=multi-arch```

## Devtron Platform

The Devtron Platform provides a complete solution for all DevOps requirements, helping to create a single centrailzed DevOps Platform for managing all applications across Kubernetes clusters. 

### Features 🤩

- End-to-End CI/CD pipeline enabling a No Code software delivery workflow for Kubernetes through Helm or GitOps with ArgoCD
- Deploy applications to multiple Kubernetes clusters on multiple cloud/on-prem from one Devtron setup
- Robust DevSecOps implementation with battle tested tools like Trivy and Clair
- A Single dashboard to cater to all debugging requirements for applications
- Deployment metrics to measure the success of the agile process. It captures MTTR, change failure rate, deployment frequency, and deployment size.

To get a more comprehensive understanding of Devtron Platform's features, please visit the [documentation](https://docs.devtron.ai/)

### Install Devtron Platform

Run the following command to install the latest version of the Devtron Platform:

```bash
helm repo add devtron https://helm.devtron.ai

helm repo update devtron

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set argo-cd.enabled=true
```
## :blue_heart: Technology
 
Devtron is built on some of the most trusted and loved technologies:
<br>
<p align="center"><img width="70%" height="70%" src="./assets/we-support.jpg"></p>

## :question: FAQs & Troubleshooting
 
- For troubleshooting Devtron please [refer to this docs page](https://docs.devtron.ai/resources/devtron-troubleshoot)
 
## :busts_in_silhouette: Community
 
Get updates on Devtron's development and chat with project maintainers, contributors, and community members
- Follow [@DevtronL on Twitter](https://twitter.com/DevtronL)
- Raise feature requests, suggest enhancements, and report bugs in our [GitHub Issues](https://github.com/devtron-labs/devtron/issues)
- Articles, Howtos, Tutorials - [Devtron Blogs](https://devtron.ai/blog/)
- Devtron is trusted by communities all across the globe. The list of organizations using Devtron can be found [here](./USERS.md).
 
### Join us at Discord channel
<p>
<a href="https://discord.gg/jsRG5qx2gp">
   <img
   src="https://invidget.switchblade.xyz/jsRG5qx2gp"
   alt="Join Devtron : Heroku for Kubernetes"
   >
</a>
</p>

## :handshake: Contribute
 
Check out our [contributing guidelines](CONTRIBUTING.md). Included, are directions for opening issues, coding standards, and notes on our development processes. We deeply appreciate your contribution.
 
Please look at our [community contributions](COMMUNITY_CONTRIBUTIONS.md) and feel free to create a video or blog around Devtron and add your valuable contribution to the list.
 
### Contributors:
 
We are deeply grateful to all our amazing contributors!
 
<a href="https://github.com/devtron-labs/devtron/graphs/contributors">
 <img src="https://contrib.rocks/image?repo=devtron-labs/devtron" />
</a>
 
## :bug: Vulnerability Reporting
 
We at Devtron, take security and our users' trust very seriously. If you believe you have found a security issue, please report it to <b>security@devtron.ai</b>.
 
## :bookmark: License
 
Devtron is licensed under [Apache License, Version 2.0](LICENSE)

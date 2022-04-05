
<p align="center"><img width="333.333" height="260" src="./assets/devtron-logo-dark-light.png">
<h1 align= "center">No-Code CI/CD Orchestrator for Kubernetes</h1>
</p>

<p align="center">A web based CI/CD Orchestrator leveraging Open Source tools to provide a No-Code, SaaS-like experience for Kubernetes
<br>
<a href="https://docs.devtron.ai/" rel="nofollow"><strong>Explore documentation »</strong></a>
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
<a href="https://discord.gg/jsRG5qx2gp"><img src="https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg" alt="Join Discord"></a>
<a href="https://goreportcard.com/badge/github.com/devtron-labs/devtron"><img src="https://goreportcard.com/badge/github.com/devtron-labs/devtron" alt="Go Report Card"></a>
<a href="./LICENSE"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License"></a>
<a href="https://bestpractices.coreinfrastructure.org/projects/4411"><img src="https://bestpractices.coreinfrastructure.org/projects/4411/badge" alt="CII Best Practices"></a>
<a href="http://golang.org"><img src="https://img.shields.io/badge/Made%20with-Go-1f425f.svg" alt="made-with-Go"></a>
<a href="http://devtron.ai/"><img src="https://img.shields.io/website-up-down-green-red/http/shields.io.svg" alt="Website devtron.ai"></a>
<a href="https://twitter.com/intent/tweet?text=Devtron%20helps%20in%20simplifying%20software delivery%20workflow%20for%20Kubernetes,%20check%20it%20out!!%20&hashtags=OpenSource,Kubernetes,DevOps,CICD,go&url=https://github.com/devtron-labs/devtron%0a"><img src="https://img.shields.io/twitter/url/http/shields.io.svg?style=social" alt="Tweet"></a>

<p align="center">
<a href="https://devtron.ai/support.html">🔥 Want to accelerate K8s adoption? Our team would love to help 100 companies do it the Devtron way! 🔥
<br>
<br>
 Apply Now 👋
 </a>
</p>
<h1></h1>

Devtron is a web based CI/CD orchestrator for Kubernetes. It integrates various Open Source tools to provide AppOps, that also includes Security Scanning, GitOps, Access Control, Debugging and Observability.

<br>
<p align="center"><img src="./assets/readme-comic.png"></p>

<h3><b>Devtron is built in a modular fashion. It consists of the below modules which can be installed independently.</b></h3>


| Module  | Features |
| :-----------: | :-----------: |
| [Hyperion](https://github.com/devtron-labs/devtron#-hyperion)  | Deploy, Observe, Manage & Debug existing Helm apps in all your clusters  |
| [Devtron](https://github.com/devtron-labs/devtron#bulb-devtron)  | Perform CI/CD, Security Scanning, GitOps, Access Control, Debugging and Observability. Comes with Hyperion included. |
 

<!--- 
- [Hyperion](https://github.com/devtron-labs/devtron#-hyperion) - Devtron's light weight module to observe, deploy, manage & debug existing Helm apps in all your clusters. Start with Hyperion, to understand how Devtron can fit into your workflow. You can always switch / upgrade to Devtron for full features, like GitOps implementation, setting up Pipelines and Security.
- [Devtron](https://github.com/devtron-labs/devtron#tada-features) - Devtron gives you all the features of the system with a complete experience - providing you with CI/CD, Security, Observability etc, from a single web-console. Hyperion module is included by default.
-->

# 🦹 Hyperion

Hyperion is Devtron's light weight module to manage Helm apps. It helps you deploy, observe, manage and debug applications deployed through Helm across multiple clusters, minimizing Kubernetes complexities. 
  
## :tada: Features

https://user-images.githubusercontent.com/66381465/159458292-a4d8e212-54b6-444f-bd6d-69645fddf966.mp4

 
<details><summary><b>Application-level Resource grouping for easier Debugging</b></summary>
<br>

- Hyperion groups your Kubernetes objects deployed via Helm charts and display them in a slick UI, for easier monitoring or debugging. Access pod logs and resource manifests right from the Hyperion UI and even edit them!

</details>
 
<details><summary> <b>Centralized Access Management</b></summary>
<br>
 
- Control and give customizable view-only, edit access to users on Project, Environment and Application levels
 
</details>
 
<details><summary> <b>Deploy, Manage and Observe on multiple clusters</b></summary>
<br>
 
- Deploy and manage Helm charts, applications across multiple Kubernetes clusters (hosted on multiple clouds / on-prem) right from a single Hyperion setup

</details>
 
<details><summary> <b>View and edit Kubernetes manifests </b></summary>
<br>
 
 - View and edit all the Kubernetes resources right from the Hyperion dashboard

</details>

Hyperion is a great way to get familiar with Devtron's UI and some of its light weight features. You can always [upgrade to Devtron full stack](https://docs.devtron.ai/hyperion/upgrade-to-devtron), that comes loaded with all the features.
 
## :rocket: Getting Started

### Install Hyperion using Helm3

To install Helm3, check [here](https://helm.sh/docs/intro/install/)

```bash
helm repo add devtron https://helm.devtron.ai
helm install devtron devtron/devtron-operator --create-namespace --namespace devtroncd --set installer.mode=hyperion
```

For those countries/users where Github is blocked, you can download [Hyperion Helm chart](https://s3-ap-southeast-1.amazonaws.com/devtron.ai/devtron-operator-latest.tgz)

```bash
wget https://s3-ap-southeast-1.amazonaws.com/devtron.ai/devtron-operator-latest.tgz
helm install devtron devtron-operator-latest.tgz --create-namespace --namespace devtroncd --set installer.mode=hyperion
```

### Hyperion Dashboard

If you did not provide a **BASE\_URL** during install or have used the default installation, Devtron creates a load balancer for you. Use the following command to get the dashboard URL.

```text
kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}'
```

Please note that it may take some time for the Cloud provider to provision the load balancer and in case of on-prem installation of Kubernetes, please use port-forward or ingress.

You will get an output, something like this

```text
[test2@server ~]$ kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}'
[map[hostname:devtronsdashboardurlhere]]
```

The hostname mentioned here \( devtronsdashboardurlhere \) is the load balancer URL from where you can access the dashboard
 
*****Hyperion admin credentials*****

For admin login, use 
<br>
Username:`admin` 
<br>
and for password run the following command

```bash
kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
```


# :bulb: Devtron

Devtron is a No-Code CI/CD Orchestrator with a complete experience - providing you with CI/CD, Security Scanning, GitOps, Access Control, Debugging and Observability from a single web-console. Hyperion module is included by default, in Devtron.

## :tada: Features
<br>
<img src="./assets/preview.gif">
<br>

<details>
<summary> 
 <b>No Code self-serve DevOps platform</b>
</summary>
 
 - Understands the domain of Kubernetes, Testing, CI/CD and SecOps
 - Reusable and composable Pipelines, which makes Workflows easy to construct and visualize
</details>

<details>
 <summary> <b>Multi-Cloud / Multi-Cluster Deployment</b></summary>

- Gives the ability to deploy your applications to multiple clusters / cloud, with unified dashboard
</details>

<details>
 <summary><b>Built-in SecOps tools and Integration</b> </summary>

- UI driven hierarchical security policy (Global, Cluster, Environment and Application) management
- Integration with [Clair](https://www.redhat.com/en/topics/containers/what-is-clair) for vulnerability scanning
</details>

<details>
<summary><b>UI enabled Application Debugging Dashboard</b></summary>

 - Application-centric view for K8s components
 - Built-in monitoring for CPU, RAM, HTTP Status Code and Latency
 - Advanced Logging, with grep and json search
 - Access all the manifests securely, for e.g. secret obfuscation
 - Auto Issue identification
</details>

<details>
 <summary><b>Enterprise grade Security and Compliance</b></summary>

- Easily manage Roles and Permissions for users through UI
</details>

<details>
<summary><b>Automated GitOps based deployment using ArgoCD</b></summary>

- Automated Git repository and application manifest management
- Reduces complexity (configuration & access control) in adopting GitOps practices
- GitOps backed by Postgres for easier analysis 
</details>

## :globe_with_meridians: Architecture:
<br>
<p align="center"><img src="./assets/Architecture.jpg"></p>

## :rocket: Getting Started

### Quick installation with default settings

This installation will use Minio for storing build logs and cache

```bash
helm repo add devtron https://helm.devtron.ai
helm install devtron devtron/devtron-operator --create-namespace --namespace devtroncd
```
For detailed setup instructions and other options, check out [Devtron setup](https://docs.devtron.ai/setup/install)

### :key: Devtron Dashboard

By default, Devtron creates a load balancer. Use the following command to get the dashboard URL.

```text
kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}'
```

Please note that it may take some time for the Cloud provider to provision the load balancer and in case of on-prem installation of Kubernetes, please use port-forward or ingress.

*****Devtron admin credentials*****

For admin login, use 
<br>
Username:`admin` 
<br>
And for the password, run the following command

```bash
kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
```


# :blue_heart: Technology

Devtron is built on some of the most trusted and loved technologies
<br>
<p align="center"><img width="70%" height="70%" src="./assets/we-support.jpg"></p>

# :video_camera: Videos

- [Devtron - A Comprehensive Overview](https://youtu.be/FB5BI3Ef7uw?t=363)
- [Viktor Farcic's review](https://youtu.be/ZKcfZC-zSMM)
- [Running an application on Devtron](https://youtu.be/bA6zgjPD_yA?t=2927)
- [Devtron Demo](https://youtu.be/ekxHV2Gje-E?t=7856)

# :muscle: Trusted By

Devtron is trusted by Enterprises and Community, all across the globe:
<br>

- [Delhivery:](https://www.delhivery.com/) Delhivery is an Indian delivery and e-commerce logistics company, that provides end-to-end Supply Chain solutions through cutting-edge technology
- [BharatPe:](https://bharatpe.com/) Bharatpe is a Indian fintech company that offers a range of products including interoperable QR code for UPI payments, POS machines for card acceptance, and small business financing
- [Livspace:](https://www.livspace.com/in) Livspace is a home interior and renovation company, that provides interior design and renovation services in Singapore and India
- [Moglix:](https://www.moglix.com/) Moglix is an industrial B2B marketplace and an e-commerce platform for industrial tools and equipment, used largely by businesses in India
- [Xoxoday:](https://www.xoxoday.com/) Xoxoday provides technology infrastructure to enable businesses to automate rewards, incentives & payouts for employees, customers & channel partners

# :question: FAQ & Troubleshooting

- Hyperion - [see here](https://docs.devtron.ai/hyperion/faqs-and-troubleshooting/hyperion-troubleshoot)
- Devtron - [see here](https://docs.devtron.ai/devtron/faqs-and-troubleshooting/devtron-troubleshoot)

# :memo: Compatibility

## Current build

- Devtron uses modified version of [Argo Rollout](https://argoproj.github.io/argo-rollouts/)
- Application metrics only works for k8s version 1.16+

# Support, Contribution and Community

## :busts_in_silhouette: Community

Get updates on Devtron's development and chat with project maintainers, contributors and community members
 
- Follow [@DevtronL on Twitter](https://twitter.com/DevtronL)
- Raise feature requests, suggest enhancements, report bugs in our [GitHub Issues](https://github.com/devtron-labs/devtron/issues)
- Articles, Howtos, Tutorials - [Devtron Blogs](https://devtron.ai/blog/)

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

Check out our [contributing guidelines](CONTRIBUTING.md). Included, are directions for opening issues, coding standards and notes on our development processes. We deeply appreciate your contribution.

Please look at our [community contributions](COMMUNITY_CONTRIBUTIONS.md) and feel free to create a video or blog around Devtron and add your valuable contribution in the list.

### Contributors:

We are deeply grateful to all our amazing contributors!

<a href="https://github.com/devtron-labs/devtron/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=devtron-labs/devtron" />
</a>

## :bug: Vulnerability Reporting

We at Devtron, take security and our users' trust very seriously. If you believe you have found a security issue, please report to <b>security@devtron.ai</b>.

# :bookmark: License

Devtron is licensed under [Apache License, Version 2.0](LICENSE)

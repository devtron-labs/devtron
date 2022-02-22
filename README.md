
<p align="center"><img width="333.333" height="260" src="./assets/devtron-logo-dark-light.png">
<h1 align= "center">Web based CI/CD Platform for Kubernetes</h1>
</p>

<p align="center">A Web based CI/CD platform leveraging open source tools to provide a No-Code SaaS-like experience for Kubernetes.
<br>
<a href="https://docs.devtron.ai/" rel="nofollow"><strong>Explore documentation »</strong></a>
<br>
<a href="https://devtron.ai/">Website</a>
·
<a href="https://devtron.ai/blog/">Blog</a>
·
<a href="https://discord.gg/jsRG5qx2gp">Join Discord</a>
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

 
<!--</p>
<p align = "center">
<a href="https://github.com/devtron-labs/devtron"><img src="https://github-readme-stats-one-bice.vercel.app/api?username=Abhinav-26&show_icons=true&include_all_commits=true&count_private=true&role=OWNER,ORGANIZATION_MEMBER,COLLABORATOR" alt="Devtron's GitHub stats"></a>
</p>-->

<p align="center">
<a href="https://devtron.ai/support.html">🔥 Want to accelerate K8s adoption? Our core team would love to help 30 companies do it the Devtron way! 🔥 Apply Now 👋</a></p>

## :book: Menu

- [Devtron?](https://github.com/devtron-labs/devtron#bulb-devtron)
- [Devtron](https://github.com/devtron-labs/devtron#computer-devtron)
- [Hyperion](https://github.com/devtron-labs/devtron#-hyperion)
- [Documentation](https://docs.devtron.ai/)
- [Compatibility Notes](https://github.com/devtron-labs/devtron#memo-compatibility-notes)
- [Community](https://github.com/devtron-labs/devtron#busts_in_silhouette-community)
- [Trusted By](https://github.com/devtron-labs/devtron#muscle-Trusted-By)
- [FAQ & Troubleshooting](https://github.com/devtron-labs/devtron#question-faq--troubleshooting)
- [Contribute](https://github.com/devtron-labs/devtron#handshake-contribute)

## :bulb: Devtron?

### Why use it?

Devtron is a Web-Based CI/CD Platform for Kubernetes. It integrates various OpenSource tools to provide a modular CI/CD platform that also includes Security Scanning, GitOps, Access Control, and Debugging/Observability.


<p align="center"><img src="./assets/readme-comic.png"></p>

<b> Devtron is built in a modular approach. These modules can be installed independently: </b>
- [Devtron](https://github.com/devtron-labs/devtron#tada-Full-Devtron-Experience) - This option gives you all the features of Devtron as a Full Experience providing you with CI, CD, security, cost, observability, stabilization. All the modules stated below are included here.
- [Hyperion](https://github.com/devtron-labs/devtron#tada-featuresfor-hyperion) - Devtron's Web-based module to manage helm apps that can be installed seperately. Install Hyperion -> manage, Observe helm apps of all your clusters. This module is also a great way to manage existing helm apps and gradually understand how Devtron fits into your workflow. You can always switch to Devtron for all the features.

## :computer: Devtron

Devtron provides a full feldged web based CI/CD platform including features like Security Scanning, GitOps, Access Control, and Debugging/Observability. Modules like Hyperion are included as additional modules here.

## :tada: Features
<br>
<img src="./assets/preview.gif">
<br>

<details>
<summary> 
 <b> No code self-serve DevOps platform </b>
  </summary>
<br>

- Workflow which understands the domain of Kubernetes, testing, CD, SecOps
- Reusable and composable pipelines so that workflows are easy to construct and visualize

</details>

<details>
 <summary> <b> Multi-cloud/Multi-cluster deployment </b></summary>
<br>

- Devtron gives the ability to deploy your applications to multiple clusters/cloud just with the same dashboard.

</details>


<details>
 <summary> <b>Built-in SecOps tools and integration</b> </summary>
<br>
 
- UI driven hierarchical security policy (global, cluster, environment, and application) for efficient policy management
- Integration with [Clair](https://www.redhat.com/en/topics/containers/what-is-clair) for vulnerability scanning

</details>
<details>
<summary> <b> UI-enabled Application debugging dashboard </b></summary>
 <br> 
 
 - Application-centric view for K8s components
 - Built-in monitoring for CPU, RAM, http status code, and latency
 - Advanced logging, with grep and json search
 - Access all manifests securely for e.g. secret obfuscation
 - Auto issue identification

</details>

<details>
 <summary> <b>Enterprise grade security and compliances </b></summary>
<br>

- Easy to control roles and permissions for users. 
- Club the users of similar roles by giving the required permissions through the User Interface.

</details>

<details>
<summary> <b>Automated GitOps based deployment using argocd </b></summary>
<br>
 
- Automated git repository and application manifest management
- Reduces complexity(configuration, access control) in adopting GitOps practices
- GitOps backed by Postgres for easier analysis 

</details>

### :blue_heart: We Support:
In addition to the features, we love supporting platforms that devs find easy to work with.
<br>
<p align="center"><img width="70%" height="70%" src="./assets/we-support.jpg"></p>


## :globe_with_meridians: Architecture:
<br>
<p align="center"><img src="./assets/Architecture.jpg"></p>



## :rocket: Getting Started

#### You can follow our detailed installation guide, using Devtron and other key functionalities, in our
[Devtron Documentation](https://docs.devtron.ai/)

### Quick installation with default settings

This installation will use Minio for storing build logs and cache.

```bash
helm repo add devtron https://helm.devtron.ai
helm install devtron devtron/devtron-operator --create-namespace --namespace devtroncd
```

#### For detailed installation instructions and other options, check out:
[devtron installation documentation](https://docs.devtron.ai/setup/install)


### :key: Access Devtron dashboard

By default, Devtron creates a loadbalancer. Use the following command to get the dashboard url:

```text
kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}'
```

*****Devtron Admin credentials*****


For admin login, use the username:`admin`. And for the password, run the following command:

```bash
kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
```

## :memo: Compatibility notes

### Current build: 

- Devtron uses modified version of [argo rollout](https://argoproj.github.io/argo-rollouts/)
- Application metrics only works for k8s 1.16+


## 🦹 Hyperion

<details>
 <summary> <b> Hyperion is one of Devtron's Web-based modules to manage helm apps that can be installed seperately too. It helps you observe, manage and debug the applications deployed through Helm across multiple clusters minimizing Kubernetes Complexities. Please expand this column to find Hyperion's features and to get Started with it:</b></summary>
<br>


## :tada: Features(For Hyperion)
 
<details><summary> <b> Application-level resource grouping for easier Debugging </b></summary>
<br>

- Hyperion groups your deployed Helm charts and display them in a slick UI for easier monitoring or debugging. Access pod logs and resource manifests right from the Hyperion UI and even edit them!

</details>
 
<details><summary> <b>  Centralized Access Management </b></summary>
<br>
 
- Control and give customizable view-only, edit access to users on Project, Environment and App level.
 
</details>
 
<details><summary> <b>  Manage and observe Multiple Clusters </b></summary>
<br>
 
- Manage Helm charts, Applications across multiple Kubernetes clusters (hosted on multiple cloud/on-prem) right from a single Hyperion setup.

</details>
 
<details><summary> <b> View and Edit Kubernetes Manifests </b></summary>
<br>
 
 - View and Edit all Kubernetes resources right from the Hyperion dashboard.

</details>

#### Side Note:

Hyperion module is also a great way to get to know Devtron's UI and some of its features. You can always switch from Hyperion to Devtron which includes all the features. [Just a Couple of Commands away.](https://github.com/devtron-labs/devtron#rocket-getting-started)
 
## :rocket: Getting Started(For Hyperion)

### Install Hyperion using Helm3

To install Helm3, please check [Installing Helm3](https://helm.sh/docs/intro/install/)

```bash
helm repo add devtron https://helm.devtron.ai
helm install devtron devtron/devtron-operator --create-namespace --namespace devtroncd --set installer.mode=hyperion
```

For those countries/users where Github is blocked , you can download the [Hyperion Helm chart](https://s3-ap-southeast-1.amazonaws.com/devtron.ai/devtron-operator-latest.tgz)


```bash
wget https://s3-ap-southeast-1.amazonaws.com/devtron.ai/devtron-operator-latest.tgz
helm install devtron devtron-operator-latest.tgz --create-namespace --namespace devtroncd --set installer.mode=hyperion
```

### Access Hyperion dashboard

If you did not provide a **BASE\_URL** during install or have used the default installation, Devtron creates a loadbalancer for you on its own. Use the following command to get the dashboard url.

```text
kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}'
```

You will get result something like below

```text
[test2@server ~]$ kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}'
[map[hostname:devtronsdashboardurlhere]]
```

The hostname mentioned here \( devtronsdashboardurlhere \) is the Loadbalancer URL where you can access the Devtron dashboard.
 
### Hyperion Admin credentials

For admin login use username:`admin` and for password run the following command.

```bash
kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
```

#### If you want to perform CI/CD, App creation present in Devtron you are always few commands away:
 
[Getting Started for Devtron](https://github.com/devtron-labs/devtron#rocket-getting-started)

 
</details>


## :video_camera: Videos:

- [Devtron - A Comprehensive Overview](https://youtu.be/FB5BI3Ef7uw?t=363)
- [Viktor Farcic(YouTuber) Review](https://youtu.be/ZKcfZC-zSMM)
- [Running an application on Devtron](https://youtu.be/bA6zgjPD_yA?t=2927)
- [Devtron Demo](https://youtu.be/ekxHV2Gje-E?t=7856)


## :busts_in_silhouette: Community

Get updates on Devtron's development and chat with the project maintainers, contributors, and community members.
- Follow [@DevtronL on Twitter](https://twitter.com/DevtronL)
- Raise feature requests, suggest enhancements, report bugs in our [GitHub issues](https://github.com/devtron-labs/devtron/issues)
- Read the [Devtron blog](https://devtron.ai/blog/)

### Join Our Discord Community
<p>
<a href="https://discord.gg/jsRG5qx2gp">
    <img 
    src="https://invidget.switchblade.xyz/jsRG5qx2gp" 
    alt="Join Devtron : Heroku for Kubernetes"
    >
</a>
 </p>
 
## :muscle: Trusted By

<details>
 <summary> <b> Devtron has been trusted by the Enterprises and community all across the globe: </b></summary>
<br>

- [Delhivery:](https://www.delhivery.com/) Delhivery Limited is one the largest and most profitable logistics company in India
- [BharatPe:](https://bharatpe.com/) Bharatpe is a business utility app to accept payments transactions in settlements.
- [Livspace:](https://www.livspace.com/in) Livspace is one-stop shop for all things home interiors and renovation services.
- [Moglix:](https://www.moglix.com/) It is an Asia-based B2B commerce company intensively inclined towards B2B procurement of industrial supplies
- [Xoxoday:](https://www.xoxoday.com/) Xoxoday helps to send rewards, perks & incentives to employees, customers and partners.<br>

</details>



## :question: FAQ & Troubleshooting:
### FAQ:

<details>
<summary> <b>1.How to resolve unauthorized error/s, while trying to save global configurations like hostname, GitOps etc. after successful devtron installation</b></summary>
<br>
A. This occurs most of the time because any one or multiple jobs get failed during installation. To resolve this, you'll need to first check which jobs have failed. Follow these steps:

- Run the following command and check which are the jobs with 0/1 completions:
```bash
kubectl get jobs -n devtroncd
```
- Note the names of the jobs with 0/1 completions and check if their pods are in running state or not by running the command:
kubectl get pods -n devtroncd
- If they are in running condition, please wait for the jobs to be completed. This may be due to internet issue. If the job is not in running condition, delete those incomplete jobs using:
kubectl delete jobs <job1-name> <job2-name> -n devtroncd..[Read More](https://github.com/devtron-labs/devtron/blob/main/Troubleshooting.md#1-how-to-resolve-unauthorized-error-while-trying-to-save-global-configurations-like-hostname-gitops-etc-after-successful-devtron-installation)
<br><br>
</details>
 
<details>
<summary> <b>2.What to do if devtron dashboard is not accessible on browser, even after successful completion of all jobs and all pods are in running mode?</b></summary>
<br>

A. Check if nats-cluster is created or not, you can check it using the following command:
```bash
kubectl get natscluster -n devtroncd
```
- You should see a natscluster with the name devtron-nats. If not, run the following command:
```bash
kubectl apply -f https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/yamls/nats-server.yaml -n devtroncd
```
- Wait util all nats pods are created, and the pods are in running condition. Once complete, delete devtron and dashboard pods. Then you should be able to access the devtron dashboard without any issues.
- If your problem is still not resolved, you can post your query in our [discord](https://discord.gg/jsRG5qx2gp) channel
<br><br>
</details>

<details>
<summary> <b>3.Not able to see deployment metrics on production environment or Not able to enable application-metrics or Not able to deploy the app after creating a configmap or secret with data-volume option enabled</b></summary>
<br>
A. Update the rollout crds to latest version, run the following command
```bash
kubectl apply -f https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/yamls/rollout.yaml -n devtroncd
```
</details>
 
### Troubleshooting:
- For Installation Troubleshooting, check this [documentation](https://docs.devtron.ai/setup/install)
- For other troubleshooting, Check the [Common troubleshooting documentation](https://docs.devtron.ai/user-guide/command-bar)


## :handshake: Contribute

Check out our [contributing guidelines](CONTRIBUTING.md). Included are directions for opening issues, coding standards, and notes on our development processes. We deeply appreciate your contributions.

Also please checkout our [community contributions](COMMUNITY_CONTRIBUTIONS.md) and feel free to create a video or blog around Devtron and add your valuable contribution in the list.

### Our Contributors:

We are deeply grateful for all our amazing contributors!    

<a href="https://github.com/devtron-labs/devtron/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=devtron-labs/devtron" />
</a>

## :bug: Vulnerability Reporting

We at Devtron take security and our users' trust very seriously. If you believe you have found a security issue in Devtron, please responsibly disclose this to us at security@devtron.ai.

## :bookmark: License
Devtron is available under the [Apache License, Version 2.0](LICENSE)

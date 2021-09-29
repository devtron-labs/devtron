
<p align="center"><img width="333.333" height="260" src="./assets/devtron-logo-dark-light.png">
<h1 align= "center">Heroku-like Platform for Kubernetes.</h1>
</p>

<p align="center">Devtron leverages popular Open-Source tools to provide a No-Code SaaS like experience for Kubernetes.
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


## :book: Menu

- [Devtron?](https://github.com/devtron-labs/devtron#bulb-devtron)
- [Features](https://github.com/devtron-labs/devtron#tada-features)
- [Getting Started](https://github.com/devtron-labs/devtron#rocket-getting-started)
- [Documentation](https://docs.devtron.ai/)
- [Videos](https://github.com/devtron-labs/devtron#memo-compatibility-notes)
- [Compatibility Notes](https://github.com/devtron-labs/devtron#memo-compatibility-notes)
- [Community](https://github.com/devtron-labs/devtron#busts_in_silhouette-community)
- [Love what you see!](https://github.com/devtron-labs/devtron#sparkling_heart-love-what-you-see)
- [FAQ & Troubleshooting](https://github.com/devtron-labs/devtron#question-faq--troubleshooting)
- [Contribute](https://github.com/devtron-labs/devtron#handshake-contribute)

## :bulb: Devtron?

### Why use it?

We have seen various tools that are used to greatly increase the ease of using Kubernetes but using these tools simultaneously is painful and hard to use. As these tools dont talk to eachother for managing different aspects of application lifecycle - CI, CD, security, cost, observability, stabilization. We built Devtron to solve this problem precisely.

<p align="center"><img src="./assets/readme-comic.png"></p>

Devtron is an OpenSource modular product providing 'seamless', 'implementation agnostic uniform interface' integrated  with OpenSource and commercial tools across life cycle. All done focusing on a slick User Experience enabling self-serve model. 
<br>
You can efficiently handle Security, Stability, Cost and more in a unified experience.


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
</details>


<details>
 <summary> <b>Built-in SecOps tools and integration</b> </summary>
<br>
 
- UI driven hierarchical security policy (global, cluster, environment and application) for efficient policy management
- Integration with [Clair](https://www.redhat.com/en/topics/containers/what-is-clair) for vulnerability scanning

</details>
<details>
<summary> <b> UI-enabled Application debugging dashboard </b></summary>
 <br> 
 
 - Application centric view for K8s components
 - Built-in monitoring for cpu, ram, http status code and latency
 - Advanced logging with grep and json search
 - Access all manifests securely for e.g. secret obfuscation
 - Auto issue identification

</details>

<details>
<summary> <b>Enterprise grade security and compliances </b></summary>
</details>

<details>
<summary> <b>Automated Gitops based deployment using argocd </b></summary>
<br>
 
- Automated git repository and application manifest management
- Reduces complexity(configuration, access control) in adopting gitops practices
- Gitops backed by Postgres for easier analysis 

</details>

## :globe_with_meridians: Architecture:
<br>
<img src="./assets/Architecture.jpg">

### :blue_heart: We Support:
In addition to the features, we love supporting platforms that devs find easy to work with.
<br>
<p align="center"><img width="660" height="216" src="./assets/we-support.jpg"></p>

## :rocket: Getting Started

#### You can follow through a detailed installation guide, using Devtron and other key functionalities of devtron in our
[Devtron Documentation](https://docs.devtron.ai/)

#### Quick installation with default settings

This installation will use Minio for storing build logs and cache.

```bash
helm repo add devtron https://helm.devtron.ai
helm install devtron devtron/devtron-operator --create-namespace --namespace devtroncd
```

#### For detailed installation instructions and other options, check out:
[devtron installation documentation](https://docs.devtron.ai/setup/install)


#### :key: Access Devtron dashboard

By default Devtron creates a loadbalancer. Use the following command to get the dashboard url.

```text
kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}'
```

*****Devtron Admin credentials*****


For admin login use username:`admin` and for password run the following command.

```bash
kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
```

## :video_camera: Videos:

- [Devtron - A Comprehensive Overview](https://youtu.be/FB5BI3Ef7uw?t=363)
- [Viktor Farcic(YouTuber) Review](https://youtu.be/ZKcfZC-zSMM)
- [Running an application on Devtron](https://youtu.be/bA6zgjPD_yA?t=2927)
- [Devtron Demo](https://youtu.be/ekxHV2Gje-E?t=7856)

## :memo: Compatibility notes

### Current build: 

- It uses modified version of [argo rollout](https://argoproj.github.io/argo-rollouts/)
- Application metrics only works for k8s 1.16+

## :busts_in_silhouette: Community

Get updates on Devtron's development and chat with the project maintainers, contributors and community members.
- Follow [@DevtronL on Twitter](https://twitter.com/DevtronL)
- Raise feature requests, suggest enhancements, report bugs at [GitHub issues](https://github.com/devtron-labs/devtron/issues)
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


## :sparkling_heart: Love What You See!

If you are loving what we are doing, Please consider giving us a star.
<br>
[![GitHub stars](https://img.shields.io/github/stars/devtron-labs/devtron)](https://github.com/devtron-labs/devtron/stargazers)
<br>
Or you can tweet about us: 
<br>
<a href="https://twitter.com/intent/tweet?text=Devtron%20helps%20in%20simplifying%20software delivery%20workflow%20for%20Kubernetes,%20check%20it%20out!!%20&hashtags=OpenSource,Kubernetes,DevOps,CICD,go&url=https://github.com/devtron-labs/devtron%0a"><img src="https://img.shields.io/twitter/url/http/shields.io.svg?style=social" alt="Tweet"></a>
<br>
Your token of gratitude will go a long way helping us reach more developers like you. ❤:)

Or you can do one better and Contribute 👏

## :question: FAQ & Troubleshooting:
### FAQ:
1.How to resolve unauthorized error while trying to save global configurations like hostname, gitops etc. after successful devtron installation
<br>
A. This occurs most of the time because any one or multiple jobs get failed during installation. To resolve this, you need to first check which are the jobs that have failed. Follow these steps :-

- Run the following command and check which are the jobs with 0/1 completions:
```bash
kubectl get jobs -n devtroncd
```
- Note down or remember the names of jobs with 0/1 completions and check if their pods are in running state still or not by running the command:
kubectl get pods -n devtroncd
- If they are in running condition, please wait for the jobs to be completed as it may be due to internet issue and if not in running condition, then delete those incomplete jobs using:
kubectl delete jobs <job1-name> <job2-name> -n devtroncd..[Read More](https://github.com/devtron-labs/devtron/blob/main/Troubleshooting.md#1-how-to-resolve-unauthorized-error-while-trying-to-save-global-configurations-like-hostname-gitops-etc-after-successful-devtron-installation)
<br><br>

2.What to do if devtron dashboard is not accessible on browser even after successful completion of all the jobs and all pods are in running mode
<br>
A. For this, you need to check if nats-cluster is created or not, you can check it using the following command:
```bash
kubectl get natscluster -n devtroncd
```
- You should see a natscluster with the name devtron-nats and if not, run the given command:
```bash
kubectl apply -f https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/yamls/nats-server.yaml -n devtroncd
```
- Wait till all the nats pods are created and the pods are in running condition. After that delete devtron and dashboard pods once and then you should be able to access the devtron dashboard without any issues.
- If your problem is still not resolved, you can post your query in our [discord](https://discord.gg/jsRG5qx2gp) channel

### Troubleshooting:
- For Installation Troubleshooting, check this [Documentation](https://docs.devtron.ai/setup/install)
- For other troubleshooting, Check the [Common troubleshooting documentation](https://docs.devtron.ai/user-guide/command-bar)


## :handshake: Contribute


Check out our [contributing guidelines](CONTRIBUTING.md). Included are directions for opening issues, coding standards, and notes on our development processes. We deeply appreciate your contributions.

### Our Contributors:

We are deeply grateful for all our amazing contributors!    

<a href="https://github.com/devtron-labs/devtron/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=devtron-labs/devtron" />
</a>

## :bug: Vulnerability Reporting

We at Devtron take security and our users' trust very seriously. If you believe you have found a security issue in Devtron, please responsibly disclose us at security@devtron.ai.

## :bookmark: License
Devtron is available under the [Apache License, Version 2.0](LICENSE)
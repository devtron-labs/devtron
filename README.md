<p align="center"><img width="200" height="156" src="./assets/logo.png"></p>
<p align="center">Devtron is an open source software delivery workflow for kubernetes written in go.
<br>
<a href="https://docs.devtron.ai/" rel="nofollow"><strong>Explore documentation »</strong></a>
<br>
<br>
<a href="https://devtron.ai/">Website</a>
·
<a href="https://devtron.ai/blog/">Blog</a>
·
<a href="https://discord.gg/72JDKy4">Join Discord</a>
·
<a href="https://twitter.com/DevtronL">Twitter</a>
</p>

## Why Devtron?
It is designed as a self-serve platform for operationalizing and maintaining applications (AppOps) on kubernetes in a developer friendly way. 
<br>
<br>
<img src="./assets/preview.gif">
<br>
<br>
### Some of the benefits  provided by devtron are: 
<details>
<summary> 
 <b> Zero code software delivery workflow </b>
  </summary>
<br>

- Workflow which understands the domain of **kubernetes, testing, CD, SecOps** so that you dont have to write scripts
- Reusable and composable components so that workflows are easy to contruct and reason through
</details>

<details>
<summary> <b> Multi cloud deployment </b></summary>
 <br> 
 
 - Deploy to multiple kubernetes cluster
 - Test on aws cloud 
   > coming soon: support for GCP and microsoft azure  
</details>
<details>
 <summary> <b> Easy dev-sec-ops integration </b> </summary>
<br>
 
- Multi level security policy at global, cluster, environment and application for efficient hierarchical policy management
- Behavior driven security policy
- Define policies and exception for kubernetes resources
- Define policies for events for faster resolution
</details>

<details>
 <summary> <b> Application debugging dashboard </b> </summary>
<br>
 
- One place for all historical kubernetes events 
- Access all manifests securely for e.g. secret obfuscation 
- ***Application metrics*** for cpu, ram, http status code and latency with comparison between new and old 
- ***Advanced logging*** with grep and json search 
- Intelligent ***correlation between events, logs*** for faster triangulation of issue 
- Auto issue identification 
</details>

<details>
<summary> <b>Enterprise grade security and compliances </b></summary>
<br>
 
- Fine grained access control; control who can edit configuration and who can deploy
- Audit log to know who did what and when
- History of all CI and CD events
- Kubernetes events impacting application
- Relevant cloud events and their impact on applications
- Advanced workflow policies like blackout window, branch environment relationship to secure build and deployment pipelines
</details>
<details>
<summary> <b> Gitops aware  </b></summary>
<br>
 
- Gitops exposed through API and UI so that you dont have to interact with git cli
- Gitops backed by postgres for easier analysis
- Enforce finer access control than git

</details>
<details>
<summary> <b>Operational insights  </b></summary>
<br>
 
- Deployment metrics to measure success of agile process. It captures mttr, change failure rate, deployment frequency, deployment size out of the box.
- Audit log to understand the failure causes
- Monitor changes across deployments and revert easily

</details>


## How to start using Devtron

#### Installing devtron wuth Helm 

```bash
$ git clone https://github.com/devtron-labs/devtron-installation-script.git
$ cd devtron-installation-script/charts
$ #modify values in values.yaml
$ helm install devtron . -f values.yaml
```


#### Access Devtron dashboard

Devtron dashboard in now available at the `BASE_URL/dashboard`, where `BASE_URL` same as provided in `values.yaml`


*****For login use username:`admin` and for password run command mentioned below.*****

```bash
$ kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
```

*****[Detail configuration options](https://docs.devtron.ai/setup/install#configuration)*****

#### Using devtron
  
- [Deploying First Application](https://docs.devtron.ai/user-guide/creating-application)
- [Deploying Helm charts](https://docs.devtron.ai/user-guide/deploy-chart)
- [Configure Security policy](https://docs.devtron.ai/user-guide/security-features)
- [Detail Userguide](https://docs.devtron.ai)



<!--
## Why another Deployment tool? 

**TODO**
-->

## Compatibility notes

- Only AWS kubernetes cluster is supported as of now
- It uses modified version of [argo rollout](https://argoproj.github.io/argo-rollouts/)
- Application metrics only works for k8s 1.16+

## Community

Get updates on Devtron's development and chat with the project maintainers, contributors and community members.

 - Join the [Discord Community](https://discord.gg/72JDKy4) 
 - Follow [@DevtronL on Twitter](https://twitter.com/DevtronL)
 - Raise feature requests, suggest enhancements, report bugs at [GitHub issues](https://github.com/devtron-labs/devtron/issues)
 - Read the [Devtron blog](https://devtron.ai/blog/)


## Contribute

Check out our [contributing guidelines](CONTRIBUTING.md). Included are directions for opening issues, coding standards, and notes on our development processes.

## Vulnerability Reporting

We at Devtron take security and our users' trust very seriously. If you believe you have found a security issue in Devtron, please responsibly disclose us at security@devtron.ai.

## License

Devtron is available under the [Apache License, Version 2.0](LICENSE)


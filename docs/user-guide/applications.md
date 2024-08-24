# Applications

{% hint style="warning" %}
Configure [Global Configurations](./global-configurations/README.md) first before creating an application or cloning an existing application.
{% endhint %}

## Introduction

The **Applications** page helps you create and manage your microservices, and it majorly consists of the following:

* [Application Listing](#application-listing)
* [Create Button](#create-button)
* [Other Options](#other-options)

### Application Listing

You can view the app name, its status, environment, namespace, and many more upfront. The apps are segregated into: [Devtron Apps](../reference/glossary.md#devtron-apps), [Helm Apps](../reference/glossary.md#helm-apps), and [ArgoCD Apps](../reference/glossary.md#argocd-apps)

![Figure 1: App Types](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/argocd/app-types.jpg)

### Create Button

You can use this to:
* [Create a Devtron app](./create-application.md)
* [Create a Helm app](./deploy-chart/deployment-of-charts.md)
* [Create a Job](./jobs/create-job.md)

### Other Options

There are additional options available for you:
* **Search and filters** to make it easier for you to find applications.
* **Export CSV** to download the data of Devtron apps (not supported for Helm apps and Argo CD apps).
* **Sync button** to refresh the app listing.

---

## View ArgoCD App Listing

{% hint style="warning" %}
### Who Can Perform This Action?
Users need super-admin permission to view/enable/disable the ArgoCD listing.
{% endhint %}

### Preface

In Argo CD, a user manages one dashboard for each Argo CD app. Therefore, with multiple apps, the process becomes cumbersome for the user to manage several dashboards.

With Devtron, you get an entire Argo CD app listing in one place. This listing includes:
* Apps deployed using [GitOps](../reference/glossary.md#gitops) on Devtron
* Other Argo CD apps present in your cluster

![Figure 2: ArgoCD App List](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/argocd/argo-cd-listing.jpg)

### Advantages

Devtron also bridges the gap for ArgoCD users by providing additional features as follows:

* **Resource Scanning**: You can scan for vulnerabilities using Devtron's [resource scanning](../user-guide/security-features.md#from-app-details) feature.

* **Single-pane View**: All Argo CD apps will show details such as their app status, environment, cluster, and namespace together in one dashboard. 

* **Feature-rich Options**: Clicking an Argo CD app will give you access to its logs, terminal, events, manifest, available resource kinds, pod restart log, and many more.

{% hint style="info" %}
### Additional References
[ArgoCD: Standalone Configuration vs Devtron Configuration](https://devtron.ai/blog/argocd-standalone-configuration-vs-devtron-configuration/#argocd-installation-and-configuration)
{% endhint %}

### Prerequisite
The cluster in which Argo CD apps exist should be added in **Global Configurations** → **Clusters and Environments**

### Enabling ArgoCD App Listing

{% embed url="https://www.youtube.com/watch?v=4KyYnsAEpqo" caption="Enabling External ArgoCD Listing" %}

1. Go to the **Resource Browser** of Devtron.

2. Select the cluster (in which your Argo CD app exists).

3. Type `ConfigMap` in the 'Jump to Kind' field.

4. Search for `dashboard-cm` using the available search bar and click it.

5. Click **Edit Live Manifest**.

6. Set the feature flag **ENABLE_EXTERNAL_ARGO_CD** to  **"true"**

7. Click **Apply Changes**.

8. Go back to the 'Jump to Kind' field and type `Pod`.

9. Search for `dashboard` pod and use the kebab menu (3 vertical dots) to delete the pod.

10. Go to **Applications** and refresh the page. A new tab named **ArgoCD Apps** will be visible.
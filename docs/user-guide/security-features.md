# Security Features

## Introduction

Devtron provides [DevSecOps](https://devtron.ai/product/devsecops) capabilities across your software development life cycle for both: the default CI/CD solution by Devtron as well as your existing CI/CD Tools.

One of the key components of DevSecOps is the detection of security risks. Currently, Devtron supports the following types of scanning:

* Image Scan
* Code Scan (not available for Helm apps) 
* Kubernetes Manifest Scan

![Figure 1: Security Scan Results](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/security-scan.jpg)

You can integrate a scanning tool of your choice. By default, Devtron integrates with Trivy using which you can scan for the following issues:

* Vulnerability
* License Risks
* Misconfigurations
* Exposed Secrets

---

## Where to Initiate the Scan

### After Code Commit

When you commit the code, it's essential to scan it before building a [container image](../reference/glossary.md#image). By scanning early, you can catch and fix problems before they become expensive or time-consuming to remediate later. 

![Figure 2: Scanning in Pre-CI Stage](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/pre-ci.gif)

1. In your application, go to **App Configuration** → **Workflow Editor**.

2. Click the CI pipeline of your preferred workflow.

3. Go to the **Pre-build stage** (tab).

4. Click **+ Add Task**.

5. Choose **Vulnerability_Scanner v1.0.0** plugin from the list.

6. Click **Update Pipeline**.

Based on the results of the scanner, you can also decide whether your CI should proceed further or not. This is possible through **Pass/Failure Condition** setting in the plugin. In the below example, we are allowing image build only if the no. of high vulnerability is zero.

![Figure 3: Setting a Condition](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/pre-ci-condition.gif)

Results of Pre-CI scan will be visible under `Code Scan` in the **App Details** page as shown below.

![Figure 4: Pre-CI Code Scan Results](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/code-scan.gif)

### After Building Container Image

Once a container image is ready, you can scan its base image libraries, stale files, compromised licenses, and many more.

There are 2 options available:
* Image scan in the Build stage (refer [Security Scans](./security-features/security-scans.md))
* Comprehensive scan in Post-Build stage

This section contains the steps for comprehensive scan.

![Figure 5: Scanning in Post-CI Stage](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/post-ci.gif)

1. Go to the **Post-build stage** (tab) of your CI pipeline.

2. Click **+ Add Task** and choose **Vulnerability_Scanner v1.0.0**.

3. Click **Update Pipeline**.

Results of Post-CI scan will be visible under `Image Scan` in the **App Details** page as shown below.

![Figure 6: Post-CI Image Scan Results](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/image-scan-1.gif)

### Before Triggering Deployment

There can be a loophole where the original image built in the CI stage gets compromised later (say, in publicly accessible repository). Therefore, you can scan the image and catch issues before deploying it. On top of that, you can also scan manifests to detect misconfigurations and exposed secrets.

![Figure 7: Scanning in Pre-CD Stage](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/pre-deploy.gif)

1. Go to the **Pre-Deployment stage** (tab) of your CD pipeline.

2. Click **+ Add Task** and choose **Vulnerability_Scanner v1.0.0**.

3. Click **Update Pipeline**.

Results of Pre-CD scan will be visible under `Image Scan` and `Kubernetes Manifest` in the **App Details** page as shown below.

![Figure 8: Pre-CD Scan Results](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/manifest-scan.gif)

### During Helm App Deployment

When you [deploy a helm chart](../user-guide/deploy-chart/deployment-of-charts.md), Devtron will scan the image associated with that helm chart and also the manifests, but unlike Devtron Apps, there is no code scan involved.

Results of helm app scan will be visible under `Image Scan` and `Kubernetes Manifest` in the **App Details** page as shown below.

![Figure 9: Helm App Scan Results](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/helm-app-scan.gif)

---

## Scans and Policies

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to enable vulnerability scanning and to define security policies in Devtron.
{% endhint %}

Devtron's Security feature has two primary sections:

1. [**Security Scans**](./security-features/security-scans.md) - You can view the vulnerabilities detected across your applications.

2. [**Security Policies**](./security-features/security-policies.md) - This allows you to define guardrails to block or allow the deployment of container images depending on the vulnerabilities detected.


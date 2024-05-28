# Security Features

## Introduction

Devtron provides DevSecOps capabilities across your software development life cycle for the following:
1. Default CI/CD solution by Devtron
2. Your existing CI/CD Tools

One of the key components of DevSecOps is the detection of security risks. Currently, Devtron supports the following types of scanning:

1. Image Scan
2. Code Scan (not available for Helm apps)
3. Kubernetes Manifest Scan

---

{% hint style="info" %}
### Prerequisite

Install any one of the following integrations for scanning vulnerabilities:
* [Clair](../user-guide/integrations/clair.md)
* Trivy
{% endhint %}

Devtron's security feature consists of two primary components:

1. [Security Scans](./security-features/security-scans.md) - By leveraging this feature, you can ensure that your containerized applications are safeguarded against known vulnerabilities.

2. [Security Policies](./security-features/security-policies.md) - This allows you to define policies to block or allow the deployment of container images depending on the vulnerabilities detected.

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to enable vulnerability scanning and to define security policies in Devtron.
{% endhint %}
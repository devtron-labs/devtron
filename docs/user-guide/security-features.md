# Security Features

{% hint style="info" %}
### Prerequisite

Install any one of the following integrations for scanning vulnerabilities:
* [Clair](../user-guide/integrations/clair.md)
* Trivy
{% endhint %}

Devtron's security feature consists of two primary components:

1. [Security Scans](./security-features/security-scans.md) - This allows you to identify and address potential security risks effectively. By leveraging this feature, you can ensure that your containerized applications are safeguarded against known vulnerabilities.

2. [Security Policies](./security-features/security-policies.md) - This allows you to define policies to block or allow the deployment of container images depending on the vulnerabilities detected.

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to enable vulnerability scanning and to define security policies in Devtron.
{% endhint %}
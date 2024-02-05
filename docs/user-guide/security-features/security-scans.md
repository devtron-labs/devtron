# Security Scans

{% hint style="info" %}
### Prerequisite

Install any one of the following integrations for scanning vulnerabilities:
* [Clair](../user-guide/integrations/clair.md)
* Trivy
{% endhint %}

Devtron's CI pipeline provides a [**Scan for vulnerabilities**](creating-application/workflow/ci-pipeline.md#scan-for-vulnerabilities) option as shown below. Once you enable this option, it will automatically scan the image for vulnerabilities.

![Figure 1: Scan for vulnerabilities](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/enable-image-scan.jpg)

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have Admin permission or above (along with access to the environment and application) to enable the **Scan for vulnerabilities** option.
{% endhint %}

---

## Checking Comprehensive Vulnerability Scan Report

To access the comprehensive security scan reports, follow these steps:

1. In the left sidebar, click **Security** and go to the `Security Scans` tab.

    {% hint style="warning" %}
    ### Who Can Perform This Action?
    Users need to have super-admin permission to view the 'Security Scans' page.
    {% endhint %}

2. Select the desired application from the available list.

    ![Figure 3: Application Security Scans](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/security-scans.jpg)

    This action provides a detailed overview of the application's security scan, including CVE IDs, severity levels of vulnerabilities, and more, as shown below.

    ![Figure 4: Detailed Scan Report](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/security-scans-report.jpg)

Each vulnerability is identified by a **CVE ID** and categorized based on **Severity**, **Package**, **Current Version**, and **Fixed In Version**.

* **CVE ID** - Refers to the Common Vulnerability ID assigned to each vulnerability.
* **Severity** - Indicates the severity of the vulnerability and can be classified as Critical, Medium, or Low.
* **Package** - Contains metadata associated with the vulnerability. The `CURRENT VERSION` refers to the specific version of the vulnerability.
* **Fixed In Version** - Displays the version name if the vulnerability has been addressed in a subsequent release; otherwise, it remains blank.

---

## Discover Vulnerabilities on the Trigger Page

Devtron provides the capability to identify vulnerabilities before image deployment in the Continuous Deployment (CD) pipeline. This ensures that potential vulnerabilities are detected and addressed early in the deployment process.

To access security vulnerability details during image deployment in Devtron, follow these steps:

1. Click **Show Source Info** option for the desired image during the deployment process.
2. Navigate to the `Security` tab.

![Figure 5: Vulnerability Details](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/show-vulnerabilities-before-deployment.jpg)

In the `Security` tab, you will find the security vulnerability details associated with the image. 

{% hint style="info" %}
Vulnerability information will only be displayed for images that have undergone vulnerability scanning. If no vulnerabilities were identified during the scan, the **Security** tab will display a zero count, indicating **Security (0)**.
{% endhint %}

---

## Accessing Vulnerability Information on the App Details Page

Devtron offers the capability to identify vulnerabilities even after an image has been deployed. By navigating to the `App Details` page, you can find comprehensive details about the vulnerabilities associated with the deployed image.

With this capability, Devtron empowers users to stay informed about the security vulnerabilities present in their deployed images.

![Figure 6: App Details Tab](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/vulnerability-app-details-v2.jpg)

Clicking the 'Details' link in the security vulnerabilities report (shown above) reveals detailed information about those found within the deployed image.

![Figure 7: Detailed List of Vulnerabilities](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/detailed+vulnerability-app-details-v2.jpg)


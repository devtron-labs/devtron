# Security Features

## Introduction

Devtron provides strong security features that help identify vulnerabilities in container images. The system scans container images thoroughly and generates reports if any vulnerabilities are found. 

Devtron's CI pipeline includes an option called [**Scan for vulnerabilities**](creating-application/workflow/ci-pipeline.md#scan-for-vulnerabilities). 

![Figure 1: Scan for vulnerabilities](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/enable-image-scan.jpg)

{% hint style="info" %}
Users need to have Admin permission or above (along with access to the environment and application) to enable the **Scan for vulnerabilities** option.
{% endhint %}

By enabling this option, the system automatically scans the container image after the image build stage. It then generates a report that highlights all the vulnerabilities present within the image. To access the scan report of all builds with vulnerability scans enabled, simply navigate to the 'Security' tab on the dashboard. There you can conveniently view the build history and all the vulnerabilities detected in the build image.

![Figure 2: Security tab](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/scan-report-on-build-history.jpg)

The scan report provides a comprehensive overview of any vulnerabilities present in the image. This allows you to identify and address potential security risks effectively. By leveraging this feature, you can ensure that your containerized applications are safeguarded against known vulnerabilities.

Devtron's Security Feature consists of two primary components:

1. [Security Scans](#security-scans)
2. [Security Policies](#security-policies)

---

## Security Scans

Devtron's security scans provide comprehensive scan reports for all applications scanned for vulnerabilities. These reports offer a detailed overview of the security status of each scanned application.

These comprehensive scan reports provide valuable insights, including information about identified vulnerabilities, their severity levels, and any corresponding Common Vulnerabilities and Exposures (CVE) entries.

{% hint style="info" %}
Users need to have super-admin permission to view the `Security Scans` page.
{% endhint %}

---

## Security Policies

Devtron's Security Policies feature allows users to define policies based on the severity levels of vulnerabilities, which include `Critical`, `Moderate`, and `Low`. Users have the flexibility to set policies that either block the deployment of container images with vulnerabilities or allow their deployment.

With this feature, users can specify their desired actions for each severity level. For example, they can choose to block any container image with `Critical` vulnerabilities, while allowing container images with `Moderate` or `Low` vulnerabilities to be deployed.

{% hint style="info" %}
Users need to have super-admin permission to define or modify security policies. 
{% endhint %}

### Checking Comprehensive Vulnerability Scan Report

To access the comprehensive security scan reports, follow these steps:

1. Navigate to the `Security` tab within Devtron.
2. Select the desired application from the available list.

![Figure 3: Application Security Scans](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/security-scans.jpg)

This action provides a detailed overview of the application's security scan, including CVE IDs, severity levels of vulnerabilities, and more, as shown below.

![Figure 4: Detailed Scan Report](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/security-scans-report.jpg)

Each vulnerability is identified by a **CVE ID** and categorized based on **Severity**, **Package**, **Current Version**, and **Fixed In Version**.

* **CVE ID** - Refers to the Common Vulnerability ID assigned to each vulnerability.
* **Severity** - Indicates the severity of the vulnerability and can be classified as Critical, Medium, or Low.
* **Package** - Contains metadata associated with the vulnerability. The current Version refers to the specific version of the vulnerability.
* **Fixed In Version** - Displays the version name if the vulnerability has been addressed in a subsequent release; otherwise, it remains blank.

### Discover Vulnerabilities on the Trigger Page

Devtron provides the capability to identify vulnerabilities before image deployment in the Continuous Deployment (CD) pipeline. This ensures that potential vulnerabilities are detected and addressed early in the deployment process.

To access security vulnerability details during image deployment in Devtron, follow these steps:

1. Click **Show source info** option for the desired image during the deployment process.
2. Navigate to the `Security` tab.

![Figure 5: Vulnerability Details](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/show-vulnerabilities-before-deployment.jpg)

In the `Security` tab, you will find the security vulnerability details associated with the image. 

{% hint style="info" %}
Vulnerability information will only be displayed for images that have undergone vulnerability scanning. If no vulnerabilities were identified during the scan, the **Security** tab will display a zero count, indicating **Security (0)**.
{% endhint %}


### Accessing Vulnerability Information on the App Details Page

Devtron offers the capability to identify vulnerabilities even after an image has been deployed. By navigating to the `App Details` page, you can find comprehensive details about the vulnerabilities associated with the deployed image.

With this capability, Devtron empowers users to stay informed about the security vulnerabilities present in their deployed images.

![Figure 6: App Details Tab](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/vulnerability-app-details.jpg)

Clicking the 'Details' link in the security vulnerabilities report (shown above) reveals detailed information about those found within the deployed image.

![Figure 7: Detailed List of Vulnerabilities](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/detailed+vulnerability-app-details.jpg)

---

## Configuring Security Policies

You can establish security policies for their vulnerabilities through the `Security Policies` tab, which can be accessed from the left pane by navigating to `Security` and selecting `Security Policies`. Policies are implemented in a hierarchical order, following a specific sequence. The order of implementation is as follows, starting from the highest level:

* Global
* Cluster
* Environment
* Application

Higher-level policies take precedence over lower-level policies, ensuring a systematic and structured enforcement of security measures.

![Figure 8: Security Policies](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/security-feature-global-security-policies.jpg)

### Examples of Defining a Policy

* Users can block all vulnerabilities
* Users can block critical vulnerabilities and allow the moderate and low vulnerabilities
* Users can block all vulnerabilities for one application and can block only critical vulnerabilities for other applications
* Users can block those vulnerabilities for which a fix is already available

---

## Configure Global Security Policy

Within the Global Security Policies, there are three options available:

| Option                    | Description                                                                                        |
| ------------------------- | ---------------------------------------------------------------------------------------------------|
| Block always               | Images containing vulnerabilities will be blocked from deployment                        |
| Block if fix is available  | Images containing vulnerabilities will be blocked if a fix is available and has not been applied |
| Allow | Images containing vulnerabilities will be allowed to be deployed regardless of whether a fix is available or not |

If critical severity levels are blocked in the Global Security Policy, the same blocking will be applied to the Cluster Security Policy. Likewise, allowing critical levels in the global policy automatically allows them in Cluster Security Policies.

However, users have the flexibility to explicitly modify these policies as desired.

![Figure 9: Configuring Global Security Policy](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/global-security-policy-1.jpg)

---

## Configure Cluster Security Policy

Cluster Security Policies offer the same three options as [Global Security Policies](#configure-global-security-policy) for handling vulnerabilities. However, an extra option called `Inherit` is available too.

When `Inherit` is selected, the policy adopts settings from higher-level options. For example, if critical severity levels are blocked globally, they will also be blocked in Cluster Security Policies. Changing the global policy to allow critical levels will also allow them in Cluster Security Policies. Explicit changes can be made to these policies.

To block critical vulnerabilities globally but allow them in specific clusters:

1. Select the desired cluster.
2. Change the critical setting to allow.
3. This change only affects the policy of the selected cluster without impacting others or the global policy.

![Figure 10: Configuring Cluster Security Policy](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/cluster-security-policy-1.jpg)

---

## Configure Environment Security Policy

Environment Security Policies, like [Cluster Security Policies](#configure-cluster-security-policy), offer four options:

* Block always
* Block if fix is available
* Allow
* Inherit

The Environment Security Policy inherits its settings from the Cluster Security Policy, following a hierarchical structure where each level inherits the policy from its upper level.

When you select an environment, it automatically adopts the policy of the associated cluster. For example, if critical-level vulnerabilities are blocked globally but allowed in the Cluster Security Policy, the Environment Security Policy will inherit this allowance. Consequently, critical-level vulnerabilities will also be allowed in the Environment Security Policy.

However, you have the flexibility to make explicit changes to the policy if needed. This empowers you to customize the policy to align with specific requirements or preferences.

![Figure 11: Configuring Environment Security Policy](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/environment-security-policy-1.jpg)

---

## Configure Application Security Policy

The Application Security Policy operates on a similar principle as other policies and offers four options:

* Block always
* Block if fix is available
* Allow
* Inherit

However, in the Application Security Policy, the policy is determined by both the Environment option and the Application option.

When modifying the policy within a development environment, the changes will be applied to all applications within that specific development environment. This means that any adjustments made to the policy settings will be consistently applied across all applications associated with that particular development environment.

This approach ensures uniformity and streamlined management of security policies within specific environments and their corresponding applications.

![Figure 12: Configuring Application Security Policy](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/application-security-policy-1.jpg)

---

## Block or Allow Specific CVE Policies

To block or allow specific Common Vulnerabilities and Exposures (CVE) policies, simply click **Add CVE Policy**.

![Figure 13: Adding CVE Policy](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/add-cve-policy-1.jpg)

A window will appear where you can enter the CVE ID and select whether to allow or block it.

![Figure 14: Allowing/Blocking a CVE ID](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/cve-popup.jpg)

This action will determine whether image deployment is allowed or blocked based on the presence of vulnerabilities matching that particular CVE ID. Any other deployment decisions will be made according to the policies set previously.

---

## Check CVE Policy

Since Devtron provides you with the CVE IDs of all the vulnerabilities identified during image scan, you can use the `Check CVE Policy` option to know more about each vulnerability. This helps you check the exposure of your system for a given CVE-ID, for future deployments.

1. Go to **Security Policies** (tab) → **Check CVE Policy**.

    ![Figure 15: Check CVE Policy](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/check-cve-1.jpg)

2. Enter the CVE ID in the search-box (e.g., CVE-2023-1194) and click **Search**.

    ![Figure 16: Entering CVE ID](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/check-cve-2.jpg)

   You may click the link provided in the search result to view the vulnerability in detail. 

{% hint style="info" %}
All users can use the **Check CVE Policy** feature.
{% endhint %}



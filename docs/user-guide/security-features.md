# Security Features

Devtron provides strong security features that help identify vulnerabilities in container images. The system scans container images thoroughly and generates reports if any vulnerabilities are found. 

Within the CI pipeline of Devtron, there is an option called [**Scan for vulnerabilities**](creating-application/workflow/ci-pipeline.md#scan-for-vulnerabilities). 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/enable-image-scan.jpg)

By enabling this option, the system automatically scans the container image after the image build stage. It then generates a report that highlights all the vulnerabilities present within the image. To access the scan report of all builds with enabled vulnerability scans, simply navigate to the 'Security' tab on the dashboard. You can conveniently view the build history and all the vulnerabilities found in the build image there.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/scan-report-on-build-history.jpg)

 The scan report provides a comprehensive overview of any vulnerabilities present in the image. This allows you to identify and address potential security risks effectively. By leveraging this feature, you can ensure that your containerized applications are safeguarded against known vulnerabilities.

Devtron's Security Feature consists of two primary components:

1. **Security Scans**
2. **Security Policies**

## Security Scans

Devtron's security scans provide comprehensive scan reports for all applications that have undergone vulnerability scanning. These reports offer a detailed overview of the security status of each scanned application.

These comprehensive scan reports provide valuable insights, including information about identified vulnerabilities, their severity levels, and any corresponding Common Vulnerabilities and Exposures (CVE) entries.

## Security Policies

Devtron's Security Policies feature allows users to define policies based on the severity levels of vulnerabilities, which include **Critical**, **Moderate**, and **Low**. Users have the flexibility to set policies that either block the deployment of container images with vulnerabilities or allow their deployment.

With this feature, users can specify their desired actions for each severity level. For example, they can choose to block any container image with Critical vulnerabilities, while allowing container images with Moderate or Low vulnerabilities to be deployed.

### Checking Comprehensive Vulnerability Scan Report

To access the comprehensive security scan reports, follow these steps:

1. Navigate to the Security tab within Devtron.
2. Select the desired application from the available list.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/security-scans.jpg)

This action will provide detailed information regarding the security scan of the application, including the CVE, Severity of the identified vulnerabilities, as you can see in the image below.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/security-scans-report.jpg)

Each vulnerability is identified by a **CVE ID** and categorized based on **Severity**, **Package**, **Current Version**, and **Fixed In Version**.

* **`CVE ID`** refers to the Common Vulnerability ID assigned to each vulnerability.
* **`Severity`** indicates the severity of the vulnerability and can be classified as Critical, Medium, or Low.
* The **`Package`** column contains metadata associated with the vulnerability. The current Version refers to the specific version of the vulnerability.
* The **`Fixed In Version`** column displays the version name if the vulnerability has been addressed in a subsequent release; otherwise, it remains blank.

### Discover Vulnerabilities on the Trigger Page

Devtron provides the capability to identify vulnerabilities before image deployment in the Continuous Deployment (CD) pipeline. This ensures that potential vulnerabilities are detected and addressed early in the deployment process.

To access security vulnerability details during image deployment in Devtron, follow these steps:

1. Click on the `Show source info` option for the desired image during the deployment process.
2. Navigate to the `Security` tab.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/show-vulnerabilities-before-deployment.jpg)

In the `Security` tab, you will find the security vulnerability details associated with the image. 

**NOTE:** Vulnerabilities will only be displayed if a vulnerability scan has been enabled for that specific image. If no vulnerabilities are visible, it indicates that a vulnerability scan has not been performed for the image.

### Accessing Vulnerability Information on the App Details Page

Devtron offers the capability to identify vulnerabilities even after an image has been deployed. By navigating to the `app details` page, you can find comprehensive details about the vulnerabilities associated with the deployed image.

With this capability, Devtron empowers users to stay informed about the security vulnerabilities present in their deployed images.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/vulnerability-app-details.jpg)

By clicking on the 'Details' link in the security vulnerabilities report, you can access detailed information about the security vulnerabilities present inside the deployed image.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/detailed+vulnerability-app-details.jpg)

## Configuring Security Policies

You can establish security policies for their vulnerabilities through the Security Policies tab, which can be accessed from the left pane by navigating to Security and selecting Security Policies. Policies are implemented in a hierarchical order, following a specific sequence. The order of implementation is as follows, starting from the highest level:

* **Global**
* **Cluster**
* **Environment**
* **Application**

Policies are implemented in a hierarchical order, with the following sequence: Global, Cluster, Environment, and Application. Higher-level policies take precedence over lower-level policies, ensuring a systematic and structured enforcement of security measures.

![](../.gitbook/assets/security-feature-global-security-policies.png)

**Some examples of how policies can be defined**

Users can block all the critical vulnerabilities and allow the moderate and low vulnerabilities or Users can block all vulnerabilities or users can block all vulnerabilities for one application and can block only Critical vulnerabilities for other applications.

## Configure Global Security Policy

Within the Global Security Policies, there are two options available: Block and Allow.
If critical severity levels are blocked in the Global Security Policy, the same blocking will be applied to the Cluster Security Policy. Similarly, if the global policy is modified to allow critical levels, it will also allow them in Cluster Security Policies. 
However, users have the flexibility to explicitly modify these policies as desired.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/global-security-policy.jpg)

## Configure Cluster Security Policy


In Global Security Policies, there are two options: `Block` and `Allow`. Cluster Security Policies have an additional option called `Inherit`.

When Inherit is selected, the policy adopts settings from higher-level options. For example, if critical severity levels are blocked globally, they will also be blocked in Cluster Security Policies. Changing the global policy to allow critical levels will also allow them in Cluster Security Policies. Explicit changes can be made to these policies.

To block critical vulnerabilities globally but allow them in specific clusters:

1. Select the desired cluster.
2. Change the critical setting to allow.
3. This change only affects the policy of the selected cluster without impacting others or the global policy.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/cluster-security-policy.jpg)

## Configure Environment Security Policy

Environment Security Policies, like Cluster Security Policies, offer three options: Block, Allow, and Inherit.

The Environment Security Policy inherits its settings from the Cluster Security Policy, following a hierarchical structure where each level inherits the policy from its upper level.

When you select an environment, it automatically adopts the policy of the associated Cluster. For example, if critical-level vulnerabilities are blocked globally but allowed in the Cluster Security Policy, the Environment Security Policy will inherit this allowance. Consequently, critical-level vulnerabilities will also be allowed in the Environment Security Policy.

However, you have the flexibility to make explicit changes to the policy if needed. This empowers you to customize the policy to align with specific requirements or preferences.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/environment-security-policy.jpg)

## Configure Application Security Policy

The Application Security Policy operates on a similar principle as other policies. However, in the Application Security Policy, the policy is determined by both the Environment option and the Application option.

When modifying the policy within a development environment, the changes will be applied to all applications within that specific development environment. This means that any adjustments made to the policy settings will be consistently applied across all applications associated with that particular development environment.

This approach ensures uniformity and streamlined management of security policies within specific environments and their corresponding applications.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/application-security-policy.jpg)

## Block or Allow Specific CVE Policies.

To block or allow specific Common Vulnerabilities and Exposures (CVE) policies, simply click on `Add CVE Policy`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/add-cve-policy.jpg)

A window will appear where you can enter the CVE ID and select whether to Allow or Block it.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/cve-popup.jpg)

This action will determine whether image deployment is allowed or blocked based on the presence of vulnerabilities matching that particular CVE ID. Any other deployment decisions will be made according to the policies set previously.


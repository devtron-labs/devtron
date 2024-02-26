# Security Policies

{% hint style="info" %}
### Prerequisite

Install any one of the following integrations for scanning vulnerabilities:
* [Clair](../../user-guide/integrations/clair.md)
* Trivy
{% endhint %}

Devtron's Security Policies feature allows users to define policies based on the severity levels of vulnerabilities, which include `Critical`, `Moderate`, and `Low`. Users have the flexibility to set policies that either block the deployment of container images with vulnerabilities or allow their deployment.

With this feature, users can specify their desired actions for each severity level. For example, they can choose to block any container image with `Critical` vulnerabilities, while allowing container images with `Moderate` or `Low` vulnerabilities to be deployed.

For in-depth instructions, refer to the [Configure Security Policies](#configuring-security-policies) section.

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to define or modify security policies. 
{% endhint %}

---

## Configuring Security Policies

You can establish security policies for their vulnerabilities through the `Security Policies` tab, which can be accessed from the left pane by navigating to `Security` and selecting `Security Policies`. 

You can define policies at the following levels:

* [Global](#configure-global-security-policy)
* [Cluster](#configure-cluster-security-policy) 
* [Environment](#configure-environment-security-policy)
* [Application](#configure-application-security-policy)

![Figure 1: Security Policies](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/security-feature-global-security-policies.jpg)

However, if you define policies at more than one level, the order of precedence would be as follows:

* Application + Environment (highest priority)
* Environment 
* Cluster
* Global

### Examples of Defining a Policy

* Users can block all vulnerabilities
* Users can block critical vulnerabilities and allow moderate and low vulnerabilities
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

![Figure 2: Configuring Global Security Policy](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/global-security-policy-1.jpg)

If critical severity levels are blocked in the Global Security Policy, the same blocking will be applied to the Cluster Security Policy. Likewise, allowing critical levels in the global policy automatically allows them in Cluster Security Policies.

However, users have the flexibility to explicitly modify these policies as desired.

---

## Configure Cluster Security Policy

Cluster Security Policies offer the same three options as [Global Security Policies](#configure-global-security-policy) for handling vulnerabilities. However, an extra option called `Inherit` is available too.

![Figure 3: Configuring Cluster Security Policy](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/cluster-security-policy-1.jpg)

When `Inherit` is selected, the policy adopts settings from higher-level options. For example, if critical severity levels are blocked globally, they will also be blocked in Cluster Security Policies. Changing the global policy to allow critical levels will also allow them in Cluster Security Policies. Explicit changes can be made to these policies.

To block critical vulnerabilities globally but allow them in specific clusters:

1. Select the desired cluster.
2. Change the critical setting to allow.
3. This change only affects the policy of the selected cluster without impacting others or the global policy.

---

## Configure Environment Security Policy

Environment Security Policies, like [Cluster Security Policies](#configure-cluster-security-policy), offer four options:

* Block always
* Block if fix is available
* Allow
* Inherit

![Figure 4: Configuring Environment Security Policy](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/environment-security-policy-1.jpg)

The Environment Security Policy inherits its settings from the Cluster Security Policy, following a hierarchical structure where each level inherits the policy from its upper level.

When you select an environment, it automatically adopts the policy of the associated cluster. For example, if critical-level vulnerabilities are blocked globally but allowed in the Cluster Security Policy, the Environment Security Policy will inherit this allowance. Consequently, critical-level vulnerabilities will also be allowed in the Environment Security Policy.

However, you have the flexibility to make explicit changes to the policy if needed. This empowers you to customize the policy to align with specific requirements or preferences. Any adjustments made to the environment policy settings will be consistently applied across all applications associated with that environment.

---

## Configure Application Security Policy

The Application Security Policy operates on a similar principle as other policies and offers four options:

* Block always
* Block if fix is available
* Allow
* Inherit

However, in the Application Security Policy, the policy is determined by both: Application and Environment

First, choose an application from the list.

![Figure 5a: Configuring Application Security Policy - Choosing an App](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/app-selection.jpg)

Next, configure a security policy for that application in the intended environment.

![Figure 5b: Configuring Application Security Policy - Choosing an Env](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/app-env-v2.jpg)

---

## Example

1. Let's say, you have defined a policy to block the deployment if critical vulnerabilities are found in a given application.

    ![Figure 6: Defining a Block Policy](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/blocked-example.jpg)

2. Now, go to the **Build & Deploy** tab of that application to select an image.

    ![Figure 7: Selecting an Image](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/select-image.jpg)

3. As you can see, security issues were found in the scanned image, hence it is not available for selection. Click **Show Source Info**.

    ![Figure 8: Blocked Deployment of Image](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/source-info.jpg)

4. The `Security` tab shows the critical vulnerabilities and the policy enforced to prevent deployment.

    ![Figure 9: Detected Vulnerabilities](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/blocked-deployment.jpg)

---

## Block or Allow Specific CVE Policies

To block or allow specific Common Vulnerabilities and Exposures (CVE) policies, simply click **Add CVE Policy**.

![Figure 10: Adding CVE Policy](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/add-cve-policy-1.jpg)

A window will appear where you can enter the CVE ID and select whether to allow or block it.

![Figure 11: Allowing/Blocking a CVE ID](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/security-features/new/cve-popup-v2.jpg)

This action will determine whether image deployment is allowed or blocked based on the presence of vulnerabilities matching that particular CVE ID. Any other deployment decisions will be made according to the policies set previously.





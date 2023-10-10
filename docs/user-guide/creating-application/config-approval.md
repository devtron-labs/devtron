# Protect Configuration

## Introduction

Since resources are created according to the configurations you enter, it's essential to restrict such configurations from direct modifications. For critical environments like production, it becomes necessary to introduce an approval flow for any edits made to the configuration files.

In Devtron, these configurations are present in the **App Configuration** tab of your application.

Any changes made to the following configurations will require approval if enabled:

- Deployment Template
- ConfigMaps
- Secrets

This stands true for both: base configuration and respective environment-level configuration.

![Figure 1a: Protecting Base Configuration](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/config-approval/base-config.jpg)

![Figure 1b: Protecting Environment-level Configuration](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/config-approval/env-config.jpg)

---

## Changing the Configuration Values

Let's say you wish to edit the deployment template of your environment (as an override)

1. Go to the `App Configuration` tab.

2. In Environment Overrides → (choose your environment) → Deployment Template 

    ![Figure 2: Selecting Values to Change](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/config-approval/edit-deployment-template.jpg)

3. You can change the value of a key to a desired value as shown below. Once done, click the **Save Changes…** button **.**

    ![Figure 3: Changed Values](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/config-approval/changed-values.jpg)

4. If the configuration is protected, your changes won't be published right away. You can do either of the following:

    * **Save as draft** : Selecting this option will save your file as a draft. You and other users can view and edit the saved draft and propose it further for approval.
    * **Save & Propose Changes** : Selecting this option will propose your changes to a configuration approver for a review.

    Since we are proposing the changes immediately, click **Propose Changes**.

    ![Figure 4: Proposing Changes](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/config-approval/propose-changes.jpg)

5. You can also view the approver(s) if you wish.

    ![Figure 5: Viewing the Approver(s)](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/config-approval/approver-list-1.jpeg)

{% hint style="info" %}
The one who performs the edits cannot approve their own changes. A different user has to review and approve.
{% endhint %}

Only one draft can exist at time and you cannot create multiple drafts. In the top-right corner, you have the option to discard the draft if you don't wish to proceed with the edits you made. 

---

## Approving the Configuration

Only a different super-admin user or someone (who is not amongst the editors of the draft), having `Configuration approver` access, can approve the changes made to the configuration files as shown below.

![Figure 6: Authorized User Approving the Changes](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/config-approval/approval-screen.jpg)

A super-admin can check whether a user has approval rights by going to **Global Configurations** → **Authorization** (dropdown) → **User Permissions**.

![Figure 7: Checking Permissions](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/config-approval/config-approver-1.jpg)


---

## Deploying with New Configuration Values

Once the approver validates and approves your configuration changes, you can proceed to deploy your application with the updated configuration.

1. Go to the **Build & Deploy** tab of your application.

2. Click **Select Image** in the deployment flow. 

    ![Figure 8: Build & Deploy Page](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/config-approval/deploy-pipeline.jpg)

3. You can view an indicator at the bottom `Config Diff from Last Deployed`. Click **Review** to view the changes.

    ![Figure 9: Checking Config Diff](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/config-approval/review-config-diff.jpg)

{% hint style="info" %}
If the new configuration is not yet approved, the changes made to the config would not be visible during deployment, it would show `No Config Diff from Last Deployed` at the bottom. In that case, check whether your changes are present in the live config or not. If your changes are absent, chances are your draft is either pending for approval or rejected (discarded).
{% endhint %}

4. Once you have verified the changes, you can click **Deploy**.

    ![Figure 10: Deploying App using New Config](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/config-approval/deploy.jpg)

{% hint style="info" %}
If you don't wish to deploy with the new changes, you can choose `Last deployed config` from the available drop-down.
{% endhint %}

## Enabling/Disabling Config Protection

1. Go to the `App Configuration` tab.

2. Click `Protect Configuration`.

3. Use the toggle button to enable the protection for the configuration of your choice (base/environment level). A protection badge would appear next to the chosen configuration.

Alternatively, unprotecting the configuration will lead to the discarding of unapproved drafts (if any).

![Figure 11a: Disabling Protection](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/config-approval/toggle-protection.jpg)

![Figure 11b: Confirming Action](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/config-approval/unprotect.jpg)


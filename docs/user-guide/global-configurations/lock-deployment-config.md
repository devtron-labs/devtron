# Lock Deployment Config

## Introduction [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

The [Deployment Template](../../reference/glossary.md#base-deployment-template) might contain certain configurations intended for the DevOps team (e.g., `ingress`), and not meant for developers to modify. 

Therefore, Devtron allows super-admins to restrict such fields from modification or deletion.

![Figure 1: Preventing Changes to Locked Keys](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/lock-dt/not-eligible-changes.jpg)

This stands true for deployment templates in:
* [Base configuration](../../user-guide/creating-application/deployment-template.md)
* [Environment-level configuration](../../user-guide/creating-application/environment-overrides.md)

**How is this different from the [Protect Configuration](../../user-guide/creating-application/config-approval.md) feature?**

The 'protect configuration' feature is meant to verify the edits by introducing an approval flow for any changes made to the configuration files, i.e., Deployment template, ConfigMaps, and Secrets. This is performed at application-level.

Whereas, the 'lock deployment config' feature goes one step further. It is meant to prevent any edits to specific keys by non-super-admins. This applies only to deployment templates and is performed at global-level.

---

## Locking Deployment Keys

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to lock deployment keys.
{% endhint %}

1. Go to **Global Configurations** → **Lock Deployment Config**. Click **Configure Lock**.

    ![Figure 2: Configure Lock Button](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/lock-dt/lock-deployment-config.jpg)

2. (Optional) Click **Refer Values.YAML** to check which keys you wish to lock.

    ![Figure 3: Values.YAML File](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/lock-dt/values-yaml.jpg)

3. Enter the keys inside the editor on the left-hand side, e.g., `autoscaling.MaxReplicas`. Use <a href="https://goessner.net/articles/JsonPath/index.html" target="_blank">JSONpath expressions</a> to enter specific keys, lists, or objects to lock.

    ![Figure 4: Referring Values.YAML File for Locking Keys](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/lock-dt/autoscaling-lock.jpg)

4. Click **Save**. 

    ![Figure 5: Saving Locked Keys](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/lock-dt/saving-locked-keys.jpg)

5. A confirmation dialog box would appear. Read it and click **Confirm**.

    ![Figure 6: Confirmation Dialog](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/lock-dt/confirmation.jpg)

---

## Result

While super-admins can directly edit the locked keys, let's look at a scenario where a user (non-super-admin) tries to edit the same in an [unprotected](../../user-guide/creating-application/config-approval.md) base deployment template.

* User can hide/unhide the locked keys as shown below.

    ![Figure 7: Hiding Locked Keys](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/lock-dt/hide-locked-keys.gif)

{% hint style="info" %}
<span><img src="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/lock-dt/mode.jpg" alt="GUI or YAML Mode"></span> <br />
If you select 'Basic' mode instead of 'Advanced (YAML)', all the keys meant for basic mode will be displayed in the GUI even if some are locked. While users can modify these keys, they cannot save the changes made to the locked keys.
{% endhint %}

* Let's assume the user edits one of the locked keys...

    ![Figure 8: Editing Locked Keys](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/lock-dt/change-locked-values.gif)

    ...and saves the changes.

    ![Figure 9: Saving Edits to Locked Keys](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/lock-dt/changing-values.jpg)

* A modal window will appear on the right highlighting the non-eligible edits.

    ![Figure 10: Eligible and Non-eligible Changes](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/lock-dt/not-eligible-changes.jpg)

* Let's assume the user edits a key that is not locked or adds a new key.

    ![Figure 11: Editing Allowed Keys](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/lock-dt/changing-allowed-values.jpg)

* The modal window will highlight the eligible edits. However, it will not let the user save those eligible edits unless the user clicks the checkbox: **Save changes which are eligible for update**.

    ![Figure 12: Saving Eligible Changes](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/lock-dt/saving-allowed-changes.jpg)

{% hint style="warning" %}
### Who Can Perform This Action?
Only a super-admin, manager, or application admin can edit the configuration values. 
{% endhint %}

* Once the user clicks the **Update** button, the permissible changes will reflect in the deployment template. 

    ![Figure 13: Updating Deployment Config](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/lock-dt/updating-changes.jpg)

    However, if it's a [protected template](../../user-guide/creating-application/config-approval.md), the user will require the approval of a [configuration approver](./user-access.md#role-based-access-levels) as shown below.

    ![Figure 14: Proposing Changes to Protected Config](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/lock-dt/proposing-changes.jpg)

The same result can be seen if the user tries to edit environment-specific deployment templates.


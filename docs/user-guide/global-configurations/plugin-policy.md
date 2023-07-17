# Plugin Policy

Devtron's `Plugin Policy` feature allows the Superadmin to define plugin policies that enforce the usage of specific plugins within workflows. 
Enforced plugin policies play a crucial role in maintaining the integrity and reliability of pipelines within Devtron. 
When a plugin policy is enforced, it becomes mandatory for application owners to configure the specified plugins in their workflows. Failure to adhere to the enforced policy may result in pipelines becoming non-functional.

## 1. Add and Configure Plugin Policies

**To add plugins from the mandatory plugins list, please follow these steps:**

**1.** Navigate to the Global Configuration from the left pane.
**2.** Select the Plugins option and click on `Add` button present on upper-right corner to add new plugin policy .

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/mandatory-plugins/add-plugin-policy.jpg
)
**3.** In the Mandatory Plugins section, you will find a list of available plugins. Select the desired plugins by checking the corresponding checkboxes.
**4.** After selecting the plugins, click `Next` to configure policies on the enforced plugins.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/mandatory-plugins/select-plugins.jpg)

**5.** **Enter a policy name:** Provide a descriptive name for the policy you are configuring.

**6.** **Enter a policy description:** Add a brief summary or description for better understanding.

**7.** **Set Policy Condition:** Define the policy condition using Applications, Environments, and Branch names. The condition is applied using the AND operation. Follow these sub-steps:

**a. Select Applications:** Choose the applications on which you want to apply the policy. If multiple applications need to be selected, the OR operation will be applied.

**b. Select Environments:** Choose the environments on which you want to apply the policy. If you need to select multiple environments, apply the OR operation. You have the option to either select 'all prod environment' or choose specific environments as per your preference.

**c. Select Branch names:** Choose the branches on which you want to apply the policy. If multiple branches need to be selected, the OR operation will be applied.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/mandatory-plugins/configure-plugin-policy.jpg)

By configuring these policy, you can define specific conditions for enforcing plugins based on selected applications, environments, and branch names. Remember that the AND operation is used to apply the conditions collectively.

## 2. Plugin Enforcement at Pipeline Stages

After configuring the policy condition, specify the stage at which each policy should be applied. The following options are available:

**Pre-build:** Apply the policy before the build stage.
**Post-build:** Apply the policy after the build stage.
**Let users decide:** Allow users to decide when to add the plugin while configuring the workflow.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/mandatory-plugins/plugin-enforcement-stages.jpg)

## 3. Managing Non-Compliance with Policies

In addition to configuring and applying plugin policies, you can determine the behavior when a policy is not adhered to. Two options are available:

**Block Build trigger:** If selected, users will be unable to trigger a build if the policy is not adhered to in the workflow. This ensures strict enforcement of the policy, preventing builds from being triggered until compliance is achieved.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/mandatory-plugins/block-trigger.jpg)

If you wish to trigger this pipeline, you can edit the CI pipeline and configure the mandatory plugins at pre-build or post-build stage. Once the plugins are configured, you can proceed to trigger the build on the pipeline.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/mandatory-plugins/configure-plugins-within-pipeline.jpg)

**"Allow build trigger with warning" option:** Users can trigger builds even if the policy is not followed, but a warning will be displayed with each build trigger to notify users of non-compliance. 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/mandatory-plugins/warning.jpg)

Build triggers without policy adherence are allowed until a specified date and time.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/mandatory-plugins/allow-until.jpg)

Additionally, there is an option to set this configuration as `Forever` if the superadmin wants to allow build triggers without policy adherence indefinitely.

## 4. Save and Enforce Policy

After configuring all the options for the plugin policy, save and activate the policy to ensure its effectiveness. Note that saving the policy alone will not make it effective. Activate the policy by checking the corresponding checkbox and then save it.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/mandatory-plugins/save-and-enforce.jpg)

## 5. Editing/Deleting Plugin Policies

To edit or delete policies, follow these steps:

Go to global configuration and navigate to Plugins.
Hover over the plugin policy you want to edit or delete.
Click on the edit or delete icon to perform the respective action.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/mandatory-plugins/edit-delete-plugin-policy.jpg)

From this section, you can also view non-compliant pipelines and take necessary actions accordingly.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/mandatory-plugins/non-compliant-pipelines.jpg)

Additionally, you can click on the `Take me there` button, which will redirect you to the pipeline edit page. From there, you can configure the mandatory plugins in accordance with the plugin policy.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/mandatory-plugins/no-compliant-pipelines.jpg)

## 6. Impact of Policy/Plugin Configuration in Different CI Scenarios

When configuring policies or plugins for pre/post CI stages in workflows involving linked CI or external CI, consider their impact on deployment, rollback, and plugin enforcement. The following behaviors should be observed:

* If a policy/plugin is configured for pre/post CI stages in workflows with linked CI, the plugin will be enforced in the parent CI, impacting its trigger.
* For workflows with external CI, the mandatory plugin cannot be enforced on the external CI itself.
* A policy change can make a plugin no longer mandatory, allowing it to become optional.
* Policies enforced on auto-triggered pipelines can notify users through configured notifications.

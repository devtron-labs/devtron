# Tags Policy

Devtron's `Tags Policy` feature enables you to assign tags to your applications. Devtron also offers the option to propagate the tags assigned to an application as labels within the associated Kubernetes resources. 

## Add Tags

To mandatory tags, follow these steps:

1. From the left pane, navigate to the Global Configuration section.
2. Select Tags within the Global Configuration section.

Once you are in the Tags section, locate the `Add Tag` button in the upper-right corner of the screen. Click on this button to proceed.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/mandatory-tags/add-tag.jpg)

Within the Add Tag section, you will find two options for tags:

* **Suggested tags:** These tags appear as suggestions when adding tags to applications.
* **Mandatory tags:** These tags are required for applications within the selected project.

To create mandatory tags, choose the second option: `Mandatory tags`. This ensures that the specified tags are mandatory for the applications within the selected project.

3. Next, choose the project(s) for which you want to create mandatory tags. You can select multiple projects at once.

4. After selecting the projects, proceed to add the mandatory tags for the selected projects.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/mandatory-tags/create-tag.jpg)

## Propagate Tags to Kubernetes Resources

By default, tags assigned to applications in Devtron are not automatically propagated to Kubernetes resources as labels. However, Devtron provides the flexibility to enable this feature if desired.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/mandatory-tags/propagation-enable-disable.jpg)

When the propagation is enabled for tags from the global configuration, the tags will be automatically propagated as labels for all applications within the projects where these tags are used. Even if tag propagation is disabled from the global configuration in Devtron, you still have the option to enable propagation at the application level.

## Create Application With Mandatory Tags

In a project where mandatory tags are enabled, it is required to provide values for those tags when creating new applications. Without providing values for the mandatory tags, it is not possible to create a new application within that project.

When mandatory tags are enabled, Devtron enforces the requirement to specify values for these tags during the application creation process. This ensures that all applications within the project adhere to the specified tag values.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/mandatory-tags/app-creation-with-tags.jpg)

## Propagate Tags to Kubernetes Resources While Creating Application

If tag propagation for a project is disabled globally, you can still enable it for individual applications. During the application creation process, you have the option to enable tag propagation specifically for that application. By doing so, the tags assigned to that application will be propagated as labels to the associated Kubernetes resources.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/mandatory-tags/propagation-enable-disable.jpg)
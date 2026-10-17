# Environment Overrides

You will see all your environments associated with an application under the `Environment Overrides` section.

![Figure 1: App Configuration → Environment Overrides](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/environment-overrides/environment-override-v2.jpg)


You can customize your `Deployment template, ConfigMap, Secrets` in Environment Overrides section to add separate customizations for different environments such as dev, test, integration, prod, etc.

---

## Deployment template - Functionality

If you want to deploy an application in a non-production environment and then in production environment, once testing is done in the non-production environment, then you do not need to create a new application for production environment. Your existing pipeline\(non-production env\) will work for both the environments with little customization in your deployment template under `Environment overrides`.

### Example customization:

In a Non-production environment, you may have specified 100m CPU resources in the deployment template but in the Production environment, you may want to have 500m CPU resources as the traffic on Pods will be higher than traffic on non-production env.

Configuring the Deployment template inside `Environment Overrides` for a specific environment will not affect the other environments because `Environment Overrides` will configure deployment templates on environment basis. And at the time of deployment, it will always pick the overridden deployment template if any.

If there are no overrides specified for an environment in the `Environment Overrides` section, the deployment template will be the one you specified in the `deployment template section` of the app creation.

*(Note: This example is meant only for a representational purpose. You can choose to add any customizations you want in your deployment templates in the `Environment Overrides` tab)*

Any changes in the configuration will not be added to the template, instead, it will make a copy of the template and lets you customize it for each particular environment. And now this overridden template will be used only for the specified Environment.

This will save you the trouble to manually create deployment files separately for each environment. Instead, all you have to do is to change the required variables in the deployment template.

---

## How to add Environment Overrides

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have [Admin role](../global-configurations/authorization/user-access.md#role-based-access-levels) or above (along with access to the environment and applications) to change perform environment override.
{% endhint %}

Go to **App Configuration** → **Environment Overrides**. For each environment you can override the following configurations:
* [Deployment Template](#deployment-template)
* [ConfigMaps](#configmaps--secrets)
* [Secrets](#configmaps--secrets)

### Deployment Template

{% hint style="info" %}
Users who are not super-admins will land on [Basic (GUI)](#basic-gui) section when they visit the **Deployment Template** page; whereas super-admins will land on [Advanced (YAML)](#advanced-yaml) section. This is just a default behavior, they can still navigate to the other section if needed.
{% endhint %}

#### Basic (GUI)

If you have a [base deployment configuration](../creating-application/deployment-template/deployment.md#2.-basic-configuration) set up at application level, the environment(s) you define for your application will also inherit those values.

However, you have the flexibility to use different values at the environment level by overriding the base configurations as shown below. 

![Figure 2: Overriding Deployment Template - GUI Method](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/environment-overrides/base-config-override.gif)

Refer [basic configuration](../creating-application/deployment-template/deployment.md#2.-basic-configuration) to know more about each field within `Basic (GUI)` section.

{% hint style="info" %}
### Want to customize the fields displayed on Basic (GUI)? [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)
Refer [Customize Basic GUI](../creating-application/deployment-template.md#customize-basic-gui) to know the process of adding, removing, and customizing the Basic (GUI) section.
{% endhint %}

#### Advanced (YAML)

Similarly, if you are an advanced user intending to tweak more values in deployment template, you may go to `Advanced (YAML)` section and edit them.

![Figure 3: Overriding Deployment Template - YAML Method](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/environment-overrides/yaml-override.jpg)

[Click here](../creating-application/deployment-template/deployment.md#3-advanced-yaml) to know more about each key-value pair within the `Advanced (YAML)` section.

{% hint style="warning" %}
**Delete Override** will discard the current overrides and the base deployment configuration will be applicable again to the environment. 
{% endhint %}

### ConfigMaps & Secrets

The same goes for `ConfigMap` and `Secrets`. You can also create an environment-specific configmap and Secrets inside the `Environment override` section.

If you want to configure your ConfigMap and secrets at the application level then you can provide them in [ConfigMaps](config-maps.md) and [Secrets](secrets.md), but if you want to have environment-specific ConfigMap and secrets then provide them under the Environment override Section. At the time of deployment, it will pick both of them and provide them inside your cluster.

To update a ConfigMap, follow the steps below:
1. In your environment, click **ConfigMaps**.
2. Click the ConfigMap you wish to update.
3. Click **Allow Override**.
4. Edit your ConfigMap.
5. Click **Save Changes**.

![Figure 2: Updating ConfigMap](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/environment-overrides/update-configmap.gif)

Similarly, you can update Secrets too as shown below.

![Figure 3: Updating Secret](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/environment-overrides/update-secret.gif)


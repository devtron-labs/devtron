# Deployment Window

## Introduction [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

Unplanned or unwanted deployments of applications can affect the services of an organization. Consequently, its business impact will be severe if such disruptions occur during peak hours or critical periods (say festive season).

Therefore, Devtron comes with a feature called 'Deployment Window' that allows you to define specific timeframes during which application deployments are either blocked or allowed in specific environments.

### Types of Deployment Window

| Name  | Blackout Window                                    | Maintenance Window |
| --------------------- | ---------------------------------------------------|--------------------|
| **Definition** | Time period during which deployments are not allowed | Only time period during deployments are allowed |
| **Use** | To block deployments when systems are already stable and running a critical business in peak hours | To allow deployments preferably during non-business hours so as to minimize any negative impact on end-users |
| **In case of overlap?** | Blackout window gets a higher priority over maintenance window | Maintenance window has a lower priority |


### Difference between a Blackout Window and Maintenance Window

Technically, both of them are different methods of restricting deployments to an environment. For example, specifying either a blackout window of [8:00 AM to 10:00 PM] or a maintenance window of [10:00 PM to 8:00 AM] essentially does the same job. You can define either of them depending on your use case.

---

## Configuring Deployment Window

(This is configured by super-admins. Details yet to be added.)

---

## Checking Deployment Window

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have view-only permission or above (along with access to the environment and application) to check deployment windows.
{% endhint %}

### Overview Page

The **Deployment window** section shows the deployment windows configured for each [environment](../../reference/glossary.md#environment).

![Figure 1: Overview Page - Deployment Window](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/overview-deployment-win.jpg)

However, if a deployment window doesn't exist for an environment, the message `No deployment windows are configured` would be displayed next to it.

You may click the dropdown icon to view the details which include:
* Type of deployment window (Blackout/Maintenance)
* Name and description
* Frequency of window (once, weekly, monthly, yearly)
* Start date and time 
* End date and time

### App Details Page

Unlike the **Overview** page which shows deployment windows for all environments, the **App Details** page shows the deployment windows for one environment at a time.

![Figure 1: App Details Page - Deployment Window](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/app-details-deployment-win.jpg)

If configured, the ongoing and upcoming deployment windows would be visible in the form of cards (in the same row that has `Application Status`, `Deployment Status`, etc.)

For example, if the super-admin has configured 4 deployment windows (say 2 Blackout and 2 Maintenance), you will see 4 cards stacked upon each other. However, no cards will be shown if deployment windows aren't configured.

You may click on any of them to view the details of the deployment windows.

---

## Result

The below functions are blocked during an ongoing blackout window or outside maintenance window.

* [Hibernation](#hibernation)
* [Restart Workloads](#restart-workloads)
* [Deployment](#deployment)
* [Deletion of CD Pipeline](#deletion-of-cd-pipeline)

{% hint style="info" %}
However, the exempted users specified in deployment window configuration can bypass the restrictions and perform the above actions.
{% endhint %}


### Hibernation

As you know, every Devtron application comes with an option to hibernate. It is generally used when you do not want your application(s) to consume resources and incur costs.

The hibernation process sets the replica count to 0; therefore, your application will become non-functional. To avoid this, hibernation of application is blocked.

### Restart Workloads

Restarting workloads is a common practice for deploying updates, fixing bugs, or recovering from errors. Old pods are terminated and new pods are launched.

Although Kubernetes handles the process smoothly, there is a possibility of interruptions or downtime. To avoid this, restarting workloads of an application is blocked when deployment is restricted.

### Deployment

Go to the `Build & Deploy` tab. The CD pipelines with restricted deployment will carry a **`DO NOT DEPLOY`** label. 

Still, if a user selects an eligible image and proceeds to deploy, it will show `Deployment is blocked` along with a list of exempted users who are allowed to deploy.

{% hint style="warning" %}
Not just manual trigger, deployments remain blocked even if the trigger mode is automatic. In such cases, if a new container image comes into picture, the user has to manually deploy it when allowed.
{% endhint %}

The `Deployment History` tab will also log whether a given deployment was initiated inside or outside the deployment window.

### Deletion of CD Pipeline

Go to **App Configuration** → **Workflow**. 

While deleting a CD pipeline doesn't necessarily affect the current state of the deployed application, it might impact future deployments and you will also lose information about past deployments, i.e., Deployment History. 

If you attempt to delete any CD pipeline with restricted deployment, it will show `Pipeline deletion is blocked`.

{% hint style="info" %}
The behavior of deployment window is same for **Application** as well as **Application Groups**, since both have almost similar screens.
{% endhint %}




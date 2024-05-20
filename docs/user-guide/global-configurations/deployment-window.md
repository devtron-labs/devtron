# Deployment Window

## Introduction [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

Unplanned or last minute deployments of applications can affect the services of an organization. Consequently, its business impact will be severe if such disruptions occur during peak hours or critical periods (say festive season or no deployment on Fridays).

Therefore, Devtron comes with a feature called 'Deployment Window' that allows you to define specific timeframes during which application deployments are either blocked or allowed in specific environments.

![Figure 1: Deployment Window](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/overview-deployment.jpg)

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

(Currently, this is configured by super-admins using APIs)

---

## Checking Deployment Window

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have view-only permission or above (along with access to the environment and application) to check deployment windows.
{% endhint %}

### Overview Page

The **Deployment window** section shows the deployment windows configured for each [environment](../../reference/glossary.md#environment).

![Figure 2: Overview Page - Deployment Window](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/deployment-window.jpg)

However, if a deployment window doesn't exist for an environment, the message `No deployment windows are configured` would be displayed next to it.

You may click the dropdown icon to view the details which include:
* Type of deployment window (Blackout/Maintenance)
* Name and description
* Frequency of window (once, weekly, monthly, yearly)
* Start date and time 
* End date and time

### App Details Page

Unlike the **Overview** page which shows deployment windows for all environments, the **App Details** page shows the deployment windows for one environment at a time.

![Figure 3: App Details Page - Deployment Window](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/app-details-deployment-win.jpg)

If configured, the ongoing and upcoming deployment windows would be visible in the form of cards (in the same row that has `Application Status`, `Deployment Status`, etc.)

For example, if the super-admin has configured 4 deployment windows (say 2 Blackout and 2 Maintenance), you will see 4 cards stacked upon each other. However, no cards will be shown if deployment windows aren't configured.

You may click on any of them to view the details of the deployment windows.

---

## Result

The below functions are blocked during an ongoing blackout window or outside maintenance window.

* [Hibernation](#hibernation)
* [Restart Workloads](#restart-workloads)
* [Deletion of Workloads](#deletion-of-workloads)
* [Deployment](#deployment)
* [Rollback](#rollback)
* [Deletion of CD Pipeline](#deletion-of-cd-pipeline)

{% hint style="info" %}
For exceptional cases, the exempted users specified in the deployment window configuration can perform the above actions. In case you have configured [SES or SMTP on Devtron](../global-configurations/manage-notification.md#notification-configurations), an email notification will be sent as well.
{% endhint %}


### Hibernation

When you hibernate an application, it becomes non-functional. To avoid this, hibernation of application is blocked.

![Figure 4a: Hibernate App](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/hibernate-1.jpg)

![Figure 4b: Hibernation Dialog](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/hibernate-2.jpg)

### Restart Workloads

Although Kubernetes handles the restart process smoothly, there is a possibility of interruptions or downtime. To avoid this, restarting workloads (say Pod, Deployment, ReplicaSet) of an application is blocked when deployment is restricted.

![Figure 5a: Restart Workload](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/restart-workloads-1.jpg)

![Figure 5b: Selecting Workloads](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/restart-workloads-2.jpg)

![Figure 5c: Restart Workloads Dialog](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/restart-workloads-3.jpg)

### Deletion of Workloads

Similar to [restart workloads](#restart-workloads), deletion of workloads might disrupt the desired state and behavior of the application, hence it is barred during a deployment block.

![Figure 6a: Workload Deletion](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/workload-deletion-1.jpg)

![Figure 6b: Workload Deletion Dialog](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/workload-deletion-2.jpg)

### Deployment

Go to the `Build & Deploy` tab. The CD pipelines with restricted deployment will carry a **`DO NOT DEPLOY`** label. 

![Figure 7: Do Not Deploy Labels](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/deployment-restricted.jpg)

Despite that, if a user selects an eligible image and proceeds to deploy, it will show `Deployment is blocked` along with a list of exempted users who are allowed to deploy.

![Figure 8a: Selecting an Image](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/select-image.jpg)

![Figure 8b: Deployment Dialog](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/blocked-deployment-dialog.jpg)

{% hint style="warning" %}
Not just manual trigger, deployments remain blocked even if the trigger mode is automatic. In such cases, if a new container image comes into picture, the user has to manually deploy once the deployment block is lifted.
{% endhint %}

The `Deployment History` tab will also log whether a given deployment was initiated inside or outside the deployment window.

![Figure 9: Deployment Log](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/deployment-log.jpg)

### Rollback

Rolling back to an older version, by using a previously deployed image, is barred during a deployment block.

![Figure 10a: Rollback Deployment](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/rollback-1.jpg)

![Figure 10b: Selecting Previously Deployed Image](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/rollback-2.jpg)

![Figure 10c: Rollback Dialog](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/rollback-3.jpg)


### Deletion of CD Pipeline

Go to **App Configuration** → **Workflow**. 

In Devtron, deleting a CD pipeline affects the current state of the deployed application. Moreover, it might impact future deployments and you will also lose information about past deployments, i.e., Deployment History. 

If you attempt to delete any CD pipeline with restricted deployment, it will show `Pipeline deletion is blocked`.

![Figure 11a: Pipeline Deletion](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/pipeline-deletion-1.jpg)

![Figure 11b: Pipeline Deletion Dialog](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/pipeline-deletion-2.jpg)

---

## Impact on Application Groups

Just like application, [application groups](../application-groups.md) are also subjected to deployment windows.

![Figure 12: Deployment Window in Application Group](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/app-group-blackout.jpg)

Let's say you have 10 applications in your application group, and a blackout window is ongoing for 3 of them. In such a case, if you deploy your application group, those 3 applications will not get deployed. Therefore, you might experience a partial success along with an option to retry the failed deployments.

![Figure 13: Partial Deployment of Application Group](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/ag-deploy.jpg)

The same stands true for other bulk actions like hibernate, unhibernate, and restart workloads.







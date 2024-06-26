# Deployment Window

## Introduction [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

Unplanned or last minute deployments of applications can affect the services of an organization. Consequently, its business impact will be severe if such disruptions occur during peak hours or critical periods (say festive season or no deployment on Fridays).

Therefore, Devtron comes with a feature called 'Deployment Window' that allows you to define specific timeframes during which application deployments are either blocked or allowed in specific environments. Moreover, actions that can potentially impact the existing deployment are also restricted, which include:

* [Hibernation](#hibernation)
* [Restart Workloads](#restart-workloads)
* [Deletion of Workloads](#deletion-of-workloads)
* [Deployment](#deployment)
* [Rollback](#rollback)
* [Deletion of CD Pipeline](#deletion-of-cd-pipeline)

However, exempted users can still perform the above actions even during blocked periods.

![Figure 1: Deployment Window](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/overview-deployment-v2.jpg)

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

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to configure deployment window.
{% endhint %}

Go to **Global Configurations** → **Deployment Window**.

![Figure 2: Deployment Window in Global Configurations](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/gc-deployment-window.jpg)

This involves two parts:
* [Creating Deployment Window](#creating-deployment-window)
* [Applying Window to Deployment Pipelines](#applying-window-to-deployment-pipelines)

### Creating Deployment Window

This involves the process of creating a blackout window or a maintenance window.

1. In the `Windows` tab, click **+ Add Window**.

    ![Figure 3: Adding Deployment Window](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/add-window.jpg)

2. Give an appropriate name to your deployment window (e.g., *weekend restrictions*) and write a brief description that explains what the deployment window does. Refer [this section](#checking-deployment-window) to view the pages where the window name and description will appear.

    ![Figure 4: Name and Description](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/name-description.jpg)

3. Choose a deployment window type, i.e., maintenance window or blackout window. 

    ![Figure 5: Selecting Deployment Window Type](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/window-type-v2.jpg)

4. In your deployment window, make sure to choose the correct time zone (by default it is determined from the browser you use).

    ![Figure 6: Selecting Timezone](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/time-zone.jpg)

{% hint style="info" %}
### Why Time Zone?
Let's say you are a super-admin located in New Delhi (GMT +5:30) and you wish to restrict midnight deployments according to the Californian timings (GMT -07:00) for your team in the US. Therefore, it's crucial to choose the correct time zone (i.e., GMT -07:00) and then add the duration (see next steps). 

This ensures that deployments occur at the intended local time, helping to avoid disruptions, and facilitating co-ordinated operations across different regions.
{% endhint %}

4. Click **+ Add duration**. 

    ![Figure 7: Adding Duration](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/add-duration.jpg)

5. The following options are available for you to enforce the deployment window:
    * **Once**: Use this to make your deployment window active between two specific date and time, e.g., 20 Jun 2024, 08:00 PM ➝ 26 Jun 2024, 05:00 PM
    * **Daily**: Use this to make your deployment window active everyday between specific timings, e.g., daily between 12:00 AM ➝ 06:00 AM
    * **Weekly**: On selected days at specific timings, e.g., Wed and Sun • 02:00 AM ➝ 05:30 AM 
    * **Weekly Range**: Between days of the week, e.g., Mon (02:00 AM) to Fri (05:30 AM) 
    * **Monthly**: On or between days of the month, e.g., Day 1 (10:30 PM) to Day 2 (06:30 AM)

    ![Figure 8: Setting Window Duration](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/set-duration.jpg)

You can also add **Start Date** and **End Date** to your recurring deployment window. 

| Option                    | When To Use                                                                      |
|---------------------------|----------------------------------------------------------------------------------|
| Start Date                | Use this to enforce restrictions of a deployment window only after a specific date|
| End Date                  | Use this to stop restrictions of a deployment window beyond a specific date       |
| Both Start Date and End Date | Use these to confine your deployment window restrictions between two dates       |

Let's say you wish to enforce a blackout window every weekend to prevent unsolicited deployments by your team. If you select a weekly range (e.g., Saturday 12:00 AM to Monday 12:00 AM) and apply the deployment window without specifying dates, the weekend restrictions will persist indefinitely. 

However, by specifying a start date and an end date (as shown below), your deployment window will have a defined validity period. This ensures that the deployment window restrictions are temporary and do not extend beyond the intended timeframe.

![Figure 9: Setting Start and End Date](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/start-end-date-v2.jpg)

{% hint style="info" %}
After clicking **Done**, you can use the **+ Add duration** button to add more than one duration (for e.g., one monthly and one weekly) in a given deployment window.
{% endhint %}

6. You can also determine the users who can take actions (say deployment) even when restrictions are in place. These can be super-admins, specific users, both, or none.

    ![Figure 10: Selecting Unrestricted Users](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/user-selection.jpg)

7. Enter a display message to show the user whose deployment gets blocked, e.g., *Try deploying on Monday - Weekend deployment is not a best practice - Contact your Admin*. This will help the user understand the restriction better.

    ![Figure 11: Writing Display Message](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/display-message.jpg)

8. Click **Save Changes**.

If required, you can edit a deployment window to modify it as shown below.

![Figure 12: Editing Window](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/edit-window.gif)

You may delete a deployment window if it's not needed anymore. If the deployment window was applied to any deployment pipeline (application + environment), the restrictions would no longer exist.

![Figure 13: Deleting Window](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/delete-window.gif)


### Applying Window to Deployment Pipelines

This involves the process of applying the deployment window you created above to your deployment pipeline(s).

1. Go to the **Apply To** tab and click the **No windows** dropdown next to the [application + environment] you wish to apply deployment window(s).

    ![Figure 14: Applying Windows](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/apply-window-single.gif)

2. Select the deployment windows from the dropdown and click **Save Changes**.

#### Bulk Apply

1. If you wish to apply deployment windows to multiple applications and environments at once, use the checkbox.

    ![Figure 15: Selecting Deployment Pipelines](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/select-apps.jpg)

    We recommend you to use the available filters (Application, Environment, Deployment Window) to simplify the process of application + environment selection.

    ![Figure 16a: Using Filters](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/filter-apply.jpg)

    ![Figure 16b: Filtered Results](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/filter-apply-2.jpg)

2. On the floating widget, click **Manage Windows**.

    ![Figure 17: Manage Windows](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/manage-window.jpg)

3. Click the **Add Deployment Windows** dropdown to choose the deployment window(s).

    ![Figure 18: Attaching Windows to Applications](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/attach-windows.gif)

4. Use the **Review Changes** option to confirm the impacted environment(s) and click **Save**.

    ![Figure 19: Checking Impacted Apps + Environments](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/review-environment.gif)

You can remove deployment window(s) applied to one or more deployment pipelines as shown below.

![Figure 20a: Removing Windows from Single Deployment Pipeline](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/remove-window-single.gif)

![Figure 20b: Removing Windows from Multiple Deployment Pipeline](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/remove-window.gif)

---

## Checking Deployment Window

{% hint style="warning" %}
### Who Can Perform This Action?
Users with view only permission or above for an application can view all deployment windows configured for its deployment pipelines.
{% endhint %}

### Overview Page

The **Deployment window** section shows the blackout and maintenance windows configured for each [environment](../../reference/glossary.md#environment) of the application.

![Figure 21: Overview Page - Deployment Window](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/deployment-window.jpg)

However, if a deployment window doesn't exist for an environment, the message `No deployment windows are configured` would be displayed next to it.

You may click the dropdown icon to view the details which include:
* Type of deployment window (Blackout/Maintenance)
* Name and description
* Frequency of window (once, weekly, monthly, yearly)
* Duration

### App Details Page

Unlike the **Overview** page which shows deployment windows for all environments, the **App Details** page does not show all deployment windows configured for the environment. It shows:
* Active deployment windows
* Upcoming deployment windows

![Figure 22: App Details Page - Deployment Window](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/app-details-deployment-win-v2.jpg)

For example, if the super-admin has configured 4 deployment windows (say 2 Blackout and 2 Maintenance), you will see 4 cards stacked upon each other. However, no cards will be shown if deployment windows aren't configured. You may click on the windows card stack to view the details of active and upcoming deployment windows.

The default time period for showing upcoming deployment windows is 90 days. You can configure this individually for blackout and maintenance windows, via ConfigMap, in the `Orchestrator` microservice as shown below:

```json
DEPLOYMENT_WINDOW_FETCH_DAYS_BLACKOUT: "90"
DEPLOYMENT_WINDOW_FETCH_DAYS_MAINTENANCE: "90"
```

![Figure 23: Additional Configuration - ConfigMap](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/deployment-window-cm.jpg)


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
The exempted users specified in the deployment window configuration can perform the above actions.
{% endhint %}


### Hibernation

When you hibernate an application, it becomes non-functional. To avoid this, hibernation of application is blocked.

![Figure 24a: Hibernate App](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/hibernate-1-v2.jpg)

![Figure 24b: Hibernation Dialog](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/hibernate-2.jpg)

### Restart Workloads

Although Kubernetes handles the restart process smoothly, there is a possibility of interruptions or downtime. To avoid this, restarting workloads (say Pod, Deployment, ReplicaSet) of an application is blocked when deployment is restricted.

![Figure 25a: Restart Workload](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/restart-workloads-1-v2.jpg)

![Figure 25b: Selecting Workloads](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/restart-workloads-2.jpg)

![Figure 25c: Restart Workloads Dialog](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/restart-workloads-3.jpg)

### Deletion of Workloads

Similar to [restart workloads](#restart-workloads), deletion of workloads might disrupt the desired state and behavior of the application, hence it is barred during a deployment block.

![Figure 26a: Workload Deletion](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/workload-deletion-1-v2.jpg)

![Figure 26b: Workload Deletion Dialog](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/workload-deletion-2.jpg)

### Deployment

Go to the `Build & Deploy` tab. The CD pipelines with restricted deployment will carry a **`DO NOT DEPLOY`** label. 

![Figure 27: Do Not Deploy Labels](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/deployment-restricted.jpg)

Despite that, if a user selects an eligible image and proceeds to deploy, it will show `Deployment is blocked` along with a list of exempted users who are allowed to deploy.

![Figure 28a: Selecting an Image](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/select-image.jpg)

![Figure 28b: Deployment Dialog](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/blocked-deployment-dialog.jpg)

{% hint style="warning" %}
Not just manual trigger, deployments remain blocked even if the trigger mode is automatic. In such cases, if a new CI image is built, the user has to manually deploy once the deployment block is lifted.
{% endhint %}

The `Deployment History` tab will also log whether a given deployment was initiated during a blackout window or outside a maintenance window.

![Figure 29: Deployment Log](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/deployment-log-v2.jpg)

### Rollback

Rolling back to an older version, by using a previously deployed image, is barred during a deployment block.

![Figure 30a: Rollback Deployment](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/rollback-1.jpg)

![Figure 30b: Selecting Previously Deployed Image](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/rollback-2.jpg)

![Figure 30c: Rollback Dialog](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/rollback-3.jpg)


### Deletion of CD Pipeline

Go to **App Configuration** → **Workflow**. 

In Devtron, deleting a CD pipeline affects the current state of the deployed application. Moreover, it might impact future deployments and you will also lose information about past deployments, i.e., Deployment History. 

If you attempt to delete any CD pipeline with restricted deployment, it will show `Pipeline deletion is blocked`.

![Figure 31a: Pipeline Deletion](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/pipeline-deletion-1.jpg)

![Figure 31b: Pipeline Deletion Dialog](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/pipeline-deletion-2.jpg)

---

## Impact on Application Groups

Just like application, [application groups](../application-groups.md) are also subjected to deployment windows.

![Figure 32: Deployment Window in Application Group](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/app-group-blackout.jpg)

Let's say you have 10 applications in your application group, and a blackout window is ongoing for 3 of them. In such a case, if you deploy your application group, those 3 applications will not get deployed. Therefore, you might experience a partial success along with an option to retry the failed deployments.

![Figure 33: Partial Deployment of Application Group](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/deployment-window/ag-deploy.gif)

The same stands true for other bulk actions like hibernate, unhibernate, and restart workloads.







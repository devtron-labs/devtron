# Resource Watcher

## Introduction [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

An incident response if delayed can impact businesses, revenue, and waste valuable engineering time. Devtron's Resource Watcher enables you to perform automated actions upon the occurrence of events:

* **Create Event** - Occurs when a new Kubernetes resource is created, for e.g., a new pod spun up to handle increased traffic.
* **Update Event** - Occurs when an existing Kubernetes resource is modified, for e.g., deployment configuration tweaked to increase the replica count.
* **Delete Event** - Occurs when an existing Kubernetes resource is deleted, for e.g., deletion of an orphaned pod. 

You can make the Resource Watcher listen to the above events and accordingly run a job you wish to get done, for e.g., increasing memory, executing a script, raising Jira ticket, emailing your stakeholders, sending Slack notifications, and many more. Since manual intervention is absent, the timely response of this auto-remediation system improves your operational efficiency.

---

## Creating a Watcher

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to create a watcher.
{% endhint %}

This page allows you to create a watcher to track events and run a job. It also shows the existing list of watchers (if any).

1. Click **+ Create Watcher**. 

    ![Figure 1: Watchers - Page](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resource-watcher/watchers-page.jpg)

2. Creating a watcher consists of 4 parts, fill all the sections one by one:
    * [Basic Details](#basic-details)
    * [Namespaces to Watch](#namespaces-to-watch)
    * [Intercept Change in Resources](#intercept-change-in-resources)
    * [Execute Runbook](#execute-runbook)

    ![Figure 2: Create Watcher - Window](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resource-watcher/create-watcher-window.jpg)

### Basic Details

Here, you can give a name and description to your watcher.

![Figure 3: Adding Name and Description of Watcher](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resource-watcher/basic-details.gif)

### Namespaces to Watch

Here, you can select the [namespaces](../reference/glossary.md#namespace) whose [Kubernetes resource](../reference/glossary.md#objects) you wish to monitor for changes. 

* You can watch the namespace(s) across **All Clusters** (existing and future). 

    ![Figure 4: Choosing Namespaces of all Clusters](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resource-watcher/all-cluster.gif)

* Or you can watch namespace(s) of **Specific Clusters**.

    ![Figure 5: Choosing Namespaces of Specific Clusters](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resource-watcher/specific-cluster.gif)

{% hint style="info" %}
In both the above options, if you choose 'Specific Namespaces', you can further decide whether to track the namespaces you enter (by clicking 'Include selections') or to track the namespaces except the ones you enter (by clicking 'Exclude selections').
{% endhint %}


### Intercept Change in Resources

Here, you can select the exact Kubernetes resource(s) you wish to track for changes (in the namespace(s) you selected in the previous step).

![Figure 6: Picking Resources to Track](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resource-watcher/intercept-changes.gif)

* You can choose the resource from the **Resource kind(s) to watch** dropdown. Enter the Group Version Kind (GVK) if it's a custom resource definition (CRD), for e.g., `install.istio.io/v1apha1/IstioOperator`

* Choose the event type your watcher should listen to: `Created`, `Updated`, `Deleted`.

* Enter a [CEL expression](https://github.com/google/cel-spec/blob/master/doc/langdef.md) to catch a specific change in the resource's manifest.

**Example**: `final.status.currentReplicas == final.spec.MaxReplicas`

{% hint style="info" %}
* **If Resource Is Created** - Use 'final'
* **If Resource Is Updated** - Both 'initial' and 'final' manifest exist
* **If Resource Is Deleted** - Use 'initial'
{% endhint %}

### Execute Runbook

Here, you can choose a job that should trigger if your watcher intercepts any changes.

![Figure 7: Choosing a Job to Trigger](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resource-watcher/execute-runbook.gif)

* Choose a job pipeline from the **Run Devtron Job pipeline** dropdown.

* Select the environment in which the job should run. It can either be `devtron-ci` or the source environment (the intercepted namespace where the event has occurred).

* If the job expects input parameters, you may add its key and value under **Runtime input parameters**. 

    During a job's execution, its container can access the initial and final resource manifest through special environment variables. These variables are:
    * `DEVTRON_INITIAL_MANIFEST`
    * `DEVTRON_FINAL_MANIFEST`

* Click **Create Watcher**. 

Your watcher is now ready to intercept the changes to the selected resources. 

---

## Viewing Intercepted Changes

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to view intercepted changes.
{% endhint %}

### Details

This page allows you to view the changes to Kubernetes resources that you have selected for tracking changes. 

![Figure 8: Intercepted Changes - Page](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resource-watcher/intercepted-changes-page.jpg)

It comes with the following items to help you locate the resource, where the event has been intercepted:

* Searchbox 
* Cluster filter 
* Namespace filter
* Action filter (event type, i.e., `Created`, `Updated`, `Deleted`)
* Watcher filter (to check the intercepted changes of a specific watcher)

You get the following details in the results shown on the page.

|Field  | Description |
|-------|-------------|
|[Change In Resource](#change-in-resource)|Describes the type of change to the Kubernetes resource along with a link to its manifest|
|[Cluster/Namespace](#namespaces-to-watch) |Shows the cluster and namespace where the tracked Kubernetes resource belongs to|
|Intercepted By    |Shows the name of the watcher that intercepted the change|
|Intercepted At    |Shows the date and time when the event occurred |
|[Job Execution](#execute-runbook)     |Shows the status of the execution of job, e.g., `In Progress`, `Succeeded`, `Failed`|
|[Logs](#job-execution-log) |Links to the job log, i.e, the `Run history` page of the job|

### Change in Resource

You can check the changes in manifest by clicking **View Manifest** in `Change In Resource` column.

![Figure 9a: Created Resource Manifest - Final Manifest](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resource-watcher/view-manifest-v1.gif)


![Figure 9b: Updated Resource - Initial and Final Manifest](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resource-watcher/view-manifest-v2.gif)


![Figure 9c: Deleted Resource - Initial Manifest](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resource-watcher/view-manifest.gif)

### Job Execution Log

You can check the logs of the job executed when the Resource Watcher intercepts any change by clicking **logs**.

![Figure 10: Job Progress](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/resource-watcher/job-exec-log.gif)

---

## Use Cases

### Live Stream Traffic Surge

A live streaming sports application experiences a surge in viewers during a major game. The Horizontal Pod Autoscaler (HPA) might not be able to handle the unexpected traffic if it's capped at a low max replica count.

1. Create a watcher named 'Live Stream Scaling Alert'.
2. Monitor updates to HPA resource in the application's namespace.
3. When `currentReplicas` count reaches `MaxReplicas`, trigger a job that contains the script to increase the replica count.

### Pod Health Monitoring

A stock trading application constantly updates stock prices for its traders. If the pods become unhealthy, traders might see incorrect stock prices leading to bad investments.

1. Create a watcher named 'Pod Health Monitor'.
2. Track the pod workload of your application, if `final.status.phase != 'Running'`, trigger a job that sends an Email/Slack alert with pod details.


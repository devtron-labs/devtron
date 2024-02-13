# CD Pipeline

After your CI pipeline is ready, you can start building your CD pipeline. Devtron enables you to design your CD pipeline in a way that fully automates your deployments. Images from CI stage can be deployed to one or more environments through dedicated CD pipelines.

## Creating CD Pipeline

Click the '**+**' sign on CI Pipeline to attach a CD Pipeline to it.

![Figure 1a: Adding CD Pipeline](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/workflow-cd-v2.jpg)

A basic `Create deployment pipeline` window will pop up.

![Figure 1b: Creating CD Pipeline](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/ca-workflow-basic-v2.jpg)

Here, you get three sections:

* [Deploy to Environment](#1-deploy-to-environment)
* [Deployment Strategy](#2-deployment-strategy)
* [Advanced Options](#3-advanced-options)

### 1. Deploy to Environment

This section expects three inputs from you:

| Setting     | Description                                                | Options                   |
| ----------- | ---------------------------------------------------------- | ------------------------- |
| Environment | Select the environment where you want to deploy your application | (List of available environments)  |
| Namespace   | Automatically populated based on the selected environment | Not Applicable                           |
| Trigger     | When to execute the deployment pipeline                   | **Automatic**: Deployment triggers automatically when a new image is available at the previous stage (build pipeline or another deployment pipeline) <br /> **Manual**: Deployment is not initiated automatically. You will have to trigger deployment with a desired image. |

### 2. Deployment Strategy

Devtron supports multiple deployment strategies depending on the [deployment chart type](../../creating-application/deployment-template.md#select-chart-from-default-charts). 

![Figure 2: Strategies Supported by Chart Type](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/chart-and-strategy.jpg)

Refer [Deployment Strategies](#deployment-strategies) to know more about each strategy in depth.

{% hint style="info" %}
The next section is [Advanced Options](#advanced-options) and it comes with additional capabilities. However, if you don't need them, you may proceed with a basic CD pipeline and click **Create Pipeline**. 
{% endhint %}

### 3. Advanced Options

This option is available at the bottom of the `Create deployment pipeline` window.

![Figure 3: Advanced Options](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/advanced-option.jpg)

Now, the window will have 3 distinct tabs, and you will see the following additions:
* [Pre-Deployment stage (tab)](#pre-deployment-stage)
* [Deployment stage (tab)](#deployment-stage)
  * [Pipeline Name (input field)](#pipeline-name)
  * [Manual approval for deployment (toggle button)](#manual-approval-for-deployment)
  * [Custom Image tag pattern (toggle button)](#custom-image-tag-pattern)
  * [Pull container image with image digest](#pull-container-image-with-image-digest)
* [Post-Deployment stage (tab)](#post-deployment-stage)

![Figure 4: Advanced Options (Expanded View)](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/cd-advanced.jpg)

{% hint style="info" %}
You can create or edit a deployment strategy in Advanced Options. Remember, only the default strategy will be used for deployment, so use the **SET DEFAULT** button to mark your preferred strategy as default after creating it.
{% endhint %}

#### Pre-Deployment Stage

If your deployment requires prior actions like DB migration, code quality check (QC), etc., you can use the `Pre-deployment stage` to configure such tasks.

![Figure 5: Pre-deployment Stage](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/cd-prebuild-v2.jpg)

1. **Tasks**

Here you can add one or more tasks. The tasks can be re-arranged using drag-and-drop and they will be executed sequentially. 

2. **Trigger Pre-Deployment Stage**

Pre-deployment stages can be configured to be executed automatically or manually.

* **Automatic** - Deployment triggers automatically when a new image is available at the previous stage (build pipeline or another deployment pipeline)
* **Manual** - Deployment is not initiated automatically. You will have to trigger deployment with a desired image.

3. **ConfigMaps & Secrets**

{% hint style="info" %}
### Prerequisites
Make sure you have added [ConfigMaps](../config-maps.md) and [Secrets](../secrets.md) in App Configuration.
{% endhint %}

If you want to use some configuration files and secrets in pre-deployment stages or post-deployment stages, then you can use the `ConfigMaps` & `Secrets` options. You will get them as a drop-down in the pre-deployment stage.

* **ConfigMaps** - Used to define configuration files.
* **Secrets** - Used to store the private data of your application.

4. **Execute tasks in application environment**

These `Pre-deployment CD / Post-deployment CD` pods can be created in your deployment cluster or the devtron build cluster. It is recommended that you run these pods in the Deployment cluster so that your scripts \(if any\) can interact with the cluster services that may not be publicly exposed.

If you want to run it within your application, tick the `Execute tasks in application environment` checkbox. Otherwise, leave it unchecked to run it within the Devtron build cluster. By default, this option is disabled, so refer [Executing in Application Environment](#enabling-execution-in-application-environment) to know the process of enabling it.

#### Deployment Stage

##### Pipeline Name

Pipeline name will be auto-generated; however, you are free to modify the name as per your requirement.

##### Manual Approval for Deployment [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

If you want only approved images to be eligible for deployment, enable the `Manual approval for deployment` option in the respective deployment pipeline. By doing so, unapproved images would be prevented from being deployed for that deployment pipeline.

{% hint style="info" %}
Currently, only super-admins can enable or disable this option.
{% endhint %}

Users can also specify the number of approvals required for each deployment, where the permissible limit ranges from one approval (minimum) to six approvals (maximum). In other words, if the image doesn't get the specified number of approvals, it will not be eligible for deployment

![Figure 6: Configuring Manual Approval of Images](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/deployment-approval-new.jpg)

To enable manual approval for deployment, follow these steps:

1. Click the deployment pipeline for which you want to enable manual approval.
2. Turn on the ‘Manual approval for deployment’ toggle button.
3. Select the number of approvals required for each deployment.

To know more about the approval process, refer [Triggering CD](../../deploying-application/triggering-cd.md#manual-approval-for-deployment). 

##### Custom Image Tag Pattern

This feature helps you append custom tags (e.g., `v1.0.0`) to readily distinguish container images within your repository. 

{% hint style="warning" %}
This will be utilized only when an existing container image is copied to another repository using the [Copy Container Image Plugin](../workflow/plugins/copy-container-image.md). The image will be copied with the tag generated by the Image Tag Pattern you defined.
{% endhint %}

1. Enable the toggle button as shown below.

    ![Figure 7: Enabling Custom Image Tag Pattern](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/cd-image-pattern.jpg)

2. Click the edit icon.

    ![Figure 8: Edit Icon](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/edit-cd-image-tag.jpg)

3. You can write an alphanumeric pattern for your image tag, e.g., **prod-v1.0.{x}**. Here, 'x' is a mandatory variable whose value will incrementally increase with every pre or post deployment trigger (that option is also available to you). You can also define the value of 'x' for the next trigger in case you want to change it. 

    ![Figure 9: Defining Tag Pattern](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/cd-image-tag.jpg)

    {% hint style="warning" %}
    Ensure your custom tag do not start or end with a period (.) or comma (,)
    {% endhint %}

4. Click **Update Pipeline**. 

To know how and where this image tag would appear, refer [Copy Container Image Plugin](../workflow/plugins/copy-container-image.md)

##### Pull Container Image with Image Digest

Although Devtron ensures that image tags remain unique, the same cannot be said if images are pushed with the same tag to the same container registry from outside Devtron. 

Therefore, to eliminate the possibility of pulling an unintended image, Devtron offers the option to pull container images using digest. Here, image digest is a unique and immutable SHA-256 string returned by the container registry when you push an image. So the image referenced by the digest will never change.

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have Admin permission or above (along with access to the environment and application) to enable this option. However, this option will be non-editable in case the super-admin has enabled pull image digest globally or for the given environment.
{% endhint %}

#### Post-Deployment Stage

If you need to run any actions for e.g., run actions like closure of Jira ticket or provide secrets after the deployment, you can configure such actions in the post-deployment stages.

Post-deployment stages are similar to pre-deployment stages. The difference is, pre-deployment executes before the CD pipeline execution and post-deployment executes after the CD pipeline execution. The configuration of post-deployment stages is similar to the pre-deployment stages.

Similar to Pre-Deployment stage, you can use [ConfigMap and Secrets](#configmaps--secrets) in post deployments as well. [Execute tasks in application environment](#execute-in-application-environment) option is available too.

![Figure 10: Post-deployment Stage](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/cd_post_build.jpg)

---

## Updating CD Pipeline

You can update the deployment stages and the deployment strategy of the CD Pipeline whenever you require it. However, you cannot change the name of a CD Pipeline or its Deployment Environment. If you want a new CD pipeline for the same environment, first delete the previous CD pipeline.

To update a CD Pipeline, go to the `App Configurations` section, Click on `Workflow editor` and then click on the CD Pipeline you want to Update.

![Figure 11: Updating CD Pipeline](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/ca-workflow-update.gif)


Make changes as needed and click on `Update Pipeline` to update this CD Pipeline.

---

## Deleting CD Pipeline

If you no longer require the CD Pipeline, you can also delete the Pipeline.

To delete a CD Pipeline, go to the App Configurations and then click on the Workflow editor. Now click on the pipeline you wish to delete. A pop-up having the CD details will appear. Verify the name and the details to ensure that you are not accidentally deleting the wrong CD pipeline and then click **Delete Pipeline** to delete it.

{% hint style="warning" %}
Deleting a CD pipeline also deletes all the K8s resources associated with it and will bring a disruption in the deployed micro-service. Before deleting a CD pipeline, please ensure that the associated resources are not being used in any production workload.
{% endhint %}

---

## Extras

### Enabling Execution in Application Environment

{% hint style="info" %}
Make sure your cluster has [devtron-agent](../../global-configurations/cluster-and-environments.md#installing-devtron-agent) installed.
{% endhint %}

Some tasks require extra permissions for the node where Devtron is installed. However, if the node already has the necessary permissions for deploying applications, there is no need to assign them again. Instead, you can enable the **Execute tasks in application environment** option for the pre-CD or post-CD steps.

To enable the `Execute tasks in application environment` option, follow these steps:

1. Go to the chart store and search for the devtron-in-clustercd chart.

  ![Figure 12: 'devtron-in-clustercd' Chart](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/devtron-incluster-chart.jpg)

2. Configure the chart according to your requirements and deploy it in the target cluster.

3. After the deployment, edit the devtron-cm configmap and add the following key-value pair:

  ```bash
  ORCH_HOST: <host_url>/orchestrator/webhook/msg/nats

  Example:

  ORCH_HOST: http://xyz.devtron.com/orchestrator/webhook/msg/nats

  ```

  `ORCH_HOST` value should be same as of `CD_EXTERNAL_LISTENER_URL` value which is passed in values.yaml.

  ![Figure 13: Configuration](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/incluster-configuration.jpg)

4. Delete the Devtron pod using the following command:

  ```bash
  kubectl delete pod -l app=devtron -n devtroncd
  ```

5. Again navigate to the chart store and search for the "migration-incluster-cd" chart.

  ![Figure 14: 'migration-incluster-cd' chart](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/migration-incluster-chart.jpg)

6. Edit the `cluster-name` and `secret name` values within the chart. The `cluster name` refers to the name used when adding the cluster in the global configuration and for which you are going to enable `Execute tasks in application environment` option.

  ![Figure 15: Configuration](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/migration-incluster.jpg)

7. Deploy the chart in any environment within the Devtron cluster. Now you should be able to enable `Execute tasks in application environment` option for an environment of target cluster.

### Deployment Strategies

A deployment strategy is a method of updating, downgrading, or creating new versions of an application. The options you see under deployment strategy depend on the selected chart type (see fig 2). Below are some deployment configuration-based strategies.

#### 1. Blue-Green Strategy

Blue-green deployments involve running two versions of an application at the same time and moving traffic from the in-production version \(the green version\) to the newer version \(the blue version\).

```markup
blueGreen:
  autoPromotionSeconds: 30
  scaleDownDelaySeconds: 30
  previewReplicaCount: 1
  autoPromotionEnabled: false
```

| Key | Description |
| :--- | :--- |
| `autoPromotionSeconds` | It will make the rollout automatically promote the new ReplicaSet to active Service after this time has passed |
| `scaleDownDelaySeconds` | It is used to delay scaling down the old ReplicaSet after the active Service is switched to the new ReplicaSet |
| `previewReplicaCount` | It will indicate the number of replicas that the new version of an application should run |
| `autoPromotionEnabled` | It will make the rollout automatically promote the new ReplicaSet to the active service |

#### 2. Rolling Strategy

A rolling deployment slowly replaces instances of the previous version of an application with instances of the new version of the application. Rolling deployment typically waits for new pods to become ready via a readiness check before scaling down the old components. If a significant issue occurs, the rolling deployment can be aborted.

```markup
rolling:
  maxSurge: "25%"
  maxUnavailable: 1
```

| Key | Description |
| :--- | :--- |
| `maxSurge` | No. of replicas allowed above the scheduled quantity |
| `maxUnavailable` | Maximum number of pods allowed to be unavailable |

#### 3. Canary Strategy

Canary deployments are a pattern for rolling out releases to a subset of users or servers. The idea is to first deploy the change to a small subset of servers, test it, and then roll the change out to the rest of the servers. The canary deployment serves as an early warning indicator with less impact on downtime: if the canary deployment fails, the rest of the servers aren't impacted.

```markup
canary:
  maxSurge: "25%"
  maxUnavailable: 1
  steps:
    - setWeight: 25
    - pause:
        duration: 15 # 1 min
    - setWeight: 50
    - pause:
        duration: 15 # 1 min
    - setWeight: 75
    - pause:
        duration: 15 # 1 min
```

| Key | Description |
| :--- | :--- |
| `maxSurge` | It defines the maximum number of replicas the rollout can create to move to the correct ratio set by the last setWeight |
| `maxUnavailable` | The maximum number of pods that can be unavailable during the update |
| `setWeight` | It is the required percent of pods to move to the next step |
| `duration` | It is used to set the duration to wait to move to the next step |

#### 4. Recreate Strategy

The recreate strategy is a dummy deployment that consists of shutting down version 'A' and then deploying version 'B' after version 'A' is turned off. 

A recreate deployment incurs downtime because, for a brief period, no instances of your application are running. However, your old code and new code do not run at the same time. It terminates the old version and releases the new one.

```markup
recreate:
```

Unlike other strategies mentioned above, 'Recreate' strategy doesn't contain keys for you to configure.

{% hint style="info" %}
Does your app have different requirements for different environments? Read [Environment Overrides](../environment-overrides.md)
{% endhint %}

### Creating Sequential Pipelines

Devtron supports attaching multiple deployment pipelines to a single build pipeline, in its workflow editor. This feature lets you deploy an image first to stage, run tests and then deploy the same image to production.

Please follow the steps mentioned below to create sequential pipelines:

1. After creating CI/build pipeline, create a CD pipeline by clicking on the `+` sign on CI pipeline and configure the CD pipeline as per your requirements.
2. To add another CD Pipeline sequentially after previous one, again click on + sign on the last CD pipeline.
3. Similarly, you can add multiple CD pipelines by clicking + sign of the last CD pipeline, each deploying in different environments.

![Figure 16: Adding Multiple CD Pipelines](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/sequential-workflow.jpg)




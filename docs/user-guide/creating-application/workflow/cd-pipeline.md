# CD Pipeline
Once you are done creating your CI pipeline, you can start building your CD pipeline. Devtron enables you to design your CD pipeline in a way that fully automates your deployments.

## Creating CD Pipeline

Click the '**+**' sign on CI Pipeline to attach a CD Pipeline to it.

![Figure 1: Adding CD Pipeline](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/workflow-cd-v2.jpg)

A basic `Create deployment pipeline` window will pop up.

![Figure 2: Creating CD Pipeline](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/ca-workflow-basic-v2.jpg)

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
| Trigger     | When to execute the deployment pipeline                   | **Automatic**: Deployment triggers automatically once the corresponding CI pipeline has been executed successfully <br /> **Manual**: You trigger a deployment for a built image |

### 2. Deployment Strategy

Devtron supports 4 types of deployment strategies. Click on `Add Deployment strategy` and select from the available options:

* Recreate
* Canary
* Blue Green
* Rolling

Refer [Deployment Strategies](#deployment-strategies) to know more about each strategy in depth.

{% hint style="info" %}
The next section is [Advanced Options](#advanced-options) that you may depending on your requirements. In case you do not need the additional features provided by that option, you may proceed and click **Create Pipeline**. You can see your newly created CD Pipeline in the **Workflow** tab attached to the corresponding CI Pipeline.
{% endhint %}

{% hint style="info" %}
One can have a single CD pipeline or multiple CD pipelines connected to the same CI Pipeline. Each CD pipeline corresponds to only one environment. In other words, any single environment of an application can have only one CD pipeline.
So, the images created by the CI pipeline can be deployed into multiple environments through different CD pipelines originating from a single CI pipeline.

If you already have one CD pipeline and want to add more, you can add them by clicking the `+` sign and then choosing the environment in which you want to deploy your application. Once a new CD Pipeline is created for the environment of your choice, you can move ahead and configure the CD pipeline as required.
{% endhint %}

### 3. Advanced Options

This option is available at the bottom of the `Create deployment pipeline` window.

![Figure 3: Advanced Options](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/advanced-option.jpg)

Now, the window will have 3 distinct tabs, and you will see the following additions:
* [Pre-Deployment stage (tab)](#1-pre-deployment-stage)
* [Deployment stage (tab)](#2-deployment-stage)
  * [Pipeline Name (input field)](#pipeline-name)
  * [Manual approval for deployment (toggle button)](#manual-approval-for-deployment)
  * [Custom Image tag pattern (toggle button)](#custom-image-tag-pattern)
* [Post-Deployment stage (tab)](#3-post-deployment-stage)

![Figure 4: Advanced Options (Expanded View)](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/cd-advanced.jpg)

{% hint style="info" %}
Advanced options has an **+ Add Strategy** option that offers you the ability to define a new deployment strategy or edit it.
{% endhint %}

#### 1. Pre-Deployment stage

Sometimes one has a requirement where certain actions like DB migration are to be executed before deployment, the `Pre-deployment stage` should be used to configure these actions.

Pre-deployment stages can be configured to be executed automatically or manually.

If you select automatic, `Pre-deployment Stage` will be triggered automatically after the CI pipeline gets executed and before the CD pipeline starts executing itself. But, if you select a manual, then you have to trigger your stage via console.

If you want to use some configuration files and secrets in pre-deployment stages or post-deployment stages, then you can use the `ConfigMaps` & `Secrets` options.

`ConfigMaps` can be used to define configuration files. And `Secrets` can be defined to store the private data of your application.

Once you are done defining Config Maps & Secrets, you will get them as a drop-down in the pre-deployment stage and you can select them as part of your pre-deployment stage.

These `Pre-deployment CD / Post-deployment CD` pods can be created in your deployment cluster or the devtron build cluster. It is recommended that you run these pods in the Deployment cluster so that your scripts \(if there are any\) can interact with the cluster services that may not be publicly exposed.

If you want to run it within your application, check the `Execute in application Environment` option. Otherwise, leave it unchecked to run it within the Devtron build cluster."

Make sure your cluster has `devtron-agent` installed if you check the `Execute in the application Environment` option.

![Figure 4: Pre-deployment Stage](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/cd_pre_build_2.jpg)

#### 2. Deployment stage

##### Pipeline Name

Pipeline name will be auto-generated; however, you are free to modify the name as per your requirement.

##### Manual approval for deployment

If you want only approved images to be eligible for deployment, enable the `Manual approval for deployment` option in the respective deployment pipeline. By doing so, unapproved images would be prevented from being deployed for that deployment pipeline.

{% hint style="info" %}
Currently, only super-admins can enable or disable this option.
{% endhint %}

Users can also specify the number of approvals required for each deployment, where the permissible limit ranges from one approval (minimum) to six approvals (maximum). In other words, if the image doesn't get the specified number of approvals, it will not be eligible for deployment

![Figure 5: Configuring Manual Approval of Images](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/deployment-approval-new.jpg)

To enable manual approval for deployment, follow these steps:

1. Click the deployment pipeline for which you want to enable manual approval.
2. Turn on the ‘Manual approval for deployment’ toggle button.
3. Select the number of approvals required for each deployment.

To know more about the approval process, refer [Triggering CD](../../deploying-application/triggering-cd.md#manual-approval-for-deployment). 

##### Custom Image tag pattern



#### 3. Post-Deployment stage

If you need to run any actions for e.g., run actions like closure of Jira ticket or provide secrets after the deployment, you can configure such actions in the post-deployment stages.

Post-deployment stages are similar to pre-deployment stages. The difference is, pre-deployment executes before the CD pipeline execution and post-deployment executes after the CD pipeline execution. The configuration of post-deployment stages is similar to the pre-deployment stages.

You can use ConfigMap and Secrets in post deployments as well, as defined in the Pre-Deployment stages.

![Figure 6: Post-deployment Stage](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/cd_post_build.jpg)


##### Execute in Application Environment

When deploying an application, we often need to perform additional tasks before or after the deployments. These tasks require extra permissions for the node where Devtron is installed. However, if the node already has the necessary permissions for deploying applications, there is no need to assign them again. Instead, you can enable the "Execute in application environment" option for the pre-CD and post-CD steps. By default, this option is disabled, and some configurations are required to enable it.

To enable the "Execute in application environment" option, follow these steps:

1. Go to the chart store and search for the devtron-in-clustercd chart.

  ![Figure 7: 'devtron-in-clustercd' Chart](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/devtron-incluster-chart.jpg)

2. Configure the chart according to your requirements and deploy it in the target cluster.

3. After the deployment, edit the devtron-cm configmap and add the following key-value pair:

  ```bash
  ORCH_HOST: <host_url>/orchestrator/webhook/msg/nats

  Example:

  ORCH_HOST: http://xyz.devtron.com/orchestrator/webhook/msg/nats

  ```

  `ORCH_HOST` value should be same as of `CD_EXTERNAL_LISTENER_URL` value which is passed in values.yaml.

  ![Figure 8: Configuration](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/incluster-configuration.jpg)

4. Delete the Devtron pod using the following command:

  ```bash
  kubectl delete pod -l app=devtron -n devtroncd
  ```

5. Again navigate to the chart store and search for the "migration-incluster-cd" chart.

  ![Figure 9: 'migration-incluster-cd' chart](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/migration-incluster-chart.jpg)

6. Edit the `cluster-name` and `secret name` values within the chart. The `cluster name` refers to the name used when adding the cluster in the global configuration and for which you are going to enable `Execute in application environment` option.

  ![Figure 10: Configuration](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/migration-incluster.jpg)

7. Deploy the chart in any environment within the Devtron cluster. Now you should be able to enable `Execute in application environment` option for an environment of target cluster.

  ![Figure 11: Execute Option](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/enabled-incluster.jpg)

---

## Updating CD Pipeline

You can update the deployment stages and the deployment strategy of the CD Pipeline whenever you require it. But, you cannot change the name of a CD Pipeline or its Deployment Environment. If you need to change such configurations, you need to make another CD Pipeline from scratch.

To update a CD Pipeline, go to the `App Configurations` section, Click on `Workflow editor` and then click on the CD Pipeline you want to Update.

![Figure 12: Updating CD Pipeline](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/ca-workflow-update.gif)


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

### Deployment Strategies

A deployment strategy is a way to make changes to an application, without downtime, in a way that a user would barely notice the changes. Several deployment configuration-based strategies are discussed in this section.

#### 1. Blue Green Strategy

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
| `scaleDownDelaySeconds` | It is used to delay scaling down the old ReplicaSet after the active Service is switched to the new ReplicaSet. |
| `previewReplicaCount` | It will indicate the number of replicas that the new version of an application should run |
| `autoPromotionEnabled` | It will make the rollout automatically promote the new ReplicaSet to the active service. |

#### 2. Rolling Strategy

A rolling deployment slowly replaces instances of the previous version of an application with instances of the new version of the application. Rolling deployment typically waits for new pods to become ready via a readiness check before scaling down the old components. If a significant issue occurs, the rolling deployment can be aborted.

```markup
rolling:
  maxSurge: "25%"
  maxUnavailable: 1
```

| Key | Description |
| :--- | :--- |
| `maxSurge` | No. of replicas allowed above the scheduled quantity. |
| `maxUnavailable` | Maximum number of pods allowed to be unavailable. |

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
| `duration` | It is used to set the duration to wait to move to the next step. |

#### 4. Recreate

The recreate strategy is a dummy deployment that consists of shutting down version A then deploying version B after version A is turned off. A recreate deployment incurs downtime because, for a brief period, no instances of your application are running. However, your old code and new code do not run at the same time.

```markup
recreate:
```

It terminates the old version and releases the new one.

[Does your app has different requirements in different Environments? Also read Environment Overrides](../environment-overrides.md)

### Creating Sequential Pipelines

Devtron now supports attaching multiple deployment pipelines to a single build pipeline, in its workflow editor. This feature lets you deploy an image first to stage, run tests and then deploy the same image to production.

Please follow the steps mentioned below to create sequential pipelines:

1. After creating CI/build pipeline, create a CD pipeline by clicking on the `+` sign on CI pipeline and configure the CD pipeline as per your requirements.
2. To add another CD Pipeline sequentially after previous one, again click on + sign on the last CD pipeline.
3. Similarly, you can add multiple CD pipelines by clicking + sign of the last CD pipeline, each deploying in different environments.

![Figure 13: Adding Multiple CD Pipelines](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-cd-pipeline/sequential-workflow.jpg)




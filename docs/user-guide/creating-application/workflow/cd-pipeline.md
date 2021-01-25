# CD Pipeline

## Create CD Pipeline

![](../../../.gitbook/assets/cd-pipeline-console.png)

Click on **“+”** sign on CI Pipeline to attach a CD Pipeline to it.

![](../../../.gitbook/assets/create-cd-pipeline%20%283%29%20%283%29.png)

One can have a single CD pipeline or multiple CD pipelines connected to the same CI Pipeline. A CD pipeline corresponds to one environment or in other words, an environment of an application can have only one CD pipeline. So images created by the CI pipeline can be deployed into multiple environments.

If you already have one CD pipeline and want to add more, you can add them by clicking on the `+` sign and selecting the environment, where you want to deploy this application to. A new CD Pipeline will be created for the environment you selected.

CD pipeline configuration has below options to be configured:

| Key | Description |
| :--- | :--- |
| Pipeline Name | Enter the name of the pipeline to be created |
| Environment | Select the environment in which you want to deploy |
| Pre-deployment stage | Run any configuration and provide secrets before the deployment |
| Deployment stage | Select the deployment type through which the CD Pipeline will be triggered by Automatic or Manual. |
| Deployment Strategy | Select the type of deployment strategy that you want to enable by clicking `Add Deployment Strategy` |
| Post-deployment stage | Run any configuration and provide secrets after the deployment |

![](../../../.gitbook/assets/configuring-cd-pipeline.png)

### 1. Pipeline Name

Inside the Pipeline name column, give a name to your Continuous deployment as per your understanding.

### 2. Deploy to Environment

Select the environment where you want to deploy your application. Once you select the environment, it will display the `Namespace` corresponding to your selected environment automatically.

### 3. Pre-deployment Stage

Sometimes you encounter a requirement where you have to Configure actions like DB migration, which you want to run before the deployment. The `Pre-deployment Stage` comes into the picture in such scenarios.

Pre-deployment stages can be configured to be executed automatically or manually.

If you select automatic, `Pre-deployment Stage` will be triggered automatically after the CI pipeline gets executed and before the CD pipeline starts executing itself. But, if you select a manual, then you have to trigger your stage via console.

And if you want to use some configuration files and secrets in pre-deployment stages or post-deployment stages, then you can make use of `Config Maps` & `Secrets` options.

`Config Maps` can be used to define configuration files. And `Secrets` can be defined to keep the secret data of your application.

Once you are done defining Config Maps & Secrets, you will get them as a drop-down in the pre-deployment stage and you can select them as part of your pre-deployment stage.

These `Pre-deployment CD / Post-deployment CD` pods can be created in the deployment cluster or in the devtron build cluster. Running these pods in a Deployment cluster is recommended so that your scripts\(if there are any\) can interact with the cluster services which may not be publicly exposed.

If you want to run it inside your application then you have to check the `Execute in application Environment` option else leave it unchecked.

Make sure your cluster has `devtron-agent` installed if you Check the `Execute in the application Environment` option.

![](../../../.gitbook/assets/cd_pre_build%20%282%29.jpg)

### 4. Deployment Stages

We support two types of deployments- Manual and Automatic. If you select automatic, it will trigger your CD pipeline automatically once your corresponding CI pipeline has built successfully.

If you have defined pre-deployment stages, then CD Pipeline will be triggered automatically after the successful build of your CI pipeline followed by the successful build of your pre-deployment stages. But if you select manual, then you have to trigger your deployment via console.

### 5. Deployment Strategy

Devtron's tool has 4 types of deployment strategies. Click on `Add Deployment strategy` and select from the available options. Options are:

\(a\) Recreate

\(b\) Canary

\(c\) Blue Green

\(d\) Rolling

### 6. Post-deployment Stages

If you want to Configure actions like Jira ticket close, that you want to run after the deployment, you can configure such actions in post-deployment stages.

Post-deployment stages are similar to pre-deployment stages. The difference is, pre-deployment executes before the CD pipeline execution and post-deployment executes after the CD pipeline execution. The configuration of post-deployment stages is similar to the pre-deployment stages.

You can use Config Map and Secrets in post deployments as well, as defined in Pre-Deployment stages

![](../../../.gitbook/assets/cd_post_build.jpg)

You have configured the CD pipeline, now click on `Create Pipeline` to save it. You can see your newly created CD Pipeline on the Workflow tab attached to the corresponding CI Pipeline.

![](../../../.gitbook/assets/create-cd-pipeline%20%283%29.png)

The CD Pipeline is created

## Update CD Pipeline

You can update the CD Pipeline. Updates such as- adding Deployment Stage, Deployment Strategy. But you cannot update the name of a CD Pipeline or it’s Deploy Environment. If you need to change such configurations, you need to make another CD Pipeline.

To Update a CD Pipeline, go to the App Configurations and then click on Workflow editor and click on your CD Pipeline you want to Update.

![](../../../.gitbook/assets/update_pipeline_cd%20%282%29.jpg)

![](../../../.gitbook/assets/edit_cd_pipeline%20%285%29%20%282%29.jpg)

Make changes as per your requirement and click on `Update Pipeline` to update this CD Pipeline.

## Delete CD Pipeline

If you no longer require the CD Pipeline, you can also Delete the Pipeline.

To Delete a CD Pipeline, go to the App Configurations and then click on the Workflow editor. Now Click on the pipeline you want to delete. A pop will be displayed with CD details. Now click on the Delete Pipeline option to delete this CD Pipeline

![](../../../.gitbook/assets/edit_cd_pipeline%20%285%29.jpg)

## Deployment Strategies

A deployment strategy is a way to make changes to an application, without downtime in a way that the user barely notices the changes. There are different types of deployment strategies like Blue/green Strategy, Rolling Strategy, Canary Strategy, Recreate Strategy. These deployment configuration-based strategies are discussed in this section.

**Blue Green Stategy**

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

**Rolling Strategy**

A rolling deployment slowly replaces instances of the previous version of an application with instances of the new version of the application. Rolling deployment typically waits for new pods to become ready via a readiness check before scaling down the old components. If a significant issue occurs, the rolling deployment can be aborted.

```markup
rolling:
  maxSurge: "25%"
  maxUnavailable: 1
```

| Key | Description |
| :--- | :--- |
| `maxSurge` | No. of replicas allowed above the scheduled qauntity. |
| `maxUnavailable` | Maximum number of pods allowed to be unavailable. |

**Canary Strategy**

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
| `setWeight` | It is the required percent of pods to move to next step |
| `duration` | It is used to set the duration to wait to move to the next step. |

**Recreate**

The recreate strategy is a dummy deployment which consists of shutting down version A then deploying version B after version A is turned off. A recreate deployment incurs downtime because, for a brief period, no instances of your application are running. However, your old code and new code do not run at the same time.

```markup
recreate:
```

It terminate the old version and release the new one.

[Does your app has different requirements in different Environments? Also read Environment Overrides](../environment-overrides.md)


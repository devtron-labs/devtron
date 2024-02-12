# CI Pipeline

{% hint style="info" %}
For Devtron version older than v0.4.0, please refer the [CI Pipeline (legacy)](./ci-pipeline-legacy.md) page.
{% endhint %}

## Creating CI Pipeline

A CI Workflow can be created in one of the following ways:

* [Build and Deploy from Source Code](#1.-build-and-deploy-from-source-code)
* [Linked Build Pipeline](#2.-linked-build-pipeline)
* [Deploy Image from External Service](#3.-deploy-image-from-external-service)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/workflow-ci-1.jpg)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/workflow-ci-2.jpg)

Each method has different use-cases that can be tailored according the needs of the organization.

### 1. Build and Deploy from Source Code

`Build and Deploy from Source Code` workflow allows you to build the container image from a source code repository.

1. From the **Applications** menu, select your application.
2. On the **App Configuration** page, select **Workflow Editor**.
3. Select **+ New Workflow**.
4. Select **Build and Deploy from Source Code**.
5. Enter the following fields on the **Create build pipeline** window:

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/ci-pipeline-v1.jpg)

| Field Name | Required/Optional | Description |
| :--- | :--- | :--- |
| Source type | Required | Source type to trigger the CI. Available options: [Branch Fixed](#source-type-branch-fixed) \| [Branch Regex](#source-type-branch-regex) \|[Pull Request](#source-type-pull-request) \| [Tag Creation](#source-type-tag-creation) |
| Branch Name | Required | Branch that triggers the CI build |
| Advanced Options | Optional | Create Pre-Build, Build, and Post-Build tasks |

#### Advanced Options

The Advanced CI Pipeline includes the following stages:

* **Pre-build stage**: The tasks in this stage are executed before the image is built.
* **Build stage**: In this stage, the build is triggered from the source code that you provide.
* **Post-build stage**: The tasks in this stage will be triggered once the build is complete.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/advanced-options.jpg)

The Pre-build and Post-build stages allow you to create Pre/Post-Build CI tasks as explained [here](./ci-build-pre-post-plugins.md).

#### Build Stage

Go to the **Build stage** tab.

![Build stage](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/build-stage-v2.jpg)

| Field Name | Required/Optional | Description |
| :--- | :--- | :--- |
| TRIGGER BUILD PIPELINE | Required | The build execution may be set to: <ul><li>**Automatically (default)**: Build is triggered automatically as the Git source code changes.</li><li>**Manually**: Build is triggered manually.</li></ul> 
| Pipeline Name | Required | A name for the pipeline |
| Source type | Required | Select the source type to build the CI pipeline: [Branch Fixed](#source-type-branch-fixed) \| [Branch Regex](#source-type-branch-regex) \| [Pull Request](#source-type-pull-request) \| [Tag Creation](#source-type-tag-creation) |
| Branch Name | Required | Branch that triggers the CI build |
| Docker build arguments | Optional | Override docker build configurations for this pipeline. <br> <ul><li>Key: Field name</li><li>Value: Field value</li></ul>

##### Source type

###### Branch Fixed

This allows you to trigger a CI build whenever there is a code change on the specified branch.

Enter the **Branch Name** of your code repository.

###### Branch Regex

`Branch Regex` allows users to easily switch between branches matching the configured Regex before triggering the build pipeline.
In case of `Branch Fixed`, users cannot change the branch name in ci-pipeline unless they have admin access for the app. So, if users with 
`Build and Deploy` access should be allowed to switch branch name before triggering ci-pipeline, `Branch Regex` should be selected as source type by a user with Admin access.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/branch-regex.jpg)

For example if the user sets the Branch Regex as `feature-*`, then users can trigger from branches such as `feature-1450`, `feature-hot-fix` etc.

###### Pull Request

This allows you to configure the CI Pipeline using the PR raised in your repository.

{% hint style="info" %}
**Prerequisites**

[Configure the webhook](#configuring-webhook) for either GitHub or Bitbucket.
{% endhint %}

{% hint style="warning" %}
The **Pull Request** source type feature only works for the host GitHub or Bitbucket Cloud for now. To request support for a different Git host, please create a GitHub issue [here](https://github.com/devtron-labs/devtron/issues).
{% endhint %}

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/ci-pipeline-3.jpg)

To trigger the build from specific PRs, you can filter the PRs based on the following keys:

| Filter key | Description |
| :--- | :--- |
| `Author` | Author of the PR |
| `Source branch name` | Branch from which the Pull Request is generated |
| `Target branch name` | Branch to which the Pull request will be merged |
| `Title` | Title of the Pull Request |
| `State` | State of the PR. Default is "open" and cannot be changed |

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/ci-pipeline-8.jpg)

Select the appropriate filter and pass the matching condition as a regular expression (`regex`).

> Devtron uses regexp library, view [regexp cheatsheet](https://yourbasic.org/golang/regexp-cheat-sheet/). You can test your custom regex from [here](https://regex101.com/r/lHHuaE/1).

Select **Create Pipeline**.

###### Tag Creation

This allows you to build the CI pipeline from a tag.

{% hint style="info" %}
**Prerequisites**

[Configure the webhook](#configuring-webhook) for either GitHub or Bitbucket.
{% endhint %}

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/ci-pipeline-9.jpg)

To trigger the build from specific tags, you can filter the tags based on the `author` and/or the `tag name`.

| Filter key | Description |
| :--- | :--- |
| `Author` | The one who created the tag |
| `Tag name` | Name of the tag for which the webhook will be triggered |

Select the appropriate filter and pass the matching condition as a regular expression (`regex`).

Select **Create Pipeline**.

{% hint style="info" %}
**(a)** You can provide pre-build and post-build stages via the Devtron tool’s console or can also provide these details by creating a file `devtron-ci.yaml` inside your repository. There is a pre-defined format to write this file. And we will run these stages using this YAML file. You can also provide some stages on the Devtron tool’s console and some stages in the devtron-ci.yaml file. But stages defined through the `Devtron` dashboard are first executed then the stages defined in the `devtron-ci.yaml` file.

**(b)** The total timeout for the execution of the CI pipeline is by default set as 3600 seconds. This default timeout is configurable according to the use case. The timeout can be edited in the configmap of the orchestrator service in the env variable as `env:"DEFAULT_TIMEOUT" envDefault:"3600"`
{% endhint %}

##### Scan for Vulnerabilities

{% hint style="info" %}
### Prerequisite
Install any one of the following integrations from Devtron Stack Manager:
* [Clair](../../../user-guide/integrations/clair.md)
* Trivy
{% endhint %}

To perform the security scan after the container image is built, enable the **Scan for vulnerabilities** toggle in the build stage.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/scan-for-vulnerabilities-v2.jpg)

##### Custom Image Tag Pattern

This feature helps you append custom tags (e.g., `v1.0.0`) to readily distinguish container images within your repository.

1. Enable the toggle button as shown below.

    ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/ci-image-tag-pattern.jpg)

2. You can write an alphanumeric pattern for your image tag, e.g., **test-v1.0.{x}**. Here, 'x' is a mandatory variable whose value will incrementally increase with every build. You can also define the value of 'x' for the next build trigger in case you want to change it.

    ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/ci-image-tag-version.jpg)

    {% hint style="warning" %}
    Ensure your custom tag do not start or end with a period (.) or comma (,)
    {% endhint %}

3. Click **Update Pipeline**.

4. Now, go to **Build & Deploy** tab of your application, and click **Select Material** in the CI pipeline.

    ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/ci-material.jpg)

5. Choose the git commit you wish to use for building the container image. Click **Start Build**.

    ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/choose-commit.jpg)

6. The build will initiate and once it is successful the image tag would reflect at all relevant screens:

    * **Build History**

    ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/build-history.jpg)

    * **Docker Registry**

    ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/docker-image.jpg)

    * **CD Pipeline (Image Selection)**

    ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/build-history.jpg)


{% hint style="info" %}
Build will fail if the resulting image tag has already been built in the past. This means if there is an existing image with tag `test-v1.0.0`, you cannot build another image having the same tag `test-v1.0.0` in the same CI pipeline. This error might occur when you reset the value of the variable `x` or when you disable/enable the toggle button for `Custom image tag pattern`.
{% endhint %}


### 2. Linked Build Pipeline

If one code is shared across multiple applications, `Linked Build Pipeline` can be used, and only one image will be built for multiple applications because if there is only one build, it is not advisable to create multiple CI Pipelines.

1. From the **Applications** menu, select your application.
2. On the **App Configuration** page, select **Workflow Editor**.
3. Select **+ New Workflow**.
4. Select **Linked Build Pipeline**.
5. Enter the following fields on the **Create linked build pipeline** screen:

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/ca-workflow-linked.jpg)

* Select the application in which the source CI pipeline is present.
* Select the source CI pipeline from the application that you selected above.
* Enter a name for the linked CI pipeline.

Click **Create Linked CI Pipeline**.

After creating a linked CI pipeline, you can create a CD pipeline. 
Builds cannot be triggered from a linked CI pipeline; they can only be triggered from the source CI pipeline. There will be no images to deploy in the CD pipeline created from the 'linked CI pipeline' at first. To see the images in the CD pipeline of the linked CI pipeline, trigger build in the source CI pipeline. The build images will now be listed in the CD pipeline of the 'linked CI pipeline' whenever you trigger a build in the source CI pipeline.

### 3. Deploy Image from External Service

For CI pipeline, you can receive container images from an external services via webhook API.

You can use Devtron for deployments on Kubernetes while using an external CI tool such as Jenkins or CircleCI. External CI feature can be used when the CI tool is hosted outside the Devtron platform. However, by using an external CI, you will not be able to use some of the Devtron features such as Image scanning and security policies, configuring pre-post CI stages etc. 


* Create a [new](https://docs.devtron.ai/usage/applications/create-application) or [clone](https://docs.devtron.ai/usage/applications/cloning-application) an application.
* To configure `Git Repository`, you can add any Git repository account (e.g., dummy account) and click **Next**.
* To configure the `Container Registry` and `Container Repository`, you can leave the fields blank or simply add any test repository and click **Save & Next**.
* On the `Base Deployment Template` page, select the `Chart type` from the drop-down list and configure as per your [requirements](https://docs.devtron.ai/usage/applications/creating-application/deployment-template) and click **Save & Next**.
* On the **Workflow Editor** page, click **New Workflow** and select **Deploy image from external service**.

* On the **Deploy image from external source** page, provide the information in the following fields:

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/deploy-image-from-external-source.jpg)

| Fields | Description |
| --- | --- |
| **Deploy to environment** | <ul><li>`Environment`: Provide the name of the [environment](../../global-configurations/cluster-and-environments#add-environment).</ul></li><ul><li>`Namepsace`: Provide the [namespace](../../global-configurations/cluster-and-environments#add-environment).</ul></li> |
| **When do you want to deploy** | You can deploy either in one of the following ways: <ul><li>`Automatic`: If you select automatic, your application will be deployed automatically everytime a new image is received.</ul></li> <ul><li>`Manual`: In case of manual, you have to select the image and deploy manually. </ul></li>|
| **Deployment Strategy** | Configure the deployment preferences for this pipeline. |

* Click **Create Pipeline**.
A new CI pipeline will be created for the external source.
To get the webhook URL and JSON sample payload to be used in external CI pipeline, click **Show webhook details**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/webhook-ci/show-webhook-details-v2.jpg)

* On the **Webhook Details** page, you have to authenticate via `API token` to allow requests from an external service (e.g. Jenkins or CircleCI).

* For authentication, only users with `super-admin` permissions can select or generate an API token:
    * You can either use **Select API Token** if you have generated an [API Token](../../global-configurations/authorization/api-tokens) under **Global Configurations**. 

    ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/webhook-ci/select-api-token-webhook-details-v2.jpg)

    * Or use **Auto-generate token** to generate the API token with the required permissions. Make sure to enter the token name in the **Token name** field.

    ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/webhook-ci/auto-generate-token-webhook-details-v2.jpg)

* To allow requests from the external source, you can request the API by using:
     * **Webhook URL**
     * **cURL Request**

#### Webhook URL

 HTTP Method: `POST`

 API Endpoint: `https://{domain-name}/orchestrator/webhook/ext-ci/{pipeline-id}`
    
 JSON Payload:

```bash
    {
    "dockerImage": "445808685819.dkr.ecr.us-east-2.amazonaws.com/orch:23907713-2"
}
 ```  

You can also select metadata to send to Devtron. Sample JSON will be generated accordingly.
You can send the Payload script to your CI tools such as Jenkins and Devtron will receive the build image every time the CI pipeline is triggered or you can use the Webhook URL, which will build an image every time CI pipeline is triggered using Devtron Dashboard.
 

#### Sample cURL Request
   
```bash
curl --location --request POST \
'https://{domain-name}/orchestrator/webhook/ext-ci/{pipeline-id}' \
--header 'Content-Type: application/json' \
--header 'token: {token}' \
--data-raw '{
    "dockerImage": "445808685819.dkr.ecr.us-east-2.amazonaws.com/orch:23907713-2"
}'
 ```

#### Response Codes

| Code    | Description |
| --- | --- |
| `200` | `app detail page url` |
| `400` | `Bad request` |
| `401` | `Unauthorized` |


#### Integrate with External Sources - Jenkins or CircleCI

 * On the Jenkins dashboard, select the Jenkins job which you want to integrate with the Devtron dashboard.
 * Go to the **Configuration** > **Build Steps**, click **Add build step**, and then click **Execute Shell**.

 ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/webhook-ci/add-build-step-jenkins-v2.jpg)

 * Enter the cURL request command.
 * Make sure to enter the `API token` and `dockerImage` in your cURL command and click **Save**.

 ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/webhook-ci/execute-shell-jenkins-v2.jpg)

  Now, you can access the images on the Devtron dashboard and deploy manually. In case, if you select **Automatic** deployment option, then your application will be deployed automatically everytime a new image is received.

  Similarly, you can also integrate with external source such as **CircleCI** by:
  
  * Select the job on the `CircleCI` dashboard and click `Configuration File`.
  * On the respective job, enter the `cURL` command and update the `API token` and `dockerImage` in your cURL command.

---

## Updating CI Pipeline

You can update the configurations of an existing CI Pipeline except for the pipeline's name.
To update a pipeline, select your CI pipeline.
In the **Edit build pipeline** window, edit the required stages and select **Update Pipeline**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/update-pipeline.jpg)

---

## Deleting CI Pipeline

You can only delete a CI pipeline if there is no CD pipeline created in your workflow.

To delete a CI pipeline, go to **App Configurations > Workflow Editor** and select **Delete Pipeline**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/delete-pipeline.jpg)



---

## Extras

### Configuring Webhook

{% hint style="info" %}
If you choose [Pull Request](#pull-request) or [Tag Creation](#tag-creation) as the [source type](#source-type), you must first configure the Webhook for GitHub/Bitbucket as a prerequisite step.
{% endhint %}

#### For GitHub

1. Go to the **Settings** page of your repository and select **Webhooks**.
2. Select **Add webhook**.
3. In the **Payload URL** field, enter the Webhook URL that you get on selecting the source type as "Pull Request" or "Tag Creation" in Devtron the dashboard.
4. Change the Content-type to `application/json`.
5. In the **Secret** field, enter the secret from Devtron the dashboard when you select the source type as "Pull Request" or "Tag Creation".

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/ci-pipeline-4.jpg)

6. Under **Which events would you like to trigger this webhook?**, select **Let me select individual events.** to trigger the webhook to build CI Pipeline.
7. Select **Branch or tag creation** and **Pull Requests**.
8. Select **Add webhook**.

#### For Bitbucket Cloud

1. Go to the **Repository settings** page of your Bitbucket repository.
2. Select **Webhooks** and then select **Add webhook**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/ci-pipeline-6.jpg)

3. Enter a **Title** for the webhook.
4. In the **URL** field, enter the Webhook URL that you get on selecting the source type as "Pull Request" or "Tag Creation" in the Devtron dashboard.
5. Select the event triggers for which you want to trigger the webhook.
6. Select **Save** to save your configurations.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/ci-pipeline-7.jpg)
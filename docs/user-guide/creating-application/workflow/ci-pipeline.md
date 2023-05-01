# CI Pipeline

> **Note:**
> 
> For Devtron version older than v0.4.0, please refer the [CI Pipeline (legacy)](./ci-pipeline-legacy.md) page.

A CI Workflow can be created in one of the following ways:

* [Build and Deploy from Source Code](#1.-build-and-deploy-from-source-code)
* [Linked Build Pipeline](#2.-linked-build-pipeline)
* [Deploy Image from External Service](#3.-deploy-image-from-external-service)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/workflow-ci.jpg)

Each method has different use-cases that can be tailored according the needs of the organization.

## 1. Build and Deploy from Source Code

`Build and Deploy from Source Code` workflow allows you to build the container image from a source code repository.

1. From the **Applications** menu, select your application.
2. On the **App Configuration** page, select **Workflow Editor**.
3. Select **+ New Workflow**.
4. Select **Build and deploy from source code**.
5. Enter the following fields on the **Create build pipeline** screen:

![](../../../.gitbook/assets/ci-pipeline-2.png)

| Field Name | Required/Optional | Description |
| :--- | :--- | :--- |
| Source type | Required | Source type to trigger the CI. Available options: [Branch Fixed](#source-type-branch-fixed) \| [Branch Regex](#source-type-branch-regex) \|[Pull Request](#source-type-pull-request) \| [Tag Creation](#source-type-tag-creation) |
| Branch Name | Required | Branch that triggers the CI build |
| Advanced Options | Optional | Create Pre-Build, Build, and Post-Build tasks |

### Advanced Options

The advanced CI Pipeline includes the following stages:

* Pre-build stage: The tasks in this stage run before the image is built.
* Build stage: In this stage, the build is triggered from the source code that you provide.
* Post-build stage: The tasks in this stage are triggered once the build is complete.

The Pre-Build and Post-Build stages allow you to create Pre/Post-Build CI tasks as explained [here](./ci-build-pre-post-plugins.md).

### Scan for vulnerabilities

To Perform the security scan after the container image is built, enable the **Scan for vulnerabilities** toggle in the build stage.

![scan-for-vulnerabilities](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/scan-for-vulnerabilities-1.png)


### Build stage

The Build stage allows you to configure a build pipeline from the source code.

1. From the **Create build pipeline** screen, select **Advanced Options**.
2. Select **Build stage**.

![Build stage](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/scan-for-vulnerabilities-2.png)

| Field Name | Required/Optional | Description |
| :--- | :--- | :--- |
| TRIGGER BUILD PIPELINE | Required | The build execution may be set to: <ul><li>**Automatically (default)**: Build is triggered automatically as the Git source code changes.</li><li>**Manually**: Build is triggered manually.</li></ul> 
| Pipeline Name | Required | A name for the pipeline |
| Source type | Required | Select the source type to build the CI pipeline: [Branch Fixed](#source-type-branch-fixed) \| [Branch Regex](#source-type-branch-regex) \| [Pull Request](#source-type-pull-request) \| [Tag Creation](#source-type-tag-creation) |
| Branch Name | Required | Branch that triggers the CI build |
| Docker build arguments | Optional | Override docker build configurations for this pipeline. <br> <ul><li>Key: Field name</li><li>Value: Field value</li></ul>

Select **Update Pipeline**.

### Source type: Branch Fixed

The **Source type** - "Branch Fixed" allows you to trigger a CI build whenever there is a code change on the specified branch.

Select the **Source type** as "Branch Fixed" and enter the **Branch Name**.

### Source type: Branch Regex

`Branch Regex` allows users to easily switch between branches matching the configured Regex before triggering the build pipeline.
In case of `Branch Fixed`, users cannot change the branch name in ci-pipeline unless they have admin access for the app. So, if users with 
`Build and Deploy` access should be allowed to switch branch name before triggering ci-pipeline, `Branch Regex` should be selected as source type by a user with Admin access.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow/branch-regex.jpg)

For example if the user sets the Branch Regex as `feature-*`, then users can trigger from branches such as `feature-1450`, `feature-hot-fix` etc.

### Configuring Webhook

> **Info**:
> If you choose "Pull Request" or "Tag Creation" as the source type, you must first configure the Webhook for GitHub/Bitbucket as a prerequisite step.

#### 1. Configure Webhook for GitHub

1. Go to the **Settings** page of your repository and select **Webhooks**.
2. Select **Add webhook**.
3. In the **Payload URL** field, enter the Webhook URL that you get on selecting the source type as "Pull Request" or "Tag Creation" in Devtron the dashboard.
4. Change the Content-type to `application/json`.
5. In the **Secret** field, enter the secret from Devtron the dashboard when you select the source type as "Pull Request" or "Tag Creation".

![](../../../.gitbook/assets/ci-pipeline-4.png)

6. Under **Which events would you like to trigger this webhook?**, select **Let me select individual events.** to trigger the webhook to build CI Pipeline.
7. Select **Branch or tag creation** and **Pull Requests**.
8. Select **Add webhook**.

#### 2. Configure Webhook for Bitbucket cloud

1. Go to the **Repository settings** page of your Bitbucket repository.
2. Select **Webhooks** and then select **Add webhook**.

![](../../../.gitbook/assets/ci-pipeline-6.png)

3. Enter a **Title** for the webhook.
4. In the **URL** field, enter the Webhook URL that you get on selecting the source type as "Pull Request" or "Tag Creation" in the Devtron dashboard.
5. Select the event triggers for which you want to trigger the webhook.
6. Select **Save** to save your configurations.

![](../../../.gitbook/assets/ci-pipeline-7.png)

### Source type: Pull Request

The **Source type** - "Pull Request" allows you to configure the CI Pipeline using the PR raised in your repository.

> Before you begin, [configure the webhook](#configuring-webhook) for either GitHub or Bitbucket.

> The "Pull Request" source type feature only works for the host GitHub or Bitbucket cloud for now. To request support for a different Git host, please create a github issue [here](https://github.com/devtron-labs/devtron/issues).

![](../../../.gitbook/assets/ci-pipeline-3.png)

To trigger the build from specific PRs, you can filter the PRs based on the following keys:

| Filter key | Description |
| :--- | :--- |
| `Author` | Author of the PR |
| `Source branch name` | Branch from which the Pull Request is generated |
| `Target branch name` | Branch to which the Pull request will be merged |
| `Title` | Title of the Pull Request |
| `State` | State of the PR. Default is "open" and cannot be changed |

![](../../../.gitbook/assets/ci-pipeline-8.png)

Select the appropriate filter and pass the matching condition as a regular expression (`regex`).

> Devtron uses regexp library, view [regexp cheatsheet](https://yourbasic.org/golang/regexp-cheat-sheet/). You can test your custom regex from [here](https://regex101.com/r/lHHuaE/1).

Select **Create Pipeline**.

### Source type: Tag Creation

The **Source type** - "Tag Creation" allows you to build the CI pipeline from a tag.

> Before you begin, [configure the webhook](#configuring-webhook) for either GitHub or Bitbucket.

![](../../../.gitbook/assets/ci-pipeline-9.png)

To trigger the build from specific tags, you can filter the tags based on the `author` and/or the `tag name`.

| Filter key | Description |
| :--- | :--- |
| `Author` | The one who created the tag |
| `Tag name` | Name of the tag for which the webhook will be triggered |

Select the appropriate filter and pass the matching condition as a regular expression (`regex`).

Select **Create Pipeline**.

> **Note**
>
> **(a)** You can provide pre-build and post-build stages via the Devtron tool’s console or can also provide these details by creating a file `devtron-ci.yaml`
> inside your repository. There is a pre-defined format to write this file. And we will run these stages using this YAML file.
> You can also provide some stages on the Devtron tool’s console and some stages in the devtron-ci.yaml file. But stages defined through the `Devtron` dashboard are
> first executed then the stages defined in the `devtron-ci.yaml` file.
>
> **(b)** The total timeout for the execution of the CI pipeline is by default set as 3600 seconds. This default timeout is configurable according to the use case. The timeout can be edited in the configmap of the orchestrator service in the env variable as `env:"DEFAULT_TIMEOUT" envDefault:"3600"`

## 2. Linked Build Pipeline

If one code is shared across multiple applications, `Linked Build Pipeline` can be used, and only one image will be built for multiple applications because if there is only one build, it is not advisable to create multiple CI Pipelines.

1. From the **Applications** menu, select your application.
2. On the **App Configuration** page, select **Workflow Editor**.
3. Select **+ New Workflow**.
4. Select **Linked Build Pipeline**.
5. Enter the following fields on the **Create linked build pipeline** screen:

![](../../../.gitbook/assets/ca-workflow-linked.png)

* Select the application in which the source CI pipeline is present.
* Select the source CI pipeline from the application that you selected above.
* Enter a name for the linked CI pipeline.

Click **Create Linked CI Pipeline**.

After creating a linked CI pipeline, you can create a CD pipeline. 
Builds cannot be triggered from a linked CI pipeline; they can only be triggered from the source CI pipeline. There will be no images to deploy in the CD pipeline created from the 'linked CI pipeline' at first. To see the images in the CD pipeline of the linked CI pipeline, trigger build in the source CI pipeline. The build images will now be listed in the CD pipeline of the 'linked CI pipeline' whenever you trigger a build in the source CI pipeline.

## 3. Deploy Image from External Service

For CI pipeline, you can receive container images from an external services via webhook API.

You can use Devtron for deployments on Kubernetes while using an external CI tool such as Jenkins or CircleCI. External CI feature can be used when the CI tool is hosted outside the Devtron platform. However, by using an external CI, you will not be able to use some of the Devtron features such as Image scanning and security policies, configuring pre-post CI stages etc. 


* Create a [new](https://docs.devtron.ai/usage/applications/create-application) or [clone](https://docs.devtron.ai/usage/applications/cloning-application) an application.
* To configure `Git Repository`, you can add any Git repository account (e.g., dummy account) and click **Next**.
* To configure the `Container Registry` and `Container Repository`, you can leave the fields blank or simply add any test repository and click **Save & Next**.
* On the `Base Deployment Template` page, select the `Chart type` from the drop-down list and configure as per your [requirements](https://docs.devtron.ai/usage/applications/creating-application/deployment-template) and click **Save & Next**.
* On the **Workflow Editor** page, click **New Workflow** and select **Deploy image from external service**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/webhook-ci/click-new-workflow.png)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/webhook-ci/select-deploy-image-from-external-service.png)

* On the **Deploy image from external source** page, provide the information in the following fields:

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/webhook-ci/deploy-image-from-external-service.png)

| Fields | Description |
| --- | --- |
| **Deploy to environment** | <ul><li>`Environment`: Provide the name of the [environment](https://docs.devtron.ai/global-configurations/cluster-and-environments#add-environment).</ul></li><ul><li>`Namepsace`: Provide the [namespace](https://docs.devtron.ai/global-configurations/cluster-and-environments#add-environment).</ul></li> |
| **When do you want to deploy** | You can deploy either in one of the following ways: <ul><li>`Automatic`: If you select automatic, your application will be deployed automatically everytime a new image is received.</ul></li> <ul><li>`Manual`: In case of manual, you have to select the image and deploy manually. </ul></li>|
| **Deployment Strategy** | Configure the deployment preferences for this pipeline. |

* Click **Create Pipeline**.
A new CI pipeline will be created for the external source.
To get the webhook URL and JSON sample payload to be used in external CI pipeline, click **Show webhook details**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/webhook-ci/show-webhook-details.jpg)

* On the **Webhook Details** page, you have to authenticate via `API token` to allow requests from an external service (e.g. Jenkins or CircleCI).

* For authentication, only users with `super-admin` permissions can select or generate an API token:
    * You can either use **Select API Token** if you have generated an [API Token](https://docs.devtron.ai/getting-started/global-configurations/authorization/api-tokens) under **Global Configurations**. 

    ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/webhook-ci/select-api-token-webhook-details.png)


    * Or use **Auto-generate token** to generate the API token with the required permissions. Make sure to enter the token name in the **Token name** field.

    ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/webhook-ci/auto-generate-token-webhook-details.png)

* To allow requests from the external source, you can request the API by using:
     * **Webhook URL**
     * **cURL Request**

### Webhook URL

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
 

### Sample cURL Request
   
```bash
curl --location --request POST \
'https://{domain-name}/orchestrator/webhook/ext-ci/{pipeline-id}' \
--header 'Content-Type: application/json' \
--header 'token: {token}' \
--data-raw '{
    "dockerImage": "445808685819.dkr.ecr.us-east-2.amazonaws.com/orch:23907713-2"
}'
 ```

### Response Codes

| Code    | Description |
| --- | --- |
| `200` | `app detail page url` |
| `400` | `Bad request` |
| `401` | `Unauthorized` |


### Integrate with External Sources - Jenkins or CircleCI

 * On the Jenkins dashboard, select the Jenkins job which you want to integrate with the Devtron dashboard.
 * Go to the **Configuration** > **Build Steps**, click **Add build step**, and then click **Execute Shell**.

 ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/webhook-ci/add-build-step-jenkins.png)

 * Enter the cURL request command.
 * Make sure to enter the `API token` and `dockerImage` in your cURL command and click **Save**.

 ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/webhook-ci/execute-shell-jenkins.jpg)

  Now, you can access the images on the Devtron dashboard and deploy manually. In case, if you select **Automatic** deployment option, then your application will be deployed automatically everytime a new image is received.

  Similarly, you can also integrate with external source such as **CircleCI** by:
  
  * Select the job on the `CircleCI` dashboard and click `Configuration File`.
  * On the respective job, enter the `cURL` command and update the `API token` and `dockerImage` in your cURL command.





## Update CI Pipeline

You can update the configurations of an existing CI Pipeline except for the pipeline's name.
To update a pipeline, select your CI pipeline.
In the **Edit build pipeline** window, edit the required stages and select **Update Pipeline**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/ci-pipeline-update.jpg)

## Delete CI Pipeline

You can only delete a CI pipeline if there is no CD pipeline created in your workflow.

To delete a CI pipeline, go to **App Configurations > Workflow Editor** and select **Delete Pipeline**.

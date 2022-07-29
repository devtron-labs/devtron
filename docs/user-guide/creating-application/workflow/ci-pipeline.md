# CI Pipeline

> **Info:**
> 
> For Devtron version older than v0.4.0, please refer the [CI Pipeline (legacy)](./ci-pipeline-legacy.md) page.

A CI Pipeline can be created in one of the three ways:

* [Continuous Integration](#1-continuous-integration)
* [Linked CI Pipeline](#2.-linked-ci-pipeline)
* [Incoming Webhook](#3.-incoming-webhook)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/workflow-ci.jpg)

Each of these methods has different use-cases that can be tailored to the needs of the organization.

## 1. Continuous Integration

Continuous Integration Pipeline allows you to build the container image from a source code repository.

1. From the **Applications** menu, select your application.
2. On the **App Configuration** page, select **Workflow Editor**.
3. Select **+ New Build Pipeline**.
4. Select **Continuous Integration**.
5. Enter the following fields on the **Create build pipeline** screen:

![](../../../.gitbook/assets/ci-pipeline-2.png)

| Field Name | Required/Optional | Description |
| :--- | :--- | :--- |
| Source type | Required | Source type to trigger the CI. Available options: [Branch Fixed](#source-type-branch-fixed) \| [Pull Request](#source-type-pull-request) \| [Tag Creation](#source-type-tag-creation) |
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
| Source type | Required | Select the source type to build the CI pipeline: [Branch Fixed](#source-type-branch-fixed) \| [Pull Request](#source-type-pull-request) \| [Tag Creation](#source-type-tag-creation) |
| Branch Name | Required | Branch that triggers the CI build |
| Docker build arguments | Optional | Override docker build configurations for this pipeline. <br> <ul><li>Key: Field name</li><li>Value: Field value</li></ul>

Select **Update Pipeline**.

### Source type: Branch Fixed

The **Source type** - "Branch Fixed" allows you to trigger a CI build whenever there is a code change on the specified branch.

Select the **Source type** as "Branch Fixed" and enter the **Branch Name**.

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

## 2. Linked CI Pipeline

If one code is shared across multiple applications, `Linked CI Pipeline` can be used, and only one image will be built for multiple applications because if there is only one build, it is not advisable to create multiple CI Pipelines.

1. From the **Applications** menu, select your application.
2. On the **App Configuration** page, select **Workflow Editor**.
3. Select **+ New Build Pipeline**.
4. Select **Linked CI Pipeline**.
5. Enter the following fields on the **Create linked build pipeline** screen:

![](../../../.gitbook/assets/ca-workflow-linked.png)

* Select the application in which the source CI pipeline is present.
* Select the source CI pipeline from the application that you selected above.
* Enter a name for the linked CI pipeline.

Select **Create Linked CI Pipeline**.

After creating a linked CI pipeline, you can create a CD pipeline. 
Builds cannot be triggered from a linked CI pipeline; they can only be triggered from the source CI pipeline. There will be no images to deploy in the CD pipeline created from the 'linked CI pipeline' at first. To see the images in the CD pipeline of the linked CI pipeline, trigger build in the source CI pipeline. The build images will now be listed in the CD pipeline of the 'linked CI pipeline' whenever you trigger a build in the source CI pipeline.

## 3. Incoming Webhook

The CI pipeline receives container images from an external source via a webhook service.

You can use Devtron for deployments on Kubernetes while using your CI tool such as Jenkins. External CI features can be used when the CI tool is hosted outside the Devtron platform.

1. From the **Applications** menu, select your application.
2. On the **App Configuration** page, select **Workflow Editor**.
3. Select **+ New Build Pipeline**.
4. Select **Incoming Webhook**.

![](../../../.gitbook/assets/ca-workflow-external.png)

| Field Name | Required/Optional | Description |
| :--- | :--- | :--- |
| **Pipeline Name** | Required | Name of the pipeline |
| **Source Type** | Required | ‘Branch Fixed’ or ‘Tag Regex’ |
| **Branch Name** | Required | Name of the branch |

* Select **Save and Generate URL**. This generates the Payload format and Webhook URL.

You can send the Payload script to your CI tools such as Jenkins and Devtron will receive the build image every time the CI Service is triggered or you can use the Webhook URL which will build an image every time CI Service is triggered using Devtron Dashboard.

## Update CI Pipeline

You can update the configurations of an existing CI Pipeline except for the pipeline's name.
To update a pipeline, select your CI pipeline.
In the **Edit build pipeline** window, edit the required stages and select **Update Pipeline**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/ci-pipeline-update.jpg)

## Delete CI Pipeline

You can only delete a CI pipeline if there is no CD pipeline created in your workflow.

To delete a CI pipeline, go to **App Configurations > Workflow Editor** and select **Delete Pipeline**.

# Workflow Editor

In the `Workflow Editor` section, you can configure a job pipeline to be executed. Pipelines can be configured to be triggered automatically or manually based on code change or time.

* After adding Git repo in the `Source Code` section, go to the `Workflow Editor`. 
* Click `Job Pipeline`.
* Provide the information in the following fields on the **Create job pipeline** page:

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/create-job/create-job-pipeline-basic.jpg)

| Field Name | Required/Optional | Description |
| :--- | :--- | :--- |
| Pipeline Name | Required | A name for the pipeline |
| Source type | Required | Source type to trigger the job pipeline. Available options: [Branch Fixed](#source-type-branch-fixed) \| [Branch Regex](#source-type-branch-regex) \|[Pull Request](#source-type-pull-request) \| [Tag Creation](#source-type-tag-creation) |
| Branch Name | Required | Branch that triggers the job pipeline. |

* Click **Create Pipeline**.

* The job pipeline is created.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/create-job/job-pipeline-created.jpg)

* To trigger job pipeline, go to the [Trigger Job](triggering-job.md) section. 

`Note`: You can create more than one job pipeline by clicking **+ Job Pipeine**.


### Source type: Branch Fixed

The **Source type** - "Branch Fixed" allows you to trigger a CI build whenever there is a code change on the specified branch.

Select the **Source type** as "Branch Fixed" and enter the **Branch Name**.

### Source type: Branch Regex

`Branch Regex` allows users to easily switch between branches matching the configured Regex before triggering the build pipeline.
In case of `Branch Fixed`, users cannot change the branch name in ci-pipeline unless they have admin access for the app. So, if users with 
`Build and Deploy` access should be allowed to switch branch name before triggering ci-pipeline, `Branch Regex` should be selected as source type by a user with Admin access.

For example if the user sets the Branch Regex as `feature-*`, then users can trigger from branches such as `feature-1450`, `feature-hot-fix` etc.

### Source type: Pull Request

The **Source type** - "Pull Request" allows you to configure the CI Pipeline using the PR raised in your repository.

> Before you begin, [configure the webhook](../creating-application/workflow/ci-pipeline.md#configuring-webhook) for either GitHub or Bitbucket.

> The "Pull Request" source type feature only works for the host GitHub or Bitbucket cloud for now. To request support for a different Git host, please create a github issue [here](https://github.com/devtron-labs/devtron/issues).


To trigger the build from specific PRs, you can filter the PRs based on the following keys:

| Filter key | Description |
| :--- | :--- |
| `Author` | Author of the PR |
| `Source branch name` | Branch from which the Pull Request is generated |
| `Target branch name` | Branch to which the Pull request will be merged |
| `Title` | Title of the Pull Request |
| `State` | State of the PR. Default is "open" and cannot be changed |

Select the appropriate filter and pass the matching condition as a regular expression (`regex`).

> Devtron uses regexp library, view [regexp cheatsheet](https://yourbasic.org/golang/regexp-cheat-sheet/). You can test your custom regex from [here](https://regex101.com/r/lHHuaE/1).

### Source type: Tag Creation

The **Source type** - "Tag Creation" allows you to build the CI pipeline from a tag.

> Before you begin, [configure the webhook](#configuring-webhook) for either GitHub or Bitbucket.

To trigger the build from specific tags, you can filter the tags based on the `author` and/or the `tag name`.

| Filter key | Description |
| :--- | :--- |
| `Author` | The one who created the tag |
| `Tag name` | Name of the tag for which the webhook will be triggered |

Select the appropriate filter and pass the matching condition as a regular expression (`regex`).


### Add Preset Plugins

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/create-job/create-job-pipeline-add-tasks.jpg)

You can also add preset plugins in your job pipeline to execute some standard tasks, such as Code analysis, Load testing, Security scanning etc. Click `Add Task` to add [preset plugins](https://docs.devtron.ai/v/v0.6/usage/applications/creating-application/ci-pipeline/ci-build-pre-post-plugins#configuring-pre-post-build-tasks).


## Update Job Pipeline

You can update the configurations of an existing Job Pipeline except for the pipeline's name.
To update a pipeline, select your job pipeline.
In the **Edit job pipeline** window, edit the required fields and select **Update Pipeline**.

## Delete Job Pipeline

You can only delete a job pipeline in your workflow.

To delete a job pipeline, go to **Configurations > Workflow Editor** and select **Delete Pipeline**.

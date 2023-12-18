# Triggering Job 

## Triggering Job Pipeline

The Job Pipeline can be triggered by selecting `Select Material`

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/create-job/trigger-job.jpg)

Job Pipelines that are set as automatic are always triggered as soon as a new commit is made to the git branch they're sensing. However, Job pipelines can always be manually triggered as and if required.

Various commits done in the repository can be seen here along with details like Author, Date etc. Select the commit that you want to trigger and then click on `Run Job` to trigger the job pipeline.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/create-job/run-job.jpg)


**Refresh** icon, refreshes Git Commits in the job Pipeline and fetches the latest commits from the `Git Repository`.

**Ignore Cache** : This option will ignore the previous build cache and create a fresh build. If selected, will take a longer build time than usual.

It can be seen that the job pipeline is triggered here and is the _Running_ state.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/create-job/click-job-details.jpg)

Click your `job pipeline` or click `Run History` to get the details about the job pipeline such as logs, reports etc.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/create-job/run-history-job.jpg)

Click `Source code` to view the details such as commit id, Author and commit message of the Git Material that you have selected for the job.

Click `Artifacts` to download the _reports_ of the job, if any.

If you have multiple job pipelines, you can select a pipeline from the drop-down list to view th details of logs, source code, or artifacts.



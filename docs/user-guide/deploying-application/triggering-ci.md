# Triggering CI Pipelines

To trigger the CI pipeline, first you need to select the Git commit for which the CI pipeline will be triggered. To select the Git commit, click on the `Select Material` button present on the CI pipeline.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/select-material-new.jpg)

Once clicked, a list will appear showing various commits made in the repository, including details such as the author name, commit date and time etc. Choose the desired commit for which you want to trigger the pipeline, and then click on "Start Build" to initiate the CI pipeline.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/trigger-build.jpg)

CI Pipelines with automatic triggers are triggereded immediately when a new commit is made to the git branch. If the trigger for a build pipeline is set to manual, it will not be automatically triggered and requires manual trigger.

The **Refresh** icon updates the Git Commits section in the CI Pipeline by fetching the latest commits from the repository. Clicking on the refresh icon ensures that you have the most recent commit available.

**Ignore Cache** : This option will ignore the previous build cache and create a fresh build. If selected, will take a longer build time than usual.

Click on the `CI Pipeline` or navigate to the `Build History` to get the CI pipeline details such as build logs, sorce code details, artifacts and vulnerability scan reports.

To access the logs of the CI Pipeline, simply click on the `Logs`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/build-logs.jpg)

To view specific details of the Git commit you've selected for the build, click on `Source`. This will provide you with information like the commit ID, author, and commit message associated with that particular commit.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/build-source.jpg)

By selecting the `Artifacts` option, you can download reports related to the tasks performed in the Pre-CI and Post-CI stages. This will allow you to access and retrieve the generated reports, if any, related to these stages. Additionally, you have the option to add tags or comments to the image directly from this section.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/tags-and-artifacts.jpg)

To check for any vulnerabilities in the build image, click on `Security`. Please note that vulnerabilities will only be visible if you have enabled the `Scan for vulnerabilities` option in the advanced options of the CI pipeline before building the image. For more information about this feature, please refer to this [documentation](../../user-guide/security-features.md).

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/security-scan-report.jpg)


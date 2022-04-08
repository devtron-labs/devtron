# Triggering CI

## Triggering CI Pipelines

The CI Pipeline can be triggered by selecting `Select Material`

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/select-material.jpg)

CI Pipelines that are set as automatic are always triggered as soon as a new commit is made to the git branch they're sensing. However, CI pipelines can always be manually triggered as and if required.

Various commits done in the repository can be seen, here along with details like Author, Date etc. Select the commit that you want to trigger and then click on `Start Build` to trigger the CI pipeline.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/start-build-1.jpg)


**Refresh** icon, refreshes Git Commits in the CI Pipeline and fetches the latest commits from the “Repository”

**Ignore Cache** : This option will ignore the previous build cache and create a fresh build. If selected, will take a longer build time than usual.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/build-history.jpg)

It can be seen that the pipeline is triggered here and is the _Running_ state.

Click on your `CI Pipeline` or click on `Build History` to get the details about the CI pipeline such as logs, reports etc.

You can read the `logs` of the CI Pipeline from here.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/build-history-logs.jpg)

Click on `Source code` to view the details such as commit id, Author and commit message of the Git Material that you have selected for the build.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/build-history-sc.jpg)

Click on `Artifacts` to download the _reports_ of the Pre-CI and Post-CI stages if any.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/build-history-artifacts.jpg)

Click on `security` to see if there is any vulnerabilitiesin the build image. You can see the vulnerabilities here only if you have enabled `Scan for vulnerabilities` before building image from advanced options of CI pipeline. To know more about this feature, follow our [documentation](https://docs.devtron.ai/devtron/user-guide/security-features).

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/build-history-security.jpg)


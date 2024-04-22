# Triggering CI Pipelines

To trigger the CI pipeline, first you need to select a Git commit. To select a Git commit, click the **Select Material** button present on the CI pipeline.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/select-material-new.jpg)

Once clicked, a list will appear showing various commits made in the repository, it includes details such as the author name, commit date, time, etc. Choose the desired commit for which you want to trigger the pipeline, and then click **Start Build** to initiate the CI pipeline.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/trigger-build.jpg)

CI Pipelines with automatic trigger enabled are triggered immediately when a new commit is made to the git branch. If the trigger for a build pipeline is set to manual, it will not be automatically triggered and requires a manual trigger.

---

## Partal Cloning Feature [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

CI builds can be time-consuming for large repositories, especially for enterprises. However, Devtron's partial cloning feature significantly increases cloning speed, reducing the time it takes to clone your source code and leading to faster build times.

**Advantages**
* Smaller image sizes
* Reduced resource usage and costs
* Faster software releases
* Improved productivity

Get in touch with us if you are looking for a way to improve the efficiency of your software development process.

The **Refresh** icon updates the Git Commits section in the CI Pipeline by fetching the latest commits from the repository. Clicking on the refresh icon ensures that you have the most recent commit available.

The **Ignore Cache** option ignores the previous build cache and creates a fresh build. If selected, will take a longer build time than usual.

---

## Passing Build Parameters [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have [Build & deploy permission](../global-configurations/authorization/user-access.md#role-based-access-levels) or above (along with access to the environment and application) to pass build parameters.
{% endhint %}

If you wish to pass runtime parameters for build job, you can provide key-value pairs before triggering the build. This will inject those key-value pairs as environment variables in CI runner pods and all its containers.

**Steps**

1. Go to the **Parameters** tab available on the screen where you select the commit.

    ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/build-parameter-tab.jpg)

2. Click **+ Add parameter**.

    ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/add-parameter.jpg)

3. Enter your key-value pair as shown below. 

    ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/key-value.jpg)

    <br /> Similarly, you may add more than one key-value pair by using the **+ Add Parameter** button.

4. Click **Start Build**.

{% hint style="info" %}
Passing build parameters is currently not supported for [Linked Build pipeline](../creating-application/workflow/ci-pipeline.md#2-linked-build-pipeline) and [External CI pipeline](../creating-application/workflow/ci-pipeline.md#3-deploy-image-from-external-service).

In case you trigger builds in bulk, you can consider passing build parameters in [Application Group](../application-groups.md).
{% endhint %}

---

## Fetching Logs and Reports

Click the `CI Pipeline` or navigate to the `Build History` to get the CI pipeline details such as build logs, source code details, artifacts, and vulnerability scan reports.

To access the logs of the CI Pipeline, simply click `Logs`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/build-logs.jpg)

To view specific details of the Git commit you've selected for the build, click on `Source`. This will provide you with information like the commit ID, author, and commit message associated with that particular commit.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/build-source.jpg)

By selecting the `Artifacts` option, you can download reports related to the tasks performed in the Pre-CI and Post-CI stages. This will allow you to access and retrieve the generated reports, if any, related to these stages. Additionally, you have the option to add tags or comments to the image directly from this section.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/tags-and-artifacts.jpg)

To check for any vulnerabilities in the build image, click on `Security`. Please note that vulnerabilities will only be visible if you have enabled the `Scan for vulnerabilities` option in the advanced options of the CI pipeline before building the image. For more information about this feature, please refer to this [documentation](../../user-guide/security-features.md).

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-ci/security-scan-report.jpg)




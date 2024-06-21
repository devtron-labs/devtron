# GitOps Configuration

{% hint style="warning" %}
The 'GitOps Configuration' page appears only if the super-admin has enabled 'Allow changing git repository for application' in [Global Configurations → GitOps](../global-configurations/gitops.md).
{% endhint %}

## Introduction

This configuration is an extension of the [GitOps](../global-configurations/gitops.md) settings present in [Global Configurations](../global-configurations/README.md) of Devtron. Therefore, make sure you read it before making any changes to your app configuration.

The application-level GitOps configuration offers the flexibility to add a custom Git repo (as opposed to Devtron auto-creating a repo for your application). 

---

## Adding Custom Git Repo for GitOps

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have [Admin permission](../../user-guide/global-configurations/authorization/user-access.md#role-based-access-levels) or above (along with access to the environment and application) to configure user-defined Git repo.
{% endhint %}

### For Devtron Apps

1. Go to **Applications** → **Devtron Apps** (tab) → (choose your app) → **App Configuration** (tab) → **GitOps Configuration**.

    ![Figure 1: App-level GitOps Config](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/gitops/app-config-gitops.jpg)

2. Assuming a GitOps repo was not added to your application earlier, you get 2 options:

    * **Auto-create repository** - Select this option if you wish to proceed with the default behavior. It will create a repository automatically, named after your application with a prefix. Thus saving you the trouble of creating one manually.
 
    * **Commit manifests to a desired repository** - Select this option if you wish to add a custom repo that is already created with your [Git provider](../global-configurations/gitops.md#supported-git-providers). Enter its link in the `Git Repo URL` field.

    ![Figure 2: Repo Creation](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/gitops/gitops-config.jpg)


{% hint style="warning" %}
GitOps repositories, whether auto-created by Devtron or added manually, are immutable. This means they cannot be modified after creation. The same is true if you have an existing CD pipeline that uses/used GitOps for deployment.
{% endhint %}

3. Click **Save**.

    ![Figure 3: Saved GitOps Config](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/gitops/saved-config.jpg)

**Note**: In case you skipped the GitOps configuration for your application and proceeded towards the [creation of a new CD pipeline](../creating-application/workflow/cd-pipeline.md#creating-cd-pipeline) (that uses GitOps), you will be prompted to configure GitOps as shown below:

![Figure 4: Incomplete GitOps Config](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/gitops/gitops-not-configured.jpg)


### For Helm Apps

You can [deploy a helm chart](../deploy-chart/overview-of-charts.md#deploying-chart) using either Helm or GitOps. Let's assume you wish to deploy `airflow` chart.

1. Select the helm chart from the [Chart Store](../deploy-chart/README.md).

    ![Figure 5: Choosing a Helm Chart](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/gitops/chart-selection.jpg)

2. Click **Configure & Deploy**.

    ![Figure 6: Configure & Deploy Button](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/gitops/configure-deploy.jpg)

3. After you enter the `App Name`, `Project`, and `Environment`; an option to choose the deployment approach (i.e., Helm or GitOps) would appear. Select **GitOps**.

{% hint style="info" %}
The option to choose between 'Helm' or 'GitOps' is only available in <img src="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg">
{% endhint %}

![Figure 7a: Deployment Approach](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/gitops/deployment-method.jpg)

![Figure 7b: Selecting GitOps Method](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/gitops/select-gitops.jpg)

4. A modal window will appear for you to enter a Git repository. Just like [Devtron Apps](#for-devtron-apps) (step 2), you get two options:
    * Auto-create repository
    * Commit manifests to a desired repository

    ![Figure 8: Adding a Repo](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/gitops/git-repository-helm-app.jpg)

5. Enter your custom Git Repo URL, and click **Save**.

    ![Figure 9: Saved GitOps Config for Helm App](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/gitops/custom-git-repo-helm-apps.jpg)

Next, you may proceed to deploy the chart.

{% hint style="warning" %}
Once you deploy a helm app with GitOps, you cannot change its GitOps repo.
{% endhint %}
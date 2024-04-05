# GitOps Configuration

{% hint style="warning" %}
The 'GitOps Configuration' page appears only if the super-admin has enabled 'Allow changing git repository for application' in [Global Configurations → GitOps](../global-configurations/gitops.md).
{% endhint %}

## Introduction

This configuration is an extension of the [GitOps](../global-configurations/gitops.md) settings present in [Global Configurations](../global-configurations/README.md) of Devtron. Therefore, make sure you read it before making any changes to your app configuration.

The application-level GitOps configuration offers the flexibility to add a custom Git repo (as opposed to Devtron auto-creating a repo for your application). 

## Adding Custom Git Repo for GitOps

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have [Admin permission](../../user-guide/global-configurations/authorization/user-access.md#role-based-access-levels) or above (along with access to the environment and application) to configure user-defined Git repo.
{% endhint %}

1. Go to **Applications** (choose your app) → **App Configuration** (tab) → **GitOps Configuration**.

    ![Figure 1: App-level GitOps Config](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/gitops/app-config-gitops.jpg)

2. Assuming a GitOps repo was not added to your application earlier, you get 2 options:

    * **Auto-create repository** - Select this option if you wish to proceed with the default behavior. This will auto-create a repository whose name will be the same as the application name.
 
    * **Commit manifests to a desired repository** - Select this option if you wish to add a custom repo that is already created with your [Git provider](../global-configurations/gitops.md#supported-git-providers). Enter its link in the `Git Repo URL` field.

    ![Figure 2: Repo Creation](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/gitops/gitops-config.jpg)


{% hint style="warning" %}
GitOps repositories, whether auto-created by Devtron or added manually, are immutable. This means they cannot be modified after creation. The same is true if you have an existing CD pipeline that uses/used GitOps for deployment.
{% endhint %}

3. Click **Save**.

    ![Figure 3: Saved GitOps Config](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/gitops/saved-config.jpg)

**Note**: In case you skipped the GitOps configuration for your application and proceeded towards the [creation of a new CD pipeline](../creating-application/workflow/cd-pipeline.md#creating-cd-pipeline)(that uses GitOps), you will be prompted to configure GitOps as shown below:

![Figure 4: Incomplete GitOps Config](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/gitops/gitops-not-configured.jpg)














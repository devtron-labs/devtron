# GitOps

## Introduction

In Devtron, you can either use Helm or GitOps (Argo CD) to deploy your applications and charts. GitOps is a branch of DevOps that focuses on using Git repositories to manage infrastructure and application code deployments.

If you use the GitOps approach, Devtron will store Kubernetes configuration files and the desired state of your applications in Git repositories.

---

## Steps to Configure GitOps

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to configure GitOps.
{% endhint %}

1. Go to **Global Configurations** → **GitOps**

   ![Figure 1: Global Configuration - GitOps](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/gitops.jpg)

2. Select any one of the [supported Git providers](#supported-git-providers) to configure GitOps. 

   ![Figure 2: Selecting a Provider](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/select-provider.jpg)

{% hint style="warning" %}
The Git provider you select for configuring GitOps might impact the following sections:
   * [Deployment Template](../creating-application/deployment-template.md)
   * [Charts](../deploy-chart/README.md)
{% endhint %}

3. Fill all the mandatory fields. Refer [supported Git providers](#supported-git-providers) to know more about the respective fields.

   ![Figure 3: Entering Git Credentials](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/git-fields.jpg)

4. In the **Directory Management in Git** section, you get the following options:

   * **Use default git repository structure**:
   
      This option lets Devtron automatically create a GitOps repository within your organization. The repository name will match your application name, and it cannot be changed. Since Devtron needs admin access to create the repository, ensure the Git credentials you provided in Step 3 have administrator rights.

   * **Allow changing git repository for application**: 
   
      Select this option if you wish to use your own GitOps repo. This is ideal if there are any confidentiality/security concerns that prevent you from giving us admin access. Therefore, the onus is on you to create a GitOps repo with your Git provider, and then [add it to the specific application](../creating-application/gitops-config.md) on Devtron. Make sure the Git credentials you provided in Step 3 have at least read/write access. Choosing this option will unlock a [GitOps Configuration](../creating-application/gitops-config.md) page under the [App Configuration](../creating-application/README.md) tab.

   ![Figure 4: Need for User-defined Git Repo](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/user-defined-git.jpg)

5. Click **Save**/**Update**. A green tick will appear on the active Git provider.

---

## Supported Git Providers

Below are the Git providers supported in Devtron for storing configuration files. 

* [GitHub](#github)
* [GitLab](#gitlab)
* [Azure](#azure)
* [Bitbucket Cloud](#bitbucket-cloud)

### GitHub

{% hint style="info" %}
### Prerequisite

1. A GitHub account
2. A GitHub organization. If you don't have one, refer [Creating Organization in GitHub](#creating-organization-in-github).
{% endhint %}

Fill the following mandatory fields:

| Field | Description |
| --- | --- |
| **Git Host** | Shows the URL of GitHub, e.g., https://github.com/ |
| **GitHub Organisation Name** | Enter the GitHub organization name. <br />If you do not have one, refer [how to create organization in Github](#how-to-create-organization-in-github). |
| **GitHub Username** | Provide the username of your GitHub account. |
| **Personal Access Token** | Provide your personal access token (PAT). It is used as an alternate password to authenticate your GitHub account. <br />If you do not have one, create a GitHub PAT [here](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token).|


### GitLab

{% hint style="info" %}
### Prerequisite

1. A GitLab account
2. A GitLab group. If you don't have one, refer [Creating Group in GitLab](#creating-group-in-gitlab).
{% endhint %}

Fill the following mandatory fields:

| Field | Description |
| --- | --- |
| **Git Host** | Shows the URL of GitLab, e.g., https://gitlab.com/ |
| **GitLab Group ID** | Enter the GitLab group ID. <br />If you do not have one, refer [GitLab Group ID](#how-to-create-organization-in-gitlab).|
| **GitLab Username** | Provide the username of your GitLab account |
| **Personal Access Token** | Provide your personal access token (PAT). It is used as an alternate password to authenticate your GitLab account. <br />If you do not have one, create a GitLab PAT [here](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html). |


### Azure

{% hint style="info" %}
### Prerequisite

1. An organization on Azure DevOps. If you don't have one, refer [this link](https://learn.microsoft.com/en-us/azure/devops/organizations/accounts/create-organization?view=azure-devops#create-an-organization).
2. A project in your Azure DevOps organization. Refer [Creating Project in Azure](#creating-project-in-azure-devops).
{% endhint %}

Fill the following mandatory fields:

| Field | Description |
| --- | --- |
| **Azure DevOps Organisation Url*** | Enter the Org URL of Azure DevOps. Format should be `https://dev.azure.com/<org-name>`, where `<org-name>` represents the organization name, e.g., [https://dev.azure.com/devtron-test](https://dev.azure.com/devtron-test)|
| **Azure DevOps Project Name** | Enter the Azure DevOps project name. <br />If you do not have one, refer [Azure DevOps Project Name](#how-to-create-azure-devops-project-name).|
| **Azure DevOps Username*** | Provide the username of your Azure DevOps account |
| **Azure DevOps Access Token*** | Provide your Azure DevOps access token. It is used as an alternate password to authenticate your Azure DevOps account. <br />If you do not have one, create a Azure DevOps access token [here](https://docs.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=preview-page). |

### Bitbucket Cloud

{% hint style="info" %}
### Prerequisite

1. A Bitbucket account
2. A workspace in your Bitbucket account. Refer [Creating Workspace in Bitbucket](#creating-workspace-in-bitbucket).

{% endhint %}

Fill the following mandatory fields:

| Field | Description |
| --- | --- |
| **Bitbucket Host** | Shows the URL of Bitbucket Cloud, e.g., https://bitbucket.org/ |
| **Bitbucket Workspace ID** | Enter the Bitbucket workspace ID. <br />If you do not have one, refer [Bitbucket Workspace Id](#how-to-create-bitbucket-workspace-id).|
| **Bitbucket Username*** | Provide the username of your Bitbucket account |
| **Personal Access Token** | Provide your personal access token (PAT). It is used as an alternate password to authenticate your Bitbucket Cloud account. <br />If you do not have one, create a Bitbucket Cloud PAT [here](https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/). |

Fill the optional field if needed:

| Field | Description |
| --- | --- |
| **Bitbucket Project Key** | Enter the Bitbucket project key. <br />If you do not have one, refer [Bitbucket Project Key](https://support.atlassian.com/bitbucket-cloud/docs/group-repositories-into-projects/). <br />Note: If the project is not provided, the repository is automatically assigned to the oldest project in the workspace. |

---

## Miscellaneous

### Creating Organization in GitHub

{% hint style="warning" %}
We do **NOT** recommend using GitHub organization that contains your source code.
{% endhint %}

{% embed url="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/github/github-gitops-latest.mp4" caption="GitHub" %}

1. Create a new account on GitHub (if you do not have one).
2. On the upper-right corner of your GitHub page, click your profile photo, then click **Settings**.
3. On the `Access` section, click  **Organizations**.
4. On the `Organizations` section, click **New organization**.
5. Pick a [plan](https://docs.github.com/en/get-started/learning-about-github/githubs-products) for your organization. You have the option to select `create free organization` also.
6. On the `Set up your organization` page, 
   * Enter the `organization account name`, `contact email`.
   * Select the option your organization belongs to.
   * Verify your account and click **Next**.
   * Your `GitHub organization name` will be created. 

7. Go to your profile and click **Your organizations** to view all the organizations you created.

For more information about the plans available for your team, see [GitHub's products](https://docs.github.com/en/get-started/learning-about-github/githubs-products). You can also refer [GitHub organization](https://docs.github.com/en/github/setting-up-and-managing-organizations-and-teams/about-organizations) official doc page for more detail.

**Note**: 
* repo - Full control of private repositories (able to access commit status, deployment status, and public repositories).
* admin:org - Full control of organizations and teams (Read and write access).
* delete_repo - Grants delete repo access on private repositories.

### Creating Group in GitLab

{% embed url="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/gitlab/gitops-gitlab-latest1.mp4" caption="GitHub" %}


1. Create a new account on GitLab (if you do not have one).
2. You can create a group by going to the 'Groups' tab on the GitLab dashboard and click `New group`.
3. Select `Create group`.
4. Enter the group name (required) and select the optional descriptions if required, and click **Create group**.
5. Your group will be created and your group name will be assigned with a new `Group ID` (e.g. 61512475).

**Note**:
* api - Grants complete read/write access to the scoped project API.
* write_repository - Allows read/write access (pull, push) to the repository.

### Creating Project in Azure DevOps

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/azure+devops/azure-new-project-v2.jpg)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/azure+devops/azure-create-new-project-v2.jpg)

1. Go to Azure DevOps and navigate to Projects.
2. Select your organization and click `New project`. 
3. On the `Create new project` page, 
   * Enter the `project name` and description of the project.
   * Select the visibility option (private or public), initial source control type, and work item process.
   * Click **Create**.
   * Azure DevOps displays the project welcome page with the `project name`.

You can also refer [Azure DevOps - Project Creation](https://docs.microsoft.com/en-us/azure/devops/organizations/projects/create-project?view=azure-devops&tabs=preview-page) official page for more details.

**Note**:
* code - Grants the ability to read source code and metadata about commits, change sets, branches, and other version control artifacts. [More information on scopes in Azure DevOps](https://docs.microsoft.com/en-us/azure/devops/integrate/get-started/authentication/oauth?view=azure-devops#scopes).

### Creating Workspace in Bitbucket

{% embed url="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/bitbucket/bitbucket-latest-gitops.mp4" caption="GitHub" %}

1. Create a new individual account on Bitbucket (if you do not have one).
2. Select your profile and settings avatar on the upper-right corner of the top navigation bar.
3. Select `All workspaces` from the dropdown menu.
4. Select the `Create workspace` on the upper-right corner of the `Workspaces` page.
5. On the `Create a Workspace` page:
  * Enter a `Workspace name`. 
  * Enter a `Workspace ID`. Your ID cannot have any spaces or special characters, but numbers and capital letters are fine. This ID becomes part of the URL for the workspace and anywhere else where there is a label that identifies the team (APIs, permission groups, OAuth, etc.).
  * Click **Create**.
6. Your `Workspace name` and `Workspace ID` will be created.

You can also refer [official Bitbucket Workspace page](https://support.atlassian.com/bitbucket-cloud/docs/what-is-a-workspace/) for more details.

**Note**:
* repo - Full control of repositories (Read, Write, Admin, Delete) access. 
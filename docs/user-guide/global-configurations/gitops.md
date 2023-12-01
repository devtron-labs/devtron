# GitOps

Devtron uses GitOps to store Kubernetes configuration files of the applications. 
For storing the configuration files and desired state of the applications, Git credentials must be provided at **Global Configurations** > **GitOps** within the Devtron dashboard.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/git-providers.jpg)

Below are the Git providers which are available in Devtron. Select one of the Git providers (e.g., GitHub) to configure GitOps:

* [GitHub](#github)
* [GitLab](#gitlab)
* [Azure](#azure)
* [BitBucket Cloud](#bitbucket-cloud)

**Note**: The Git provider you select for configuring GitOps will impact the following sections:
* Deployment Template, [click here](https://docs.devtron.ai/user-guide/creating-application/deployment-template) to learn more.
* Charts, [click here](https://docs.devtron.ai/user-guide/deploy-chart) to learn more.



## GitHub

If you select `GitHuB` as your git provider, please provide the information in the following fields to configure GitOps:

| Fields | Description |
| --- | --- |
| **Git Host** | This field shows the URL of the selected Git provider. <br>As an example: https://github.com/ for GitHub.</br> |
| **GitHub Organisation Name** | Enter the GitHub organization name.<br>If you do not have one, create using [how to create organization in Github](#how-to-create-organization-in-github).</br> |
| **GitHub Username** | Provide the username of your GitHub account. |
| **Personal Access Token** | Provide your personal access token (PAT). It is used as an alternate password to authenticate your GitHub account.<br>If you do not have one, create a GitHub PAT [here](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token).</br> |

### How to create organization in GitHub

**Note**: We do NOT recommend using GitHub organization which contains your source code.


{% embed url="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/github/github-gitops-latest.mp4" caption="GitHub" %}

1. Create a new account on GitHub (if you do not have one).
2. On the upper-right corner of your GitHub page, click your profile photo, then click `Settings`.
3. On the `Access` section, click  `Organizations`.
4. On the `Organizations` section, click `New organization`.
5. Pick a [plan](https://docs.github.com/en/get-started/learning-about-github/githubs-products) for your organization. You have the option to select `create free organization` also.
6. On the `Set up your organization` page, 
   * Enter the `organization account name`, `contact email`.
   * Select the option your organization belongs to.
   * Verify your account and click `Next`.
   * Your `GitHub organization name` will be created. 

7. Go to your profile and click `Your organizations` to view all the organizations you created.

For more information about the plans available for your team, see [GitHub's products](https://docs.github.com/en/get-started/learning-about-github/githubs-products). You can also refer [GitHub organization](https://docs.github.com/en/github/setting-up-and-managing-organizations-and-teams/about-organizations) official doc page for more detail.

**Note**: 
* repo - Full control of private repositories (able to access commit status, deployment status and public repositories).
* admin:org - Full control of organizations and teams (Read and write access).
* delete_repo - Grants delete repo access on private repositories.


## GitLab

If you select `GitLab` as your git provider, please provide the information in the following fields to configure GitOps:

| Fields | Description |
| --- | --- |
| **Git Host** | This field shows the URL of the selected Git provider. <br>As an example:<br>https://gitlab.com/ for GitLab. |
| **GitLab Group ID** | Enter the GitLab group ID.<br>If you do not have one, create using [GitLab Group ID](#how-to-create-organization-in-gitlab).</br> |
| **GitLab Username** | Provide the username of your GitLab account. |
| **Personal Access Token** | Provide your personal access token (PAT). It is used as an alternate password to authenticate your GitLab account.<br>If you do not have one, create a GitLab PAT [here](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html).</br> |


### How to create organization in GitLab


{% embed url="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/gitlab/gitops-gitlab-latest1.mp4" caption="GitHub" %}


1. Create a new account on GitLab (if you do not have one).
2. You can create a group by going to the 'Groups' tab on the GitLab dashboard and click `New group`.
3. Select `Create group`.
4. Enter the group name (required) and select the optional descriptions, if require and click `Create group`.
5. Your group will be created and your group name will be assigned with a new `Group ID` (e.g. 61512475).


**Note**:
* api - Grants complete read/write access to the scoped project API.
* write_repository - Allows read/write access (pull, push) to the repository.


## Azure

If you select `GitAzureLab` as your git provider, please provide the information in the following fields to configure GitOps:

| Fields | Description |
| --- | --- |
| **Azure DevOps Organisation Url*** | This field displays the URL of the chosen Git provider. For Azure, the format would be `https://dev.azure.com/<org-name>`, where `<org-name>` represents the organization name <br> As an example, consider [https://dev.azure.com/devtron-test](https://dev.azure.com/devtron-test), where "devtron-test" is the organization name. |
| **Azure DevOps Project Name** | Enter the Azure DevOps project name.<br>If you do not have one, create using [Azure DevOps Project Name](#how-to-create-azure-devops-project-name).</br> |
| **Azure DevOps Username*** | Provide the username of your Azure DevOps account. |
| **Azure DevOps Access Token*** | Provide your Azure DevOps access token. It is used as an alternate password to authenticate your Azure DevOps account.<br>If you do not have one, create a Azure DevOps access token [here](https://docs.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=preview-page).</br> |


### How to create Azure DevOps project name

**Note**: You need an organization before you can create a project. If you have not created an organization yet, create one by following the instructions in [Sign up, sign in to Azure DevOps](https://learn.microsoft.com/en-us/azure/devops/user-guide/sign-up-invite-teammates?view=azure-devops), which also creates a project. Or see [Create an organization or project collection](https://learn.microsoft.com/en-us/azure/devops/organizations/accounts/create-organization?view=azure-devops).

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/azure+devops/azure-new-project.jpg)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/azure+devops/azure-create-new-project.jpg)

1. Go to Azure DevOps and navigate to Projects.
2. Select your organization and click `New project`. 
3. On the `Create new project` page, 
   * Enter the `project name` and description of the project.
   * Select the visibility option (private or public), initial source control type, and work item process.
   * Click `Create`.
   * Azure DevOps displays the project welcome page with the `project name`.

You can also refer [Azure DevOps Project Name](https://docs.microsoft.com/en-us/azure/devops/organizations/projects/create-project?view=azure-devops&tabs=preview-page) official doc page for more detail.

**Note**:
* code - Grants the ability to read source code and metadata about commits, change sets, branches, and other version control artifacts. [More Information on scopes in Azure devops](https://docs.microsoft.com/en-us/azure/devops/integrate/get-started/authentication/oauth?view=azure-devops#scopes).


## Bitbucket Cloud

If you select `Bitbucket Cloud` as your git provider, please provide the information in the following fields to configure GitOps:

| Fields | Description |
| --- | --- |
| **Bitbucket Host** | This field shows the URL of the selected Git provider. <br>As an example:<br>https://bitbucket.org/ for Bitbucket. |
| **Bitbucket Workspace ID** | Enter the Bitbucker workspace ID.<br>If you do not have one, create using [Bitbucket Workspace Id](#how-to-create-bitbucket-workspace-id).</br> |
| **Bitbucket Project Key** | Enter the Bitbucket project key.<br>If you do not have one, create using [Bitbucket Project Key](https://support.atlassian.com/bitbucket-cloud/docs/group-repositories-into-projects/).</br><br>Note: This field is not mandatory. If the project is not provided, the repository is automatically assigned to the oldest project in the workspace.</br> |
| **Bitbucket Username*** | Provide the username of your Bitbucket account. |
| **Personal Access Token** | Provide your personal access token (PAT). It is used as an alternate password to authenticate your Bitbucket Cloud account.<br>If you do not have one, create a Bitbucket Cloud PAT [here](https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/).</br> |


### How to create Bitbucket workspace ID

{% embed url="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/bitbucket/bitbucket-latest-gitops.mp4" caption="GitHub" %}

1. Create a new individual account on Bitbucket (if you do not have one).
2. Select your profile and settings avatar on the upper-right corner of the top navigation bar.
3. Select `All workspaces` from the dropdown menu.
4. Select the `Create workspace` on the upper-right corner of the `Workspaces` page.
5. On the `Create a Workspace` page:
  * Enter a `Workspace name`. 
  * Enter a `Workspace ID`. Your ID cannot have any spaces or special characters, but numbers and capital letters are fine. This ID becomes part of the URL for the workspace and anywhere else where there is a label that identifies the team (API's, permission groups, OAuth, etc.).
  * Click `Create`.
6. Your `Workspace name` and `Workspace ID` will be created.

You can also refer [Bitbucket Workspace Id](https://support.atlassian.com/bitbucket-cloud/docs/what-is-a-workspace/) official doc page for more detail.

**Note**:
* repo - Full control of repositories (Read, Write, Admin, Delete) access. 

Click **Save** to save your GitOps configuration details.

**Note** : A Green tick will appear on the active GitOp provider.

# GitOps

`GitOps` is a branch of DevOps that focuses on using git repositories to manage infrastructure and application code deployments.

## Benefits of GitOps Configuration

Devtron uses GitOps and stores configurations in git; Git Credentials can be entered at `Global Configuration > GitOps` which is used by Devtron for configuration management and storing desired state of the application configuration. 

Some of the major benefits for configuring GitOps are:

* Using GitOps can help development teams to solve a number of systemic issues through the implementation of new infrastructure configurations.
* With GitOps, whenever there is any divergence between Git and what's running in a cluster, developers are alerted. 
* GitOps allows delivery pipelines to seamlessly roll out changes to infrastructure initiated through Git.
* While working as a team, team members can collaborate with one another to easily identify and correct errors within a given time.


Areas impacted by GitOps are:

* Deployment Template, [click here](https://docs.devtron.ai/user-guide/creating-application/deployment-template) to learn more.
* Charts, [click here](https://docs.devtron.ai/user-guide/deploy-chart) to learn more.


![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/git-providers.jpg)

## Add Git Configuration 

To add GitOps, go to the `Global Configurations` section and click `Gitops`.

Below are the Git providers which are available in Devtron. Select one of the Git providers (e.g., GitHub) to configure GitOps:

* [GitHub](#github)
* [GitLab](#gitlab)
* [Azure](#azure)
* [BitBucket Cloud](#bitbucket-cloud)

### GitHub

If you select `GitHuB` as your git provider, please provide the information in the following fields to configure GitOps:

| Fields | Description |
| --- | --- |
| **Git Host** | This field shows the URL of the selected Git provider. <br>As an example:<br>https://github.com/ for GitHub. |
| **GitHub Organisation Name** | Enter the GitHub organization name.<br>If you do not have one, create using [Github Organization Name](https://docs.github.com/en/github/setting-up-and-managing-organizations-and-teams/about-organizations)</br>. |
| **GitHub Username** | Provide the username of your GitHub account. |
| **Personal Access Token** | Provide your personal access token (PAT). It is used as an alternate password to authenticate your GitHub account.<br>If you do not have one, create a GitHub PAT [here](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token)</br>. |


**Note**: 
* repo - Full control of private repositories (able to access commit status, deployment status and public repositories).
* admin:org - Full control of organizations and teams (Read and write access).
* delete_repo - Grants delete repo access on private repositories.


### GitLab

If you select `GitLab` as your git provider, please provide the information in the following fields to configure GitOps:

| Fields | Description |
| --- | --- |
| **Git Host** | This field shows the URL of the selected Git provider. <br>As an example:<br>https://gitlab.com/ for GitLab. |
| **GitLab Group ID** | Enter the GitLab group ID.<br>If you do not have one, create using [GitLab Group ID](https://docs.gitlab.com/ee/user/group/)</br>. |
| **GitLab Username** | Provide the username of your GitLab account. |
| **Personal Access Token** | Provide your personal access token (PAT). It is used as an alternate password to authenticate your GitLab account.<br>If you do not have one, create a GitLab PAT [here](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html)</br>. |

**Note**:
* api - Grants complete read/write access to the scoped project API.
* write_repository - Allows read/write access (pull, push) to the repository.


### Azure

If you select `GitAzureLab` as your git provider, please provide the information in the following fields to configure GitOps:

| Fields | Description |
| --- | --- |
| **Azure DevOps Organisation Url*** | This field shows the URL of the selected Git provider. <br>As an example:<br>https://dev.azure.com/ for Azure. |
| **Azure DevOps Project Name** | Enter the Azure DevOps project name.<br>If you do not have one, create using [Azure DevOps Project Name](https://docs.microsoft.com/en-us/azure/devops/organizations/projects/create-project?view=azure-devops&tabs=preview-page)</br>. |
| **Azure DevOps Username*** | Provide the username of your Azure DevOps account. |
| **Azure DevOps Access Token*** | Provide your Azure DevOps access token. It is used as an alternate password to authenticate your Azure DevOps account.<br>If you do not have one, create a Azure DevOps access token [here](https://docs.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=preview-page)</br>. |

**Note**:
* code - Grants the ability to read source code and metadata about commits, change sets, branches, and other version control artifacts. [More Information on scopes in Azure devops](https://docs.microsoft.com/en-us/azure/devops/integrate/get-started/authentication/oauth?view=azure-devops#scopes).


### Bitbucket Cloud

If you select `Bitbucket Cloud` as your git provider, please provide the information in the following fields to configure GitOps:

| Fields | Description |
| --- | --- |
| **Bitbucket Host** | This field shows the URL of the selected Git provider. <br>As an example:<br>https://bitbucket.org/ for Bitbucket. |
| **Bitbucket Workspace ID** | Enter the Bitbucker workspace ID.<br>If you do not have one, create using [Bitbucket Workspace Id](https://support.atlassian.com/bitbucket-cloud/docs/what-is-a-workspace/)</br>. |
| **Bitbucket Project Key** | Enter the Bitbucket project key.<br>If you do not have one, create using [Bitbucket Project Key](https://support.atlassian.com/bitbucket-cloud/docs/group-repositories-into-projects/)</br>.<br>Note: This field is not mandatory. If the project is not provided, the repository is automatically assigned to the oldest project in the workspace.</br> |
| **Bitbucket Username*** | Provide the username of your Bitbucket account. |
| **Personal Access Token** | Provide your personal access token (PAT). It is used as an alternate password to authenticate your Bitbucket Cloud account.<br>If you do not have one, create a Bitbucket Cloud PAT [here](https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/)</br>. |

**Note**:
* repo - Full control of repositories (Read, Write, Admin, Delete) access. 

Click **Save** to save your GitOps configuration details.

**Note** : A Green tick will appear on the active GitOp provider.

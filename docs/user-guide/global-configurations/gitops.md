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

To add GitOps, go to the `Gitops` section of `Global Configurations`. Click **GitOps**.

Below are the Git providers which are available in Devtron. Select one of the Git providers (e.g., GitHub) to configure GitOps:

* **GitHub**
* **GitLab**
* **Azure**
* **BitBucket Cloud**

Provide the information in the following fields to configure GitOps:

| Fields | Description |
| --- | --- |
| **Git Host** | This field shows the URL of the selected Git provider. <br>As an example:<br>https://github.com/ for GitHub, https://gitlab.com/ for GitLab, https://dev.azure.com/ for Azure and https://bitbucket.org/ for BitBucket.</br> |
| **GitHub Organisation Name** | Enter the GitHub organization name.<br>If you do not have one, create using:ul><li>[Github Organization Name](https://docs.github.com/en/github/setting-up-and-managing-organizations-and-teams/about-organizations)</li></ul> <ul><li>[Gitlab group Id](https://docs.gitlab.com/ee/user/group/)</li></ul><ul><li>[Azure DevOps Project Name](https://docs.microsoft.com/en-us/azure/devops/organizations/projects/create-project?view=azure-devops&tabs=preview-page)</li></ul><ul><li>Bitbucket Workspace Id](https://support.atlassian.com/bitbucket-cloud/docs/what-is-a-workspace/)</li></ul>. |
| **GitHub Username** | Provide the username of your git account. |
| **Personal Access Token** | A personal access token (PAT) is used as an alternate password to authenticate into your git accounts. |



### 2. GitHub Organization Name / GitLab Group ID / Azure DevOps Project Name / BitBucket Workspace ID:

In the case of GitHub provide `Github Organization Name*`. Learn more about [Github organization Name](https://docs.github.com/en/github/setting-up-and-managing-organizations-and-teams/about-organizations). <br />
In the case of Gitlab provide `Gitlab group Id*`. Learn more about [Gitlab group Id](https://docs.gitlab.com/ee/user/group/). <br />
Similarly in the case of Azure provide `Azure DevOps Project Name*`. Learn more about [Azure DevOps Project Name](https://docs.microsoft.com/en-us/azure/devops/organizations/projects/create-project?view=azure-devops&tabs=preview-page). <br />
For Bitbucket Cloud, provide `Bitbucket Workspace Id*`. Learn more about [Bitbucket Workspace Id](https://support.atlassian.com/bitbucket-cloud/docs/what-is-a-workspace/).

### 3. BitBucket Project Key: 

This field is non-mandatory and is only to be filled when you have chosen `Bitbucket Cloud` as your git provider. If not provided, the oldest project in the workspace will be used. Learn more about [Bitbucket Project Key](https://support.atlassian.com/bitbucket-cloud/docs/group-repositories-into-projects/).
### 4. Git access credential

Provide Git `Username` and `Personal Access Token` of your git account. 

**\(a\) Username** 
Username for your git account.

**\(b\) Personal Access Token**  
A personal access token (PAT) is used as an alternate password to authenticate into your git accounts. 

#### For GitHub [Creating a GitHub Personal Access Token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token):

* repo - Full control of private repositories(Access commit status , Access deployment status and Access public repositories).
* admin:org - Full control of organizations and teams(Read and write access).
* delete_repo - Grants delete repo access on private repositories.

#### For GitLab [Creating a GitLab Personal Access Token](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html):

* api - Grants complete read/write access to the scoped project API.
* write_repository - Allows read-write access (pull, push) to the repository. 

#### For Azure DevOps [Creating a Azure DevOps Access Token](https://docs.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=preview-page): 

* code - Grants the ability to read source code and metadata about commits, change sets, branches, and other version control artifacts.
[More Information on scopes in Azure devops](https://docs.microsoft.com/en-us/azure/devops/integrate/get-started/authentication/oauth?view=azure-devops#scopes).

#### For BitBucket Cloud [Creating a Bitbucket Cloud Personal Access Token (App passwords)](https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/)

* repo - Full control of repositories (Read, Write, Admin, Delete access). 

Click on Save to save your gitOps configuration details.
 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/bitbucket-cloud.jpg)

> Note: A Green tick will appear on the active gitOps provider.

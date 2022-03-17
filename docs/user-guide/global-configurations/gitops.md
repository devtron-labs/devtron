# GitOps

## Why Devtron takes GitOps Configuration?
Devtron uses GitOps and stores configurations in git; Git Credentials can be entered at `Global Configuration > GitOps` which is used by Devtron for configuration management and storing desired state of the application configuration. 
In case GitOps is not configured, Devtron cannot deploy any application or charts. 


Areas impacted by GitOps are:

* Deployment Template, [click here](https://docs.devtron.ai/user-guide/creating-application/deployment-template) to learn more.
* Charts, [click here](https://docs.devtron.ai/user-guide/deploy-chart) to learn more.


## Add Git Configuration

Select the GitOps section of global configuration. At the top of the section, four Git providers are available.

* **GitHub**
* **GitLab**
* **Azure**
* **BitBucket Cloud**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/git-providers.jpg)

Select one of the Git provider.

**GitHub**

In the case of `GitHub` git provider you need to provide the following inputs as given below:

## GitHub host
This field is filled by default, Showing the URL of the selected git provider, here - https://github.com for GitHub

### GitHub organization name
In the case of GitHub provide `Github Organization Name*`. Learn more about [Github organization Name](https://docs.github.com/en/github/setting-up-and-managing-organizations-and-teams/about-organizations). <br />

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/github-org.jpg)

### GitHub username

Username for your github account.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/github-final.jpg)

### [GitHub Personal Access Token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token)

A personal access token (PAT) is used as an alternate password to authenticate into your git accounts.

* repo - Full control of private repositories(Access commit status , Access deployment status and Access public repositories).
* admin:org - Full control of organizations and teams(Read and write access).
* delete_repo - Grants delete repo access on private repositories.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/github-token.jpg)


**GitLab**

In the case of `GitLab` git provider you need to provide the following inputs as given below:

### GitLab host

This field is filled by default, Showing the URL of the selected git provider, here - https://gitlab.com for GitLab

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/gitlab-final.jpg)

### Gitlab group id

In the case of Gitlab provide `Gitlab group Id*`. Learn more about [Gitlab group Id](https://docs.gitlab.com/ee/user/group/). <br />

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/gitlab-group-id.jpg)

### GitLab username

Username for your gitlab account.

### [GitLab Personal Access Token](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html)

* api - Grants complete read/write access to the scoped project API.
* write_repository - Allows read-write access (pull, push) to the repository. 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/gitlab-token.jpg)


**Azure**

In the case of `Azure DevOps` git provider you need to provide the following inputs as given below:

### Azure Devops organization url

Create the organization from your Azure Devops account. 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/azure-org-url.jpg)

### Azure DevOps Project Name

In the case of Azure provide `Azure DevOps Project Name*`. Learn more about [Azure DevOps Project Name](https://docs.microsoft.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/azure-project.jpg)

### Azure Devops username

Provide the username of your azure devops account.

### [Azure DevOps Access Token](https://docs.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=preview-page)

* code - Grants the ability to read source code and metadata about commits, change sets, branches, and other version control artifacts.
[More Information on scopes in Azure devops](https://docs.microsoft.com/en-us/azure/devops/integrate/get-started/authentication/oauth?view=azure-devops#scopes).

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/azure+token.jpg)


**Bitbucket**

In the case of `Bitbucket` git provider you need to provide the following inputs as given below:

### BitBucket Host

This field is filled by default, Showing the URL of the selected git provider,here https://bitbucket.org for BitBucket

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/bitbucket-final.jpg)

### BitBucket Workspace ID

For Bitbucket Cloud, provide `Bitbucket Workspace Id*`. Learn more about [Bitbucket Workspace Id](https://support.atlassian.com/bitbucket-cloud/docs/what-is-a-workspace/).

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/bitbucket-ws.jpg)

### BitBucket Project Key

This field is non-mandatory and is only to be filled when you have chosen `Bitbucket Cloud` as your git provider. If not provided, the oldest project in the workspace will be used. Learn more about [Bitbucket Project Key](https://support.atlassian.com/bitbucket-cloud/docs/group-repositories-into-projects/).

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/bitbucket-project-key.jpg)

### Bitbucket username

Provide the username of your bitbucket account.

### [Bitbucket Cloud Personal Access Token (App passwords)](https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/)

* repo - Full control of repositories (Read, Write, Admin, Delete access). 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/bitbucket-password.jpg)


Click on Save to save your gitOps configuration details.
 

> Note: A Green tick will appear on the active gitOps provider.

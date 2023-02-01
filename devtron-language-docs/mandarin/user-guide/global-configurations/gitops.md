# GitOps
Devtron使用GitOps存储应用程序的Kubernetes设置文件。
为了存储应用程序的设置文件和所需状态，Git凭据必须在**全局设置** > **GitOps的**在Devron仪表板内。

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/git-providers.jpg)

下面是在Devron中可用的Git提供者。选择其中一个Git提供程序（例如GitHub）来设置GitOps：

* [GitHub](#github)
* [GitLab](#gitlab)
* [天蓝色](#azure)
* [BitBucket Cloud](#bitbucket-cloud)

**注意**：您为设置GitOps选择的Git提供程序将影响以下部分：

* 部署模板，[点击此处](https://docs.devtron.ai/user-guide/creating-application/deployment-template)了解更多。
* 图表，[点击此处](https://docs.devtron.ai/user-guide/deploy-chart)了解更多。
## GitHub
如果选择`GitHuB`作为您的Git提供商，请提供以下字段中的信息以设置GitOps：

|字段|说明|
| :-: | :-: |
|**Git主机**|此字段显示所选Git提供程序的URL。<br>作为一个例子：对于GitHub为https://github.com/。</br>|
|**GitHub组织名称**|输入GitHub组织名称。<br>如果您没有一个，创建使用[如何在Github中创建组织](#how-to-create-organization-in-github)。</br>|
|**GitHub用户名**|提供您的GitHub帐户的用户名。|
|**个人访问令牌**|提供您的个人访问令牌（PAT）。它用作验证GitHub帐户的备用密码。<br>如果您没有一个，在[此处](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token)创建一个GitHub PAT。</br>|

### 如何在GitHub中创建组织
**注意**：我们不建议使用包含源代码的GitHub组织。

{% embed url="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/github/github-gitops-latest.mp4" caption="GitHub" %}

1. 在GitHub上创建一个新帐户（如果您没有）。
1. 在GitHub页面的右上角，单击您的个人资料照片，然后点击`设置`。
1. 在`查阅资料`部分，点击`机构名称`。
1. 在`机构名称`部分，点击`新组织`。
1. 为您的组选择一个[图则](https://docs.github.com/en/get-started/learning-about-github/githubs-products)织。您也可以选择`创建自由组织`。
1. 在`设置您的组织`页面，
   * 输入`组织帐户名称`，`联络电邮`。
   * 选择您的组织所属的选项。
   * 验证您的帐户并点击`下一个`。
   * 您的`GitHub组织名称`将被创建。
1. 转到您的个人资料，然后点击`您的机构`查看您创建的所有组织。

有关可供您的团队使用的计划的详细信息，请参阅[GitHub的产品](https://docs.github.com/en/get-started/learning-about-github/githubs-products)。您也可以参考[GitHub组织](https://docs.github.com/en/github/setting-up-and-managing-organizations-and-teams/about-organizations)官方文档页面以获取更多细节。

**注意**：

* repo - 完全控制私有存储库（能够访问提交状态，部署状态和公共存储库）。
* admin:org - 组织和团队的完全控制（读写访问）。
* delete\_repo - 授予删除私有存储库上的存储库访问权限。
## GitLab
如果选择`GitLab`作为您的Git提供商，请提供以下字段中的信息以设置GitOps：

|字段|说明|
| :-: | :-: |
|**Git主机**|此字段显示所选Git提供程序的URL。<br>作为一个例子：<br>对于GitLab为https://gitlab.com/。|
|**GitLab组ID**|输入GitLab组ID。<br>如果您没有一个，创建使用[GitLab组ID](#how-to-create-organization-in-gitlab)。</br>|
|**GitLab用户名**|提供您的GitLab帐户的用户名。|
|**个人访问令牌**|提供您的个人访问令牌（PAT）。它用作验证GitLab帐户的备用密码。<br>如果您没有一个，在[此处](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html)创建一个GitLab PAT。</br>|

### 如何在GitLab中创建组织
{% embed url="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/gitlab/gitops-gitlab-latest1.mp4" caption="GitHub" %}

1. 在GitLab上创建一个新帐户（如果您没有）。
1. 您可以通过转到GitLab仪表板上的"组"选项卡创建一个组，然后点击`新组别`。
1. 选择`创建组`。
1. 输入组名称（必填），并选择可选的描述（如果需要），然后点击`创建组`。
1. 您的组将被创建，您的组名将被分配一个新的`组ID`（例如61512475）。

**注意**：

* api - 授予对作用域项目API的完全读/写访问权限。
* write\_repository - 允许对存储库进行读/写访问（拉、推）。
## 天蓝色
如果选择`GitAzureLab`作为您的Git提供商，请提供以下字段中的信息以设置GitOps：

|字段|说明|
| :-: | :-: |
|**Azure DevOps组织Url**\*|此字段显示所选Git提供程序的URL。<br>作为一个例子：<br>对于Azure为https://dev.azure.com/。|
|**Azure DevOps项目名称**|输入Azure DevOps项目名称。<br>如果您没有一个，创建使用[Azure DevOps项目名称](#how-to-create-azure-devops-project-name)。</br>|
|**Azure DevOps用户名**\*|提供Azure DevOps帐户的用户名。|
|**Azure DevOps访问令牌**\*|提供Azure DevOps访问令牌。它用作验证Azure DevOps帐户的备用密码。<br>如果没有，在[此处](https://docs.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=preview-page)创建Azure DevOps访问令牌。</br>|

### 如何创建Azure DevOps项目名称
**注意**：您需要一个组织才能创建项目。如果您还没有创建一个组织，请按照以下说明创建一个组织[注册，登录到Azure DevOps](https://learn.microsoft.com/en-us/azure/devops/user-guide/sign-up-invite-teammates?view=azure-devops)，这也创建了一个项目。或查看[创建组织或项目集合](https://learn.microsoft.com/en-us/azure/devops/organizations/accounts/create-organization?view=azure-devops)。

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/azure+devops/azure-new-project.jpg)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/azure+devops/azure-create-new-project.jpg)

1. 转到Azure DevOps并导航到项目。
1. 选择您的组织并点击`新项目`。
1. 在`创建新项目`页面，
   * 输入`项目名称`和项目的描述。
   * 选择可见性选项（私有或公共）、初始源代码管理类型和工作项进程。
   * 点击`创建`。
   * Azure DevOps显示项目欢迎页面，其中包含`项目名称`。

您也可以参考[Azure DevOps项目名称](https://docs.microsoft.com/en-us/azure/devops/organizations/projects/create-project?view=azure-devops&tabs=preview-page)官方文档页面以获取更多细节。

**注意**：

* code - 授予读取有关提交、更改集、分支和其他版本控制工件的源代码和元数据的能力。[有关Azure devops中的作用域的详细信息](https://docs.microsoft.com/en-us/azure/devops/integrate/get-started/authentication/oauth?view=azure-devops#scopes)。
## BitBucket Cloud
如果选择`Bitbucket Cloud`作为您的Git提供商，请提供以下字段中的信息以设置GitOps：

|字段|说明|
| :-: | :-: |
|**Bitbucket主机**|此字段显示所选Git提供程序的URL。<br>作为一个例子：<br>对于Bitbucket为https://bitbucket.org/。|
|**Bitbucket工作区ID**|输入Bitbucker工作区ID。<br>如果您没有一个，创建使用[Bitbucket工作区Id](#how-to-create-bitbucket-workspace-id)。</br>|
|**Bitbucket项目密钥**|输入Bitbucket项目密钥。<br>如果您没有一个，创建使用[Bitbucket项目密钥](https://support.atlassian.com/bitbucket-cloud/docs/group-repositories-into-projects/)。</br><br>注意：此字段不是强制性的。如果未提供项目，则存储库会自动分配给工作区中最旧的项目。</br>|
|**Bitbucket用户名**\*|提供您的Bitbucket帐户的用户名。|
|**个人访问令牌**|提供您的个人访问令牌（PAT）。它用作验证Bitbucket Cloud帐户的备用密码。<br>如果您没有，在[此处](https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/)创建一个Bitbucket Cloud PAT。</br>|

### 如何创建Bitbucket工作区ID
{% embed url="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/bitbucket/bitbucket-latest-gitops.mp4" caption="GitHub" %}

1. 在Bitbucket上创建一个新的个人帐户（如果您没有）。
1. 在顶部导航栏的右上角选择您的个人资料和设置头像。
1. 从下拉菜单选择`所有工作区`。
1. 在右上角的`工作区`页选择`创建工作区`。
1. 在`创建工作区`页面：
* 输入一个`工作区名称`。
* 输入一个`工作区ID`。您的ID不能有任何空格或特殊字符，但数字和大写字母可以。此ID成为工作区URL的一部分，以及任何其他有标识团队的标签（API、权限组、OAuth等）的地方。
* 点击`创建`。
6. 您的`工作区名称`和`工作区ID`将被创建。

您也可以参考[Bitbucket工作区Id](https://support.atlassian.com/bitbucket-cloud/docs/what-is-a-workspace/)的官方文档页面以获取更多细节。

**注意**：

* repo - 完全控制存储库（读、写、管理、删除）访问。

点击**保存**以保存您的GitOps设置详细信息。

**注意**：活动的GitOp提供程序上会出现绿色勾号。

# Git帐户
Git帐户允许您将代码源与Devtron连接。您将能够使用这些Git帐户使用CI管道构建代码。
## 添加Git帐户
要添加Git帐户，请转到 `Git帐户` 的部分 `全局配置`。点击 **添加Git帐户**。

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/git-accounts/git-accounts.jpg)

提供以下字段中的信息以添加Git帐户：

|字段|说明|
| :- | :- |
|`名称`|为Git提供程序提供名称。<br>注意：此名称将在应用程序配置中提供> [Git仓库](../creating-application/git-material.md) 下拉列表。</br>|
|`Git主机`|它是托管相应应用程序Git存储库的Git提供程序。<br>注意：默认情况下, `比特桶,比特桶` 和 `GitHub` 在下拉列表中可用。您可以通过点击添加多个 `[+添加Git主机]`。</br>|
|`URL`|提供Git主机`URL`。<br>作为一个例子：对于GitHub为<https://github.com>，对于GitLab为 <https://gitlab.com> 等。|
|`身份验证类型`|Devtron支持三种类型的身份验证：<ul><li>**用户认证：**如果选择 `用户认证` 作为身份验证类型，则必须提供 `用户名称` 和 `密码`或 `自动令牌` 用于版本控制帐户的身份验证。</li></ul> <ul><li>**匿名：**如果选择 `匿名` 作为身份验证类型，则不需要提供 `用户名称` 和 `密码`。<br>注意：如果身份验证类型设置为 `匿名`，只有公共Git存储库将是可访问的。</li></ul><ul><li>**SSH密钥：**如果你选择 `SSH密钥` 作为身份验证类型，则必须提供 `私有SSH密钥` 对应于您的版本控制帐户中添加的公钥。</li></ul>|

## 更新Git帐户
更新Git帐户：

1. 点击要更新的Git帐户。
1. 更新所需的更改。
1. 点击 `更新资料` 保存更改。

只能在一种身份验证类型或一种协议类型中进行更新，即HTTPS（匿名或用户身份验证）和SSH。您可以从`匿名`更新到`用户认证`，反之亦然，但不能从`匿名`或 `用户认证`更新到`SSH`，反之亦然。

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/git-accounts/update-git-accounts.jpg)

注意：

* 您可以启用或禁用Git帐户。启用的Git帐户将在应用程序配置> [Git仓库](../creating-application/git-material.md)。

![](../../user-guide/global-configurations/images/git-account-enable-disable.jpg)

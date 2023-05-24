# Git Accounts

Git Accounts allow you to connect your code source with Devtron. You will be able to use these git accounts to build the code using the CI pipeline.

## Add Git Account

To add git account, go to the `Git accounts` section of `Global Configurations`. Click **Add git account**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/git-accounts/git-accounts.jpg)

Provide the information in the following fields to add your git account:

| Field | Description |
| :--- | :--- |
| `Name` | Provide a name to your Git provider.<br>Note: This name will be available on the App Configuration > [Git repository](../creating-application/git-material.md) drop-down list.</br> |
| `Git host` |  It is the git provider on which corresponding application git repository is hosted.<br>Note: By default, `Bitbucket` and `GitHub` are available in the drop-down list. You can add many as you want by clicking `[+ Add Git Host]`.</br>  |
| `URL` | Provide the Git host `URL`.<br>As an example: [https://github.com](https://github.com) for Github, [https://gitlab.com](https://gitlab.com) for GitLab etc. |
| `Authentication Type` | Devtron supports three types of authentications:<ul><li>**User auth:** If you select `User auth` as an authentication type, then you must provide the `Username` and `Password`or `Auth token` for the authentication of your version control account.</li></ul> <ul><li>**Anonymous:** If you select `Anonymous` as an authentication type, then you do not need to provide the `Username` and `Password`.<br>Note: If authentication type is set as `Anonymous`, only public git repository will be accessible.</li></ul><ul><li>**SSH Key:** If you choose `SSH Key` as an authentication type, then you must provide the `Private SSH Key` corresponding to the public key added in your version control account.</li></ul> |



## Update Git Account

To update the git account:

1. Click the git account which you want to update. 
2. Update the required changes.
3. Click `Update` to save the changes.

Updates can only be made within one Authentication type or one protocol type, i.e. HTTPS (Anonymous or User Auth) & SSH. You can update from `Anonymous` to `User Auth` & vice versa, but not from `Anonymous` or `User Auth` to `SSH` and vice versa.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/git-accounts/update-git-accounts.jpg)

Note:
* You can enable or disable a git account. Enabled git accounts will be available on the App Configuration > [Git repository](../creating-application/git-material.md).


![](../../user-guide/global-configurations/images/git-account-enable-disable.jpg)

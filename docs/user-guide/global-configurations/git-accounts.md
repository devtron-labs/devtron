# Git Accounts

Git Accounts allow users to connect your code source with Devtron. Users be able to use these git accounts to build code using the CI pipeline.

## Git Account Configuration

`Global Configuration` helps to add a Git provider. Click on `Add git account` button at the top of the Git Account Section. To add a new git provider, add the details as mentioned below.

1. Name
2. Git Host
3. URL
4. Authentication type


![](../../user-guide/global-configurations/images/git-accounts.jpg)

### 1. Name

Provide a `Name` to your Git provider. This name will be displayed in the the Git Provider drop-down inside the Git Material configuration section.

### 2. Git Host

It is the git provider on which corresponding application git repository is hosted. By default users will get Bitbucket and GitHub but they can add many as you want clicking on **[+ Add Git Host]**.

### 3. URL

Provide the `URL`. **For example**- [https://github.com](https://github.com) for Github, [https://gitlab.com](https://gitlab.com) for GitLab, etc.

### 4. Authentication type

Here provide the type of authentication required by your version controller. Devtron support three types of authentications. Users can choose the one that suits the best.

* **Anonymous**

If authentication type is set as `Anonymous` then users do not need to provide any username, password/authentication token or SSH key. Just click on `Save` to save the git account provider details. If authentication type is set as `Anonymous`, only public git repository will be accessible.

![](../../user-guide/global-configurations/images/git-accounts-anonymous.jpg)

* **User Auth**

If users select `User Auth` then they have to provide the `Username` and either of `Password` or `Auth Token` for the authentication of your version controller account. Click on `Save` to save the git account provider details.

![](../../user-guide/global-configurations/images/git-accounts-user-auth.jpg)

* **SSH Key**

If users choose `SSH Key` then they have to provide the `Private SSH Key` corresponding to the public key added in their version controller account. Click on `Save` to save the git account provider details.

![](../../user-guide/global-configurations/images/git-accounts-ssh.jpg)

## Update Git Account

Users can update their saved git account settings at any point on time. They just need to click on the git account which they want to update. Make the required changes and click on `Update` to save the changes.

Updates can only be made within one Authentication type or one protocol type, i.e. HTTPS(Anonymous or User Auth) & SSH. Users can update from Anonymous to User Auth & vice versa, but not from Anonymous/User Auth to SSH or reverse.

![](../../user-guide/global-configurations/images/git-account-update.jpg)

### Note:

Users can enable and disable the git account settings. If they enable it, then they can see that enabled git account in the drop-down of [Git repository](../creating-application/git-material.md).

![](../../user-guide/global-configurations/images/git-account-enable-disable.jpg)

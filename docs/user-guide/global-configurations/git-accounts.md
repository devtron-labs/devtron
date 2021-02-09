# Git Accounts

Global configurations are used to configure your Git Providers, Docker Registry, Kubernetes clusters, different environments, User Management, and different Notifications for your application. We have segregated all the configurations and we will discuss how to configure all these one by one.

## Git Account Configuration

`Global Configuration` helps you to add a Git provider. Click on `Add git account` button at the top of the Git Account Section. To add a new git provider and provide three inputs as below.

1. Name
2. URL
3. Authentication type


![](../../.gitbook/assets/gc-add-gitaccount%20%284%29.png)

### 1. Name

Provide a `Name` to your Git provider. This will be added as one option in the Git Provider drop-down inside the Git Material section.

### 2. URL

Provide the `URL`. **For example**- github.com for Github, [https://gitlab.com](https://gitlab.com) for GitLab, etc.

### 3. Authentication type

Here you have to provide the type of Authentication required by your version controller. We support three types of authentications, You can choose the one that suits you the best.

* **Anonymous**

If you select `Anonymous` then you do not have to provide any username, password, and authentication token. Just click on `Save` to save your git account provider details.

![](../../.gitbook/assets/gc-git-account-configure%20%283%29.png)

* **Password/Auth token**

If you select Password/Auth token then you have to provide the `Access Token` for the authentication of your version controller account inside the Access token box. Click on `Save` to save your git account provider details.

![](../../.gitbook/assets/git_account_gc3%20%283%29.png)

* **User Auth**

If you choose `User auth` then you have to provide the `Username` and `Password` of your version controller account. Click on `Save` to save your git account provider details.

![](../../.gitbook/assets/gc-git-account-configure-user-auth.png)

## Update Git Account

You can update your saved git account settings at any point in time. Just click on the git account which you want to update. Make the required changes and click on `Update` to save you changes.

![](../../.gitbook/assets/git_account_gc5.png)

### Note:

You can enable and disable your git account setting. If you enable it, then you will be able to see that enabled git account in the drop-down of Git provider.

![](../../.gitbook/assets/gc-added-git-accounts%20%283%29.png)


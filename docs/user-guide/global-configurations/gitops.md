# Gitops

This feature in Global Configuration allows you to select Git account.


## Add Git Configuration

Select the Gitops section of global configuration. At the top of the section two git tabs are available.

* **GitHub**
* **GitLab**

![](../../.gitbook/assets/gc-gitops-tab.png)

Select one of the git tab. To add git you need to provide three inputs as below:
1. Git Host
2. Gitlab Group id / Github Organization id
3. Git access credential

### 1. Git Host: 

This field is filled by default, Showing url of selected tab. For example- https://github.com for Github, https://gitlab.com for GitLab

### 2. Github Organization Id:

Provide Git id to your account. In case of `Github` provide Github organization id. Similarly for `Gitlab`
provide Gitlab group id.

![](../../.gitbook/assets/gc-gitops-id.png)

### 3. Git access credential

Provide Git `Username` and `Token` of your git account. Click on Save to save your gitops configuration details.
 

![](../../.gitbook/assets/gc-gitops-save.png)


> Note: In case of any Invalid information, error will be shown while saving `Deployment Template`.

Learn more about [Deployment Template](https://docs.devtron.ai/user-guide/creating-application/deployment-template)
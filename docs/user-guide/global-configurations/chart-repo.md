# Chart Repository

This feature allows you to add more chart repositories to Devtron. Once added they will be available in the Discover section of the Chart Store. 


## Add Chart Repository

Select the Chart Repository section of global configuration and click on `Add Repository` button at the top of the Chart Repository Section to add new chart, you need to provide three inputs as below.

1. Name
2. URL
3. Authentication type

![](../../.gitbook/assets/gc-add-chart.png)

### 1. Name

Provide a `Name` to your Chart Repository. This name is added as prefix to the name of the chart in the listing on the helm chart section of application.

![](../../.gitbook/assets/gc-chart-name-highlight.png)

### 2. URL

Provide the `URL` of your Version Controller. **For example**- github.com for Github, [https://gitlab.com](https://gitlab.com) for GitLab, etc.

### 3. Authentication type

Here you have to provide the type of Authentication required by your version controller. We are supporting three types of authentications, you can choose any one from them.


* **Anonymous**

If you select `Anonymous` then you do not have to provide any username, password, and authentication token. Just click on `Save` to save your chart repository details.

![](../../.gitbook/assets/gc-chart-configure-anonymous.png)

* **Password/Auth token**

If you select Password/Auth token then you have to provide the `Access Token` for the authentication of your version controller account inside the Access token box. Click on `Save` to save your chart repository details.

![](../../.gitbook/assets/gc-chart-config-password.png)

* **User Auth**

If you choose `User Auth` then you have to provide the `Username` and `Password` of your version controller account. Click on `Save` to save your chart repository details.

![](../../.gitbook/assets/gc-chart-configure-user.png)

## Update Chart Repository

You can update your saved chart repository settings at any point in time. Click on the chart repository which you want to update. Make changes and click on `Update` to save you changes.

![](../../.gitbook/assets/gc-edit-chart.png)

### Note:

You can enable and disable your chart repository setting. If you enable it, then you will be able to see that enabled chart in `Discover` serction of `Chart Store`.

![](../../.gitbook/assets/gc-chart-list.png)
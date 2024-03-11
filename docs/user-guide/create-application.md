# Create a New Application

* On the Devtron dashboard, select **Applications**.
* On the upper-right corner of the screen, click **Create**.
* Select **Custom app** from the drop-down list.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/create-app-1.jpg)

A new application can be created from one of the following options:

* Custom App
* [From Chart Store](../user-guide/deploy-chart/README.md)


## Create Custom App

To create a new application from the custom app, select **Custom app**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/create-application.jpg)

* In the **Create application** window, enter an **App Name** and select a **Project**.
* Select either:<ul><li>**Create from scratch** to create an application from scratch, or<li>**Clone existing application** to clone an existing application.</ul></li>
* If you select **Create from scratch**, select the project from the drop-down list.<br>`Note`: You have to add [project under Global Configurations](./global-configurations/projects.md). Only then, it will appear in the drop-down list here.
* If you select **Clone existing application**, select an app you want to clone from and the project from the drop-down list.<br>`Note`: You have to add [project under Global Configurations](./global-configurations/projects.md). Only then, it will appear in the drop-down list here.</br>


## Tags

`Tags` are key-value pairs. You can add one or multiple tags in your application. 

**Propagate Tags** 
When tags are propagated, they are considered as labels to Kubernetes resources. Kubernetes offers integrated support for using these labels to query objects and perform bulk operations e.g., consolidated billing using labels. You can use these tags to filter/identify resources via CLI or in other Kubernetes tools.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/propagate-tags.jpg)

* Click `+ Add tag` to add a new tag.
* Click the symbol <img src="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/donot-propagate.jpg"  height="10"> on the left side of your tag to propagate a tag.<br>`Note`: Dark grey colour in symbol specifies that the tags are propagated.
* To remove the tags from propagation, click the symbol <img src="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/propagate-dark.jpg" height="10"> again.
* Click `Save`.



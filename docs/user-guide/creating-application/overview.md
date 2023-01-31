# Overview

The `Overview` section contains the brief information of the application, any added tags, configured external links and deployment details of the particular application. 
In this section, you can also [change project of your application](#change-project-of-your-application) and [manage tags](#manage-tags) if you added them while creating application.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/overview-latest.jpg)


The following details are provided on the **Overview** page:

| Fields | Description |
| :---    |     :---       |
| **App Name**  | Displays the name of the application. |
| **Created on** | Displays the day, date and time the application was created. |
| **Created by**  | Displays the email address of a user who created the application. |
| **Project**   | Displays the currect project of the application. You can change the project by selecting a different project from the drop-down list. |


## Change Project of your Application

You can change the project of your application by clicking **Project** on the `Overview` section.

1. Click `Project`. 
2. On the `Change project` dialog box, select the different project you want to change from the drop-down list.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/overview/change-project-app.jpg)


Click **Save**. The application will be moved to the selected project.

**Note**: If you change the project:
* The current users will lose the access to the application.
* The users who already have an access to the selected project, will get an access to the application automatically.

## Manage Tags

`Tags` are key-value pairs. You can add one or multiple tags in your application. When tags are propagated, they are considered as labels to Kubernetes resources. Kubernetes offers integrated support for using these labels to query objects and perform bulk operations e.g., consolidated billing using labels. You can use these tags to filter/identify resources via CLI or in other Kubernetes tools.

`Manage tags` is the central place where you can create, edit, and delete tags. You can also propagate tags as labels to Kubernetes resources for the application.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/manage-tags-latest.jpg)

* Click `Edit tags`.
* On the `Manage tags` page, click `+ Add tag` to add a new tag.
* Click `X` to delete a tag.
* Click the symbol <img src="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/donot-propagate.jpg" height="10"> on the left side of your tag to propagate a tag.<br>`Note`: Dark grey colour in symbol specifies that the tags are propagated.
* To remove the tags from propagation, click the symbol <img src="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/propagate-dark.jpg" height="10"> again.
* Click `Save`.

The changes in the tags will be reflected in the `Tags` on the `Overview` section.
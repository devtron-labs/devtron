# Overview

The `Overview` section contains the brief information of the job, any added tags, and deployment details of the particular job. 
In this section, you can also [change project of your job](#change-project-of-your-job) and [manage tags](#manage-tags) if you added them while creating job.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/create-job/overview-job.jpg)


The following details are provided on the **Overview** page:

| Fields | Description |
| :---    |     :---       |
| **Job Name**  | Displays the name of the job. |
| **Created on** | Displays the day, date and time the job was created. |
| **Created by**  | Displays the email address of a user who created the job. |
| **Project**   | Displays the current project of the job. You can change the project by selecting a different project from the drop-down list. |


## Change Project of your Job

You can change the project of your job by clicking **Project** on the `Overview` section.

1. Click `Project`. 
2. On the `Change project` dialog box, select the different project you want to change from the drop-down list.

Click **Save**. The job will be moved to the selected project.

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





 

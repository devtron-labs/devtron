# Create a New Job

* On the Devtron dashboard, select **Jobs**.
* On the upper-right corner of the screen, click **Create**.
* Select **Job** from the drop-down list.
* **Create job** page opens.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/create-job/select-create-job-latest.jpg)


## Create Job

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/create-job/create-job-page.jpg)


Provide below information on the `Create job` page:

| Fields | Description |
| --- | --- |
| **Job Name** | User-defined name for the job in Devtron. |
| **Description** | Enter the description of a job. |
| **Registry URL** | This is the URL of your private registry in Quay. E.g. `quay.io` |
| **Select one of them** |  <ul><li>**Create from scratch** :Select the project from the drop-down list.<br>`Note`: You have to add [project under Global Configurations](https://docs.devtron.ai/global-configurations/projects). Only then, it will appear in the drop-down list here.</li><li>**Clone existing application**: Select an app you want to clone from and the project from the drop-down list.<br>`Note`: You have to add [project under Global Configurations](https://docs.devtron.ai/global-configurations/projects). Only then, it will appear in the drop-down list here.</li></ul> |

**Note**: Do not forget to modify git repositories and corresponding branches to be used for each Job Pipeline if required.


### Tags

`Tags` are key-value pairs. You can add one or multiple tags in your application. 

**Propagate Tags** 
When tags are propagated, they are considered as labels to Kubernetes resources. Kubernetes offers integrated support for using these labels to query objects and perform bulk operations e.g., consolidated billing using labels. You can use these tags to filter/identify resources via CLI or in other Kubernetes tools.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/propagate-tags.jpg)

* Click `+ Add tag` to add a new tag.
* Click the symbol <img src="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/donot-propagate.jpg"  height="10"> on the left side of your tag to propagate a tag.<br>`Note`: Dark grey colour in symbol specifies that the tags are propagated.
* To remove the tags from propagation, click the symbol <img src="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/propagate-dark.jpg" height="10"> again.
* Click `Save`.



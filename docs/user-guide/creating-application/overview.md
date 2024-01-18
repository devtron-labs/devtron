# Overview

The `Overview` section contains the brief information of the application, any added tags, configured external links and deployment details of the particular application. 
In this section, you can also [change project of your application](#change-project-of-your-application) and [manage tags](#manage-tags) if you added them while creating application.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/overview/overview-latest-1.jpg)


The following details are provided on the **Overview** page:

| Fields | Description |
| :---    |     :---       |
| **App Name**  | Displays the name of the application. |
| **Created on** | Displays the day, date and time the application was created. |
| **Created by**  | Displays the email address of a user who created the application. |
| **Project**   | Displays the current project of the application. You can change the project by selecting a different project from the drop-down list. |


## Change Project of your Application

You can change the project of your application by clicking **Project** on the `Overview` section.

1. Click `Project`. 
2. On the `Change project` dialog box, select the different project you want to change from the drop-down list.

  ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/overview/change-project-app-1.jpg)

3. Click **Save**. The application will be moved to the selected project.

{% hint style="info" %}
### If you change the project:

The current users will lose the access to the application.

The users who already have an access to the selected project, will get an access to the application automatically.
{% endhint %}


## Manage Tags

`Tags` are key-value pairs. You can add one or multiple tags in your application. When tags are propagated, they are considered as labels to Kubernetes resources. Kubernetes offers integrated support for using these labels to query objects and perform bulk operations e.g., consolidated billing using labels. You can use these tags to filter/identify resources via CLI or in other Kubernetes tools.

`Manage tags` is the central place where you can create, edit, and delete tags. You can also propagate tags as labels to Kubernetes resources for the application.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/overview/manage-tags-latest-1.jpg)

* Click `Edit`.
* On the `Manage tags` page, click `+ Add tag` to add a new tag.
* Click `X` to delete a tag.
* Click the symbol <img src="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/donot-propagate.jpg" height="10"> on the left side of your tag to propagate a tag.<br>

{% hint style="info" %}
Dark grey colour in symbol specifies that the tags are propagated.
{% endhint %}

* To remove the tags from propagation, click the symbol <img src="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/propagate-dark.jpg" height="10"> again.

* Click `Save`.

The changes in the tags will be reflected in the `Tags` on the `Overview` section.


## Configure PersistentVolumeClaim (PVC) for Build Time Optimization

 A PersistentVolumeClaim (PVC) volume is a request for storage, which is used to mount a PersistentVolume (PV) into a Pod. In order to optimize build time, you can configure PVC in your application.

If you want to optimize build time for the multiple target platforms (e.g., arm64, amd64), mounting a PVC will provide volume directly to a pod which helps in shorter build time by storing build cache. Mounting a PVC into a pod will provide storage for build cache which will not impact the normal build where the image is built on the basis of architecture and operating system of the K8s node on which CI is running.

### Create PVC file

* The following configuration file describes persistent volume claim e.g.,`cache-pvc.yaml`, where you have to define the metadata `name` and `storageClassname`.

```bash
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: cache-pvc # here comes the name of PVC
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: # here comes storage class name
  resources:
    requests:
      storage: 30Gi
```

* Create the PersistentVolumeClaim by running the following command:

```bash
kubectl apply -f https://k8s.io/examples/pods/storage/pv-claim.yaml -n {namespace}
```

For more detail, refer [Kubernetes PVC](https://kubernetes.io/docs/tasks/configure-pod-container/configure-persistent-volume-storage/#create-a-persistentvolumeclaim).


## Configure PVC

In order to configure PVC:
* Go to the `Overview` section of your application.
* On the right-corner, click `Edit`.

  ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/overview/pvc-edit-tags-1.jpg)

* For app level PVC mounting, enter the following:<ul><li>key:`devtron.ai/ci-pvc-all`</li><li>value: metadata name (e.g., `cache-pvc)` which you define on the [PVC template](#create-pvc-file).</li></ul>`Note`: This PVC mounting will impact all the build pipelines of the application.

  ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/overview/manage-tags-pvc-1.jpg)

* For pipeline level, enter the following:<ul><li>key:`devtron.ai/ci-pvc-{pipelinename}`</li><li>value: metadata name which you define on the [PVC template](#create-pvc-file).</li></ul>`Note`: This PVC mounting will impact only the particular build pipeline.

To know the `pipelinename` detail, go to the `App Configuration`, click `Workflow Editor` the pipeline name will be on the `Build` pipeline as shown below.

  ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/overview/pipeline-name-pvc-1.jpg)

  ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/overview/pipeline-level-tag.jpg)

* Click `Save`.





 

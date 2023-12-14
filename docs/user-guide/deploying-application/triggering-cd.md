# Triggering CD Pipelines

After CI pipeline is complete, you can trigger the CD pipeline.

1. Go to the `Build & Deploy` tab of your application and click the `Select Image` button in the CD pipeline.

    ![Figure 1: Select Image Button](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/select-image.jpg)

2. Select an image to deploy and then click `Deploy` to trigger the CD pipeline.

    ![Figure 2: Selecting an Image](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/deploy.jpg)

The currently deployed images are tagged as `Active on <Environment name>`.

## Manual Approval for Deployment

When [manual approval is enabled](../creating-application/workflow/cd-pipeline.md#4-manual-approval-for-deployment) for the deployment pipeline configured in the workflow, you are expected to request for an image approval before each deployment. Alternatively, you can deploy images that have already been approved once.

When no approved image is available or if the image is already deployed, you will not see any image available for deployment upon clicking the `Select Image` button.

![Figure 3: No Approved Image](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/no-approved-image-v2.jpg)

### Request for Image Approval

Users need to have [Build & deploy permission](../user-guide/global-configurations/authorization/user-access.md#role-based-access-levels) or above (along with access to the environment and application) to request for an image approval.

To request an image approval, follow these steps:

1. Navigate to the `Build & Deploy` page, and click the **Approval for deployment** button.

    ![Figure 4: Approval Button](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/deployment-approval-button-v2.jpg)

2. Click the **Request Approval** button present on the image for which you want to request approval and the click **Submit Request**.

    ![Figure 5: Requesting Approval](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/request-approval-v2.jpg)

    In case you have configured [SES or SMTP on Devtron](../global-configurations/manage-notification.md#notification-configurations), you can directly choose the approver(s) from the list of approvers as shown below.

    ![Figure 6: Choosing Approvers](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/approver-list.jpg)

In case you wish to cancel the approval request, you can do so from the `Approval pending` tab as shown in the below image.

![Figure 7: 'Approval pending' tab](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/cancel-approval.jpg)


### Approve Image Approval Request

By default, super-admin users are considered as the default approvers. Users who build the image and/or request for its approval, cannot self-approve even if they have super-admin privileges.

Users with `Approver` permission for the specific application and environment can also approve a deployment. This permission can be granted to users from [`User Permissions`](../global-configurations/authorization/user-access.md#role-based-access-levels) present in [Global Configurations](../global-configurations/README.md).

In case [SES](../global-configurations/manage-notification.md#manage-ses-configurations) or [SMTP](../global-configurations/manage-notification.md#manage-smtp-configurations) is configured in Devtron, and the user has chosen the approvers while raising an image approval request, the approvers would receive an email notification as shown below:

![Figure 8: Email Notification to the Approver](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/email-notification.jpg)


To approve an image approval request, follow these steps:

1. Go to the `Build & Deploy` page and click the **Approval for deployment** button.

    ![Figure 9: Approval Button](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/deployment-approval-button-v2.jpg)

2. Switch to the `Approval pending` tab. Here, you will find all the images that are awaiting approval.

    ![Figure 10: List of Pending Approvals](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/approval-pending-tab.jpg)

3. Click **Approve** followed by **Approve Request** button.

    ![Figure 11: Approving a Request](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/approve-request-v2.jpg)

### Deploying Approved Image

Users need to have [Build & deploy permission](../user-guide/global-configurations/authorization/user-access.md#role-based-access-levels) or above (along with access to the environment and application) to select and deploy an approved image.

To deploy an approved image, follow these steps:

1. Navigate to the `Build & Deploy` tab and click the `Select Image` button. 

    ![Figure 12: Select Image Button](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/select-image.jpg)

2. You will find all the approved images listed under the `Approved images` section. From the list, you can select the desired image and deploy it to your environment.

    ![Figure 13: List of Approved Images](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/approved-images-v2.jpg)

3. The status of the current deployment can be viewed by clicking **App Details**. 

    ![Figure 14: 'App Details' Screen](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/app-status.jpg)

The status initially appears as `Progressing` for approximately 1-2 minutes, and then gradually transitions to `Healthy` state based on the deployment strategy.

Here, triggering CD pipeline is successful and the deployment is in `Healthy` state.

To further diagnose the deployments, [click here](../debugging-deployment-and-monitoring.md)


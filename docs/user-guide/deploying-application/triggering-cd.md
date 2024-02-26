# Triggering CD Pipelines

After CI pipeline is complete, CD pipeline can be triggered by clicking `Select Image`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/select-image.jpg)

Select an image to deploy and then click `Deploy` to trigger the CD pipeline.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/deploy.jpg)

The currently deployed images are tagged as `Active on <Environment name>`.

## Manual Approval for Deployment

When manual approval is enabled for the deployment pipeline, you are required to request image approval before each deployment. Alternatively, you can only deploy images that have already been approved once.

When no approved image is available or if the image is already deployed, you will not see any image available for deployment upon clicking the `Select Image` button.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/no-approved-image.jpg)

### Request For Image Approval

To request for image approval, please follow these steps:

1. Navigate to `Build & Deploy` page, and click on the `Approval for deployment` button.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/deployment-approval-button.jpg)

2. Click on `Request Approval` button present on the image for which you want to request approval and the click on `Submit Request`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/request-approval.jpg)

**NOTE**: You have the option to cancel the approval request from the `Approval Pending` section.

### Approve Image Approval Request

To approve an image approval request, please follow these steps:

1. Go to the `Build & Deploy` page and click the **Approval for deployment** button.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/deployment-approval-button.jpg)

2. Switch to the `Approval Pending` tab. Here, you will find all the images that are awaiting approval.

3. Click the **Approve** button, and then **Approve Request** to grant approval to the request.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/approve-request.png)

### Approvers for Deployment

By default, super admin users are considered as the default approvers. However, any user can be granted approver permission from the user-access section.

All users with `Approver` permission for the specific application and environment are authorized to approve deployments.

Please note that the user who built the image and/or requested approval cannot self-approve, even if they have super-admin privileges.

`Approver` role can be provided to users via [`User Permissions`](../global-configurations/authorization/user-access.md#role-based-access-levels)

After the images have been approved, navigate to the `Build and Deploy` tab and click on `Select Image`. You will find all the approved images listed under the `Approved Images` section. From this section, you can select the desired approved image and deploy it to the environment.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/approved-images.jpg)

The status of the current deployment can be viewed by clicking **App Details** that will show the `Progressing`state for 1-2 minutes and then gradually shows `Healthy` state, based on the deployment strategy.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/app-status.jpg)

Here, triggering CD pipeline is successful and the deployment is in "Healthy" state.

[To further diagnose deployments, Click here](../debugging-deployment-and-monitoring.md)


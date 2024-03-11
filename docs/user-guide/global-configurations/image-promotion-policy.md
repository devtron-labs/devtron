# Image Promotion Policy

## Introduction [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

An ideal deployment workflow may consist of multiple stages (e.g., SIT, UAT, Prod environment).

![Figure 1: Workflow on Devtron](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/image-promotion/sample-cd-workflow.jpg)

If you have built such a [workflow](../creating-application/workflow/README.md), your CI image will sequentially traverse and deploy to each environment until it reaches the target environment. However, if there's a critical issue you wish to address urgently (through a hotfix) on production, navigating the standard workflow might feel slow and cumbersome.

Therefore, Devtron offers a feature called 'Image Promotion Policy' that allows you to directly promote an image to the target environment, bypassing the intermediate stages in your workflow including:

* [Pre-CD](../creating-application/workflow/cd-pipeline.md#pre-deployment-stage) and [Post-CD](../creating-application/workflow/cd-pipeline.md#post-deployment-stage) of the intermediate stages
* All [approval nodes](../creating-application/workflow/cd-pipeline.md#manual-approval-for-deployment) of the intermediate stages

![Figure 2: Promoting an Image](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/image-promotion/image-promotion-visual.jpg)


---

## Creating an Image Promotion Policy

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to create an image promotion policy.
{% endhint %}

1. Go to **Global Configurations** → **Image Promotion Policy**.

2. Click **Create Policy** on the top-right.

3. Give a name to the policy and write a brief description, preferably explaining what it does.

4. Under **Image Filter Condition**, you can enter the conditions which your image promotion should be subjected to (e.g., *`branchName.startsWith('hotfix')`*) 

{% hint style="info" %}
Use **View filter criteria** to check the supported variables.
{% endhint %}  

5. You can specify either a pass condition, fail condition, or both conditions using [Common Expression Language (CEL)](https://github.com/google/cel-spec/blob/master/doc/langdef.md):
    * **Pass Condition**: Images that match this condition will be eligible for promotion to the target environment.
    * **Fail Condition**: Images that match this condition will NOT be eligible for promotion to the target environment.

{% hint style="info" %}
If an image matches both pass and fail conditions, the priority of the fail condition will be higher. Therefore, such image will NOT be eligible for promotion to the target environment.
{% endhint %}  

{% hint style="info" %}
If you don't define both pass and fail conditions, all images will be eligible for promotion.
{% endhint %}  

6. (Optional) If required, you can setup approval requirements for this policy. If **Approval for Image Promotion Policy** is enabled, an [approval will be required for an image]((#approving-image-promotion-request)) to be directly promoted to the target environment. Only the users having 'Image Promotion Approver' role (for the application and environment) will be able to approve the image promotion request.

 * **Number of Approvals (1-6)**: Specify the number of approvals required to promote an image. This can vary from one approval (minimum) to six approvals (maximum).

 * **Checkboxes for who can approve**: As a super-admin, you also have options to control the approval of image promotion and its deployment. These are available in the form of checkboxes.

    ![Figure 3: Controlling Approvals](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/image-promotion/control-approval.jpg)

6. Click **Save Changes**.

---

## Applying an Image Promotion Policy

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to apply an image promotion policy.
{% endhint %}

Here, you can decide the application(s) and environment(s) for which image promotion is allowed. 

1. Go to the **Apply Policy** tab.

2. Click the `Promotion Policy` dropdown next to the application, and choose the policy you wish to apply.

3. A confirmation dialog box would appear. Click **Confirm**.

### Performing Bulk Action

1. You can also apply a policy to multiple applications and environments ('App+Env' combinations) in bulk.

2. Simply use the checkboxes to select the desired application + environment combinations.

3. You will see a draggable floating widget. Click **Change Policy** in the widget and select a desired policy to be applied to all your selections.

Moreover, there are three filters available to make the selections easier for you:
* Application
* Environment
* Policy

---

## Result

### Promoting Image to Target Environment

{% hint style="warning" %}
### Who Can Perform This Action?
Users with build & deploy permission or above (for the application and target environment) can raise a promotion request (if the applied policy has 'Approval for Image Promotion Policy' enabled).
{% endhint %}

Here, you can promote images to the target environment(s).  

1. Go to the **Build & Deploy** tab of your application.

2. Click the **Promote** button next to the workflow in which the you wish to promote the image. Please note, the button will appear only if image promotion is allowed for any environment used in that workflow.

3. In the `Select Image` tab, you will see a list of images. However, you can use the **Show Images from** dropdown to decide the image you wish to promote. It can be an image either from the CI pipeline or an image that has passed from a particular environment in the workflow (the image must have passed all stages of deployment).

4. Use the **SELECT** button on the image, and click **Promote to...**

5. Select one or more target environments using the checkbox.

6. Click **Promote Image**. 

The image will get promoted to the target environment and it will be visible in the list of images in that target environment. However, if the super-admin enforced an approval mechanism, required number of approvals would be required for the image to be promoted.

{% hint style="warning" %}
In case you have configured [SES or SMTP on Devtron](../global-configurations/manage-notification.md#notification-configurations), an email notification will be sent to the approvers.
{% endhint %}

7. If approval(s) are required for image promotion, you may check the status of your request in the `Approval Pending` tab.

### Approving Image Promotion Request

{% hint style="warning" %}
### Who Can Perform This Action?
User needs to be a direct promotion approver or a super-admin to approve an image promotion request.
{% endhint %} 

1. Go to the **Build & Deploy** tab of your application.

2. Click the **Promote** button next to the workflow. The button will appear only if image promotion is allowed for any environment used in that workflow.

3. Go to the `Approval Pending` tab to see the list of images requiring approval. By default, it shows a list of all images whose promotion request is pending with you.  

{% hint style="info" %}
All the images will show the source from which it is being promoted, i.e., CI stage or intermediate stage (environment).
{% endhint %}

4. Click **Approve for...** to choose the target environments to which it can be promoted.

5. Click **Approve**.

You can also use the **Show requests** dropdown to filter the image promotion requests for a specific target environment.

### Deploying a Promoted Image

{% hint style="warning" %}
### Who Can Perform This Action?
Users with build & deploy permission or above for the application and environment can deploy the promoted image.
{% endhint %}

If a user has approved the promotion request for an image, they may or may not be able to deploy depending upon the [policy configuration](#creating-an-image-promotion-policy).

However, a promoted image does not automatically qualify as a deployable image. It must fulfill all configured requirements ([Image Deployment Approval](../creating-application/workflow/cd-pipeline.md#manual-approval-for-deployment), [Filter Conditions](./filter-condition.md), etc.) of the target environment for it to be deployed.

You can check the deployment of promoted images in the **Deployment History** of your application.
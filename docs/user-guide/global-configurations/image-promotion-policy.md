# Image Promotion Policy

## Introduction [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

An ideal CD pipeline may consist of multiple stages (e.g., SIT, UAT, Prod environment). If you have built such a [workflow](../creating-application/workflow/README.md), your CI image will sequentially traverse and deploy to each environment until it reaches the target environment. However, if there's a critical issue you wish to address urgently (through a hotfix) on production, navigating the standard workflow might feel slow and cumbersome.

Therefore, Devtron offers a feature called 'Image Promotion Policy' that allows you to directly promote an image to the target environment, bypassing the intermediate stages in your pipeline including:

* [Pre-CD](../creating-application/workflow/cd-pipeline.md#pre-deployment-stage) and [Post-CD](../creating-application/workflow/cd-pipeline.md#post-deployment-stage) of the intermediate stages
* All [approval nodes](../creating-application/workflow/cd-pipeline.md#manual-approval-for-deployment) of the intermediate stages

---

## Creating an Image Promotion Policy

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to create an image promotion policy.
{% endhint %}

1. Go to **Global Configurations** → **Image Promotion Policy**.

2. Click **Create Policy** on the top-right.

3. Give a name to the policy and write a brief description, preferably explaining what it does.

4. Under **Image Filter Condition**, you can specify either a pass condition, fail condition, or both conditions using [Common Expression Language (CEL)](https://github.com/google/cel-spec/blob/master/doc/langdef.md):
    * **Pass Condition**: Only those images that satisfy your condition will be promoted and displayed in the target environment.
    * **Fail Condition**: Images that fail the condition will not be directed to the target environment.

    **Example**: *`branchName.startsWith('hotfix')`*

{% hint style="info" %}
Use **View filter criteria** to check the supported variables.
{% endhint %}  

5. (Optional) If required, you can also introduce an approval flow using the toggle button: **Approval for Image Promotion Policy**. As a result, the [user selecting an image for promotion](#selecting-image-for-promotion) will [require approval](#approving-an-image-promotion-request).

 * **Number of Approvals (1-6)**: Specify the number of approvals required to promote an image. This can vary from one approval (minimum) to six approvals (maximum).

 * **Checkboxes for who can approve**: By default, users who [request an image promotion](#selecting-image-for-promotion) cannot approve their own request. The same is true for those who trigger the build. Whereas, the ones who [approve an image promotion request](#approving-an-image-promotion-request) are barred from deploying the promoted image. However, as a super-admin, you can use the checkboxes to remove such restrictions for better control.

6. Click **Save Changes**.

---

## Applying an Image Promotion Policy

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to apply an image promotion policy.
{% endhint %}

Here, you can decide the application(s) and environment(s) for which image promotion is allowed. 

1. Go to the **Apply Policy** tab.

2. Use the checkbox to select an application.

3. Click the `Promotion Policy` dropdown within it and choose the policy you wish to apply.

4. A confirmation dialog box would appear. Click **Confirm**.

### Performing Bulk Action

You can also apply a policy to more than one application. Simply use the checkboxes to select the applications. You can do this even if there are many applications spanning multiple pages. You will see a draggable floating widget.

Moreover, there are three filters available to make it easier for you to make the selections easier:
* Application
* Environment
* Policy

---

## Result

### Selecting Image for Promotion

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have build & deploy permission or above (along with access to the application and target environment) to select an image for promotion.
{% endhint %}

Here, you can promote images to the target environment(s).  

1. Go to the **Build & Deploy** tab of your application.

2. Click the **Promote** button next to the pipeline eligible for image promotion.

3. In the `Select Image` tab, you will see a list of images. However, you can use the **Show Images from** dropdown to decide the image you wish to promote. It can be an image either from the CI pipeline or from any one of the intermediate stages (environments).

4. Use the **SELECT** button on the image, and click **Promote to...**

5. Select one or more target environments using the checkbox.

6. Click **Promote Image**. 

The image will get promoted to the target environment and it will be visible in the list of images in that target environment. However, if the super-admin enforced an approval mechanism, an email notification will be sent to the approvers. 

You may check the status of your request in the `Approval Pending` tab.

### Approving an Image Promotion Request

{% hint style="warning" %}
### Who Can Perform This Action?
User needs to be a direct promotion approver or a super-admin to approve an image promotion request.
{% endhint %} 

1. Go to the **Build & Deploy** tab of your application.

2. Click the **Promote** button next to the pipeline that has direct promotion enabled.

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
Users need to have build & deploy permission or above (along with access to the application and target environment) to deploy an image.
{% endhint %}

Once an image is successfully promoted to the target environment, you may deploy it. 

However, a promoted image does not automatically qualify as a deployable image. If it is not visible in the list of eligible images for deployment, you might have a [Filter Condition defined at a global level]((./filter-condition.md)) that is blocking the deployment.

You can check the deployment of promoted images in the **Deployment History** of your application.
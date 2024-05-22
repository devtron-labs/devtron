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

You can create a policy using our APIs or through Devtron CLI. To get the latest version of the **devtctl** binary, please contact your enterprise POC or reach out to us directly for further assistance.

Here is the CLI approach:

**Syntax**:
```
devtctl create imagePromotionPolicy \
    --name="example-policy" \
    --description="This is a sample policy that promotes an image to production environment" \
    --passCondition="true" \
    --failCondition="false" \
    --approverCount=0 \
    --allowRequestFromApprove=false \
    --allowImageBuilderFromApprove=false \
    --allowApproverFromDeploy=false \
    --applyPath="path/to/applyPolicy.yaml"
```

**Arguments**:

* `--name` (required): The name of the image promotion policy.
* `--description` (optional): A brief description of the policy, preferably explaining what it does.
* `--passCondition` (optional): Specify a condition using [Common Expression Language (CEL)](https://github.com/google/cel-spec/blob/master/doc/langdef.md). Images that match this condition will be eligible for promotion to the target environment.
* `--failCondition` (optional): Images that match this condition will NOT be eligible for promotion to the target environment.
* `--approverCount` (optional): The number of approvals required to promote an image (0-6). Defaults to 0 (no approvals).
* `--allowRequestFromApprove` (optional): (Boolean) If true, user who raised the image promotion request can approve it. Defaults to false.
* `--allowImageBuilderFromApprove` (optional): (Boolean) If true, user who triggered the build can approve the image promotion request. Defaults to false.
* `--allowApproverFromDeploy` (optional): (Boolean) If true, user who approved the image promotion request can deploy that image. Defaults to false.
* `--applyPath` (optional): Specify the path to the YAML file that contains the list of applications and environments to which the policy should be applicable.

{% hint style="info" %}
If an image matches both pass and fail conditions, the priority of the fail condition will be higher. Therefore, such image will NOT be eligible for promotion to the target environment.
{% endhint %}  

{% hint style="info" %}
If you don't define both pass and fail conditions, all images will be eligible for promotion.
{% endhint %}  

---


<!-- {% hint style="warning" %}
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

6. (Optional) If required, you can setup approval requirements for this policy. If **Approval for Image Promotion Policy** is enabled, an [approval will be required for an image]((#approving-image-promotion-request)) to be directly promoted to the target environment. Only the users having 'Artifact promoter' role (for the application and environment) will be able to approve the image promotion request.

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

--- -->

## Applying an Image Promotion Policy

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to apply an image promotion policy.
{% endhint %}

You can apply a policy using our APIs or through Devtron CLI. Here is the CLI approach:

* Create a YAML file and give it a name (say `applyPolicy.yaml`). Within the file, define the applications and environments to which the image promotion policy should apply, as shown below.

    {% code title="applyPolicy.yaml" overflow="wrap" lineNumbers="true" %}
    ```yaml
    apiVersion: v1
    kind: artifactPromotionPolicy
    spec:
      payload:
        applicationEnvironments:
        - appName: "app1"
            envName: "env-demo"
        - appName: "app1"
            envName: "env-staging"
        - appName: "app2"
            envName: "env-demo"
        applyToPolicyName: "example-policy"
    ```
    {% endcode %}

    Here, `applicationEnvironments` is a dictionary that contains the application names (app1, app2) and the corresponding environment names (env-demo/env-staging) where the policy will apply. In the `applyToPolicyName` key, enter the value of the `name` argument you used earlier while [creating the policy](#creating-an-image-promotion-policy).

* Apply the policy using the following CLI command:

    ```
    devtctl apply policy -p="path/to/applyPolicy.yaml"
    ```


## Result

### Promoting Image to Target Environment

{% hint style="warning" %}
### Who Can Perform This Action?
Users with build & deploy permission or above (for the application and target environment) can promote an image if the image promotion policy is enabled.
{% endhint %}

Here, you can promote images to the target environment(s).  

1. Go to the **Build & Deploy** tab of your application.

2. Click the **Promote** button next to the workflow in which the you wish to promote the image. Please note, the button will appear only if image promotion is allowed for any environment used in that workflow.

    ![Figure 3: Promote Button](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/image-promotion/promote-button.jpg)

3. In the `Select Image` tab, you will see a list of images. Use the **Show Images from** dropdown to filter the list and choose the image you wish to promote. This can be either be an image from the CI pipeline or one that has successfully passed all stages (e.g., pre, post, if any) of that particular environment.

    ![Figure 4: Selecting an Image](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/image-promotion/show-images.jpg)

4. Use the **SELECT** button on the image, and click **Promote to...**

5. Select one or more target environments using the checkbox.

    ![Figure 5: Selecting the Destination Environment](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/image-promotion/selecting-env.jpg)

6. Click **Promote Image**. 

The image's promotion to the target environment now depends on the approval settings in the image promotion policy. If the super-admin has enforced an approval process, the image requires the necessary number of approvals before promotion. On the other hand, if the super-admin has not enforced approval, the image will be automatically promoted since there is no request phase involved.

{% hint style="warning" %}
In case you have configured [SES or SMTP on Devtron](../global-configurations/manage-notification.md#notification-configurations), an email notification will be sent to the approvers.
{% endhint %}

7. If approval(s) are required for image promotion, you may check the status of your request in the `Approval Pending` tab.

### Approving Image Promotion Request

{% hint style="warning" %}
### Who Can Perform This Action?
Only the users having [Artifact promoter](./user-access.md#role-based-access-levels) role (for the application and environment) or superadmin permissions will be able to approve the image promotion request.
{% endhint %} 

1. Go to the **Build & Deploy** tab of your application.

2. Click the **Promote** button next to the workflow.

3. Go to the `Approval Pending` tab to see the list of images requiring approval. By default, it shows a list of all images whose promotion request is pending with you. 

    ![Figure 6: Checking Pending Approvals](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/image-promotion/pending-approvals.jpg)

{% hint style="info" %}
All the images will show the source from which it is being promoted, i.e., CI stage or intermediate stage (environment).
{% endhint %}

4. Click **Approve for...** to choose the target environments to which it can be promoted.

5. Click **Approve**.

You can also use the **Show requests** dropdown to filter the image promotion requests for a specific target environment.

![Figure 7: Show Env-specific Promotion Requests](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/image-promotion/show-requests.jpg)

If there are pending promotion requests, you can approve them as shown below:

![Figure 8: Approving Image Promotion Requests](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/image-promotion/image-promo-approval.gif)

### Deploying a Promoted Image

{% hint style="warning" %}
### Who Can Perform This Action?
Users with build & deploy permission or above for the application and environment can deploy the promoted image.
{% endhint %}

If a user has approved the promotion request for an image, they may or may not be able to deploy depending upon the [policy configuration](#creating-an-image-promotion-policy).

However, a promoted image does not automatically qualify as a deployable image. It must fulfill all configured requirements ([Image Deployment Approval](../creating-application/workflow/cd-pipeline.md#manual-approval-for-deployment), [Filter Conditions](./filter-condition.md), etc.) of the target environment for it to be deployed.

In the **Build & Deploy** tab of your application, click **Select Image** for the CD pipeline, and choose your promoted image for deployment.

![Figure 9: Deploying Promoted Image](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/image-promotion/deploying-promoted-image.jpg)

You can check the deployment of promoted images in the **Deployment History** of your application. It will also indicate the pipeline from which the image was promoted and deployed to the target environment.

![Figure 10: Deployment History - Checking Image Source](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/image-promotion/promoted-image-deploy-log.jpg)
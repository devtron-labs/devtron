# Applying Labels and Comments

## Introduction

Typically in a CI pipeline, you [build container images](./triggering-ci.md), and the number of images gradually increases over a period of time. Devtron's image labels and comments feature helps you to mark and recall specific images from the repository by allowing you to add special instructions or notes to them. 

For example:
* You can label an image as `non-prod` to indicate that it is meant for 'Dev' or 'QA' environments, but not for production.
* Add `hotfix image only` label to indicate a one-time patch on production.
* Comments like `This image is buggy and shouldn't be used for deployment` to caution other users from deploying an unwanted image.

![Figure 1: Labels and Comments](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/tag-comment/tag-and-comment.jpg)

Such labels and comments will be visible only within Devtron, and will not propagate to your [container registry](../../reference/glossary.md#containeroci-registry) (say Docker Hub), unlike custom [image tag pattern](../creating-application/workflow/ci-pipeline.md#custom-image-tag-pattern). You may use it to simplify the management and [selection of container images](./triggering-cd.md#deploying-approved-image) for deployment.

{% hint style="warning" %}
Tagging labels and comments are supported only for images in workflows with at least one production deployment pipeline. In Devtron, you can go to **Global Configurations** → **Clusters & Environments** to identify a production environment by checking the 'Prod' label.
{% endhint %}

---

## Adding Labels & Comments

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have [Build & deploy permission](../global-configurations/authorization/user-access.md#role-based-access-levels) or above (along with access to the environment and application) to add labels and comments.
{% endhint %}

You can add labels and comments from the following pages:

* [From Build & Deploy](#from-build--deploy)
* [From Build History](#from-build-history)
* [From Deployment History](#from-deployment-history) (only after deployment)
* [From App Details](#from-app-details) (only after deployment)

{% hint style="warning" %}
You can add multiple labels to an image. but each label can be used only once 'per image, per application'. You may use it in an image of other application though. <br />
Refer [Deleting Labels](#deleting-labels-and-comments) if you commit a mistake while adding labels.
{% endhint %}
 
### From Build & Deploy

![Figure 2: Adding Labels and Comments - 'Build & Deploy' Page](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/tag-comment/tag-build.gif)

### From Build History

![Figure 3: Adding Labels and Comments - 'Build History' Page](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/tag-comment/tag-build-history.gif)

### From Deployment History

![Figure 4: Adding Labels and Comments - 'Deployment History' Page](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/tag-comment/tag-deployment-history.gif)

### From App Details

![Figure 5: Adding Labels and Comments - 'App Details' Page](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/tag-comment/tag-app-details.gif)

---

## Deleting Labels & Comments

### Soft-Delete Labels

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have [Build & deploy permission](../global-configurations/authorization/user-access.md#role-based-access-levels) or above (along with access to the environment and application) to perform soft deletion of labels.
{% endhint %}

This action marks the label as invalid but doesn't delete the label. Therefore, you can recover it again but you cannot reuse it for other image (unless it's a different application).

1. Click the edit option.
2. Use the (-) icon to strike off the label. This icon is available on the left-side of a label.
3. Click **Save**. 

![Figure 6: Soft Deletion of a Label](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/tag-comment/soft-delete-tag.gif)

### Hard-Delete Labels

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to perform hard deletion of labels.
{% endhint %}

This action deletes the label permanently and makes it available for reuse in same/other image of the given application.

1. Click the edit option.
2. Use the (x) icon to permanently remove the label. This icon is available on the right-side of a label.
3. Click **Save**.

![Figure 7: Hard Deletion of a Label](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/tag-comment/hard-delete-tag.gif)

### Removing Comments

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have [Build & deploy permission](../global-configurations/authorization/user-access.md#role-based-access-levels) or above (along with access to the environment and application) to remove comments.
{% endhint %}

If you wish to permanently remove a comment, do the following:

1. Click the edit option.
2. Empty the content of an existing comment.
3. Click **Save**.

![Figure 8: Removing a Comment](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/tag-comment/remove-comment.gif)

---

## Extra Use Case

If you use [Application Groups](../application-groups.md) to deploy in bulk, image labels (if added) will be available as filters for you to quickly locate the container image.

![Figure 9: Application Groups - Filter by Image Label](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/tag-comment/ag-image-filter.gif)

This will be helpful in scenarios (say release package) where you wish to deploy multiple applications at once, and you have already labelled the intended images of the respective applications.


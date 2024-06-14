# Applying Labels and Comments

## Introduction

Typically in a CI pipeline, you build [container images](../../reference/glossary.md#image), and the number of images gradually increases over a period of time.

Assume `azurecr.io/web-server` is your container [image repository](../../reference/glossary.md#repo) and it has multiple images. Now, differentiating one `web-server` image from another or remembering a specific `web-server` image from the repository might become a tedious task. 

Devtron provides you an option to add image labels and comments to your container images for easy identification. These data will remain only on Devtron, and will not propagate to the container registry, unlike [custom image tag pattern](../creating-application/workflow/ci-pipeline.md#custom-image-tag-pattern).

![Figure 1: Labels and Comments](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/tag-comment/tag-and-comment.jpg)

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
You can add multiple labels to an image. Once added, they cannot be used for another image. <br />
Refer [Deleting Labels](#deleting-labels--comments) if you commit a mistake while adding labels.
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

This action marks the label as invalid but doesn't delete the label. Therefore, you can recover it again.

1. Click the edit option.
2. Use the (-) icon to strike off the label. This icon is available on the left-side of a label.
3. Click **Save**. 

![Figure 6: Soft Deletion of a Label](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/tag-comment/soft-delete-tag.gif)

### Hard-Delete Labels

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to perform hard deletion of labels.
{% endhint %}

This action deletes the label and makes it available for reuse in same or any other image.

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


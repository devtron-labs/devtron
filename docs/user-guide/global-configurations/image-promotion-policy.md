# Image Promotion Policy

## Introduction [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

An ideal CD pipeline may consist of multiple stages (e.g., SIT, UAT, Prod environment). If you have built such a workflow, your CI image will sequentially traverse and deploy to each environment until it reaches the target environment (i.e. production). However, if there's a critical issue you wish to address urgently on production, navigating the standard workflow might feel slow and cumbersome.

Therefore, Devtron offers a feature called 'Image Promotion Policy' that grants you the ability to directly promote an image to the target environment, bypassing the intermediate stages in your pipeline.

---

## Creating an Image Promotion Policy

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to create an image promotion policy.
{% endhint %}

---

## Applying an Image Promotion Policy

### Selecting Image for Promotion

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have build & deploy permission or above (along with access to the application and target environment) to select an image for promotion.
{% endhint %}

### Approving an Image Promotion

{% hint style="warning" %}
### Who Can Perform This Action?
User needs to be an hotfix-deployment approver or super-admin to approve an image promotion policy.
{% endhint %}

---

## Viewing Deployed Promoted Image
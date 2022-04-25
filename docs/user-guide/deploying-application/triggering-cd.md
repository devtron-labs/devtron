# Triggering CD

## Triggering CD Pipelines

After CI pipeline is complete, CD pipeline can be triggered by clicking on `Select Image`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/select-image.jpg)

Select an image to deploy and then click on `Deploy` to trigger the CD pipeline.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/deploy.jpg)

The current deployed images are tagged as `Deployed on <Environment name>`.

The status of the current deployment can be viewed by Clicking on **App Details** that will show the `Progressing`state for 1-2 minutes and then gradually shows `Healthy` state or `Hibernating` state, based on the deployment strategy.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/triggering-cd/app-status.jpg)

Here, triggering CD pipeline is successful and the deployment is in "Healthy" state.

[To further diagnose deployments, Click here](../debugging-deployment-and-monitoring.md)


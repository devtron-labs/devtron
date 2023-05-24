# Workflow

Workflow is a logical sequence of different stages used for continuous integration and continuous deployment of an application.

![](../../../.gitbook/assets/arora1%20%282%29.gif)

Click on `New Build Pipeline` to create a new workflow

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow/new-build-pipeline-1.jpg)

On clicking `New Build Pipeline`, three options appear as mentioned below:

* **Continuous Integration:** Choose this option if you want Devtron to build the image of source code.
* **Linked CI Pipeline:** Choose this option if you want to use an image created by an existing CI pipeline in Devtron.
* **Incoming Webhook:** Choose this if you want to build your image outside Devtron, it will receive a docker image from an external source via the incoming webhook.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow/workflow-ci.jpg)

Then, create CI/CD Pipelines for your application.

To know how to create the CI pipeline for your application, click on: [Create CI Pipelines](ci-pipeline.md)

To know how to create the CD pipeline for your application, click on: [Create CD Pipelines](cd-pipeline.md)




# Base Deployment Template


A deployment configuration is a manifest of the application. It defines the runtime behavior of the application.
You can select one of the default deployment charts or custom deployment charts which are created by super admin.

To configure a deployment chart for your application, do the following steps:

* Go to **Applications** and create a new application.
* Go to **App Configuration** page and configure your application.
* On the **Base Deployment Template** page, select the drop-down under **Chart type**.


## Select chart from Default Charts

You can select a default deployment chart from the following options:

1. [Deployment](https://docs.devtron.ai/usage/applications/creating-application/deployment-template/deployment) (Recommended)
1. [Rollout Deployment](deployment-template/rollout-deployment.md)
2. [Job & CronJob](deployment-template/job-and-cronjob.md)
3. Knative


![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/deployment-template/deployment-chart.png)


## Select chart from Custom Charts

Custom charts are added by users with `super admin` permission from the [Custom charts](../global-configurations/custom-charts.md) section.

You can select the available custom charts from the drop-down list. You can also view the description of the custom charts in the list.

![Select custom chart](https://devtron-public-asset.s3.us-east-2.amazonaws.com/custom-charts/use-custom-chart.png)

### Upload Custom Chart

A [custom chart](../global-configurations/custom-charts.md) can be uploaded by a super admin.

## Application Metrics

Enable **show application metrics** toggle to view the application metrics on the **App Details** page.

![Show application metrics](https://devtron-public-asset.s3.us-east-2.amazonaws.com/custom-charts/show-application-metrics.png)

> **IMPORTANT**: Enabling Application metrics introduces a sidecar container to your main container which may require some additional configuration adjustments. We recommend you to do load test after enabling it in a non-production environment before enabling it in production environment.

Select **Save & Next** to save your configurations.

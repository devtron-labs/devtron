# Devtron Charts

A deployment configuration is a manifest for the application. It defines the runtime behavior of the application.

Devtron includes deployment template for both default as well as custom charts created by a super admin.

To configure a deployment chart for your application:

* Go to **Applications** and create a new application.
* Go to **App Configuration** page and configure your application.
* On the **Deployment Template** page, select the drop-down under **Chart type**.

You can select a chart in one of the following ways:

## Select chart by Devtron

1. [Rollout Deployment](deployment-template/rollout-deployment.md)
2. [Job and Cronjob](deployment-template/job-and-cronjob.md)
3. Knative

## Select chart from custom charts

Custom charts are added by a super admin from the [custom charts](../global-configurations/custom-charts.md) section.

Users can select the available custom charts from the drop-down list.

![Select custom chart](https://devtron-public-asset.s3.us-east-2.amazonaws.com/custom-charts/use-custom-chart.png)

### Upload custom chart

A [custom chart](../global-configurations/custom-charts.md) can be uploaded by a super admin.

## Application Metrics

Enable **show application metrics** toggle to view the application metrics on the **App Details** page.

![Show application metrics](https://devtron-public-asset.s3.us-east-2.amazonaws.com/custom-charts/show-application-metrics.png)

> **IMPORTANT**: Enabling Application metrics introduces a sidecar container to your main container which may require some additional configuration adjustments, we recommend you to do load test after enabling it in a non-prod environment before enabling it in production environment.

Select **Save** to save your configurations.
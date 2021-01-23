# Deploying MongoDB Helm Chart

## Introduction

Let's assume that you are building an application which needs mongoDB.

![](../../../.gitbook/assets/mongo%20%282%29.jpg)

Deploying applications as Helm Charts is the easiest way to create applications on Devtron.

This guide will introduce you to how to deploy the mongoDB's Helm chart.

## 1. Discover the Chart from the Chart Store

Select the `Charts` section from the left pane, you will be landed to the `Chart Store` page. Click on `Discover` and find `stable/mongodb-replicaset` Helm Chart.

![](../../../.gitbook/assets/first%20%281%29.jpg)

## 2. Configure the Chart

After selecting the stable/mongodb helm chart, click on `Deploy`

![](../../../.gitbook/assets/second%20%282%29.jpg)

Enter the following details, to deploy mongoDB chart:

| Key | Description |
| :--- | :--- |
| `App Name` | Name of the Chart |
| `Project` | Select the name of your Project in which you want to deploy the chart |
| `Environment` | Select the environment in which you want to deploy the chart |
| `Chart Version` | Select the latest Chart Version |
| `Chart Value` | Select the latest default value or create a custom value |

### Configure `values.yaml`

In this example `replicas` is set to **1** and `persistenceVolume` is set to **false**. You can configure it according to your project's requirements.

To learn about different parameters used in the chart, you can check [Documentation of mongodb Helm chart](https://hub.helm.sh/charts/bitnami/mongodb)

![](../../../.gitbook/assets/mongo1%20%282%29.jpg)

Click on `Deploy` after you have finished configuring the chart.

## 3. Check the Status of Deployment

After clicking on `Deploy` you will land on a page, that shows the Status of the deployment of the Chart.

The Status of the chart should be `Healthy`. It might take few seconds after initiating the deployment of the chart.

![](../../../.gitbook/assets/mongo4%20%281%29.png)

In case the Status of the deployment is `Degraded` or takes a long time to get deployed.

Click on the `Status` or check the logs of the pods to debug the issue.

## 4. Extract the Service name

Copy the service name, it will be used to connect your application to mongoDB.

![](../../../.gitbook/assets/mongo6%20%281%29.png)


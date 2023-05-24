# Chart Operations 

## Introduction

Discover, Create, Deploy, Update, Upgrade, Delete charts.

## 1. Discover the chart from the Chart Store

Select the `Charts` section from the left pane, you will be landed to the `Chart Store` page. 
Search `nginx` or any other charts in search filter.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/deployment-of-charts/charts-1.jpg)

Click on chart and it will redirect you to `Chart Details` page where you can see a number of instances deployed by using the same chart.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/deployment-of-charts/charts-2.jpg)

## 2. Configure the Chart

After selecting the version and values, click on `Deploy`

Enter the following details, to deploy chart:

| Key | Description |
| :--- | :--- |
| `App Name` | Name of the Chart Unique|
| `Project` |  Project in which you want to deploy the chart |
| `Environment` | Environment in which you want to deploy the chart |
| `Chart Version` | Chart version |
| `Chart Value` | Latest default value or create a custom value |

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/deployment-of-charts/charts-3.jpg)

you can choose any chart version, values and update it on values.yaml

Click on `Deploy` to deploy the Chart

## 3. Status of Deployment 

After clicking on `Deploy` you will land on a page that shows the status of the deployment of the Chart.

The status of the chart should be `Healthy`. It might take a few seconds after initiating the deployment of the chart.
In case the status of the deployment is `Degraded` or takes a long time to get deployed, click on `Details` in `Application Status` section on the same page or check the logs of the pods to debug the issue.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/deployment-of-charts/charts-4-2.jpg)

1. Shows status of deployed chart.

2. Shows the controller service accounts being used.

3. Clicking on `values` will land you on the page where you can update, upgrade or delete chart.

4. Clicking on `View Chart` will land you to the page where you can see all the running instances of this chart.

To see deployment history of Helm application, click on `Deployment history` from `App details` page.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/deploy-charts/overview-of-charts/overview-of-charts-9.jpg)


## 4. Update or Upgrade Chart

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/deployment-of-charts/charts-5-2.jpg)

For update you can change its `chart version` or `values.yaml` and then click on `Update And Deploy`.

For upgrade click on `Repo/Chart` field and search any chart name like `nginx-ingress` and change values corresponding to that chart and Click on `Update And Deploy`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/deployment-of-charts/charts-6-2.jpg)


After an update or upgrade you again will land on `App Detail` page, where you can check pods and service name.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/deployment-of-charts/charts-7-2.jpg)

## 5. Chart Details and Delete Charts

By clicking on `View Chart` in `Chart Used` section on `App Details` page, it will redirect you to `Chart Details` page where you can see number of instances installed by this chart and also you can delete the chart instance from here.


![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/deployment-of-charts/charts-8-2.jpg)

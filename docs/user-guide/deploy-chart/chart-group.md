# Chart Operations 

## Introduction

Using Devtron UI, one or more Helm charts can be grouped and deployed together with a single click.

## 1. Create Group 

1. In the left pane, select `Charts`.
2. On the `Chart Store` page, select `Create Group` from the upper-right corner.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/chart-group/chart-group-1.jpg)

3. In the `Create Chart Group` screen, enter `name` and `description`(optional) for the chart group, and then select `Create Group`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/chart-group/chart-group-2.jpg)

Once you create the group, you can now select and add the charts to this chart group.

## 2. Add Charts To Group 

1. To add a chart to the group, click the `+` sign at the top-right corner of a chart, and then select `Save`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/chart-group/chart-group-3.jpg)

2. Click on `Group Detail` to see all the running instances and group details. You can also edit the chart group from here.

## 3.Bulk Deploy and Edit Option for Charts

You can see all the charts in the chart group in the right panel. 
1. Select `Deploy to..`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/chart-group/chart-group-4.jpg)

2. In the `Deploy Selected Charts`, select the `Project` and `Deploy to Environment` values where you want to deploy the chart group.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/chart-group/chart-group-5.jpg)

3. Select `Advanced Options` for more deploy options, such as editing the `values.yaml` or changing the `Environment` and `Project` for each chart.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/chart-group/chart-group-6.jpg)

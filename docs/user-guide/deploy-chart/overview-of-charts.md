# Overview Of Charts

## Deploying Charts

Charts can be deployed individually or by creating a group of Charts. Both methods are mentioned here.

### Deploying Chart

To deploy any chart or chart group, visit the `Charts` section from the left panel and then select the chart that you want to use.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/overview-of-charts/overview-of-charts-1.jpg)

Click on `README.md` to get more ideas about the configurations of the chart.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/overview-of-charts/overview-of-charts-2.jpg)

Select the Chart Version that you want to use and Chart Value, you can either use the Default Values or Custom Values.

To know about Custom Values, Click On: [Custom Values](overview-of-charts.md#custom-values)

The configuration values can be edited in the section given below Chart Version.

| Key | Description |
| :--- | :--- |
| `App Name` | Name of the app |
| `Project` | Project of the app |
| `Environment` | Environment of the app to be deployed in |
| `Chart Version` | Version of the chart to be used |

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/overview-of-charts/overview-of-charts-3.jpg)

Readme.md present on the left can be used by the user to set configuration values.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/overview-of-charts/overview-of-charts-4.jpg)

Click on `Deploy Chart` to deploy the chart.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/overview-of-charts/overview-of-charts-5-2.jpg)

Click on `App Details` to see the status and details of the deployed chart and click on `Values` to reconfigure the deployment.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/overview-of-charts/overview-of-charts-6-2.jpg)

Configuration values can be edited over here by the help of Readme.md.

Click on `Update And Deploy` to update new settings.
You can also see deployment history of Helm application and values.yaml corresponding to the deployment by clicking on `Deployment history`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/overview-of-charts/overview-of-charts-9-2.jpg)

### Custom Values

You can use the default values or create Custom value by clicking on `Create Custom`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/overview-of-charts/overview-of-charts-7.jpg)

You can name your Custom Value, select the Chart Version and change the configurations in YAML file.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/overview-of-charts/overview-of-charts-8-2.jpg)

Click on `Save Template` to save the configurations.

### Deploying Chart Groups

You can deploy multiple applications and work with them simultaneously by creating `Chart Group`.
To create chart group click on `Create Group`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/chart-group/chart-group-1.jpg)

Add the `Group Name` and `Description`(optional), and select `Create Group`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/chart-group/chart-group-2.jpg)

You can select the Charts that you want to add to your Chart Group by clicking on '+' sign. You also can add multiple copies of the same chart in the chart group according to your requirements.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/chart-group/chart-group-3.jpg)

Select the `Version` and `Values` for your charts.
You can use Default Values or the Custom Values, just make sure the value that you select for the chart is compatible with the version of the chart that you are using.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/overview-of-charts/overview-of-charts-9.jpg)

To edit the chart configuration, Click on `Edit`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/overview-of-charts/overview-of-charts-10.jpg)

You can `Add` more charts or `Delete` charts from your existing Chart Group.
After making any changes, click on `Save` to save changes for the Chart Group.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/overview-of-charts/overview-of-charts-11.jpg)

If you wish to edit the chart configuration of any chart in the chart group, click on that Chart and edit the configurations in YAML file. You also can edit the `App Name`, `Chart Version`, `Values`, `Deploy Environment` and the YAML file from here.

| Key | Description |
| :--- | :--- |
| `App Name` | Name of the app |
| `Project` | Name of Project in which app has to be created |
| `Environment` | Name of the Environment in which app has to be deployed |
| `Chart Version` | Select the Version of the chart to be used |

After changing the configurations, click on `Deploy` to initiate the deployment of the chart in the Chart Group.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/overview-of-charts/overview-of-charts-12.jpg)


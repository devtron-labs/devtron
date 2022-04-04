# Overview Of Charts

## Deploying Charts

Charts can be deployed individually or by creating a group of Charts. Both methods are mentioned here.

### Deploying Chart

To deploy any chart or chart group, visit the `Charts` section from left panel and then select the chart that you want to use.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/deploy-charts/overview-of-charts/overview-of-charts-1.jpg)

Click on `README.md` to get more idea about the configurations of the chart.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/deploy-charts/overview-of-charts/overview-of-charts-2.jpg)

Select the Chart Version that you want to use and Chart Value, you can either use the Default Values or Custom Values.

To know about Custom Values, click on [Custom Values](overview-of-charts.md#custom-values)

The configuration values can be edited in the section given below Chart Version.

| Key | Description |
| :--- | :--- |
| `App Name` | Name of the app |
| `Environment` | Environment of the app to be deployed in |
| `Chart Version` | Version of the chart to be used |

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/deploy-charts/overview-of-charts/overview-of-charts-3.jpg)

Readme.md present on the left can be used by the user to set configuration values.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/deploy-charts/overview-of-charts/overview-of-charts-4.jpg)


Click on `Deploy Chart` to deploy the chart.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/deploy-charts/overview-of-charts/overview-of-charts-5.jpg)

Click on `App Details` to see the status and details of the deployed chart  and click on `Values` to reconfigure the deployment.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/deploy-charts/overview-of-charts/overview-of-charts-6.jpg)

Configuration values can be edited over here by the help of Readme.md .

Select Update And Deploy to update new settings.

You can also see deployment history of Helm application by clicking on `Deployment history` from `App details` page.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/deploy-charts/overview-of-charts/overview-of-charts-9.jpg)


### Custom Values

You can use the default values or create Custom value by clicking on `Create Custom`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/deploy-charts/overview-of-charts/overview-of-charts-7.jpg)

You can name your Custom Value, select the Chart Version and change the configurations in YAML file.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/deploy-charts/overview-of-charts/overview-of-charts-8.jpg)

Click on `Save Template` to save the configurations.

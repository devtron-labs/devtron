# Deploying Mysql Helm Chart

## Introduction

`stable/mysql` Helm chart bootstraps a single node MySQL deployment on a Kubernetes cluster using the Helm package manager.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/deploying-mysql-helm-chart/mysql-1.jpg)

## 1. Discover MySQL chart from Chart Store

Select `Charts` from the left panel to visit the `Chart Store` page. You will see numerous of charts on the page from which you have to find `stable/mysql` chart. You also can use the search bar to search the MySQL chart.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/deploying-mysql-helm-chart/mysql-2.jpg)

## 2. Configure the Chart

After selecting the `stable/mysql` Helm chart, click on `Deploy`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/deploying-mysql-helm-chart/mysql-3.jpg)

Enter the following details, to deploy MySQL chart:

| Key | Description |
| :--- | :--- |
| `App Name` | Name of the Chart |
| `Project` | Select the name of your Project in which you want to deploy the chart |
| `Environment` | Select the environment in which you want to deploy the chart |
| `Chart Version` | Select the latest Chart Version |
| `Chart Value` | Select the default value or create a custom value |

### Configure `values.yaml`

Set the following parameters in the chart, to be later used to connect MySQL with your Django Application.

| Parameters | Description |
| :--- | :--- |
| `mysqlRootPassword` | Password for the root user. Ignored if existing secret is provided |
| `mysqlDatabase` | Name of your MySQL database |
| `mysqluser` | Username of new user to create |
| `mysqlPassword` | Password for the new user. Ignored if existing secret is provided |

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/deploying-mysql-helm-chart/mysql-4.jpg)

Click on `Deploy Chart` to deploy the Chart.

## 3. Check the Status of Deployment

After clicking on `Deploy` you will be redirected to app details page where you can see deployment status of the chart. The Status of the chart should be `Healthy`. It might take few seconds after initiating the deployment of the chart.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/deploying-mysql-helm-chart/mysql-5.jpg)

In case the Status, of the deployment is `Degraded` or takes a long time to get deployed.
Click on the `Status` or check the logs of the pods to debug the issue.

## 4. Extract the Service Name

Copy the service name, it will be used to connect your application to MySQL.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploy-chart/deploying-mysql-helm-chart/mysql-6.jpg)


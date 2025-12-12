# Deploying Prometheus Helm Chart

## Introduction

Let's assume that you have created an application and want to monitor and get the application logs. You can deploy Prometheus using the `prometheus-community/kube-prometheus-stack` Helm chart and connect it to your application.

<p align="center">
  <img src="https://github.com/Rajdeep1311/devtron/assets/113296626/a4819bb3-4732-4c36-b3b4-5e64d3acbbdb">
</p> 

This guide will introduce you to how to deploy Prometheus's Helm chart.

## 1. Discover the Chart from the Chart Store

Visit the `Chart Store` page by clicking on `Charts` present on the left panel and find the `prometheus-community/kube-prometheus-stack` Helm Chart.
You also can search the Prometheus chart using the search bar.

![Screenshot (409)](https://github.com/Rajdeep1311/devtron/assets/113296626/c12956e2-d23d-4edd-b1cd-ed62a8a6e563)

## 2. Configure the Chart

After selecting the `prometheus-community/kube-prometheus-stack` Helm chart, click on `Configure & Deploy`.

![Screenshot (410)](https://github.com/Rajdeep1311/devtron/assets/113296626/0aac4e4f-d46c-4bd9-a979-4207542eb5bb)

Enter the following details before deploying the mongoDB chart:

| Key | Description |
| :--- | :--- |
| `App Name` | Name of the Chart |
| `Project` | Select the name of the Project in which you want to deploy the chart |
| `Environment` | Select the environment in which you want to deploy the chart |
| `Chart Version` | Select the latest Chart Version |
| `Chart Value` | Select the latest default value or create a custom value |

### Configure `values.yaml`

You can configure the `values.yaml` according to your project's requirements.
To learn about different parameters used in the chart, you can check the [Documentation of Prometheus Helm chart](https://artifacthub.io/packages/helm/prometheus-community/prometheus)

![Screenshot (411)](https://github.com/Rajdeep1311/devtron/assets/113296626/54eba3ea-912d-480c-8a1e-f1682cd0b197)

Click on `Deploy Chart` once you have finished configuring the chart.

## 3. Check the Status of Deployment

After clicking on `Deploy Chart`, you will be redirected to the `App Details` page that shows the deployment status of the chart. The Status of the chart should be `Healthy`. It might take a few seconds after initiating the deployment.

![Screenshot (413)](https://github.com/Rajdeep1311/devtron/assets/113296626/b4d2b386-3601-4ee7-b412-b5887676f09a)

![Screenshot (414)](https://github.com/Rajdeep1311/devtron/assets/113296626/065320fb-8009-4d9b-a304-dbb5c70894cd)


In case the status of the deployment is `Degraded` or takes a long time to get deployed, click on `Status` or check the logs of the pods to debug the issue.

## 4. Extract the Service name

You have to copy the service name, it will be used to connect your application to Prometheus.

![Screenshot (461)](https://github.com/Rajdeep1311/devtron/assets/113296626/961ab677-3d7c-497b-9c2b-6b23b9cffc22)



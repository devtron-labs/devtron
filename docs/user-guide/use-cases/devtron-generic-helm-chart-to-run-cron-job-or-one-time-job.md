# Devtron Generic Helm Chart To Run Cron Job Or One Time Job

**Devtron also supports Job and Cronjob pipelines. If you need to regularly update the image and configurations of your cronjob/job, you should prefer to create a pipeline,To know more about this you can refer the link** [cronjob/job documentation](../creating-application/deployment-template/job-and-cronjob.md).

## Using Devtron-generic-Helm Chart to run Cron Job or One Time job

You can discover over 200 Charts from the Devtron Chart store to perform different tasks such as to deploy a YAML file.

You can use Devtron's generic helm chart to run the CronJobs or one time Job.

Select the `devtron-charts/devtron-generic-helm` chart from the Devtron Chart Store.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/devtron-generic-helm-chart-to-run-cron-job-or-one-time-job/use-case-chart-store.jpg)

Select the Chart Version and the Chart Value of the Chart.

And, then click on `Deploy`

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/devtron-generic-helm-chart-to-run-cron-job-or-one-time-job/use-case-deploy-chart.jpg)

**Configure devtron-generic-helm chart**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/devtron-generic-helm-chart-to-run-cron-job-or-one-time-job/gc-4.jpg)

Click on **Deploy Chart**

| Key | Description |
| :--- | :--- |
| `App Name` | Name of the app |
| `Project` | Name of the Project |
| `Environment` | Select the Environment in which you want to deploy app |
| `Chart Version` | Select the Version of the chart |
| `Chart Values` | Select the Chart Value or Create a Custom Value |

In **values.yaml**, you can specify the YAML file that schedules the CronJob for your application.


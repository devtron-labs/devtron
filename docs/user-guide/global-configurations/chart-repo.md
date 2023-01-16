# Chart Repository

You can add more chart repositories to Devtron. Once added, they will be available in the `All Charts` section of the [Chart Store](https://docs.devtron.ai/usage/deploy-chart/overview-of-charts).

**Note**: After the successful installation of Devtron, click `Refetch Charts` to sync and download all the default charts listed on the dashboard.

## Add Chart Repository

To add chart repository, go to the `Chart repositories` section of `Global Configurations`. Click **Add repository**.

You can either select:
 * `Public repository` or
 * `Private repository`

If you select `Public repository`, provide below information in the following fields:

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/chart-repo/add-chart-repository-public.jpg)

| Fields | Description |
| --- | --- |
| **Name** | Provide a `Name` of your chart repository. This name is added as prefix to the name of the chart in the listing on the helm chart section of application. |
| **URL** | This is the URL of your chart repository. E.g. https://charts.bitnami.com/bitnami |


If you select `Private repository`, provide below information in the following fields:

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/chart-repo/add-chart-repository-private.jpg)

| Fields | Description |
| --- | --- |
| **Name** | Provide a `Name` of your chart repository. This name is added as prefix to the name of the chart in the listing on the helm chart section of application. |
| **URL** | This is the URL of your chart repository. E.g. https://charts.bitnami.com/bitnami |
| **Username** | Provide the username of your chart repository. |
| **Password** | Enter the password of your chart repository |
| **Secure with TLS** | Enable this field to encrypt your data. |


## Update Chart Repository

You can also update your saved chart repository settings. 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/chart-repo/update-chart-repository.jpg)

1. Click the chart repository which you want to update. 
2. Modify the required changes and click `Update` to save you changes.

**Note**: 
* You can perform a dry run to validate the below chart repo configurations by clicking `Validate`.
* You can enable or disable your chart repository. If you enable it, then you will be able to see the enabled chart in `All Charts` section of the [Chart Store](https://docs.devtron.ai/usage/deploy-chart/overview-of-charts).


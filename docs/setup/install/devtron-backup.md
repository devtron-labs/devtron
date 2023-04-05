Regular backups for Devtron postgres and ArgoCD are an important part of disaster recovery plan, as they can help protect against data loss due to any reason.

We can store the Devtron backup either on AWS S3 or on Azure containers. Both ways of taking Devtron backup is described below.

1. Go to the devtron chart store and search for `devtron-backups` chart.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/backup/backup-chart-search.jpg)

2. Select the `devtron-backups` and click `Configure & Deploy`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/backup/backup-chart-configure.jpg)

## For AWS S3

To store Devtron backups on AWS S3, please follow the mentined steps:

1. Create a user having permission to S3.
2. Get the access key and secret-access key.
3. Create a bucket for devtron backup.
6. Configure `devtron-backups` chart (for AWS S3) like : 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/backup/aws-backup-config.jpg)

7. Deploy chart. Devtron backup will be keep on uploading to the AWS S3 on scheduled time.


## For Azure Containers

To store Devtron backups on AWS S3, please follow the mentined steps:

1. Create a storage account
2. Create two containers in that storage account.
3. Go to Security + networking –> Access Key section.
4. Copy the Access Key.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/backup/azure-sa-key.jpg)

5. Configure `devtron-backups` chart (for AWS S3) like : 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/backup/azure-backup-config.jpg)

6. Deploy chart. But before deploying the backup chart, make sure that `AWS.enabled` is set as `false`. Devtron backup will be keep on uploading to the configured azure containers on scheduled time.
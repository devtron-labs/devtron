## Devtron Backup

Regular backups for Devtron PostgreSQL and ArgoCD are crucial components of a disaster recovery plan, as they protect against potential data loss due to unforeseen circumstances. This documentation provides instructions on how to take backups of Devtron and store them either on AWS S3 or Azure containers.

1. Go to the devtron chart store and search for `devtron-backups` chart.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/backup/backup-chart-search.jpg)

2. Select the `devtron-backups` and click `Configure & Deploy`.
3. Now follow either of the options described below according to your Cloud provider.

### AWS S3 Backup

To store Devtron backups on AWS S3, please follow these steps:

1. Create an S3 bucket to store the Devtron backup, you can configure the bucket to delete all the objects older than 15/30 days.
2. Create a user with sufficient permissions to push to the S3 bucket created in step 1.
3. Obtain the access key and secret access key for the created user.
4. Configure the `devtron-backups` chart for AWS S3 by selecting the appropriate options:

![AWS Backup Configuration](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/backup/aws-backup-config.jpg)

5. Deploy the chart, and the Devtron backup will be automatically uploaded to the AWS S3 bucket at the scheduled intervals.

### Azure Containers Backup

To store Devtron backups on Azure Containers, please follow these steps:

1. Create a storage account in Azure.
2. Within the storage account, create two containers for the Devtron backup.
3. Navigate to Security + Networking > Access Key section in Azure and copy the Access Key:

![Azure Storage Account Key](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/backup/azure-sa-key.jpg)

4. Configure the `devtron-backups` chart for Azure Containers by providing the Access Key:

![Azure Backup Configuration](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/backup/azure-backup-config.jpg)

5. Before deploying the backup chart, ensure that `AWS.enabled` is set to `false`. This will ensure that Devtron backup will be automatically uploaded to the configured Azure containers on the scheduled intervals.

By following these steps, you can ensure that your Devtron data is securely backed up and protected against any potential data loss, enabling you to recover quickly in case of emergencies.

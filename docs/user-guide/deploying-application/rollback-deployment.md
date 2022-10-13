# Rollback Deployment


Deployments can be rolled back manually. After a deployment is completed, you can manually rollback a stage to a previously deployed image by retaining the same configuration or changing the configuration.


![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/rollback-deployment/rollback-icon.jpg)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/rollback-deployment/select-config-to-deploy.jpg)

1. Select `Rollback` in your deployed pipeline.

2. On the `Rollback` page, select a configuration to deploy from the list:

| Configurations | Description |
| --- | --- |
| **Last saved config** | Deploy the image with the latest saved configuration. |
| **Last deployed config** | Deploy the image with the last deployed configuration. |
| **Config deployed with selected image** | Deploy the image with the configuration of a newly selected image. |

3. Once you select the previously deployed image and the configuration, review the difference between `Last Deployed Configuration` and the selected configuration.

4. You can also edit the configuration values on the `Deployment Template` and click `Deploy`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/rollback-deployment/config-diff.jpg)

The selected previously deployed image will be deployed.

**Note**: 
- There will be no difference in configuration if you select `Last deployed config` from the list.
- When you select `Config deployed with selected image` and  if the configuration is missing in the selected previously deployed image, it will show as `Config Not Available`. In such cases, you can use `Last saved config` or `Last deployed config`.




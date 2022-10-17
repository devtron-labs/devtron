# Rollback Deployment


Deployments can be rolled back manually. After a deployment is completed, you can manually rollback a stage to a previously deployed image by retaining the same configuration or changing the configuration.

**As an example**: You have deployed three different images V1, V2 and V3. And image V3 is currently deployed on your environment. Now, if you want to roll back the deployment from image V3 to image V1, you can retain the same configuration of the selected image V1 or change the configuration to `last saved config` or `last deployed config` of image V3.


![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/rollback-deployment/rollback-icon.jpg)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/rollback-deployment/select-config-to-deploy.png)

1. Select `Rollback` in your deployed pipeline.

2. On the `Rollback` page, select a configuration to deploy from the list:

| Configurations | Description |
| --- | --- |
| **Last saved config** | Deploy the image with the latest saved configuration. |
| **Last deployed config** | Deploy the image with the last deployed configuration. [As an example](#rollback-deployment): `The configuration of image V3`.|
| **Config deployed with selected image** | Deploy the image with the configuration of a newly selected image. [As an example](#rollback-deployment): `The configuration of image V1`. |

3. Once you select the previously deployed image and the configuration, review the difference between `Last Deployed Configuration` and the selected configuration.

4. Click `Deploy`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/rollback-deployment/config-diff.jpg)

The selected previously deployed image will be deployed.

**Note**: 
- There will be no difference in the configuration if you select `Last deployed config` from the list.
- When you select `Config deployed with selected image` and  if the configuration is missing in the selected previously deployed image, it will show as `Config Not Available`. In such cases, you can select either `Last saved config` or `Last deployed config`.




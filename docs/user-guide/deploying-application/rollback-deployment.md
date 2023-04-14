# Rollback Deployment


Deployments can be rolled back manually. After a deployment is completed, you can manually rollback to a previously deployed image by retaining the same configuration or changing the configuration.

**As an example**, You have deployed four different releases as follows:

| Image | Configuration | Release |
| --- | --- | --- |
| V1 | C1 | R1 |
| V2 | C2 | R2 |
| V3 | C2 | R3 |
| V3 | C3 | R4 |
| V3 | C4 (saved but not deployed)| -  |

If you want to roll back from V3 image to V2 image, then you have the following options:

| Configuration Option | Image | Configuration |
| --- | --- | --- |
| Config deployed with selected image | V2 | C2 |
| Last deployed config | V2 | C3 |
| Last saved config | V2 | C4 |


![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/rollback-deployment/rollback-icon.jpg)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/rollback-deployment/select-config-to-deploy.png)

1. Select `Rollback` in your deployed pipeline.

2. On the `Rollback` page, select a configuration to deploy from the list:

| Configurations | Description |
| --- | --- |
| **Last saved config** | Deploy the image with the latest saved configuration. |
| **Last deployed config** | Deploy the image with the last deployed configuration. [As an example](#rollback-deployment): `The configuration C3`.|
| **Config deployed with selected image** | Deploy the configuration which was deployed with the selected image. [As an example](#rollback-deployment): `The configuration C2`. |

3. Once you select the previously deployed image and the configuration, review the difference between `Last Deployed Configuration` and the selected configuration.

4. Click `Deploy`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/deploying-application/rollback-deployment/config-diff.jpg)

The selected previously deployed image will be deployed.

**Note**: 
- There will be no difference in the configuration if you select `Last deployed config` from the list.
- When you select `Config deployed with selected image` and  if the configuration is missing in the selected previously deployed image, it will show as `Config Not Available`. In such cases, you can select either `Last saved config` or `Last deployed config`.




# Build and Deploy

Each time you push a change to your application through GitHub, your application goes through a process to be built and deployed On the `Build and Deploy` section of the `Application Groups`, you can trigger build or deploy more than one application at the same time. You can also `Select all apps` to trigger build or deploy of all the applications. 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/application-groups/build-and-deploy-app-groups.jpg)


There are two main steps for building and deploying applications:

* [Triggering CI](triggering-ci.md)

* [Triggering CD](triggering-cd.md)


## Application with pre-deployment/post deployment pipeline

The application with the pre-deployment or post-deployment pipiline can also be selected by triggering:

* [Trigger Pre-deployment stage](https://docs.devtron.ai/v/v0.6/usage/applications/creating-application/cd-pipeline#3.-pre-deployment-stage)
* [Trigger Deployment](https://docs.devtron.ai/v/v0.6/usage/applications/creating-application/cd-pipeline#4.-deployment-stage)
* [Trigger Post-deployment stage](https://docs.devtron.ai/v/v0.6/usage/applications/creating-application/cd-pipeline#5.-post-deployment-stage)


![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/application-groups/pre-deployment-app-groups.jpg)


You can also rollback the deployment. Refer [Rollback Deployment](rollback-deployment.md) for detail.


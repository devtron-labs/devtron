#  Override Build Configuration

Within the same application, you can override a `container registry`, `container image` and `target platform` during the build pipeline, which means the images built for non-production environment can be included to the non-production registry and the images for production environment can be included to the production registry.


![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/select-build-override.jpg)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/build-allow-override.jpg)

To override a container registry, container image or target platform:

* Go to **Applications** and select your application from the **Devtron Apps** tabs.
* On the **App Configuration** tab, select **Workflow Editor**.
* Open the build pipeline of your application.
* Click **Allow Override** to:
   * Select the new container registry from the drop-down list.
   * Or, [create and build the new container image](../creating-application/docker-build-configuration.md#build-the-container-image) with different options.
   * Or, set a [new target platform](../creating-application/docker-build-configuration.md#set-target-platform-for-the-build) from the drop-down list or enter a new target platform.

* Click **Update Pipeline**.

The overridden container registry/container image location/target platform will be reflected on the [Build Configuration](docker-build-configuration.md) page. You can also see the number of build pipelines for which the container registry/container image location/target platform is overridden.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/build-configuration-overridden.jpg)



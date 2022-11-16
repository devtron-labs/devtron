#  Override Container Registry and Container Image

Within the same application, you can override a `Container registry` and `Container image` during the build pipeline, which means the images built for non-production environment can be included to the non-production registry and the images for production environment can be included to the production registry.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/override-container-registry-image.png)

To override a container registry and container image:
1. Open the build pipeline of your application.
2. Click **Allow Override** to select the new container registry and also to create and build the new container image.
3. On **Store container image** section:

| Fields | Description |
| --- | --- |
| **Container Registry** | Select the container registry from the drop-down list. |
| **Container Repository** | Enter the name of the container repository. |

To create and build a new container image with different options, refer [build the container image](https://docs.devtron.ai/usage/applications/creating-application/docker-build-configuration#build-the-container-image) section.

4. **Update Pipeline**.

The overriden container registry / container image location will be reflected on the [Build Configuration](docker-build-configuration.md) page. You can also see the number of build pipelines for which the container registry / container image location is overriden.


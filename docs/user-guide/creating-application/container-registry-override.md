#  Override Container Registry

Within the same application, you can override a container registry during the build pipeline, which means the images built for non-production environment can be included to the non-production registry and the images for production environment can be included to the production registry.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/override-container-registries.png)

To override a container registry:
1. Open the build pipeline of your application.
2. Select **Advance options**.
3. Click **Allow Override** to select the new container registry.
4. On **Registry to store container images** section:

| Fields | Description |
| --- | --- |
| **Container registry** | Select the container registry from the drop-down list. |
| **Container Repository** | Enter the name of the container repository. |

5. On **Docker file location** section:

| Fields | Description |
| --- | --- |
| **Select repository containing docker file** | Select the repository which contains your docker file. |
| **Docker file path (relative) *** | Enter the Docker file path. |

6. **Update Pipeline**.

The overriden container registry / docker file location will be reflected on the [Docker build configuration](https://docs.devtron.ai/usage/applications/creating-application/docker-build-configuration) page. You can also see the number of build pipelines for which the container registry / docker file location is overriden.


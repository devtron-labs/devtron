# Container Registries

The global configuration helps you add your `Container Registry`. In the container registry, you provide credentials of your registry, where your images will be stored.

## Add Container Registry configuration:

Go to the `Container Registry` section of `Global Configuration`. Click on `Add Container registry`.

You will see below the input fields to configure the container registry.

* Name
* Registry type
* Registry URL
* Set as default

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/docker-registries/First-page-registry.JPG)

### Name

Provide a name to your registry, this name will be shown to you in Docker Build Config as a drop-down.

### Registry type

Select type of Registry from the drop-down, Currently we are Supporting Multiple Types of Container Registry across Global Platforms. Default `ECR` is Selected

Registries we are Supporting are-:

  * Elastic Container Registry (ECR)
  * Docker-Hub
  * Azure Container Registry (ACR)
  * Google Artifact Registry (GAR)
  * Google Container Registry (GCR)
  * <a href= #Quay>Quay</a>
  * Others

### Registry URL

 You have to provide the URL of your registry. Create your registry and provide the URL of that registry in the URL box.

### Registry Type- ECR:

You have to provide the below information if you select the registry type as ECR.

* **Access key ID**

Inside the Access key ID box, provide your AWS access key.

* **Secret access key**

Provide your AWS secret access key ID.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/docker-registries/ECR+.JPG)


### Registry Type- Docker Hub 

You have to provide the below information if you select the registry type as Docker Hub.

* **Username**

Give the username of the docker hub account you used for creating your registry in.

* **Password**

Give the password/[token](https://docs.docker.com/docker-hub/access-tokens/) corresponding to your docker hub account.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/docker-registries/docker-hub+copy.JPG)


### Azure Container Registry (ACR)

 Service principal authentication method can be used to authenticate with username and password. Please follow [link](https://docs.microsoft.com/en-us/azure/container-registry/container-registry-auth-service-principal) for getting the Username and password for this registry.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/docker-registries/Azure-registry.jpg)


### Google Artifact Registry (GAR) 

JSON key file authentication method can be used to authenticate with username and password. Please follow [link](https://cloud.google.com/artifact-registry/docs/docker/authentication#json-key) for getting username and password for this registry. Please remove all the white spaces from JSON key and wrap it in a single quote while putting it in the password field.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/docker-registries/Artifact-registry.JPG)


### Google Container Registry (GCR)
 
JSON key file authentication method can be used to authenticate with username and password. Please follow [link](https://cloud.google.com/container-registry/docs/advanced-authentication#json-key) for getting the username and password for this registry. Please remove all the white spaces from json key and wrap it in single quote while putting in the password field.  

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/docker-registries/GCR.JPG)


<section id="Quay"></section>
### Quay Container Registry

You have to provide the below information if you select the registry type as Quay.

* **Username**

Give the username of your account, where you have created your registry in.

* **Password**

Give the password corresponding to the username of your registry.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/docker-registries/Quay.JPG)


### Registry Type Others:

You have to provide the below information if you select the registry type as others.

* **Username**

Give the username of your account, where you have created your registry in.

* **Password**

Give the password corresponding to the username of your registry.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/docker-registries/other-registry.JPG)


### Advance Registry Url connection options:

* If you enable the `Allow Only Secure Connection` option, then this registry allows only secure connections.

* If you enable the `Allow Secure Connection With CA Certificate` option, then you have to upload/provide private CA certificate (ca.crt).

* If the container registry is insecure (for eg :the SSL certificate is expired), then you enable the `Allow Insecure Connection` option.

Now click on `Save` to save the configuration of the `Container registry`.

### Note:

You can use any registry which can be authenticated using `docker login -u <username> -p <password> <registry-url>`. However,these registries might provide a more secure way for authentication, which we will support later.

### Set as default:

If you enable the `Set as default` option, then this registry name will be set as default in the `Container Registry` section inside the `Docker build config` page. This is optional. You can keep it disabled.





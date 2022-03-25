# Container Registries

The global configuration helps you add your `Container Registry`. In the container registry, you provide credentials of your registry, where your images will be stored.

## Add Container Registry configuration:

Go to the `Container Registry` section of `Global Configuration`. click on `Add Container registry`.

You will see below the input fields to configure the container registry.

* Name
* Registry type
* Registry URL
* Set as default

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/docker-registries/First-page-registry.JPG)

### Name

Provide a `Name` to your registry, this name will be shown to you in `Docker Build Config` as a drop-down.

### Registry type

Select `Registry type` from the drop-down, currently we are supporting multiple types of container registry across different global platforms. by default `ECR` is selected

Registries we are supporting are-:

  * <a href= #ECR>Elastic Container Registry (ECR)</a>
  * <a href= #Docker>Docker-Hub</a>
  * <a href= #Azure>Azure Container Registry (ACR)</a>
  * <a href= #GAR>Google Artifact Registry (GAR)</a>
  * <a href= #GCR>Google Container Registry (GCR)</a>
  * <a href= #Quay>Quay</a>
  * <a href= #others>Others</a>

### Registry URL

 You have to provide the `Registry URL` of your registry. create your registry and provide the URL of that registry in the URL box.

<section id="ECR"></section>

### Elastic Container Registry (ECR):

You have to provide the below information if you select the registry type as `ECR`.

* **Access key ID**

Inside the `Access key ID` box, provide your AWS access key.

* **Secret access key**

Inside the `Secret access key` box, provide your AWS secret access kek ID.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/docker-registries/ECR+.JPG)


<section id="Docker"></section>

### Docker-Hub:

You have to provide the below information if you select the registry type as Docker Hub.

* **Username**

Give the `Username` of the docker hub account you used for creating your registry in.

* **Password**

Give the password/[token](https://docs.docker.com/docker-hub/access-tokens/) corresponding to your docker hub account.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/docker-registries/docker-hub+copy.JPG)


<section id="Azure"></section>

### Azure Container Registry (ACR):

 Service principal authentication method can be used to authenticate with username and password. Please follow [link](https://docs.microsoft.com/en-us/azure/container-registry/container-registry-auth-service-principal) for getting the Username and password for this registry.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/docker-registries/Azure-registry.jpg)


<section id="GAR"></section>

### Google Artifact Registry (GAR):

JSON key file authentication method can be used to authenticate with username and password. Please follow [link](https://cloud.google.com/artifact-registry/docs/docker/authentication#json-key) for getting username and password for this registry. Please remove all the white spaces from JSON key and wrap it in a single quote while putting it in the password field.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/docker-registries/Artifact-registry.JPG)


<section id="GCR"></section>

### Google Container Registry (GCR):
 
JSON key file authentication method can be used to authenticate with username and password. Please follow [link](https://cloud.google.com/container-registry/docs/advanced-authentication#json-key) for getting the username and password for this registry. Please remove all the white spaces from json key and wrap it in single quote while putting in the password field.  

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/docker-registries/GCR.JPG)


<section id="Quay"></section>

### Quay Container Registry:

You have to provide the below information if you select the registry type as Quay.

* **Username**

Give the `Username` of your account, where you have created your registry in.

* **Password**

Give the `Token` corresponding to the username of your registry.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/docker-registries/Quay.JPG)

<section id="others"></section>

### Docker-HubOthers:

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





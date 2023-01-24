# Container Registries

Container registries are used for storing images built by the CI Pipeline. You can configure the container registry using any container registry provider of your choice. It allows you to build, deploy and manage your container images with easy-to-use UI. 

When configuring an application, you can choose the specific container registry and repository in the App Configuration > [Build Configuration](https://docs.devtron.ai/usage/applications/creating-application/docker-build-configuration) section.


## Add Container Registry:

To add container registry, go to the `Container Registry` section of `Global Configurations`. Click **Add Container Registry**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/add-container-registry.jpg)

Provide the information in the following fields to add container registry.

| Fields | Description |
| --- | --- |
| **Name** | Provide a name to your registry, this name will be shown to you on the [Build Configuration](https://docs.devtron.ai/usage/applications/creating-application/docker-build-configuration) in the drop-down list. |
| **Registry Type** | Select the registry type from the drop-down list:<br><ul><li>[ECR](#registry-type-ecr)</li></ul><ul><li>[Docker](#registry-type-docker)</li></ul><ul><li>[Azure](#registry-type-azure)</li></ul><ul><li>[Artifact Registry (GCP)](#registry-type-artifact-registry-gcp)</li></ul><ul><li>[GCR](#registry-type-google-container-registry-gcr)</li></ul><ul><li>[Quay](#registry-type-quay)</li></ul><ul><li>[Other](#registry-type-other)</li></ul>`Note`: For each **Registry Type**, the credential input fields are different. |
| **Registry URL** | Provide the URL of your registry. |
| **Set as default registry** | Enable this field to set as default registry hub for your images. |

  

### Registry Type: ECR

Amazon ECR is an AWS-managed container image registry service.
The ECR provides resource-based permissions to the private repositories using AWS Identity and Access Management (IAM). ECR allows both Key-based and Role-based authentications.

Before you begin, create an [IAM user](https://docs.aws.amazon.com/AmazonECR/latest/userguide/get-set-up-for-amazon-ecr.html) and attach the ECR policy according to the authentication type.

Provide below information if you select the registry type as `ECR`. 

| Fields | Description |
| --- | --- |
| **Name** | User-defined name for the registry in Devtron. |
| **Registry Type** | Select **ECR**. |
| **Registry URL** | This is the URL of your private registry in AWS.<br>For example, the URL format is: `https://xxxxxxxxxxxx.dkr.ecr.<region>.amazonaws.com`. `xxxxxxxxxxxx` is your 12-digit AWS account ID.</br> |
| **Authentication Type** | Select one of the authentication types:<ul><li>**EC2 IAM Role**: Authenticate with workernode IAM role and attach the ECR policy (AmazonEC2ContainerRegistryFullAccess) to the cluster worker nodes IAM role of your Kubernetes cluster.</li></ul><ul><li>**User Auth**: It is key-based authentication and attach the ECR policy (AmazonEC2ContainerRegistryFullAccess) to the [IAM user](https://docs.aws.amazon.com/AmazonECR/latest/userguide/get-set-up-for-amazon-ecr.html).<ul><li>`Access key ID`: Your AWS access key</li></ul><ul><li>`Secret access key`: Your AWS secret access key ID</li></ul> |
| **Set as default registry** | Enable this field to set `ECR` as default registry hub for your images. |

Click **Save**.


### Registry Type: Docker 

Provide below information if you select the registry type as `Docker`.

| Fields | Description |
| --- | --- |
| **Name** | User-defined name for the registry in Devtron. |
| **Registry Type** | Select **Docker**. |
| **Registry URL** | This is the URL of your private registry in Docker. E.g. `docker.io` |
| **Username** | Provide the username of the docker hub account you used for creating your registry. |
| **Password/Token** | Provide the password/[Token](https://docs.docker.com/docker-hub/access-tokens/) corresponding to your docker hub account. It is recommended to use `Token` for security purpose. |
| **Set as default registry** | Enable this field to set `Docker` as default registry hub for your images. |

Click **Save**.

### Registry Type: Azure

For registry type: Azure, the service principal authentication method can be used to authenticate with username and password. Please follow [link](https://docs.microsoft.com/en-us/azure/container-registry/container-registry-auth-service-principal) for getting username and password for this registry.

Provide below information if you select the registry type as `Azure`.

| Fields | Description |
| --- | --- |
| **Name** | User-defined name for the registry in Devtron. |
| **Registry Type** | Select **Azure**. |
| **Registry URL/Login Server** | This is the URL of your private registry in Azure. E.g. `xxx.azurecr.io` |
| **Username/Registry Name** | Provide the username of Azure container registry. |
| **Password** | Provide the password of Azure container registry. |
| **Set as default registry** | Enable this field to set `Azure` as default registry hub for your images. |

Click **Save**.


### Registry Type: Artifact Registry (GCP) 

JSON key file authentication method can be used to authenticate with username and service account JSON file. Please follow [link](https://cloud.google.com/artifact-registry/docs/docker/authentication#json-key) to get username and service account JSON file for this registry. 

**Note**: Please remove all the white spaces from json key and wrap it in single quote while putting in `Service Account JSON File` field.


Provide below information if you select the registry type as `Artifact Registry (GCP)`.

| Fields | Description |
| --- | --- |
| **Name** | User-defined name for the registry in Devtron. |
| **Registry Type** | Select **Artifact Registry (GCP)**. |
| **Registry URL** | This is the URL of your private registry in Artifact Registry (GCP). E.g. `region-docker.pkg.dev` |
| **Username** | Provide the username of Artifact Registry (GCP) account. |
| **Service Account JSON File** | Provide the Service Account JSON File of Artifact Registry (GCP). |
| **Set as default registry** | Enable this field to set `Artifact Registry (GCP)` as default registry hub for your images. |

Click **Save**.

### Registry Type: Google Container Registry (GCR) 

JSON key file authentication method can be used to authenticate with username and service account JSON file. Please follow [link](https://cloud.google.com/container-registry/docs/advanced-authentication#json-key) for getting username and service account JSON file for this registry. Please remove all the white spaces from json key and wrap it in single quote while putting in the `Service Account JSON File` field.  
 
Provide below information if you select the registry type as `GCR`.

| Fields | Description |
| --- | --- |
| **Name** | User-defined name for the registry in Devtron. |
| **Registry Type** | Select **GCR**. |
| **Registry URL** | This is the URL of your private registry in GCR. E.g. `gcr.io` |
| **Username** | Provide the username of your GCR account. |
| **Service Account JSON File** | Provide the Service Account JSON File of GCR. |
| **Set as default registry** | Enable this field to set `GCR` as default registry hub for your images. |

Click **Save**.

### Registry Type: Quay

Provide below information if you select the registry type as `Quay`.

| Fields | Description |
| --- | --- |
| **Name** | User-defined name for the registry in Devtron. |
| **Registry Type** | Select **Quay**. |
| **Registry URL** | This is the URL of your private registry in Quay. E.g. `quay.io` |
| **Username** | Provide the username of the Quay account. |
| **Password/Token** | Provide the password of your Quay account. |
| **Set as default registry** | Enable this field to set `Quay` as default registry hub for your images. |

Click **Save**.


### Registry Type: Other


Provide below information if you select the registry type as `Other`.

| Fields | Description |
| --- | --- |
| **Name** | User-defined name for the registry in Devtron. |
| **Registry Type** | Select **Other**. |
| **Registry URL** | This is the URL of your private registry. |
| **Username** | Provide the username of your account where you have created your registry. |
| **Password/Token** | Provide the Password/Token corresponding to the username of your registry. |
| **Set as default registry** | Enable this field to set as default registry hub for your images. |

Click **Save**.

#### Advance Registry URL Connection Options:

* If you enable the `Allow Only Secure Connection` option, then this registry allows only secure connections.
* If you enable the `Allow Secure Connection With CA Certificate` option, then you have to upload/provide private CA certificate (ca.crt).
* If the container registry is insecure (for eg : SSL certificate is expired), then you enable the `Allow Insecure Connection` option.

**Note**: You can use any registry which can be authenticated using `docker login -u <username> -p <password> <registry-url>`. However these registries might provide a more secured way for authentication, which we will support later.


## Pull an Image from a Private Registry

You can create a Pod that uses a `Secret` to pull an image from a private container registry. You can use any private container registry of your choice. As an example: [Docker Hub](https://www.docker.com/products/docker-hub).

Super admin users can decide if they want to auto-inject registry credentials or use a secret to pull an image for deployment to environments on specific clusters.

To manage the access of registry credentials, click **Manage**.

There are two options to manage the access of registry credentials:

| Fields | Description |
| --- | --- |
| **Do not inject credentials to clusters** | Select the clusters for which you do not want to inject credentials. |
| **Auto-inject credentials to clusters** | Select the clusters for which you want to inject credentials. |

You can choose one of the two options for defining credentials:

* [Use Registry Credentials](#use-registry-credentials)
* [Specify Image Pull Secret](#specify-image-pull-secret) 

### Use Registry Credentials

If you select **Use Registry Credentials**, the clusters will be auto-injected with the registry credentials of your registry type. As an example, If you select `Docker` as Registry Type, then the clusters will be auto-injected with the `username` and `password/token` which you use on the Docker Hub account.
Click **Save**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/use-registry-credentials.jpg)


### Specify Image Pull Secret

You can create a Secret by providing credentials on the command line.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/specify-image-pull-secret-latest.png)

Create this Secret naming it `regcred` (as an example):

```bash
kubectl create -n <namespace> secret docker-registry regcred --docker-server=<your-registry-server> --docker-username=<your-name> --docker-password=<your-pword> --docker-email=<your-email>
```

where:
* `namespace` is your virtual cluster. E.g., devtron-demo
* `your-registry-server` is your Private Docker Registry FQDN. Use https://index.docker.io/v1/ for DockerHub.
* `your-name` is your Docker username.
* `your-pword` is your Docker password.
* `your-email` is your Docker email.

You have successfully set your Docker credentials in the cluster as a Secret called `regcred`.

**Note**: Typing secrets on the command line may store them in your shell history unprotected, and those secrets might also be visible to other users on your PC during the time when kubectl is running.

Enter the `Secret` name in the field and click **Save**.






  
















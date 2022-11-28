# Container Registries

Container registries are used for storing images built by the CI Pipeline. You can configure the container registry using any container registry provider of your choice. It allows you to build, deploy and manage your container images with easy-to-use UI. 

When configuring an application, you can choose the specific container registry and repository in the App Configuration > [Build Configuration](user-guide/creating-application/docker-build-configuration.md) section.

Provide the information in the following fields to configure the container registry from App Configuration > [Build Configuration](user-guide/creating-application/docker-build-configuration.md) section.

| Fields | Description |
| --- | --- |
| **Container Registry** | Select the container registry from the drop-down list. |
| **Container Repository** | Enter the name of the container repository. |


## Add Container Registry:

To add container registry, go to the `Container Registry` section of `Global Configurations`. Click **Add Container Registry**.

| Fields | Description |
| --- | --- |
| **Name** | Provide a name to your registry, this name will be shown to you in Build Configuration in the drop-down list. |
| **Registry Type** | Select the registry type from the drop-down list. E.g., Docker. |
| **Registry URL** | Provide the URL of your registry. |
| **Set as default registry** | Enable this field to set as default registry hub for your images. |

* For each **Registry Type**, the credential input fields are different. Please see the table below to know the required credential inputs as per the selected registry type.

| Registry Type | Credentials |
| --- | --- |
| **ECR** | Select one of the authentication types:<ul><li>**EC2 IAM Role**</li></ul> <ul><li>**User Auth**<ul><li>`Access key ID`: </li></ul><ul><li>`Secret access key`</li></ul></li></ul>|
| **Docker** | <ul><li>`Username`</li></ul> <ul><li>`Password/Token (Recommended:Token)`</li></ul> |
| **Azure**  | <ul><li>`Username/Registry Name`</li></ul> <ul><li>`Password`</li></ul>  |
| **Artifact Registry (GCP)**  | <ul><li>`Username`</li></ul> <ul><li>`Service Account JSON File*`</li></ul>  |
| **GCR**  | <ul><li>`Username`</li></ul> <ul><li>`Service Account JSON File*`</li></ul>  |
| **Quay**  | <ul><li>`Username`</li></ul> <ul><li>`Token`</li></ul>  |
| **Other**  | <ul><li>`Username`</li></ul> <ul><li>`Password/Token`</li></ul>  |

  
![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/add-container-registry.jpg)


Please read each `registry type` in detail for you to help in choosing the right container registry for your application development needs.

* [Registry Type: ECR](#registry-type-ecr)
* [Registry Type: Docker](#registry-type-docker)
* [Registry Type: Google Container Registry (GCR)](#registry-type-google-container-registry-gcr)
* [Registry Type: Artifact Registry (GCP)](#registry-type-artifact-registry-gcp)
* [Registry Type: Others](#registry-type-others)


### Registry Type: ECR

To add an Amazon Elastic Container Registry (ECR), select the `ECR` Registry type.
Amazon ECR is an AWS-managed container image registry service.
The ECR provides resource-based permissions to the private repositories using AWS Identity and Access Management (IAM).
ECR allows both Key-based and Role-based authentications.

Before you begin, create an [IAM user](https://docs.aws.amazon.com/AmazonECR/latest/userguide/get-set-up-for-amazon-ecr.html), and attach only ECR policy ( AmazonEC2ContainerRegistryFullAccess ) if using Key-based auth. Or attach the ECR policy ( AmazonEC2ContainerRegistryFullAccess) to the cluster worker nodes IAM role of your Kubernetes cluster if using Role-based access.

| Fields | Description |
| --- | --- |
| **Name** | User-defined name for the registry in Devtron |
| **Registry Type** | Select **ECR** |
| **Registry URL** | This is the URL of your private registry in AWS. <br></br> For example, the URL format is: `https://xxxxxxxxxxxx.dkr.ecr.<region>.amazonaws.com`. <br></br>`xxxxxxxxxxxx` is your 12-digit AWS account Id. |
| **Authentication Type** | <br></br> * **EC2 IAM role**: Authenticate with workernode IAM role. <br></br> * **User Auth**: Authenticate with an authorization token <br></br>  - **Access key ID**: Your AWS access key. <br></br>  - **Secret access key**: Your AWS secret access key ID. |

![ECR Role-based authentication](https://devtron-public-asset.s3.us-east-2.amazonaws.com/Container-registeries/ECR-IAM-auth-role-based.png)

![ECR Key-based authentication](https://devtron-public-asset.s3.us-east-2.amazonaws.com/Container-registeries/ECR_user-auth-key-based.png)

To set the `ECR` as the default registry hub for your images, enable the field **[x] Set as default registry** and then click **Save**.

To use the ECR container image, go to the **Applications** page and select your application, and then select **App Configuration > [Build Configuration](./../creating-application/docker-build-configuration.md)**.

### Registry Type: Docker 

You have to provide below information if you select the registry type as `Docker`.

* **Username**

Provide the username of the docker hub account you used for creating your registry.

* **Password**

Provide the password/[Token](https://docs.docker.com/docker-hub/access-tokens/) corresponding to your docker hub account. 
It is recommended to use `Token` for security purpose.

![](../../user-guide/global-configurations/images/Container_Registry_DockerHub.jpg)

### Registry Type: Google Container Registry (GCR) 

You have to provide below information if you select the registry type as `GCR`.
JSON key file authentication method can be used to authenticate with username and service account JSON file. Please follow [link](https://cloud.google.com/container-registry/docs/advanced-authentication#json-key) for getting username and service account JSON file for this registry. Please remove all the white spaces from json key and wrap it in single quote while putting in the `Service Account JSON File` field.  
 
![](../../user-guide/global-configurations/images/Container_Registry_gcr.jpg)

### Registry Type: Artifact Registry (GCP) 

You have to provide below information if you select the registry type as `Artifact Registry (GCP)`.
JSON key file authentication method can be used to authenticate with username and service account JSON file. Please follow [link](https://cloud.google.com/artifact-registry/docs/docker/authentication#json-key) for getting username and service account JSON file for this registry. Please remove all the white spaces from json key and wrap it in single quote while putting in `Service Account JSON File` field.


### Registry Type: Azure
Service principal authentication method can be used to authenticate with username and password. Please follow [link](https://docs.microsoft.com/en-us/azure/container-registry/container-registry-auth-service-principal) for getting username and password for this registry.


### Registry Type: Others

You have to provide below information if you select the registry type as `Others`.

* **Username**

Provide the username of your account, where you have created your registry in.

* **Password/Token**

Provide the Password/Token corresponding to the username of your registry.

### Set as default registry:

If you enable the `Set as default registry` option, then the registry name will be set as default in the `Container Registry` section on the App Configuration >`Build Configuration` page. This field is optional. You can keep it disabled.

### Advance Registry URL Connection Options:

* If you enable the `Allow Only Secure Connection` option, then this registry allows only secure connections.
* If you enable the `Allow Secure Connection With CA Certificate` option, then you have to upload/provide private CA certificate (ca.crt).
* If the container registry is insecure (for eg : SSL certificate is expired), then you enable the `Allow Insecure Connection` option.

#### Note:
You can use any registry which can be authenticated using `docker login -u <username> -p <password> <registry-url>`. However these registries might provide a more secured way for authentication, which we will support later.


## Pull an Image from a Private Registry

You can create a Pod that uses a `Secret` to pull an image from a private container image registry or repository. There are many private registries in use. This task uses [Docker Hub](https://www.docker.com/products/docker-hub) as an example registry.

Super admin users can decide if they want to auto-inject registry credentials or use a secret to pull an image for deployment to environments on specific clusters.

To manage the access of registry credentials, click **Manage**.

There are two options to manage the access of registry credentials:

| Fields | Description |
| --- | --- |
| **Do not inject credentials to clusters** | Select the clusters for which you do not want to inject credentials. |
| **Auto-inject credentials to clusters** | Select the clusters for which you want to inject credentials. |

You can choose one of the two options for defining credentials:

* [User Registry Credentials](https://docs.devtron.ai/v/v0.6/getting-started/global-configurations/docker-registries#user-registry-credentials)
* [Specify Image Pull Secret](https://docs.devtron.ai/v/v0.6/getting-started/global-configurations/docker-registries#specify-image-pull-secret) 

### Use Registry Credentials

If you select **Use Registry Credentials**, the clusters will be auto-injected with the registry credentials of your registry type. As an example: If you select `Docker` as Registry Type and `docker.io` as Registry URL, the registry credentials of the clusters will be the `username` and `password` which you define.
Click **Save**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/use-registry-credentials.jpg)


### Specify Image Pull Secret

You can create a Secret by providing credentials on the command line.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/specify-image-pull-secret-latest.png)

Create this Secret, naming it `regcred`:

```bash
kubectl create -n <namespace> secret docker-registry regcred --docker-server=<your-registry-server> --docker-username=<your-name> --docker-password=<your-pword> --docker-email=<your-email>
```

where:
* <namespace> is your virtual cluster. E.g., devtron-demo
* <your-registry-server> is your Private Docker Registry FQDN. Use https://index.docker.io/v1/ for DockerHub.
* <your-name> is your Docker username.
* <your-pword> is your Docker password.
* <your-email> is your Docker email.

You have successfully set your Docker credentials in the cluster as a Secret called `regcred`.

**Note**: Typing secrets on the command line may store them in your shell history unprotected, and those secrets might also be visible to other users on your PC during the time when kubectl is running.

Enter the `Secret` name in the field and click **Save**.


## How to resolve if Deployment Status shows Failed or Degraded

If the deployment status shows `Failed` or `Degraded`, then the cluster is not able to pull container image from the private registry. In that case, the status of pod shows `ImagePullBackOff`.

The failure of deployment can be one of the following reasons:

* Provided credentials may not have permission to pull container image from registry.
* Provided credentials may be invalid.

You can resolve the `ImagePullBackOff` issue by clicking **How to resolve?** which will take you to the **App Details** page.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/how-to-resolve-latest1.png)


To provide the auto-inject credentials to the specific clusters for pulling the image from the private repository, click **Manage Access** which will take you to the **Container Registries** page. 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/manage-access-latest.jpg)

1. On the **Container Registries** page, select the docker registry and click **Manage**.
2. In the **Auto-inject credentials to clusters**, click **Confirm to edit** to select the specific cluster or all clusters for which you want to auto-inject the credentials to and click **Save**.
3. Redeploy the application after allowing the access.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/auto-inject-to-clusters.jpg)


## Integrating With External Container Registry

If you want to use a private registry for container registry other than ecr, this will be used to push image and then create a secret in same environment to pull the image to deploy. To create secret, go to charts section and search for chart ‘dt-secrets’ and configure the chart. Provide an App Name and select the Project and Environment in which you want to deploy this chart and then configure the values.yaml as shown in example. The given example is for DockerHub but you can configure similarly for any container registry that you want to use.

```yaml
name: regcred
type: kubernetes.io/dockerconfigjson
labels:
 test: chart
secrets:
 data:
   - key: .dockerconfigjson
     value: '{"auths":{"https://index.docker.io/v1/":{"username":"<username>","password":"<password>}}}'
```     

The `name` that you provide in values.yaml ie. `regcred` is name of the secret that will be used as `imagePullSecrets` to pull the image from docker hub to deploy. To know how `imagePullSecrets` will be used in the deployment-template, please follow the [documentation](https://docs.devtron.ai/devtron/user-guide/creating-application/deployment-template/rollout-deployment#imagepullsecrets).















# Container/OCI Registry

While [container registries](https://docs.devtron.ai/resources/glossary#container-registry) are typically used for storing [images](https://docs.devtron.ai/resources/glossary#image) built by the CI Pipeline, an OCI registry can store container images as well as other artifacts such as [helm charts](https://docs.devtron.ai/resources/glossary#helm-charts-packages). In other words, all container registries are OCI registries, but not all OCI registries are container registries.

You can configure a container registry using any registry provider of your choice. It allows you to build, deploy, and manage your container images or charts with easy-to-use UI. 


## Add Container Registry

1. From the left sidebar, go to **Global Configurations** → **Container/OCI Registry**.

    ![Figure 1: Container/OCI Registry](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/add-registry.jpg)

2. Click **Add Registry**.

    ![Figure 2: Add a Registry](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/add-container-registry-1.jpg)

3. Choose a provider from the **Registry provider** dropdown. View the [Supported Registry Providers](#supported-registry-providers).

4. Choose the Registry type:
    * **Private Registry**: Choose this if your images or artifacts are hosted or should be hosted on a private registry restricted to authenticated users of that registry. Selecting this option requires you to enter your registry credentials (username and password/token).
    * **Public Registry**: Unlike private registry, this doesn't require your registry credentials. Only the registry URL and repository name(s) would suffice.

5. Assuming your registry type is private, here are few of the common fields you can expect:

    | Fields | Description |
    | --- | --- |
    | **Name** | Provide a name to your registry, this name will appear in the **Container Registry** drop-down list available within the [Build Configuration](https://docs.devtron.ai/usage/applications/creating-application/docker-build-configuration) section of your application|
    | **Registry URL** | Provide the URL of your registry in case it doesn't come prefilled (do not include `oci://`, `http://`, or `/https://` in the URL) |
    | **Authentication Type** | The credential input fields may differ depending on the registry provider, check [Registry Providers](#supported-registry-providers) |
    | **Push container images** | Tick this checkbox if you wish to use the repository to push container images. This comes selected by default and you may untick it if you don't intend to push container images after a CI build. If you wish to to use the same repository to pull container images too, read [Registry Credential Access](#registry-credential-access). |
    | **Push helm packages** | Tick this checkbox if you wish to push helm charts to your registry |
    | **Use as chart repository** | Tick this checkbox if you want Devtron to pull helm charts from your registry and display them on its chart store. Also, you will have to provide a list of repositories (present within your registry) for Devtron to successfully pull the helm charts. |
    | **Set as default registry** | Tick this checkbox to set your registry as the default registry hub for your images or artifacts |

6. Click **Save**.


## Supported Registry Providers

### ECR

Amazon ECR is an AWS-managed container image registry service.
The ECR provides resource-based permissions to the private repositories using AWS Identity and Access Management (IAM). ECR allows both Key-based and Role-based authentications.

Before you begin, create an [IAM user](https://docs.aws.amazon.com/AmazonECR/latest/userguide/get-set-up-for-amazon-ecr.html) and attach the ECR policy according to the authentication type.

Provide the following additional information apart from the common fields:

| Fields | Description |
| --- | --- |
| **Registry URL** | Example of URL format: `xxxxxxxxxxxx.dkr.ecr.<region>.amazonaws.com` where `xxxxxxxxxxxx` is your 12-digit AWS account ID |
| **Authentication Type** | Select one of the authentication types:<ul><li>**EC2 IAM Role**: Authenticate with workernode IAM role and attach the ECR policy (AmazonEC2ContainerRegistryFullAccess) to the cluster worker nodes IAM role of your Kubernetes cluster.</li></ul><ul><li>**User Auth**: It is a key-based authentication, attach the ECR policy (AmazonEC2ContainerRegistryFullAccess) to the [IAM user](https://docs.aws.amazon.com/AmazonECR/latest/userguide/get-set-up-for-amazon-ecr.html).<ul><li>`Access key ID`: Your AWS access key</li></ul><ul><li>`Secret access key`: Your AWS secret access key ID</li></ul> |


### Docker 

Provide the following additional information apart from the common fields:

| Fields | Description |
| --- | --- |
| **Username** | Provide the username of the Docker Hub account you used for creating your registry. |
| **Password/Token** | Provide the password/[Token](https://docs.docker.com/docker-hub/access-tokens/) corresponding to your docker hub account. It is recommended to use `Token` for security purpose. |


### Azure

For Azure, the service principal authentication method can be used to authenticate with username and password. Visit this [link](https://docs.microsoft.com/en-us/azure/container-registry/container-registry-auth-service-principal) to get the username and password for this registry.

Provide the following additional information apart from the common fields:

| Fields | Description |
| --- | --- |
| **Registry URL/Login Server** | Example of URL format: `xxx.azurecr.io` |
| **Username/Registry Name** | Provide the username of your Azure container registry |
| **Password** | Provide the password of your Azure container registry |


### Artifact Registry (GCP) 

JSON key file authentication method can be used to authenticate with username and service account JSON file. Visit this [link](https://cloud.google.com/artifact-registry/docs/docker/authentication#json-key) to get the username and service account JSON file for this registry. 

{% hint style="warning" %}
Remove all the white spaces from JSON key and wrap it in a single quote before pasting it in `Service Account JSON File` field
{% endhint %}

Provide the following additional information apart from the common fields:

| Fields | Description |
| --- | --- |
| **Registry URL** | Example of URL format: `region-docker.pkg.dev` |
| **Service Account JSON File** | Paste the content of the service account JSON file |


### Google Container Registry (GCR) 

JSON key file authentication method can be used to authenticate with username and service account JSON file. Please follow [link](https://cloud.google.com/container-registry/docs/advanced-authentication#json-key) to get the username and service account JSON file for this registry. 

{% hint style="warning" %}
Remove all the white spaces from JSON key and wrap it in single quote before pasting it in `Service Account JSON File` field
{% endhint %}

### Quay

Provide the following additional information apart from the common fields:

| Fields | Description |
| --- | --- |
| **Username** | Provide the username of your Quay account |
| **Token** | Provide the password of your Quay account |


### Other

Provide below information if you select the registry type as `Other`.

| Fields | Description |
| --- | --- |
| **Registry URL** | Enter the URL of your private registry |
| **Username** | Provide the username of your account where you have created your registry |
| **Password/Token** | Provide the password or token corresponding to the username of your registry |
| **Advanced Registry URL Connection Options** | <ul><li>**Allow Only Secure Connection**: Tick this option for the registry to allow only secure connections</li></ul><ul><li>**Allow Secure Connection With CA Certificate**: Tick this option for the registry to allow secure connection by providing a private CA certificate (ca.crt)</li></ul><ul><li>**Allow Insecure Connection**: Tick this option to make an insecure communication with the registry (for e.g., when SSL certificate is expired)</li></ul> |

{% hint style="info" %}
You can use any registry which can be authenticated using `docker login -u <username> -p <password> <registry-url>`. However these registries might provide a more secured way for authentication, which we will support later.
{% endhint %}


## Registry Credential Access

You can create a Pod that uses a [Secret](https://docs.devtron.ai/resources/glossary#secrets) to pull an image from a private container registry. You can use any private container registry of your choice, for e.g., [Docker Hub](https://www.docker.com/products/docker-hub).

Super-admin users can decide if they want to auto-inject registry credentials or use a secret to pull an image for deployment to environments on specific clusters.

1. To manage the access of registry credentials, click **Manage**.

There are two options to manage the access of registry credentials:

| Fields | Description |
| --- | --- |
| **Do not inject credentials to clusters** | Select the clusters for which you do not want to inject credentials |
| **Auto-inject credentials to clusters** | Select the clusters for which you want to inject credentials |

2. You can choose one of the two options for defining credentials:

* [Use Registry Credentials](#use-registry-credentials)
* [Specify Image Pull Secret](#specify-image-pull-secret) 

### Use Registry Credentials

If you select **Use Registry Credentials**, the clusters will be auto-injected with the registry credentials of your registry type. As an example, If you select `Docker` as Registry Type, then the clusters will be auto-injected with the `username` and `password/token` which you use on the Docker Hub account.

Click **Save**.

![Figure 3: Using Registry Credentials](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/use-registry-credentials-1.jpg)


### Specify Image Pull Secret

You can create a Secret by providing credentials on the command line.

![Figure 4: Using Image Pull Secret](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/specify-image-pull-secret-1.jpg)

Create this Secret and name it `regcred` (let's say):

```bash
kubectl create -n <namespace> secret docker-registry regcred --docker-server=<your-registry-server> --docker-username=<your-name> --docker-password=<your-pword> --docker-email=<your-email>
```

where,

* **namespace** is your sub-cluster, e.g., devtron-demo
* **your-registry-server** is your Private Docker Registry FQDN. Use https://index.docker.io/v1/ for Docker Hub.
* **your-name** is your Docker username
* **your-pword** is your Docker password
* **your-email** is your Docker email

You have successfully set your Docker credentials in the cluster as a Secret called `regcred`.

{% hint style="warning" %}
Typing secrets on the command line may store them in your shell history unprotected, and those secrets might also be visible to other users on your PC during the time when kubectl is running.
{% endhint %}

Enter the Secret name in the field and click **Save**.






  
















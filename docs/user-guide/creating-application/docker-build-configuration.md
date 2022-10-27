 # Build Configuration

In this section, we will provide information on the `Build Configuration`.

 Build configuration is used to create and push docker images in the container registry of your application. You will provide all the docker related information to build and push docker images on the `Build Configuration` page.

Only one docker image can be created for multi-git repository applications as explained in the [Git Repository](git-material.md) section.

![](../../.gitbook/assets/create-docker.gif)

For **build configuration**, you must provide information in the sections as given below:

* [Store Container Image](https://docs.devtron.ai/usage/applications/creating-application/docker-build-configuration#store-container-image)
* [Build the Container Image](https://docs.devtron.ai/usage/applications/creating-application/docker-build-configuration#build-the-container-image)
* [Advanced Options](https://docs.devtron.ai/usage/applications/creating-application/docker-build-configuration#advanced-options)

## Store Container Image

The following fields are provided on the **Store Container Image** section:

| Field | Description |
| --- | --- |
| **Container Registry** | Select the container registry from the drop-down list or you can click **Add Container Registry**. This registry will be used to [store docker images](../global-configurations/docker-registries.md). |
| **Container Repository** | Enter the name of your container repository, preferably in the format `username/repo-name`. The repository that you specify here will store a collection of related docker images. Whenever an image is added here, it will be stored with a new tag version. |


**If you are using docker hub account, you need to enter the repository name along with your username. For example - If my username is *kartik579* and repo name is *devtron-trial*, then enter kartik579/devtron-trial instead of only devtron-trial.**

![](../../.gitbook/assets/docker-configuration-docker-hub.png)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/docker-build-configuration/docker-build-config-container-registry1.png)


## Build the Container Image

In order to deploy the application, we must build the docker images to configure a fully operational container environment.

You can choose one of the following options to build your docker image:
* **I have a Dockerfile**
* **Create Dockerfile**
* **Build without Dockerfile**

### Build Docker Image when you have a Dockerfile

A `Dockerfile` is a file that you create which in turn produces a Docker image when you build it.

| Field | Description |
| --- | --- |
| **Select repository containing Dockerfile** | Select the Git checkout path of your repository. This repository is the same which you defined on the [Git Repository](https://docs.devtron.ai/usage/applications/creating-application/git-material) section. |
| **Docker file path (relative)** | Enter a relative file path where your docker file is located in Git repository. Ensure that the dockerfile is available on this path. |

### Build Docker Image by creating Dockerfile

With the option **Create Dockerfile**, you can create a `Dockerfile` from the available templates.
You can edit any selected Dockerfile template as per your build configuration requirements.

| Field | Description |
| --- | --- |
| **Language** | Select the programming language (e.g., `Java`, `Go`, `Python`, `Node` etc.) from the drop-down list you want to create a dockerfile as per compatibility to your system.<br>**Note** We will be adding other programming languages in the future releases.</br>|
| **Framework** | Select the framework (e.g., `Maven`, `Gradle` etc.) of the selected programming language.<br>**Note** We will be adding other frameworks in the future releases.</br>|

### Build Docker Image without Dockerfile

With the option **Create Dockerfile**, you can create a `Dockerfile` from the available templates.
You can edit any selected Dockerfile template as per your build configuration requirements.

| Field | Description |
| --- | --- |
| **Repo containing code you want to build** | Select the Git checkout path of your repository. This repository is the same which you defined on the [Git Repository](https://docs.devtron.ai/usage/applications/creating-application/git-material) section..|
| **Project path (relative)** | Enter a relative project path where your project is located in Git repository. Ensure that the project is available on this path. <br>**Note** For multiple projects, provide a path for the builder.</br>|
| **Language** | Select the programming language (e.g., `Java`, `Go`, `Python`, `Node` etc.) from the drop-down list you want to cbuild your container image as per the compatibility to your system.<br> **Note**: We will be adding other programming languages in the future releases.</br>|
| **Version** | Select the version of the selected programming language. You can also select **Autodetect** to auto-select the compatible version.|
| **Select a builder** | Select a builder which will bring a base build stack (e.g., ubuntu-18, ubuntu-20 ) along with buildpacks compatible to the selected language. <ul><li>**Heroku**</li></ul><ul><li>**GCR**</li></ul><ul><li>**Packeto**</li></ul>|


#### Build Environment Arguments

You can Key/Value pair by clicking **Add Parameter**.

| Field | Description |
| --- | --- |
| **Key** | Define the key parameter for your [docker build](https://docs.docker.com/engine/reference/commandline/build/#options).|
| **Value** | Define the value for the specified key for your [docker build](https://docs.docker.com/engine/reference/commandline/build/#options). |
   

**Note** This fields are optional. If required, it can be overridden at [CI step](../deploying-application/triggering-ci.md).


## Advanced Options

### Set Target Platform for the build

Using this option, you can build images for a specific or multiple **architectures and operating systems (target platforms)**. You can select the target platform from the drop-down list or can type to select a customized target platform.

![Select target platform from drop-down](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/docker-build-configuration/set-target-platform.png)

![Select custom target platform](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/docker-build-configuration/set-target-platform-2.png)

Before selecting a customized target platform, please ensure that the architecture and the operating system are supported by the `registry type` you are using, otherwise build will fail. Devtron uses BuildX to build images for mutiple target Platforms, which requires higher CI worker resources. To allocate more resources, you can increase value of the following parameters in the `devtron-cm` configmap in `devtroncd` namespace.

- LIMIT_CI_CPU 
- REQ_CI_CPU
- REQ_CI_MEM
- LIMIT_CI_MEM

To edit the `devtron-cm` configmap in `devtroncd` namespace:
```
kubectl edit configmap devtron-cm -n devtroncd 
```



If target platform is not set, Devtron will build image for architecture and operating system of the k8s node on which CI is running.

The Target Platform feature might not work in minikube & microk8s clusters as of now.



 Docker build arguments is a collapsed view including
   * Key
   * Value

### Key-value

This field will contain the key parameter and the value for the specified key for your [docker build](https://docs.docker.com/engine/reference/commandline/build/#options). This field is Optional. If required, this can be overridden at [CI step](../deploying-application/triggering-ci.md).


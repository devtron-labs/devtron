 # Docker Build Configuration

In the previous step, we discussed `Git Configurations`. In this section, we will provide information on the `Docker Build Configuration`.

Docker build configuration is used to create and push docker images in the docker registry of your application. You will provide all the docker related information to build and push docker images in this step.

Only one docker image can be created even for multi-git repository applications as explained in the [previous step](git-material.md).

![](../../.gitbook/assets/create-docker.gif)

To add **docker build configuration**, You need to provide three sections as given below:

* **Image store**
* **Checkout path**
* **Advanced**

## Image Store
In Image store section, You need to provide two inputs as given below: 
1. Docker registry
2. Docker repository

### 1. Docker Registry
Select the docker registry that you wish to use. This registry will be used to [store docker images](../global-configurations/docker-registries.md).

### 2. Docker Repository
In this field, add the name of your docker repository. The repository that you specify here will store a collection of related docker images. Whenever an image is added here, it will be stored with a new tag version.

**If you are using docker hub account, you need to enter the repository name along with your username. For example - If my username is *kartik579* and repo name is *devtron-trial*, then enter kartik579/devtron-trial instead of only devtron-trial.**

![](../../.gitbook/assets/docker-configuration-docker-hub.png)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/docker-build-configuration/docker-build-config-1.jpg)


## Checkout path 
Checkout path including inputs:
1. Git checkout path
2. Docker file (relative)

### 1. Git checkout path
In this field, you have to provide the Git checkout path of your repository. This repository is the same that you had defined earlier in git configuration details.

### 2. Docker File Path
Here, you provide a relative path where your docker file is located. Ensure that the dockerfile is present on this path.

## Advanced 
 Docker build arguments is a collapsed view including
   * Key
   * Value

### Key-value
This field will contain the key parameter and the value for the specified key for your [docker build](https://docs.docker.com/engine/reference/commandline/build/#options). This field is Optional. \(If required, this can be overridden at [CI step](../deploying-application/triggering-ci.md) later\)


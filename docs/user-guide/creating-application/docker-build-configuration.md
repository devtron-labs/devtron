# Docker Build Configuration

In the previous step, we discussed git configurations. In this section, we will provide information on the docker build config.

Docker build configuration is used to create and push docker images in the docker registry for your application. You will provide docker related information to build and push docker images in this step.

Only one docker image can be created even for multi-git repository application as explained in the [previous step](git-material.md).

![](../../.gitbook/assets/docker-configuration%20%283%29.gif)

Here you can see, 5 options are present to configure your **docker build**.

1. Repository
2. Docker file path
3. Docker Registry
4. Docker Repository
5. Docker build arguments
   * Key
   * Value

| Options | Description |
| :--- | :--- |
| `Repository` | Provide the checkout path of the repository in this column, which you had defined earlier in git configuration details |
| `Docker File Path` | Provide a relative path for your docker file. Dockerfile should be present on this path. |
| `Docker Registry` | Select the docker registry you wish to use, which will be used to [store docker images](../global-configurations/docker-registries.md). |
| `Docker Repository` | Name of your docker repository that will store a collection of related images. Every image is stored with a new tag version. |
| `Key-value` | The key parameter and the value for a given key for your docker build. This is Optional. \(this can be overridden at [CI step](../deploying-application/triggering-ci.md) later\) |


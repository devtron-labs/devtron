# Application Groups

`Application Groups` is a new functionality in Devtron. In the `Application Groups` section, you will find all the applications which are created under the same environment. 
As an example, if you create two applications `app-A` and `app-B` and add them in the same [environment](https://docs.devtron.ai/v/v0.6/global-configurations/cluster-and-environments#add-environment) (e.g., `devtron-demo`), you will find both the applications in the same application group (environment name). 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/application-groups/application-groups.jpg)

| Fields | Description |
| --- | --- |
| **Environment** | The name of the environment. |
| **Namespace** | The name of the namespace. |
| **Cluster** | The name of the cluster. |
| **Applications** | Number of applications created in the same environment. |

You can search by using the `environment` name or you can also filter by the `cluster` name as shown in the above image.

The benefits of using `Application Groups` in Devtron:
- To deploy mulitple microservices.
- To perform bulk CI/CD across multiple applciations under the same environment.

`Note`: You must have a [super admin](https://docs.devtron.ai/v/v0.6/global-configurations/authorization/user-access#role-based-access-levels) permission to access all the applications in the `Application Groups`. Permission users with `admin` can access only the limited applications which they created.
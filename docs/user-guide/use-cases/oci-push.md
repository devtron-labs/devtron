# Push Helm Charts to OCI Registry

## Introduction

Similar to Devtron's [OCI Pull feature](./oci-pull.md) that fetches [helm charts](../../reference/glossary.md#helm-chartspackages) from [Container/OCI registries](../../reference/glossary.md#containeroci-registry), Devtron supports the opposite too, i.e., pushing of helm charts to your OCI registry. 

This is possible through isolated clusters that facilitate virtual deployments. In other words, it generates a helm package that you can use to deploy in clusters not connected to Devtron.

{% hint style="info" %}
Devtron doesn't support pushing helm packages to [chart repositories](../global-configurations/chart-repo.md)
{% endhint %}

**Pre-requisites**

* Helm Chart(s)
* 'Build and Deploy' access or greater (check [role-based access levels](../global-configurations/authorization/user-access.md#role-based-access-levels))
* OCI-compliant Registry (e.g. Docker Hub and [many more](../global-configurations/container-registries.md#supported-registry-providers))

You must [add your OCI registry](../global-configurations/container-registries.md) to Devtron with the `Push helm packages` option enabled. 

---

## Configuring an Isolated Cluster

1. Go to **Global Configurations** → **Clusters & Environments**.

2. Click the **Add Cluster** button on the top-right corner. 

    ![Figure 1: Adding a Cluster](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/oci-push/add-cluster.jpg)

3. Select **Add Virtual Cluster** (2nd option). 

    ![Figure 2: Adding an Isolated Cluster](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/oci-push/adding-cluster.jpg)

4. Add a cluster name (let's say, *demo*) and click **Save Cluster**.

5. Since the newly created cluster has no environments, click **Add Environment**.

6. Add an environment name and namespace. Click **Save**. 

    ![Figure 3: Adding an Environment](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/oci-push/adding-env.jpg)

You have successfully configured an isolated cluster.

![Figure 4: Isolated Cluster Successfully Created](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/oci-push/added-env.jpg)

---

## Configuring Application Workflow

1. In the left sidebar, go to **Applications**.

2. Search and click your application from the list of Devtron Apps. 

    {% hint style="info" %}
    If you haven't created an application already, refer the [Applications section](../applications.md)
    {% endhint %}

3. In your application, go to **App Configuration** → **Workflow Editor**

4. After creating a [CI build pipeline](../creating-application/workflow/ci-pipeline.md) add a [deployment pipeline](../creating-application/workflow/cd-pipeline.md). 

    ![Figure 5: Adding a CD Pipeline](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/oci-push/workflow-editor.jpg)

5. In the **Environment** dropdown, choose the isolated cluster environment you created in the previous section. 

    ![Figure 6: Choosing an Environment](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/oci-push/env-selection.jpg)

6. You have two options to push the generated helm package:
    * **Do not push** - A link to download the helm package will be available when you trigger a deployment. However, it will not push the helm package to the OCI registry.
    * **Push to registry** - This will generate and push the helm package to the OCI registry. Upon selecting this option, you will get two more fields:
        * **Registry** - Choose the OCI registry to which the helm chart package must be pushed. Only those registries that have `Push helm packages` enabled will be shown in the dropdown.
        * **Repository** - Write the repository name in the format `username/repo-name`. You can find the username from your registry provider account.
    
    ![Figure 7: Choosing a Registry](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/oci-push/create-cd2.jpg)

7. Click **Create Pipeline** or **Update Pipeline**.

---

## Building and Deploying the Application

{% hint style="info" %}
Refer [Deploying Application](https://docs.devtron.ai/usage/applications/deploying-application) to know in detail 
{% endhint %}

1. In your application, go to the **Build & Deploy** tab. 

2. In the 'CI Build' step, click **Select Material**.

3. Choose a commit and click **Start Build**.

4. Once the build is successful, click **Select Image** in the 'Deployment' step.

5. Choose the image and click **Deploy to Virtual Env** (this is an optional step if the deployment execution type was set to automatic in Workflow Editor).

---

## Viewing the Generated Helm Package

The generated helm chart package can be viewed in three places:

### App Details Page

![Figure 8: App Details Page](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/oci-push/app-details-page.jpg)

### Deployment History Page

![Figure 9: Deployment Artifacts](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/oci-push/deployment-history-page.jpg)

### Your OCI Registry

![Figure 10a: OCI Registry Page](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/oci-push/pushed-artifacts.jpg)

![Figure 10b: Pushed Helm Chart](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/oci-push/helm-chart.jpg)








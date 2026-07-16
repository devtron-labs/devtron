# Deploy an Image from Azure Container Registry

You can use an existing Azure Container Registry (ACR) with Devtron in either of the following ways:

* [Build an image manually in Devtron and push it to ACR](#build-and-push-an-image-to-acr)
* [Deploy an image that already exists in ACR](#deploy-an-existing-acr-image)

## Build and Push an Image to ACR

Use this method when Devtron should build the image from source code, but the build should start only when a user triggers it.

### Prerequisites

* An ACR repository and credentials that can push images to it.
* A Git repository containing the application source code.

### Steps

1. Go to **Global Configurations** → **Container/OCI Registry** and [add the ACR registry](../global-configurations/container-registries.md#azure). Use the ACR login server, for example, `myregistry.azurecr.io`, and configure [Registry Credential Access](../global-configurations/container-registries.md#registry-credential-access) for the target cluster.
2. In the application's [Build Configuration](../creating-application/docker-build-configuration.md#store-container-image), select the ACR registry and enter the target container repository.
3. In **App Configuration** → **Workflow Editor**, create a **Build and Deploy from Source Code** workflow.
4. In the build stage, set **Trigger Build Pipeline** to **Manually** and create the pipeline.
5. Go to **Build & Deploy**, click **Select Material**, choose the Git commit, and click **Start Build**. Devtron builds the image and pushes it to the configured ACR repository.
6. In the CD pipeline, click **Select Image**, choose the image produced by the build, and click **Deploy**.

## Deploy an Existing ACR Image

Use this method when the image was built outside Devtron and is already available in ACR. Devtron registers the image through an External CI webhook; it does not rebuild or copy the image.

### Prerequisites

* The complete image reference, including its tag, for example, `myregistry.azurecr.io/backend:v1.0.0`.
* Pull access from the target Kubernetes cluster to the ACR registry.

For an AKS cluster that already has pull access to ACR, no additional registry secret is required. Otherwise, create an image pull secret in the namespace used by the Devtron environment:

```bash
kubectl create secret docker-registry acr-pull-secret \
    --namespace <target-namespace> \
    --docker-server=<registry-name>.azurecr.io \
    --docker-username=<username> \
    --docker-password=<password>
```

Add the secret to the application's **Base Deployment Template**:

```yaml
imagePullSecrets:
  - name: acr-pull-secret
```

{% hint style="warning" %}
Avoid entering credentials directly in commands that might be stored in shell history. Use your organization's approved secret-management method where possible.
{% endhint %}

### Create a Manual Deployment Pipeline

1. Create or select an application and configure its **Base Deployment Template**.
2. Go to **App Configuration** → **Workflow Editor** and click **+ New Workflow**.
3. Select **Deploy image from external service**.
4. Select the target environment and set **When do you want to deploy** to **Manual**.
5. Configure the deployment strategy and click **Create Pipeline**.
6. Click **Show webhook details** and select or generate an API token. Only a super-admin can select or generate the token from this page.

### Register the Existing Image

Use the cURL request generated on the **Webhook Details** page, or send the following request with the complete ACR image reference:

```bash
curl --location --request POST \
    'https://<devtron-host>/orchestrator/webhook/ext-ci/<pipeline-id>' \
    --header 'Content-Type: application/json' \
    --header 'api-token: <api-token>' \
    --data-raw '{
        "dockerImage": "myregistry.azurecr.io/backend:v1.0.0"
    }'
```

A successful request adds the image to the manual deployment pipeline. Go to **Build & Deploy**, click **Select Image**, select the registered ACR image, and click **Deploy**.

{% hint style="info" %}
With a manual deployment pipeline, receiving the webhook does not deploy the image automatically. It makes the image available for selection in the CD pipeline.
{% endhint %}

If the deployment reports `ImagePullBackOff`, verify that the image reference is correct and that the target namespace can authenticate to ACR. See [Registry Credential Access](../global-configurations/container-registries.md#registry-credential-access) for more information.

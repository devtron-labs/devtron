# Secrets

Secrets and configmaps both are used to store environment variables but there is one major difference between them: Configmap stores key-values in normal text format while secrets store them in base64 encrypted form. Devtron platform hides the data of secrets for the normal users and it is only visible to the users having edit permission.

Secret objects let you store and manage sensitive information, such as passwords, authentication tokens, and ssh keys. Embedding this information in secrets is safer and more flexible than putting it verbatim in a Pod definition or in a container image.

## Configure Secret

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/creating-applications-secrets-1.jpg)

Click on `Add Secret` to add a new secret.


![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/creating-applications-secrets-2.jpg)

| Key | Description |
| :--- | :--- |
| `Name` | Provide a name to your Secret |
| `Data Type` | Provide the Data Type of your secret. To know about different Data Types available click on [Data Types](secrets.md#data-types) |
| `Data Volume` | Specify if there is a need to add a volume that is accessible to the Containers running in a pod. |
| `Use secrets as Environment Variable` | Select this option if you want to inject Environment Variables in your pods using Secrets. |
| `Use secrets as Data Volume` | Select this option if you want to configure a Data Volume that is accessible to Containers running in a pod. Ensure that you provide a Volume mount path for the same. |
| `Key-Value` | Provide a key and the corresponding value of the provided key. |

## Volume Mount Path

Specify the Volume Mount folder path in `Volume Mount Path`, a path where the data volume needs to be mounted. This volume will be accessible to the Containers running in a pod.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/creating-applications-secrets-3.jpg)

## Sub Path
For multiple files mount at the same location you need to check sub path `bool` field, it will use the file name (key) as sub path. 
Sub Path feature is not applicable in case of external configmap except
AWS Secret Manager, AWS System Manager and Hashi Corp Vault, for these cases `Name (Secret key)` as sub path will be picked up automatically. 

## File Permission
File permission will be provide at the configmap level not on the each key of the configmap. it will take 3 digit standard permission for the file.


Click on `Save Secret` to save the secret.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/creating-applications-secrets-4.jpg)

You can see the Secret is added.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/creating-applications-secrets-5.gif)

## Update Secrets

You can update your secrets anytime later, but you cannot change the name of your secrets. If you want to change your name of secrets then you have to create a new secret.

To update secrets, click on the secret you wish to update.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/creating-applications-secrets-6.jpg)

Click on `Update Secret` to update your secret.

## Delete Secret

You can delete your secret. Click on your secret and click on the `delete sign` to delete your secret.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/creating-applications-secrets-7.jpg)

## Data Types

There are five Data types that you can use to save your secret.

* **Kubernetes Secret**: The secret that you create using Devtron.
* **Kubernetes External Secret**: The secret data of your application is fetched by Devtron externally. Then the Kubernetes External Secret is converted to Kubernetes Secret.
* **AWS Secret Manager**: The secret data of your application is fetched from AWS Secret Manager and then converted to Kubernetes Secret from AWS Secret. 
* **AWS System Manager**: The secret data for your application is fetched from AWS System Secret Manager and all the secrets stored in AWS System Manager are converted to Kubernetes Secret.
* **Hashi Corp Vault**: The secret data for your application is fetched from Hashi Corp Vault and the secrets stored in Hashi Corp Vault are converted to Kubernetes Secret.

*Note: The conversion of secrets from various data types to Kubernetes Secrets is done within Devtron and irrespective of the data type, after conversion, the Pods access `secrets` normally.*

## External Secrets

In some cases, it may be that you already have secrets for your application on some other sources and you want to use that on devtron. External secrets are fetched by devtron externally and then converted to kubernetes secrets. 

### Kubernetes External Secret

The secret that is already created and stored in the environment and being used by devtron externally is referred here as `Kubernetes External Secret`. For this option, devtron will not create any secret by itself but they can be used within the pods. Before adding secret from kubernetes external secret, please make sure that secret with the same name is present in the environment. To add secret from kubernetes external secret, follow the steps mentioned below:

1. Navigate to `Secrets` of the application.
2. Click on `Add Secret` to add a new secret.
3. Select `Kubernetes External Secret` from dropdown of `Data type`.
4. Provide a name to your secret. Devtron will search secret in the environment with the same name that you mention here. 

### AWS Secret Manager

Before adding any external secrets on devtron, `kubernetes-external-secrets` must be installed on the target cluster. Kubernetes External Secrets allows you to use external secret management systems (e.g., AWS Secrets Manager, Hashicorp Vault, etc) to securely add secrets in Kubernetes.

#### Installing kubernetes-external-secrets Using Chart

To install the chart with the release named my-release:

```bash
$ helm install my-release external-secrets/kubernetes-external-secrets
```
To install the chart with AWS IAM Roles for Service Accounts:

```bash
$ helm install my-release external-secrets/kubernetes-external-secrets --set securityContext.fsGroup=65534 --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"='arn:aws:iam::111111111111:role/ROLENAME'
```
#### Adding Secrets From AWS Secret Manager

To add secrets from AWS secret manager, navigate to `Secrets` of the application and follow the steps mentioned below :

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/creating-applications-secrets-8.jpg)

1. Click on `Add Secret` to add a new secret.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/creating-applications-secrets-9.jpg)

2. Select `AWS Secret Manager` from dropdown of `Data type`.

3. Provide a name to your secret.

4. Select how you want to use the secret. You many leave it selected as environment variable and also you may leave `Role ARN` empty.

5. In `Data` section, you will have to provide data in key-value format.

All the required field to pass your data to fetch secrets on devtron are described below :

| Key | Description |
| :--- | :--- |
|`key`| Secret key in backend |
|`name`| Name for this key in the generated secret |
|`property`| Property to extract if secret in backend is a JSON object |
|`isBinary`| Set this to true if configuring an item for a binary file stored else set false |

#### Adding Secrets in AWS Secret Manager

To add secrets in AWS secret manager, do the following steps :

Step 1: Go to AWS secret manager console.
Step 2: Click on `Store a new secret`.
Step 3: Add and save your secret.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/creating-applications-secrets-10.jpg)

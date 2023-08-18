# AWS Secrets Manager

To add secrets from **AWS Secrets Manager**, we need to create a generic Kubernetes secret for AWS authentication.

Create a Kubernetes secret in the namespace in which the application is to be deployed using base64 encoded AWS access-key and secret-access-key. You can use a Devtron generic chart for it.

**Note**: You don't have to create the Kubernetes secret every time you create external secret for the respective namespace.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/aws-secret-generic-chart.jpg)

After creating the generic secret, navigate to `Secrets` section of the application and follow the steps mentioned below :

**1. Click `Add Secret` to add a new secret**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/add-secrets.jpg)

**2. Select `AWS Secret Manager` under `External Secret Operator` (ESO) from the dropdown of `Data type`**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/aws-secret-manager-dropdown.jpg)

**3. Configure the secret**

| Key | Description |
| :--- | :--- |
| `region` | AWS region in which secret is created |
| `accessKeyIDSecretRef.name` | Name of secret created that would be used for authentication|
| `accessKeyIDSecretRef.key` | In generic secret created for AWS authentication, variable name in which base64 encoded AWS access-key is stored |
| `secretAccessKeySecretRef.name` | Name of secret created that would be used for authentication|
| `secretAccessKeySecretRef.key` | In generic secret created for AWS authentication, variable name in which base64 encoded secret-access-key is stored|
| `secretKey` | Key name to store secret |
| `key` | AWS Secrets Manager secret name |
| `property` | AWS Secrets Manager secret key |

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/aws-eso.jpg)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/aws-external-secret.jpg)


**4. Save the secret**


## ESO AWS secrets Manager Setup with Devtron using ClusterSecretsStore

ClusterSecretStore provides a secure and centralized storage solution for managing and accessing sensitive information, such as passwords, API keys, certificates, and other credentials, within a cluster or application environment.

**Requirement:** Devtron deployment template chart version should be 4.17 and above.

To setup ESO AWS secrets manager with Devtron using ClusterSecretsStore, follow the mentined steps:

**1. Create a secret for AWS authentication**

Create a Kubernetes secret in any namespace using base64 encoded AWS access-key and secret-access-key. You can use the devtron generic chart for this.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/aws-secret-generic-chart.jpg)

**2. Create a `ClusterSecretStore`**

Create a `ClusterSecretStore` using the secret created for AWS authentication in step 1.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/clustersecretstore-yaml.jpg)

**3. Create a secret in the application using ESO AWS Secrets Manager**

Go to the application where you want to create an external secret. Navigate to secrets section under application configuration and create a secret using ESO AWS Secrets Manager.


![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/app-secret-clustersecretstore.jpg)

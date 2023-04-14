# AWS Secrets Manager

To add secrets from **AWS Secrets Manager**, we need to create a generic Kubernetes secret for AWS authentication.

Create a Kubernetes secret in the namespace in which the application is to be deployed using base64 encoded AWS access-key and secret-access-key. You can use devtron generic chart for this.

**Note:** You don't have to create the Kubernetes secret everytime you create External Secret for the respective namespace.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/aws-secret-generic-chart.jpg)

After creating the generic secret, navigate to `Secrets` section of the application and follow the steps mentioned below :

**1. Click `Add Secret` to add a new secret.**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/add-secrets.jpg)

**2. Select `AWS Secret Manager` under `External Secret Operator` (ESO) from the dropdown of `Data type`.**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/aws-secret-manager-dropdown.jpg)

**3. Configure secret:**

| Key | Description |
| :--- | :--- |
| `region` | AWS region in which secret is created. |
| `accessKeyIDSecretRef.name` | Name of secret created that would be used for authentication.|
| `accessKeyIDSecretRef.key` | In generic secret created for AWS authentication, variable name in which base64 encoded AWS access-key is stored. |
| `secretAccessKeySecretRef.name` | Name of secret created that would be used for authentication.|
| `secretAccessKeySecretRef.key` | In generic secret created for AWS authentication, variable name in which base64 encoded secret-access-key is stored.|
| `secretKey` | Key name to store secret. |
| `key` | AWS Secrets Manager secret name. |
| `property` | AWS Secrets Manager secret key. |

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/aws-eso.jpg)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/aws-external-secret.jpg)


**4. Save secret.**

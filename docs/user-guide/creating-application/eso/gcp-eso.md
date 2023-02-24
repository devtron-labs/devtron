# GCP Secrets Manager

To add secrets from **Google Secrets Manager**, follow the steps mentioned below :

**1. Go to Google cloud console and create a Service Account.**


![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/google-sa.jpg)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/google-sa-create.jpg)

**2. Assign roles to the service account.**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/google-sa-roles.jpg)

**3.  Add and create a new key.**
        
![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/google-key.jpg)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/google-key-create.jpg)


**5. Create a Kubernetes secret in the namespace in which the application is to be deployed using base64 encoded service account key.**

You can use devtron generic chart for this. 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/gcp-auth-generic.jpg)

**6. After creating the generic secret, navigate to `Secrets` section of the application and click `Add Secret` to add a new secret.**


![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/add-secrets.jpg)

**7. Select `Google Secrets Manager` under `External Secret Operator` (ESO) from the dropdown of `Data type`.**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/gcp-secret-manager-dropdow.jpg)

**8. Configure secret:**

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/gcp-es-configure.jpg)

| Key | Description |
| :--- | :--- |
| `secretAccessKeySecretRef.name` | Name of secret created that would be used for authentication.|
| `secretAccessKeySecretRef.key` | In generic secret created for GCP authentication, variable name in which base64 encoded service account key is stored.|
| `ProjectID` | GCP Project ID where secret is created. |
| `secretKey` | Key name to store secret. |
| `key` | GCP Secrets Manager secret name. |


![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/gcp-secret.jpg)

**9. Save secret.**

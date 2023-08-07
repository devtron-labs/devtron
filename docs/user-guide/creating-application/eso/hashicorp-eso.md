# HashiCorp Vault

To incorporate secrets from **HashiCorp Vault**, you need to create a generic Kubernetes secret that will be used for vault authentication. This involves creating a Kubernetes secret in the specific namespace where your application will be deployed. 
The secret should store the base64-encoded password or token obtained from vault. To simplify the process, you can utilize the Devtron generic chart. An example yaml is given below:

```yaml
apiVersion: v1
kind: Secret
type: Opaque
data:
   token: <vault-password>
metadata:
   name: vault-token
   namespace: <namespace>
```

**Note**: Please note that you don't need to create the Kubernetes secret every time you create an External Secret for the corresponding namespace.

Once you have created the generic secret, follow these steps in the application's Secrets section:

**1. Create a new secret**

To add a new secret to the application, go to the `App Configuration` section of the application. Then, navigate to the left pane and select the `Secrets` option and click the **Add Secret** button.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/hc-add-secret.jpg)

**2. Select `HashiCorp Vault` as the External Secret Operator**

After clicking the **Add Secret** button, select `HashiCorp Vault` from the dropdown menu for the `Data type` option. Provide a name for the secret you are creating, and then proceed to configure the external secret as described in the next step.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/hc-secret-type.jpg)

**3. Configure the secret**

To configure the external secret that will be fetched from HashiCorp Vault for your application, you will need to provide specific details using the following key-value pairs:

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/hc-secret-configure.jpg)


| Key | Description |
| :--- | :--- |
| `vault.server` | Server is the connection address for the Vaultserver, e.g: "https://vault.example.com:8200" |
| `vault.path` | Specify the path where the secret is stored in Vault |
| `tokenSecretRef.name` | Enter the name of the secret that will be used for authentication |
| `tokenSecretRef.key` | Specify the key name within the secret that contains the token |
| `secretKey` | Provide a name for the secret in Kubernetes |
| `key` | Enter the name of the secret in Vault |
| `property` | Specify the key within the Vault secret |

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/tokenSecretRef.jpg)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/hc-eso.jpg)

**4. Save the secret**

After configuring the external secret from HashiCorp Vault, proceed to save the secret by clicking the **Save** button. 

By following the steps mentioned above and configuring these values correctly, you can seamlessly fetch and utilize external secrets from HashiCorp Vault within your application environment by deploying the application.


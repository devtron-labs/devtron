# API Tokens

API tokens are the access tokens for authentication. Instead of using username and password, it can be used for programmatic access to API. It allows users to generate API tokens with the desired access. Only super admin users can generate API tokens and see the generated tokens.

## Generate API Token

To generate API tokens, go to `Global Configurations -> Authorization -> API tokens` and click `Generate New Token`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-1.png)

* Enter a name for the token.
* Add Description. 
* Select an expiration date for the token (7 days, 30 days, 60 days, 90 days, custom and no expiration).


![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-2.png)

* To select a custom expiration date, select `Custom` from the drop-down list. In the adjacent field, you can select your custom expiration date for the API token.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-custom.png)


* You can assign permission to the token either with:

   * **Super admin permission**: To generate a token with super admin permission, select `Super admin permission`.

   * **Specific permissions**: Selecting `Specific permissions` option allows you to generate a token with a specific role for:<ul><li>`Devtron Apps`</li></ul><ul><li>`Helm Apps`</li></ul><ul><li>`Kubernetes Resources`</li></ul><ul><li>`Chart Groups`</li></ul>


![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-specific-permissions-latest1.jpg)
  

* Click `Generate Token`.

A pop-up window will appear on the screen from where you can copy the API token.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-5.png)

## Use API Token

Once Devtron API token has been generated, you can use this token to request Devtron APIs using any API testing tool like Jmeter, Postman, Citrus. Using Postman here as an example.

Open Postman. Enter the request URL with `POST` method and under HEADERS, enter the API token as shown in the image below.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-6.png)


In the `Body` section, provide the API payload as shown below and click `Send`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-7.png)

As soon as you click `Send`, the created application API will be triggered and a new Devtron app will be created as provided in the payload.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-8.png)


## Update API Token

To set a new expiration date or to make changes in permissions assigned to the token, we need to update the API token in Devtron.
To update the API token, click the token name or click on the edit icon.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-9.png)

To set a new expiration date, you can regenerate the API token. Any scripts or applications using this token must be updated. To regenerate a token, click `Regenerate token`. 

A pop-up window will appear on the screen from where you can select a new expiration date.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-10.png)

Select a new expiration date and click `Regenerate token`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-11.png)

This will generate a new token with a new expiration date.

To update API token permissions, give the permissions as you want to and click  `Update Token`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-3.png)


## Delete API Token

To delete an API token, click `delete` icon. Any applications or scripts using this token will no longer be able to access the Devtron API.








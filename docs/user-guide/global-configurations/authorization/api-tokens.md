# API Tokens
API tokens are like ordinary OAuth access tokens. They can be used instead of username and password for programmatic access to API.
API token allows users to generate API tokens with the desired access. Only super admin users can generate tokens and see generated tokens.

## Generate API Token

To generate API tokens, go to global configurations -> Authorizations -> API tokens and click on Generate New Token.
![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-1.png)

1. Enter a name for the token
2. Add Description. 

3. Select an Expiration date for the token(7 days, 30 days, 60 days, 90 days, custom and no expiration) 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-2.png)

To select a custom expiration date, select `Custom` from the drop-down. This will pop-up a calender from where you can select your custom expiration date for the API token.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-12.png) 

4. Assign  permissions to the token. To generate a token with super admin permission, select super admin permission. 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-4.png)

Or select specific permission if you want to generate a token with a specific role over a particular Devtron app or Helm app or chart group.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-3.png)

5. Now click on Generate Token.



A pop-up window will appear over the screen from where you can copy the API token.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-5.png)


## Use API Token

Once devtron api token has been generated, you can use this token to hit devtron apis using any api testing tool like Jmeter, postman, citrus. Using postman here.

Open postman. Enter the request URL with POST method and under HEADERS, enter the API token as shown in the image below.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-6.png)


Now, under body, provide the api payload as shown below and click on Send.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-7.png)

As soon as you click on send, the create application api will be triggered and a new devtron app will be created as you mentioned in the payload.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-8.png)


## Update API Token

To set a new expiration date or to make changes in permissions assigned to the token, we need to update the API token.
To update the API token, click over the token name or click on the edit icon.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-9.png)

To set a new expiration date, you can regenerate the API token. Any scripts or applications using this token will need to be updated. To regenerate a token, click on regenerate token. A pop-up window will appear on the screen from where you can select a new expiration date and then click on `regenerate token`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-10.png)

Select a new expiration date and click on regenerated token.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-11.png)

This will generated a new token with new expiration date.


To update API token permissions, give the permissions as you want to and click on update token.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-3.png)



## Delete API Token

To delete an API token, click on the delete icon. Any applications or scripts using this token will no longer be able to access the Devtron API.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/api-token/api-token-14.png)






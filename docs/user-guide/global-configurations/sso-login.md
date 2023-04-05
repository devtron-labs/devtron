# SSO Login Services

Once Devtron is installed, it has a built-in `admin` user with super admin privileges with unrestricted access to all Devtron resources. We recommended to use a user with super admin privileges for initial and global configurations only and then switch to local users or configure SSO integration.

Only users with [super-admin](https://docs.devtron.ai/v/v0.6/global-configurations/authorization/user-access#role-based-access-levels) privileges can create SSO configuration. Devtron uses [Dex](https://dexidp.io/docs/connectors/google/) for authenticating a user against the identity provider.


To add/edit SSO configuration, go to the `SSO Login Services` section of `Global Configurations`. 

## Supported SSO Providers

Below are the SSO providers which are available in Devtron. Select one of the SSO providers (e.g., GitHub) to configure SSO:

`Google` 
`GitHub` 
`GitLab`
`Microsoft`
`LDAP`  
`OpenID Connect`
`OpenShift` 

Dex implements connectors that target specific `identity providers` for each connector configuration. You must have a created account for the corresponding identity provider and registered an app for client key and secret.

Refer the following documents for more detail.
* https://dexidp.io/docs/connectors/
* https://dexidp.io/docs/connectors/google/


### 1. Create New SSO Configuration

Make sure that you have a [super admin access](https://docs.devtron.ai/global-configurations/authorization/user-access#assign-super-admin-permission).

* Go to the `Global Configurations` -&gt; `SSO Login Services` and click any `SSO Provider` of your choice.
* In the `URL` field, enter the valid Devtron application `URL` where it is hosted.
* For providing `redirectURI` or `callbackURI` registered with the SSO provider, you can either select `Configuration` or `Sample Script`.
* Provide the `client ID` and `client Secret` of your SSO provider (e.g. If you select `Google` as SSO provider, then you must enter `$GOOGLE_CLIENT_ID` and `$GOOGLE_CLIENT_SECRET` in the `client ID` and `client Secret` respectively.)
* Select `Save` to create and activate SSO Login Service.


**Note**: 
* Only single SSO login configuration can be active at one time. Whenever you create or update any SSO configuration, it will be activated and used by Devtron and previous configurations will be deleted.
* Except for the domain substring, URL and redirectURI remains same.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/sso-login-services.jpg)


### 2. Update SSO Configuration

You can change SSO configuration anytime by updating the configuration and click `Update`.
**Note**: In case of configuration change, all users will be logged out of Devtron and will have to login again.


### 3. Configuration Payload

* `type` : Any platform name such as (Google, GitLab, GitHub etc.) 
* `name` : Identity provider platform name 
* `id` : Identity provider platform which is a unique ID in string. (Refer to [dexidp.io](https://dexidp.io/)
* `config` : User can put connector details for this key. Platforms may not have same structure but common configurations are `clientID`, `clientSecret`, `redirectURI`.
* `hostedDomains` : Domains authorized for SSO login.
 
 ### 3. Google SSO Integration In Case Of VPN

 If you want configure SSO login for the Devtron dashboard that is restricted through the VPN, then you need to whitelist the VPN IP under `loadBalancerSourceRanges:` either way i.e. by editing the `devtron-service` or by whitelisting the IP from the `ingress controller`. 

![Whitelist VPN IP in devtron-service](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/vpn-devtron-service.jpg)

![Whitelist VPN IP in ingress controller](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/vpn-ingress.jpg)
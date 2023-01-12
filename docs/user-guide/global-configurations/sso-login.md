# SSO LOGIN 

Once Devtron is installed, it has a built-in `admin` user with super admin privileges with unrestricted access to all Devtron resources. We recommended to use a user with super adminn privileges only for initial and global configurations and then switch to local users or configure SSO integration.

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


### 1. Create new SSO Configuration

Login as a user with super admin privileges and go to `Global Configurations` -&gt; `SSO Login Services` and click any `Identity Provider` and fill the configuration. 

Add valid devtron application `URL` where it is hosted.

Fill correct `redirect URL` or `callback URL` from which you have registered with the identity provider in the previous step along with the `client id` and `client secret` shared by the identity provider.

Only single SSO login configuration can be active at one time. Whenever you create or update any SSO config, it will be activated and used by the system and previous configurations will be deleted.

Except for the domain substring, URL and redirectURI should be the same as in the screenshots.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/sso-login-services.jpg)

Select `Save` to create and activate SSO login.

### 2. Update SSO Configuration

You can change SSO configuration anytime by updating the configuration and click `Save`.
**Note**: In case of configuration change, all users will be logged out of the system and will have to login again.

### 3. Configuration Payload

* `type` : oidc or any platform name such as (google, gitlab, github etc) 
* `name` : identity provider platform name 
* `id` : identity provider platform unique id in string. (refer to dexidp.io)
* `config` : user can put connector details into this key. platforms may not have same structure but commons are clientID, clientSecret, redirectURI.
* `hostedDomains` : domains authorized for SSO login.

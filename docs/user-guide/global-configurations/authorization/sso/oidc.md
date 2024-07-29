# OIDC

## Sample Configuration

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/oidc.jpg)

---

## Values You Would Require at SSO Provider

Devtron provides a sample configuration out of the box. There are some values that you need to either get from your SSO provider or give to your SSO provider.

### Values to Fetch

* clientID

* clientSecret

    ![Fetching Client ID and Secret](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/secret/oidc-id-secret.jpg)

### Values to Provide

* redirectURI (provided in SSO Login Services by Devtron)

    ![Copying Redirect URI from Devtron](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/redirect/oidc-redurl.jpg)

    ![Pasting Redirect URI](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/redirect/oidc-redirect.jpg)

---

## Reference

* [View Okta Documentation](https://developer.okta.com/docs/guides/find-your-app-credentials/main/)

* [Configure Keycloak SSO](../sso/keycloak.md)

* [Configure Okta SSO](../sso/okta.md)

* [View Dex IdP Documentation](https://dexidp.io/docs/connectors/oidc/)













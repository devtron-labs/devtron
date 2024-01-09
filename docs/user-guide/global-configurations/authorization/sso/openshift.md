# Openshift

## Sample Configuration

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/openshift.jpg)

---

## Values You Would Require at SSO Provider

Devtron provides a sample configuration out of the box. There are some values that you need to either get from your SSO provider or give to your SSO provider.

### Values to Fetch

* clientID

    ![Fetching Client ID](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/secret/openshift-id-secret1.jpg)

* clientSecret

    ![Fetching Secret](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/secret/openshift-id-secret2.jpg)

### Values to Provide

* redirectURI (already provided in SSO Login Services by Devtron)

    ![Copying Redirect URI from Devtron](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/redirect/openshift-redurl.jpg)

    ![Pasting Redirect URI](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/redirect/openshift-redirect.jpg)

---

## Reference

* [View Openshift Documentation](https://docs.openshift.com/container-platform/4.14/authentication/configuring-oauth-clients.html)

* [View Dex IdP Documentation](https://dexidp.io/docs/connectors/openshift/)








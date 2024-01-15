# Microsoft

## Sample Configuration

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/microsoft.jpg)

---

## Values You Would Require at SSO Provider

Devtron provides a sample configuration out of the box. There are some values that you need to either get from your SSO provider or give to your SSO provider.

### Values to Fetch

* clientID

* tenantID (required only if you want to use Azure AD for auto-assigning permissions)

    ![Fetching Client ID and Tenant ID](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/secret/ms-id-secret1.jpg)

* clientSecret

    ![Fetching Secret](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/secret/ms-id-secret2.jpg)

### Values to Provide

* redirectURI (provided in SSO Login Services by Devtron)

    ![Copying Redirect URI from Devtron](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/redirect/microsoft-redurl.jpg)

    ![Pasting Redirect URI](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/redirect/ms-redirect.jpg)

---

## Reference

* [View Microsoft Documentation](https://learn.microsoft.com/en-us/entra/identity-platform/quickstart-register-app)

* [View Dex IdP Documentation](https://dexidp.io/docs/connectors/microsoft/)

---

## Auto-assign Permissions [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

{% hint style="info" %}
Make sure to add tenantID in the SSO configuration field without fail.
{% endhint %}

Since Microsoft supports <a href="https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/get-started/virtual-dc/active-directory-domain-services-overview" target="_blank">Active Directory (AD)</a>
, this feature further simplifies the onboarding process of organizations having a large headcount of users. It also eliminates repetitive permission assignment by automatically mapping your Azure AD groups to Devtron's [Permission Groups](../permission-groups.md) during single sign-on (SSO) login.

![Enabling Permission Auto-assignment](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/secret/auto-grant.jpg)

If you've defined groups in your Active Directory, you can create corresponding permission groups in Devtron with the same names. When members of those Active Directory groups first log in to Devtron, they'll automatically inherit the permissions from their Devtron permission group. This means you can't manually adjust or add [individual permissions for users](../user-access.md) mapped to a permission group.

{% hint style="warning" %}
SSO login requires exact matching between Devtron permission group names and AD groups. Any discrepancies or missing groups will prevent successful login.

Once you save the configuration with this feature enabled, existing user permissions will be cleared and the future permissions will be managed through [permission groups](../permission-groups.md) linked to Azure Active Directory (Microsoft Entra ID) groups.
{% endhint %}

{% hint style="info" %}
If your AD permissions aren't reflecting in Devtron, a quick sign-out and sign-in can resolve the issue.
{% endhint %}







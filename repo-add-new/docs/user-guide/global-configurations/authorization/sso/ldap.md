# LDAP

## Sample Configuration

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/ldap.jpg)

---

## Values to fetch from LDAP

Devtron provides a sample configuration out of the box. Here are some values you need to fetch from your LDAP.

* bindDN
* bindPW
* baseDN

---

## Reference

[What is LDAP](https://www.okta.com/identity-101/what-is-ldap/)

---

## Auto-assign Permissions [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

Since LDAP supports creation of User Groups, this feature simplifies the onboarding process of organizations having a large headcount of users. It also eliminates repetitive permission assignment by automatically mapping your LDAP User groups to Devtron's [Permission Groups](../permission-groups.md) during single sign-on (SSO) login.

![Enabling Permission Auto-assignment](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/secret/auto-grant-ldap.jpg)

If you've created user groups in LDAP, you can create corresponding permission groups in Devtron with the same names. When members of those user groups first log in to Devtron, they'll automatically inherit the permissions from their Devtron permission group. This means you can't manually adjust or add [individual permissions for users](../user-access.md) mapped to a permission group.

{% hint style="warning" %}
SSO login requires exact matching between Devtron permission group names and LDAP user groups. Any discrepancies or missing groups will prevent successful login.

Once you save the configuration with this auto-assign feature enabled, existing user permissions will be cleared and the future permissions will be managed through [Permission Groups](../permission-groups.md) linked to LDAP user groups.
{% endhint %}

{% hint style="info" %}
If you're missing some permissions that you know you should have, try logging out and signing back in to Devtron. This will refresh your permissions based on your latest LDAP user group.
{% endhint %}










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

{% hint style="info" %}
Only a super-admin can enable this feature.
{% endhint %}

Since LDAP supports creation of User Groups, this feature further simplifies the onboarding process of organizations having a large headcount of users. It also eliminates repetitive permission assignment by automatically mapping your LDAP User groups to Devtron's [Permission Groups](../permission-groups.md) during single sign-on (SSO) login.

![Enabling Permission Auto-assignment](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/secret/auto-grant-ldap.jpg)

In other words, if you have created user groups in LDAP representing different roles or access levels, you can use the same group name to create a permission group in Devtron. When users belonging to a specific user group of LDAP log in to Devtron for the first time, they will automatically get the pre-decided permissions granted to them. This also means that you cannot change or add separate [permissions for a user](../user-access.md) mapped to a permission group.

{% hint style="warning" %}
SSO login will fail if there's any mismatch between the name of permission group created in Devtron with the user group present in LDAP (or if the group is absent altogether).

Once you save the configuration with this auto-assign feature enabled, existing user permissions will be cleared and the future permissions will be managed through [Permission Groups](../permission-groups.md) linked to LDAP user groups.
{% endhint %}

{% hint style="info" %}
Fetching and updating of permissions occurs during login. Therefore, if you have some permissions at your LDAP user group that aren't reflecting in Devtron, you may have to sign out and sign in to Devtron again.
{% endhint %}










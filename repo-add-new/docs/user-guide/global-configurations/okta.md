# Example - Okta SSO

## Prerequisites

A verified account on [Okta](https://www.okta.com/). Okta activates your account only if email verification is successful.

Here's a reference guide to set up your Okta org and application: [Link](https://developer.okta.com/docs/guides/oie-embedded-common-org-setup/go/main/)

## Tutorial

{% embed url="https://www.youtube.com/watch?v=i-7IWkg6Ipk" caption="Okta App Setup" %}

## Steps on Okta Admin Console

Once your Okta org is set up, create an app integration on Okta to get a Client ID and Client Secret.

1. In the Admin Console, go to **Applications** → **Applications**.

2. Click **Create App Integration**.

3. Select **OIDC - OpenID Connect** as the **Sign-in method**.

{% hint style="info" %}
OIDC stands for OpenID Connect. [Click here](https://www.okta.com/openid-connect/) to read more.
{% endhint %}

4. Select **Web** as the application type and click **Next**.

5. On the **App Integration** page:
    * Give a name to your application.
    * Select the **Interaction Code** and **Refresh Token** checkbox.
    * Now go to Devtron's Global Configurations → SSO Login Services → OIDC.
    * Copy the redirect URI given in the helper text (might look like: https://xxx.xxx.xxx/xxx/callback).
    * Return to the Okta screen, and remove the prefilled value in **Sign-in redirect URIs**.
    * Paste the copied URI in **Sign-in redirect URIs**.
    * Click **Save**.

6. On the **General** tab:
    * Note the **Client ID** value.
    * Click the **Edit** option.
    * In Client Authentication, choose **Client Secret**.
    * Click **Save**.
    * Click **Generate new secret**.
    * Note the **Client Secret** value.


## Steps on Devtron

1. Go to the Global Configurations → SSO Login Services → OIDC.
2. In the **URL** field, enter the Devtron application URL (a valid https link) where it is hosted.
3. Under `Configuration` tab, locate the config object, and provide the `clientID` and `clientSecret` of the app integration you created on Okta.
4. Add a key `insecureSkipEmailVerified: true`. Note that this key is only required for Okta SSO. For other types of OIDC SSO, refer [OIDC supported configurations](https://dexidp.io/docs/connectors/oidc/).
5. Provide `issuer` value as `https://${yourOktaDomain}`. Replace `${yourOktaDomain}` with your domain on Okta as shown in the video.
6. For providing `redirectURI` or `callbackURI` registered with the SSO provider, you can either select `Configuration` or `Sample Script`. Note that the redirect URI is already given in the helper text (as seen in the previous section).
7. Click **Save** to create and activate Okta SSO login.

Now your users will be able to log in to Devtron using the Okta authentication method. Note that existing signed-in users will be logged out and they have to log in again using their OIDC account.

## Sample Configuration

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/sample-config-okta.jpg)



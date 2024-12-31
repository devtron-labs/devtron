# Keycloak

## Prerequisites

* Install and [configure Keycloak](https://www.keycloak.org/guides#getting-started) on your server or cloud environment.
* Create a new [realm in Keycloak](https://www.keycloak.org/getting-started/getting-started-kube#_create_a_realm) for your application.

---

## Steps on Keycloak Admin Console

### Creating a Client

Here, we will add Devtron as a client for using Keycloak SSO.

1. In the Admin Console, go to **Clients** and click **Create client**.

    ![Figure 1: Creating Client on Keycloak](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/keycloak/create-client.gif)

2. Within **General Settings**:
    * Enter `devtron` in the **Client ID** field. We will use this ID while configuring SSO later in Devtron.
    * Enter `Devtron` in the **Name** field.

    ![Figure 2: Client ID and Name](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/keycloak/client-id.jpg)

3. Within **Capability config**, turn on **Client Authentication**.

    ![Figure 3: Enabling Client Authentication Toggle](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/keycloak/enable-client-auth.gif)


4. Within **Login settings**, enter `https://<DEVTRON_BASE_URL>/orchestrator/api/dex/callback` in the following fields. 
    * **Valid redirect URIs**
    * **Valid post logout redirect URIs**
    * **Web origins**

    [Click here](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/keycloak/base-url.jpg) to know where to find `DEVTRON_BASE_URL`.

    ![Figure 4: Entering Callback/Redirect URIs](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/keycloak/redirect-url.gif)

5. Click **Save**.

### Getting Client Secret

Here, we will obtain the secret we need while configuring SSO in Devtron.

1. Go to the **Credentials** tab of the client you created.

    ![Figure 5: Obtaining Client Secret](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/keycloak/client-secret.gif)

2. Use the copy button next to the **Client Secret** field and paste it somewhere for future reference. 

### Creating Users

Here, we will create a user that can log in to Devtron via SSO. We will assign a username and password that the user can enter while logging in to Devtron via Keycloak SSO.

1. In the Admin Console, go to **Users** and click **Add user**.

2. Give a username (e.g., *usertest*) in the **Username** field and enter the user's email address (e.g., *usertest@example.com*) in the **Email** field.

    ![Figure 6: Creating User Data](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/keycloak/create-user.gif)

3. Click **Create**. Your user creation will be successful.

4. Go to the **Credentials** tab of the user you created.

    ![Figure 7: Adding User Password](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/keycloak/set-user-password.gif)

5. Click **Set password**.

6. Enter the password and confirm it.

7. Click **Save**.

### Retrieving Issuer URL

Here, we will obtain the Issuer URL we need while configuring SSO in Devtron.

1. In the Admin Console, go to **Realm settings**.

2. In the **General** tab, scroll down to the **Endpoints** field, and click the **OpenID Endpoint Configuration** link.

    ![Figure 8: OpenID Endpoint Configuration Link](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/keycloak/endpoint-config.gif)

3. This will open a new page, copy the value of the key named `issuer`, and paste it somewhere for future reference. 

    ![Figure 9: Locating Issuer URL](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/keycloak/issuer-url.jpg)

---

## Steps on Devtron

### Configuring OIDC SSO

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to configure SSO.
{% endhint %}

Here, we will set up an OIDC SSO and enter the values we obtained in the [previous section](#steps-on-keycloak-admin-console).

1. Go to **Global Configurations** → **SSO Login Services** → **OIDC**.

    ![Figure 10: Choosing OIDC SSO](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/keycloak/oidc.jpg)

2. Below the URL field, take the help of the **Click to use** option to populate the exact URL if the displayed one is incorrect.

    ![Figure 11: Populating Correct Orchestrator URL](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/keycloak/click-to-use.jpg)

3. In the **Configuration** editor, do the following:
    * In the `issuer` field, paste the URL you got while [retrieving issuer URL](#retrieving-issuer-url).
    * In the `clientID` field, paste the ID you entered while [creating the client](#creating-a-client).
    * In the `clientSecret` field, paste the secret you got under [client credentials tab](#getting-client-secret).
    * In the `redirectURI` field, make sure to enter the same redirect URI you gave in step 4 of [client creation](#creating-a-client).

    ![Figure 12: Sample Keycloak SSO Config](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/keycloak/keycloak-sso-config.jpg)

4. Click **Save** or **Update** to activate Keycloak SSO login. 

### Adding Users

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to add users.
{% endhint %}

Here, we will add the user we created in the Keycloak Admin Console. If this step is skipped, the user might not be able to log in to Devtron via Keycloak.

1. Go to **Global Configurations** → **Authorization** → **User Permissions**.

2. Click **+ Add Users**.

    ![Figure 13: Adding Users to Devtron](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/keycloak/add-user.jpg)

3. In the **Email addresses** field, enter the email address of the user you created in Keycloak.

    ![Figure 14: Entering User Data and Permissions](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/keycloak/add-email.jpg)

4. Assign necessary permissions to this new user. Refer [user permissions](../user-access.md) to know more.

5. Click **Save**.

Now, you may log out and test the Keycloak OIDC login method using the [user credentials](#creating-users). Clicking the **Login with Oidc** button will land you on Keycloak's login page.

![Figure 15a: Login using OIDC method](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/keycloak/login-oidc.jpg) 


![Figure 15b: Keycloak's Login Page](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/sso-login-service/keycloak/keycloak-login.jpg)

{% hint style="warning" %}
### Note
Kindly get in touch with us if you encounter any issues while logging out of Keycloak on Devtron as it might be buggy.
{% endhint %}
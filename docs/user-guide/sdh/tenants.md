# Tenants

## Introduction [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

This section allows you to add new [tenants](./README.md#tenants) and map environments to these [installations](./README.md#installations) to ensure updates [rollout](./README.md#rollout) correctly. 

---

## Adding Tenant

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to add tenants.
{% endhint %}

This involves the creation of new organizations where you wish to deploy s/w updates. Whenever you are onboarding a new client, you add them as a tenant. 

1. Click **+ Add Tenant**.

    ![Figure 1: Adding a Tenant](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/sdh/tenants-tab.jpg)

2. Enter a name in **Tenant display name** field, e.g., `flareup.xyz`. Once set, you can rename it later (if needed).

3. Add a unique identifier to your tenant in **Tenant ID** field, e.g., `flareup123`. Once set, you cannot change it later.

4. (Optional) Add a description of the tenant.

    ![Figure 2: Saving Tenant Details](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/sdh/add-tenant.jpg)

5. Click **Save**.

---

## Adding Installation

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to add installations to a tenant.
{% endhint %}

This involves setting up [installation](./README.md#installations) for different environments, such as Prod, Development, and QA environments. You can consider these as licenses/installations your client has subscribed for.

1. Click the tenant you created.

    ![Figure 3: Clicking Your Tenant](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/sdh/tenant-list.jpg)

2. Click **+ Add Installation**.

    ![Figure 4: Adding an Installation](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/sdh/add-installation.jpg)

3. Enter a name in **Installation display name** field, e.g., `Flareup Prod`

4. Enter an **Installation ID**, e.g., `flareup-prod-1`

    ![Figure 5: Entering Installation ID](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/sdh/installation-creation.jpg)

5. Click **Save**.

---

## Mapping Environments

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to map environments.
{% endhint %}

This involves mapping customer's environments to the tenant installation so that your updates are deployed to the correct environments. If you haven't created an environment yet, refer [Clusters & Environments](../global-configurations/cluster-and-environments.md#add-environment).

1. Click **Map Environment**.

    ![Figure 6: Map Environment Button](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/sdh/map-environment.jpg)

2. Use the checkbox to choose the environments to map to the tenant installation. Note that, you cannot map an environment that is already mapped to another tenant installation.

    ![Figure 7: Mapping Environment to Your Installation](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/sdh/map-environments.jpg)

    Here, we have mapped `doc1` and `doc2` environments to the production installation. 

3. Click **Save**

Next, you need to set up a release on [Release Hub](./release-hub.md). If you have correctly mapped the customer's environments to an installation, and if you choose applications that already have those environments (say `doc1` or `doc2` or both) in their deployment pipeline, you can [rollout](./README.md#rollout) your release.
# Pull Helm Charts from OCI Registry

## Introduction

Devtron supports the installation of [Helm charts](../../reference/glossary.md#helm-chartspackages) from both: Helm [repos](../../reference/glossary.md#repo) and [Container/OCI registries](../../reference/glossary.md#containeroci-registry). Unlike Helm repos, OCI registries do not have an index file to discover all the charts. However, Devtron makes it easier for you to populate your charts from multiple sources to the [chart store](../../reference/glossary.md#chart-store).

**Prerequisites**

* Helm Chart(s)
* OCI-compliant Registry (e.g. Docker Hub and [many more](../global-configurations/container-registries.md#supported-registry-providers))

You must [add your OCI registry](../global-configurations/container-registries.md) to Devtron with the `Use as chart repository` option enabled. 

---

## Tutorial

{% embed url="https://www.youtube.com/watch?v=9imC5MMz9gs" caption="Pulling Charts from an OCI Registry to Devtron" %}

---

## Populating your Charts to the Chart Store

1. Go to **Global Configurations** → **Container/OCI Registry**.

2. Search your OCI registry in the list, and click it.

3. In the **List of repositories**, add the chart repo(s). The format should be `username/chartname`. You can find the username from your registry provider account.

    ![Figure 1: Adding Chart Repos](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/oci-pull/container-config.jpg)

4. Click **Save** or **Update**.

5. From the left sidebar, go to **Chart Store**.

6. You can find your chart(s) either by using the search bar or by selecting your chart source.

    ![Figure 2: Searching your Chart](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/oci-pull/chart-search.jpg)

You have successfully pulled your charts to the chart store.

![Figure 3: Uploaded Helm Charts](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/oci-pull/chart-list.jpg)

### Unable to find your Charts?

Deprecated charts won't show up in the Chart Store unless you enable the **Show deprecated charts** filter as shown below

![Figure 4: Checking Deprecated Charts](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/oci-pull/deprecated.jpg)

Or, you may try performing a resync as shown below

![Figure 5: Performing a Resync](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/oci-pull/chart-sync.jpg)

---


## Removing your Chart from the Chart Store

1. Go to your OCI registry settings in Devtron.

2. In the **List of repositories** field, remove the unwanted chart repo.

    ![Figure 6: Removing a Chart Repo](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/oci-pull/remove-chart-repo.jpg)

3. Click **Update**.

The removed chart would no longer appear in the Chart Store.

{% hint style="info" %}
Deleting a chart repo from your OCI registry will not lead to the removal of chart from the Chart Store
{% endhint %}



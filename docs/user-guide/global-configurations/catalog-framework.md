# Catalog Framework

## Introduction [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

Ideally, a microservice/application should contain detailed information so that its users know what it does, how to use it, as well as all its technical specs. Access to such data makes it easier for engineers to quickly discover relevant microservices and also its dependent services. 

With Catalog Framework, you as a [super-admin](../global-configurations/authorization/user-access.md#role-based-access-levels) can decide the data you expect from the managers of different resource types. In other words, you can create a whole JSON schema that would ultimately render a form for the resource owners to fill.

Therefore, there are two parts involved in the creation of a desirable resource catalog:

* [Defining a Schema](#defining-a-schema) 
* [Filling the Details](#filling-the-details)

Currently, Devtron supports catalog framework for the following resource types (a.k.a. resource kind):

* [Devtron applications](../../reference/glossary.md#devtron-apps)
* [Helm applications](../../reference/glossary.md#helm-apps)
* [Clusters](../../reference/glossary.md#cluster)
* [Jobs](../../reference/glossary.md#job)

---

## Defining a Schema

Only a super-admin can create/edit a schema.

1. Go to **Global Configurations** → **Catalog Framework**.

2. Choose a resource type, for which you wish to define a schema, for e.g., Devtron applications.

    ![Figure 1: Choosing a Schema](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/catalog-framework.jpg)

3. You can edit the schema name and description.

4. There is a sample schema available for you to create your own customized schema. Using this schema, you can decide the input types that renders within the form, for e.g., a dropdown of enum values, a boolean toggle button, text field, label, and many more.

    ![Figure 2a: Using Sample Schema](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/schema.jpg)

    ![Figure 2b: Expected Future Output](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/rendering.jpg)

5. After defining your schema, click **Review Changes**.

6. You get a side-by-side comparison (diff) highlighting the changes you made.

    ![Figure 3: Change Diff](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/changed-schema.jpg)

7. Click **Save**.

Similarly, you can define schemas for other resource types.

{% hint style="warning" %}
If you edit a field (within an existing schema) for which users have already filled the data, that data will be erased. You will get a prompt (as shown below) confirming whether you want to proceed with your changes.

{% endhint %}

---

## Filling the Details

Once a catalog schema exists for a resource type, its corresponding form would be available in the overview section of that resource type. 

1. Since we defined a schema for Devtron applications in the above example, go to the **Overview** tab of your application (any Devtron application). Click the **Edit** button within the `About` section.

    ![Figure 4: Unfilled Details](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/app-overview.jpg)

2. The schema created for Devtron applications would render into an empty form as shown below.

    ![Figure 5: Rendered Empty Form](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/edit-catalog.jpg)

3. Fill as many details as an application owner to the best of your knowledge and click **Save**.

    ![Figure 6: Filled Form](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/filled-catalog.jpg)

4. Your saved data would be visible in a GUI format (and also in JSON format) as shown below.

    ![Figure 7: App Catalog Data](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gui-app-catalog.jpg)

This catalog data would be visible to all the users who have access to the application, but its data can be edited only by the resource owners (in this case, app admin/managers).








# Catalog Framework

## Introduction [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

Ideally, all resources such as microservices, clusters, jobs, pods, etc. should contain detailed information so that its users know what each of those resources do, how to use them, as well as all their technical specs. Access to such data makes it easier for engineers to quickly discover and understand the relevant resources.

To achieve this, Devtron supports a feature known as **Catalog Framework**. Using this, you as a [super-admin](../global-configurations/authorization/user-access.md#role-based-access-levels) can decide the data you expect from the managers of different resource types. In other words, you can create a custom JSON schema that would ultimately render a form for the resource owners to fill. Once the form is filled, a GUI output will appear as shown below.

![Sample Catalog Data for an App](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/catalog-framework/sample-app-catalog.jpg)

Currently, Devtron supports catalog framework for the following resource types (a.k.a. resource kind):

* [Devtron applications](../../reference/glossary.md#devtron-apps)
* [Helm applications](../../reference/glossary.md#helm-apps)
* [Clusters](../../reference/glossary.md#cluster)
* [Jobs](../../reference/glossary.md#job)

There are two parts involved in the creation of a desirable resource catalog:

1. [Defining a Schema](#defining-a-schema) 
2. [Filling the Schema-Generated Form](#filling-the-schema-generated-form)

---

## Defining a Schema

Only a super-admin can create/edit a schema.

1. Go to **Global Configurations** → **Catalog Framework**.

2. Choose a resource type, for which you wish to define a schema, for e.g., Devtron applications.

    ![Figure 1: Choosing a Schema](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/catalog-framework/catalog-framework.jpg)

3. You can edit the schema name and description.

4. There is a sample schema available for you to create your own customized schema. Using this schema, you can decide the input types that renders within the form, for e.g., a dropdown of enum values, a boolean toggle button, text field, label, and many more.

    ![Figure 2a: Using Sample Schema](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/catalog-framework/schema.jpg)

    ![Figure 2b: Expected Future Output](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/catalog-framework/rendering.jpg)

5. After defining your schema, click **Review Changes**.

6. You get a side-by-side comparison (diff) highlighting the changes you made.

    ![Figure 3: Change Diff](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/catalog-framework/changed-schema.jpg)

7. Click **Save**.

Similarly, you can define schemas for other resource types.

**Note**: If you edit a field (within an existing schema) for which users have already filled the data, that data will be erased. You will receive a prompt (as shown below) to confirm whether you want to proceed with the changes.

![Figure 4: Indication of Existing Data](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/catalog-framework/existing-filled-data.jpg)


---

## Filling the Schema-Generated Form

Once a catalog schema exists for a resource type, its corresponding form would be available in the overview section of that resource type. 

1. Since we defined a schema for Devtron applications in the above example, go to the **Overview** tab of your application (any Devtron application). Click the **Edit** button within the `About` section.

    ![Figure 5: Unfilled Details](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/catalog-framework/app-overview.jpg)

2. The schema created for Devtron applications would render into an empty form as shown below.

    ![Figure 6: Rendered Empty Form](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/catalog-framework/edit-catalog.jpg)

3. Fill as many details as an application owner to the best of your knowledge and click **Save**.

    ![Figure 7: Filled Form](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/catalog-framework/filled-catalog.jpg)

4. Your saved data would be visible in a GUI format (and also in JSON format) as shown below.

    ![Figure 8: App Catalog Data](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/catalog-framework/gui-app-catalog.jpg)

This catalog data would be visible to all the users who have access to the application, but its data can be edited only by the resource owners (in this case, app admin/managers).








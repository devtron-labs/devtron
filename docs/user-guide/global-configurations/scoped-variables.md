# Using Scoped Variables in Devtron

## Introduction

In any piece of software or code, variables are used for holding data such as numbers or strings. Variables are created by declaring them, which involves specifying the variable's name and type, followed by assigning it a value.

Devtron offers super-admins the capability to define scoped variables (key-value pairs). It means, while the key remains the same, its value can change depending on the following context: 

* **Global**: Variable values will be universally same for everybody in Devtron.
* **Cluster**: Variable values might differ for each Kubernetes cluster. [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)
* **Environment**: Variable values might differ for each environment within a cluster, e.g., staging, dev, prod. [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)
* **Application**: Variable values might differ for each application. [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)
* **Environment + Application**: Variable values might differ for each application on a specific environment. [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

**Advantages of using scoped variables**

* **Reduces repeatability**: Configuration management team can centrally maintain the static data.
* **Simplifies bulk edits**: All the places that use a variable get updated when you change the value of the variable.
* **Keeps data secure**: You can decide the exposure of a variable's value to prevent misuse or leakage of sensitive data.

---

## How to Define a Scoped Variable

On Devtron, a super-admin can download a YAML template. It will contain a schema for defining the variables.

### Download the Template

1. From the left sidebar, go to **Global Configurations** → **Scoped Variables**

2. Click **Download template**. 

    ![Figure 1: Downloading the Template](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/scoped-variables/template-downloader.jpg)

3. Open the downloaded template using any code editor (say VS Code). 

### Enter the Values

The YAML file contains key-value pairs that follow the below schema:

| Field                    | Type    | Description                                                                     |
| ------------------------ | ------- | ------------------------------------------------------------------------------- |
| `apiVersion`             | string  | The API version of the resource (comes pre-filled)                              |
| `kind`                   | string  | The kind of resource (i.e. Variable, comes pre-filled)                          |
| `spec`                   | object  | The complete specification object containing all the variables                  |
| `spec.name`              | string  | Unique name of the variable, e.g. *DB_URL*                                      |
| `spec.shortDescription`  | string  | A short description of the variable (up to 120 characters)                      |
| `spec.notes`             | string  | Additional details about the variable (will not be shown on UI)                 |
| `spec.isSensitive`       | boolean | Whether the variable value is confidential (will not be shown on UI if true)    |
| `spec.values`            | array   | The complete values object containing all the variable values as per context    |

The `spec.values` array further contains the following elements:

| Field                    | Type    | Description                                                                                          |
| ------------------------ | ------- | ---------------------------------------------------------------------------------------------------- |
| `category`                                      | string | The context, e.g., Global, Cluster, Application, Env, ApplicationEnv           |
| `value`                                         | string | The value of the variable                                                      |
| `selectors`                                     | object | A set of selectors that restrict the scope of the variable                     |
| `selectors.attributeSelectors`                  | object | A map of attribute selectors to values                                         |
| `selectors.attributeSelectors.<selector_key>`   | string | The key of the attribute selector, e.g., *ApplicationName*, *EnvName*, *ClusterName* |
| `selectors.attributeSelectors.<selector_value>` | string | The value of the attribute selector                                            |


Here's a truncated template containing the specification of two variables for your understanding:

    apiVersion: devtron.ai/v1beta1
    kind: Variable
    spec:

    # First example of a variable
    - name: DB_URL 
    shortDescription: My application's customers are stored
    notes: The DB is a MySQL DB running version 7.0. The DB contains confidential information.
    isSensitive: true 
    values: 
        - category: Global 
        value: mysql.example.com 

    # Second example of a variable
    - name: DB_Name
    shortDescription: My database name to recognize the DB
    notes: NA
    isSensitive: false 
    values: 
        - category: Global 
        value: Devtron 
        - category: ApplicationEnv 
        value: app1-p
        selectors:
            attributeSelectors:
            ApplicationName: MyFirstApplication
            EnvName: prod

### Upload the Template

1. Once you save the YAML file, go back to the screen where you downloaded the template.

2. Use the file uploader utility to upload your YAML file.

    ![Figure 2: Uploading the Template](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/scoped-variables/file-uploader.jpg)

3. The content of the file will be uploaded for you to review and edit. Click **Review Changes**.

    ![Figure 3: Reviewing the YAML file](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/scoped-variables/file-preview.jpg)

4. You may check the changes between the last saved file and the current one before clicking **Save**.

    ![Figure 4: Saving the file](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/scoped-variables/save.jpg)

5. Click the **Variable List** tab to view the variables. Check the [How to Use a Scoped Variable](#how-to-use-a-scoped-variable) section to know more.

    ![Figure 5: List of Variables](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/scoped-variables/variable-list.jpg)

---

## How to Edit an Exiting Scoped Variable

Only a super-admin can edit existing scoped variables.

**Option 1**: Directly edit using the UI

![Figure 6: Editing from UI](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/scoped-variables/edit.jpg)

**Option 2**: Reupload the updated YAML file

![Figure 7: Reuploading New File](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/scoped-variables/reupload.jpg)

{% hint style="warning" %}
Reuploading the YAML file will replace the previous file, so any variable that existed in the previous file but not in the latest one will be lost
{% endhint %}

---

## How to Use a Scoped Variable

Once a variable is defined, it can be used by your authorized users on Devtron. A scoped variable widget would appear only on the screens that support its usage. 

Currently, the widget is shown only on `Edit build pipeline` and `Edit deployment pipeline` screens under **App Configuration** → **Workflow Editor**.

![Figure 8: Unexpanded Widget](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/scoped-variables/widget1.jpg)

Upon clicking on the widget, a list of variables will be visible. 

![Figure 9: Expanded Widget](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/scoped-variables/widget-expanded.jpg)

Use the copy button to copy a relevant variable of your choice.

![Figure 10: Copying a Variable](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/scoped-variables/display-value.jpg)

It would appear in the following format upon pasting it within an input field: `@{{variable-name}}`

![Figure 11: Pasting a Variable](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/scoped-variables/paste-value.jpg)

---

## Order of Precedence

When multiple values are associated with a scoped variable, the precedence order is as follows, with the highest priority at the top:

1. Environment + App [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)
2. App [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)
3. Environment [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)
4. Cluster [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)
5. Global

### Example

![Figure 12: Variable key in Red, Variable value in Green](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/scoped-variables/key-values.jpg)


1. **Environment + App:** This is the most specific scope, and it will take precedence over all other scopes. For example, the value of `DB name` variable for the `app1` application in the `prod` environment would be `app1-p`, even though there is a global `DB name` variable set to `Devtron`. If a variable value for this scope is not defined, the **App** scope will be checked.
2. **App:** This is the next most specific scope, and it will take precedence over the `Environment`, `Cluster`, and `Global` scopes. For example, the value of `DB name` variable for the `app1` application would be `project-tahiti`, even though the value of `DB name` exists in lower scopes. If a variable value for this scope is not defined, the **Environment** scope will be checked.
3. **Environment:** This is the next most specific scope, and it will take precedence over the `Cluster` and `Global` scopes. For example, the value of `DB name` variable in the `prod` environment would be `devtron-prod`, even though the value of `DB name` exists in lower scopes. If a variable value for this scope is not defined, the **Cluster** scope will be checked. 
4. **Cluster:** This is the next most specific scope, and it will take precedence over the `Global` scope. For example, the value of `DB name` variable in the `gcp-gke` cluster would be `Devtron-gcp`, even though there is a global `DB name` variable set to `Devtron-gcp`. If a variable value for this scope is not defined, the **Global** scope will be checked. 
5. **Global:** This is the least specific scope, and it will only be used if no variable values are found in other higher scopes. The value of `DB name` variable would be `Devtron`.

---

## List of Predefined Variables

There are some system variables that exist by default in Devtron that you can readily use if needed:

* **DEVTRON_NAMESPACE**: Provides name of the [namespace](../../reference/glossary.md#namespace)
* **DEVTRON_CLUSTER_NAME**: Provides name of the [cluster](../global-configurations/cluster-and-environments.md) configured on Devtron
* **DEVTRON_ENV_NAME**: Provides name of the [environment](../../reference/glossary.md#environment)
* **DEVTRON_IMAGE_TAG**: Provides [image tag](https://docs.docker.com/engine/reference/commandline/tag/) associated with the [container image](../../reference/glossary.md#image)
* **DEVTRON_IMAGE**: Provides full image path of the container image, e.g., `gcr.io/k8s-minikube/kicbase:v0.0.39`
* **DEVTRON_APP_NAME**: Provides name of the [application on Devtron](../create-application.md)

{% hint style="info" %}
Currently, these variables do not appear in the scoped variable widget, but you may use them. 
{% endhint %}








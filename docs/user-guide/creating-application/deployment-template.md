# Base Deployment Template


A deployment configuration is a manifest of the application. It defines the runtime behavior of the application.
You can select one of the default deployment charts or custom deployment charts which are created by super admin.

To configure a deployment chart for your application, do the following steps:

* Go to **Applications** and create a new application.
* Go to **App Configuration** page and configure your application.
* On the **Base Deployment Template** page, select the drop-down under **Chart type**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/deployment-template/deployment-chart-v3.jpg)

---

## Selecting a Chart Type

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have [Admin role](../global-configurations/authorization/user-access.md#role-based-access-levels) or above to select a chart.
{% endhint %}

{% hint style="warning" %}
### Note
After you select and save a chart type for a given application, you won't be able to change it later. Make sure to choose the correct chart type before saving. You can select a chart from [Devtron Charts](#from-devtron-charts) or other [Deployment Charts](#from-deployment-charts).
{% endhint %}

### From Devtron Charts

You can select a default deployment chart from the following options:

1. [Deployment](deployment-template/deployment.md) (Recommended)
2. [Rollout Deployment](deployment-template/rollout-deployment.md)
3. [Job & CronJob](deployment-template/job-and-cronjob.md)
4. [StatefulSet](deployment-template/statefulset.md)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/deployment-template/select-devtron-chart.gif)

### From Deployment Charts

{% hint style="warning" %}
This option will be available only if a custom chart exists. If it doesn't, a user with `super admin` permission may upload one in [Global Configurations → Deployment Charts](../global-configurations/deployment-charts.md).
{% endhint %}

You can select an available custom chart as shown below. You can also view the description of the custom charts in the list.

![Selecting Custom Chart](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/deployment-template/select-custom-chart.gif)

---

## Selecting a Chart Version

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have [Admin role](../global-configurations/authorization/user-access.md#role-based-access-levels) or above to select a chart version.
{% endhint %}

Once you select a chart type, choose a chart version using which you wish to deploy the application.

![Selecting Chart Version](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/deployment-template/chart-version.jpg)

Devtron uses helm charts for deployments and it maintains multiple chart versions based on the features it supports.

One can see available chart versions in the drop-down. You can select any chart version as per your requirements. By default, the latest version of the helm chart is selected.

Every chart version has its own YAML file that provides specifications for your application. To make it easy to use, we have created templates for the YAML file and have added some variables inside the YAML. You can provide or change the values of these variables as per your requirement.

---

## Configuring the Chart

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have [Admin role](../global-configurations/authorization/user-access.md#role-based-access-levels) or above to configure a chart. However, super-admins can lock keys in base deployment template to prevent non-super-admins from modifying them. Refer [Lock Deployment Configuration](../global-configurations/lock-deployment-config.md) to know more.
{% endhint %}

### Using Basic GUI

If you are not an advanced user, you may use the **Basic (GUI)** section to configure your chosen chart.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/deployment-template/basic-gui.jpg)

By default, the following fields are available for you to modify in the **Basic (GUI)** section:

| Fields | Description |
| :---    |     :---       |
| **Arguments**  | Enable the `Arguments` to pass one or more argument values. By default, it is in the `disabled` state. |
| **Command**  | Enable the `Command` to pass one or more command values. By default, it is in the `disabled` state. |
| **HTTP Request Routes** | Enable the `HTTP Request Routes` to define `Host`, and `Path`. By default, it is in the `disabled` state.<ul><li> **Host**: Domain name of the server. </li><li>**Path**: Path of the specific component in the host that the HTTP wants to access.</li></ul> You can define multiple paths as required by clicking **Add path**.|
| **Resources**  | Here, you can tweak the requests and limits of the CPU resource and RAM resource as per the application. |
| **Autoscaling** | Define the autoscaling parameters to automatically scale your application's deployment based on resource utilization.<ul><li>**Maximum Replicas**: The maximum number of replicas your application can scale up to.</li><li>**Minimum Replicas**: The minimum number of replicas your application should run at any time.</li><li>**Target CPU Utilization Percentage**: The average CPU utilization across all pods that will trigger scaling.</li><li>**Target Memory Utilization Percentage**: The average memory utilization across all pods that will trigger scaling.</li></ul>|
| **Environment Variables** (**Key/Value**)  | Define `key/value` by clicking **Add variable**. <ul><li> **Key**: Define the key of the environment.</li><li>**Value**: Define the value of the environment.</li></ul> You can define multiple env variables by clicking **Add EnvironmentVariables**.  |
| **Container Port** | The internal port on which the container listens for HTTP requests. Specify the container port and optionally the service port that maps to it. |
| **Service** | Configure the service that exposes your application to the network.<ul><li>**Type**: Specify the type of service (e.g., ClusterIP, NodePort, LoadBalancer).</li><li>**Annotations**: Add custom annotations to the service for additional configuration.</li></ul>|
| **Readiness Probe** | Define the readiness probe to determine when a container is ready to start accepting traffic.<ul><li>**Path**: The HTTP path that the readiness probe will access.</li><li>**Port**: The port on which the readiness probe will access the application.</li></ul>|
| **Liveness Probe** | Define the liveness probe to check if the container is still running and to restart it if it is not.<ul><li>**Path**: The HTTP path that the liveness probe will access.</li><li>**Port**: The port on which the liveness probe will access the application.</li></ul>|
| **Tolerations** | Define tolerations to allow the pods to be scheduled on nodes with matching taints.<ul><li>**Key**: The key of the taint to tolerate.</li><li>**Operator**: The relationship between the key and the value (e.g., Exists, Equal).</li><li>**Value**: The value of the taint to match.</li><li>**Effect**: The effect of the taint to tolerate (e.g., NoSchedule, NoExecute).</li></ul>|
| **ServiceAccount** | Specify the service account for the deployment to use, allowing it to access Kubernetes API resources.<ul><li>**Create**: Toggle to create a new service account.</li><li>**Name**: The name of the service account to use.</li></ul>|

Click **Save Changes**. If you want to do additional configurations, then click the **Switch to Advanced** button or **Advanced (YAML)** button for modifications.

{% hint style="warning" %}
### Note
* If you change any values in the 'Basic (GUI)', then the corresponding values will change in 'Advanced (YAML)' too.
* Users who are not super-admins will land on 'Basic (GUI)' section when they visit **Base Deployment Template** page; whereas super-admins will land on 'Advanced (YAML)' section. This is just a default behavior; therefore, they can still navigate to the other section if needed.
{% endhint %}

#### Customize Basic GUI [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

{% hint style="warning" %}
### Who Can Perform This Action?
Superadmin can define and apply custom deployment schema.
{% endhint %}

By default, the `Basic (GUI)` section comes with multiple predefined fields as seen earlier [in the table](#2-basic-configuration). However, if you wish to display a different set of fields to your team, you can modify the whole section as per your requirement.

This is useful in scenarios where:
* Your team members find it difficult to understand and edit the [Advanced (YAML)](#3-advanced-yaml) section.
* You frequently edit certain fields in Advanced (YAML), which you expect to remain easily accessible in Basic (GUI) section.
* You don't require some fields in Basic (GUI) section.
* You need the autonomy to keep the Basic (GUI) unique for applications/clusters/environments/charts, or display the same Basic (GUI) everywhere.

{% hint style="info" %}
There are two ways you can customize the Basic GUI, use any one of the following:
1. From [Deployment Charts](../global-configurations/deployment-charts.md#editing-gui-schema-of-deployment-charts) section
2. Using APIs (explained below)
{% endhint %}

{% embed url="https://www.youtube.com/watch?v=09VP1I-WvUs" caption="JSON-driven Deployment Schema" %}

You can pass a custom JSON (deployment schema) of your choice through the following API. You may need to run the API with the `POST` method if you are doing it for the first time.

```
PUT {{DEVTRON_BASEURL}}/orchestrator/deployment/template/schema
```

{% code title="Sample API Request Body" overflow="wrap" lineNumbers="true" %}

```json
{
  "name": "schema-1",
  "type": "JSON",
  "schema": "{\"type\":\"object\",\"properties\":{\"args\":{\"type\":\"object\",\"title\":\"Arguments\",\"properties\":{\"value\":{\"type\":\"array\",\"items\":{\"type\":\"string\"},\"title\":\"Value\"},\"enabled\":{\"type\":\"boolean\",\"title\":\"Enabled\"}}},\"command\":{\"type\":\"object\",\"title\":\"Command\",\"properties\":{\"value\":{\"type\":\"array\",\"items\":{\"type\":\"string\"},\"title\":\"Value\"},\"enabled\":{\"type\":\"boolean\",\"title\":\"Enabled\"}}},\"resources\":{\"type\":\"object\",\"title\":\"Resources(CPU&RAM)\",\"properties\":{\"limits\":{\"type\":\"object\",\"required\":[\"cpu\",\"memory\"],\"properties\":{\"cpu\":{\"type\":\"string\"},\"memory\":{\"type\":\"string\"}}},\"requests\":{\"type\":\"object\",\"properties\":{\"cpu\":{\"type\":\"string\"},\"memory\":{\"type\":\"string\"}}}}},\"autoscaling\":{\"type\":\"object\",\"title\":\"Autoscaling\",\"properties\":{\"MaxReplicas\":{\"type\":[\"integer\",\"string\"],\"title\":\"MaximumReplicas\",\"pattern\":\"^[a-zA-Z0-9-+\\\\/*%_\\\\\\\\s]+$\"},\"MinReplicas\":{\"type\":[\"integer\",\"string\"],\"title\":\"MinimumReplicas\",\"pattern\":\"^[a-zA-Z0-9-+\\\\/*%_\\\\\\\\s]+$\"},\"TargetCPUUtilizationPercentage\":{\"type\":[\"integer\",\"string\"],\"title\":\"TargetCPUUtilizationPercentage\",\"pattern\":\"^[a-zA-Z0-9-+\\\\/*%_\\\\\\\\s]+$\"},\"TargetMemoryUtilizationPercentage\":{\"type\":[\"integer\",\"string\"],\"title\":\"TargetMemoryUtilizationPercentage\",\"pattern\":\"^[a-zA-Z0-9-+\\\\/*%_\\\\\\\\s]+$\"}}},\"EnvVariables\":{\"type\":\"array\",\"items\":{\"type\":\"object\",\"properties\":{\"key\":{\"type\":\"string\"},\"value\":{\"type\":\"string\"}}},\"title\":\"EnvironmentVariables\"},\"ContainerPort\":{\"type\":\"array\",\"items\":{\"type\":\"object\",\"properties\":{\"port\":{\"type\":\"integer\"}}},\"title\":\"ContainerPort\"}}}",
  "selectors": [
    {
      "attributeSelector": {
        "category": "APP",
        "appNames": ["my-demo-app"]
      }
    },
    {
      "attributeSelector": {
        "category": "ENV",
        "envNames": ["env1", "env2", "env3"]
      }
    },
    {
      "attributeSelector": {
        "category": "CLUSTER",
        "clusterNames": ["cluster1", "cluster2", "cluster3"]
      }
    },
    {
      "attributeSelector": {
        "category": "CHART_REF",
        "chartVersions": [
          {
            "type": "Deployment",
            "version": "1.0.0"
          }
        ]
      }
    },
    {
      "attributeSelector": {
        "category": "APP_ENV",
        "appEnvNames": [
          {
            "appName": "my-demo-app",
            "envName": "devtron"
          }
        ]
      }
    }
  ]
}

```
{% endcode %}

1. In the `name` field, give a name to your schema, e.g., *schema-1*
2. Enter the `type` as JSON.
3. The `schema` field is for entering your custom deployment schema. Perform the following steps:
    * To create a custom schema of your choice, you may use [RJSF JSON Schema Tool](https://rjsf-team.github.io/react-jsonschema-form/). 
    * Copy the final JSON and stringify it using any free online tool. 
    * Paste the stringified JSON in the `schema` field of the API request body.
    * Send the API request. If your schema already exists, use the `PUT` method instead of `POST` in the API call.
4. The `attributeSelector` object helps you choose the scope at which your custom deployment schema will take effect.
    | Priority | Category Scope  | Description                                                                |
    |----------|-----------------|----------------------------------------------------------------------------|
    | 1 (High) | APP_ENV         | Specific to an application and its environment                             |
    | 2        | APP             | Applies at the application level if no specific environment is defined     |
    | 3        | ENV             | Applies to specific deployment environment                                 |
    | 4        | CHART_REF       | Applies to all applications using a specific chart type and version        |
    | 5        | CLUSTER         | Applies across all applications and environments within a specific cluster |
    | 6        | GLOBAL          | Universally applies if no other more specific schemas are defined          |


### Using Advanced (YAML)

If you are an advanced user wishing to perform additional configurations, you may switch to **Advanced (YAML)** for modifications.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/deployment-template/advanced-yaml.jpg)

Refer the respective templates to view the YAML details.
* [Deployment](deployment-template/deployment.md)
* [Rollout Deployment](deployment-template/rollout-deployment.md)
* [Job & CronJob](deployment-template/job-and-cronjob.md)
* [StatefulSet](deployment-template/statefulset.md)

---

## Application Metrics

Depending on the chart type and version you select, application metrics of your application may be viewed. <br />
This includes: 
* Status codes 2xx, 3xx, 5xx
* Throughput
* Latency
...and many more

Enable **Show application metrics** toggle to view the application metrics on the **App Details** page.

![Show application metrics](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/deployment-template/show-application-metrics-v2.jpg)

> **IMPORTANT**: Enabling application metrics introduces a sidecar container to your main container which may require some additional configuration adjustments. We recommend you to do load test after enabling it in a non-production environment before enabling it in production environment.

Select **Save & Next** to save your configurations.



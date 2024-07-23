# Base Deployment Template


A deployment configuration is a manifest of the application. It defines the runtime behavior of the application.
You can select one of the default deployment charts or custom deployment charts which are created by super admin.

To configure a deployment chart for your application, do the following steps:

* Go to **Applications** and create a new application.
* Go to **App Configuration** page and configure your application.
* On the **Base Deployment Template** page, select the drop-down under **Chart type**.

---

## Select chart from Default Charts

You can select a default deployment chart from the following options:

1. [Deployment](deployment-template/deployment.md) (Recommended)
2. [Rollout Deployment](deployment-template/rollout-deployment.md)
3. [Job & CronJob](deployment-template/job-and-cronjob.md)
4. [StatefulSet](deployment-template/statefulset.md)


![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/deployment-template/deployment-chart.png)

---

## Select chart from Custom Charts

Custom charts are added by users with `super admin` permission from the [Custom charts](../global-configurations/custom-charts.md) section.

You can select the available custom charts from the drop-down list. You can also view the description of the custom charts in the list.

![Select custom chart](https://devtron-public-asset.s3.us-east-2.amazonaws.com/custom-charts/use-custom-chart.png)

A [custom chart](../global-configurations/custom-charts.md) can be uploaded by a super admin.

---

##

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

### Customize Basic GUI [![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

{% hint style="warning" %}
### Who Can Perform This Action?
Superadmin can define and apply custom deployment schema using API
{% endhint %}

By default, the `Basic (GUI)` section comes with multiple predefined fields as seen earlier [in the table](#2-basic-configuration). However, if you wish to display a different set of fields to your team, you can modify the whole section as per your requirement.

{% embed url="https://www.youtube.com/watch?v=09VP1I-WvUs" caption="JSON-driven Deployment Schema" %}

This is useful in scenarios where:
* Your team members find it difficult to understand and edit the [Advanced (YAML)](#3-advanced-yaml) section.
* You frequently edit certain fields in Advanced (YAML), which you expect to remain easily accessible in Basic (GUI) section.
* You don't require some fields in Basic (GUI) section.
* You need the autonomy to keep the Basic (GUI) unique for applications/clusters/environments/charts, or display the same Basic (GUI) everywhere.

This is possible by passing a custom JSON (deployment schema) of your choice through the following API. You may need to run the API with the `POST` method if you are doing it for the first time.

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


---

## Application Metrics

Enable **show application metrics** toggle to view the application metrics on the **App Details** page.

![Show application metrics](https://devtron-public-asset.s3.us-east-2.amazonaws.com/custom-charts/show-application-metrics.png)

> **IMPORTANT**: Enabling Application metrics introduces a sidecar container to your main container which may require some additional configuration adjustments. We recommend you to do load test after enabling it in a non-production environment before enabling it in production environment.

Select **Save & Next** to save your configurations.

{% hint style="warning" %}
Super-admins can lock keys in base deployment template to prevent non-super-admins from modifying those locked keys. Refer [Lock Deployment Configuration](../global-configurations/lock-deployment-config.md) to know more.
{% endhint %}

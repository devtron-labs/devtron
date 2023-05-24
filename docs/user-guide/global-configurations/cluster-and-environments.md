# Clusters and Environments

You can add your existing Kubernetes clusters and environments on the `Clusters and Environments` section. You must have a [super admin](https://docs.devtron.ai/global-configurations/authorization/user-access#assign-super-admin-permissions) access to add a cluster.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/cluster-and-environments.png)

## Add Cluster:

To add cluster, go to the `Clusters & Environments` section of `Global Configurations`. Click **Add cluster**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/add-clusters.png)

Provide the information in the following fields to add your kubernetes cluster:

| Field | Description |
| :--- | :--- |
| `Name` | Enter a name of your cluster. |
| `Server URL` |  Server URL of a cluster.<br>Note: We recommended to use a [self-hosted URL](#benefits-of-self-hosted-url) instead of cloud hosted URL.</br>  |
| `Bearer Token` | Bearer token of a cluster. |

### Get Cluster Credentials

>**Prerequisites:** `kubectl` must be installed on the bastion.

**Note**: We recommend to use a self-hosted URL instead of cloud hosted URL. Refer the benefits of [self-hosted URL](#benefits-of-self-hosted-url).

You can get the **`Server URL`** & **`Bearer Token`** by running the following command depending on the cluster provider:

{% tabs %}
{% tab title="k8s Cluster Providers" %}
If you are using EKS, AKS, GKE, Kops, Digital Ocean managed Kubernetes, run the following command to generate the server URL and bearer token:
```bash
curl -o https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/kubernetes_export_sa.sh \ 
&& bash kubernetes_export_sa.sh cd-user  devtroncd
```
{% endtab %}
{% tab title="Microk8s Cluster" %}
If you are using a **`microk8s cluster`**, run the following command to generate the server URL and bearer token:

```bash
curl -o https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/kubernetes_export_sa.sh && sed -i 's/kubectl/microk8s kubectl/g' \
kubernetes_export_sa.sh && bash kubernetes_export_sa.sh cd-user \
devtroncd
```
{% endtab %}
{% endtabs %}

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/generate-cluster-credentials.png)

### Benefits of Self-hosted URL

* Disaster Recovery: 
  * It is not possible to edit the server URL of a cloud specific provider. If you're using an EKS URL (e.g.` *****.eu-west-1.elb.amazonaws.com`), it will be a tedious task to add a new cluster and migrate all the services one by one. 
  * But in case of using a self-hosted URL (e.g. `clear.example.com`), you can just point to the new cluster's server URL in DNS manager and update the new cluster token and sync all the deployments.

* Easy Cluster Migrations: 
  * In case of managed Kubernetes clusters (like EKS, AKS, GKE etc) which is a cloud provider specific, migrating your cluster from one provider to another will result in waste of time and effort. 
  * On the other hand, migration for a  self-hosted URL is easy as the URL is of single hosted domain independent of the cloud provider.


### Configure Prometheus (Enable Applications Metrics)

If you want to see application metrics against the applications deployed in the  cluster, Prometheus must be deployed in the cluster. Prometheus is a powerful tool to provide graphical insight into your application behavior.

>**Note:** Make sure that you install `Monitoring (Grafana)` from the `Devtron Stack Manager` to configure prometheus.
If you do not install `Monitoring (Grafana)`, then the option to configure prometheus will not be available. 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/enable-app-metrics.png)

Enable the application metrics to configure prometheus and provide the information in the following fields:

| Field | Description |
| :--- | :--- |
| `Prometheus endpoint` | Provide the URL of your prometheus. |
| `Authentication Type` | Prometheus supports two authentication types:<ul><li>**Basic:** If you select the `Basic` authentication type, then you must provide the `Username` and `Password` of prometheus for authentication.</li></ul> <ul><li>**Anonymous:** If you select the `Anonymous` authentication type, then you do not need to provide the `Username` and `Password`.<br>Note: The fields `Username` and `Password` will not be available by default.</li></ul> |
| `TLS Key` & `TLS Certificate` | `TLS Key` and `TLS Certificate` are optional, these options are used when you use a customized URL. |

Now, click `Save Cluster` to save your cluster on Devtron.

### Installing Devtron Agent

Your Kubernetes cluster gets mapped with Devtron when you save the cluster configurations. Now, the Devtron agent must be installed on the added cluster so that you can deploy your applications on that cluster.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/install-devtron-agent.png)

When the Devtron agent starts installing, click `Details` to check the installation status.

![Install Devtron Agent](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/gc-cluster-agents.jpg)

A new window pops up displaying all the details about the Devtron agent.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/cluster_gc5.jpg)

## Add Environment

Once you have added your cluster in the `Clusters & Environments`, you can add the environment by clicking `Add environment`.

A new environment window pops up.

| Field | Description |
| :--- | :--- |
| `Environment Name` | Enter a name of your environment. |
| `Enter Namespace` | Enter a namespace corresponding to your environment.<br>**Note**: If this namespace does not already exist in your cluster, Devtron will create it. If it exists already, Devtron will map the environment to the existing namespace.</br> |
| `Environment Type` | Select your environment type:<ul><li>`Production`</li></ul> <ul><li>`Non-production`</li></ul>Note: Devtron shows deployment metrics (DORA metrics) for environments tagged as `Production` only. |

Click `Save` and your environment will be created. 


![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/gc-cluster-add-environment.jpg)


## Update Environment

* You can also update an environment by clicking the environment.
* You can change `Production` and `Non-Production` options only.
* You cannot change the `Environment Name` and `Namespace Name`.
* Make sure to click **Update** to update your environment.

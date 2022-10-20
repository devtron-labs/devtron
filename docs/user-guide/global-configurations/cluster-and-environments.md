# Cluster And Environments

You can add your existing Kubernetes clusters and environments here.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/cluster-and-environments.png)

## Add Cluster:

To add a cluster on devtron, you must have superadmin access.

Navigate to the `Global Configurations` → `Clusters and Environments` on devtron and click `Add Cluster`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/add-clusters.png)

Provide the below information to add your kubernetes cluster:

| Field | Description |
| :--- | :--- |
| `Name` | Enter a name of your cluster |
| `Server URL` | API server URL of the cluster |
| `Bearer Token` | Bearer token of the cluster |

#### Get Cluster Credentials

>**Prerequisites:** `kubectl` and `jq` should be installed on the bastion.

You can get the **`Server URL`** & **`Bearer Token`** by running the following command.

{% tabs %}
{% tab title="k8s cluster providers" %}
```bash
curl -O https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/kubernetes_export_sa.sh && bash kubernetes_export_sa.sh cd-user devtroncd https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/clusterrole.yaml
```
{% endtab %}
{% tab title="Microk8s clusters" %}
If you are using a **`microk8s cluster`**, run the following command to generate the server url and bearer token:

```bash
curl -O https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/kubernetes_export_sa.sh && sed -i 's/kubectl/microk8s kubectl/g' kubernetes_export_sa.sh && bash kubernetes_export_sa.sh cd-user devtroncd https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/clusterrole.yaml
```
{% endtab %}
{% endtabs %}

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/generate-cluster-credentials.png)

#### Server URL

Provide the server URL of your kubernetes cluster. It is recommended to use a self-hosted URL instead of cloud hosted. Self-hosted URL will provide the following benefits:

**\(a\) Disaster Recovery -** It is not possible to edit the server-url of a cluster. So if you're using an eks url, For eg- ` *****.eu-west-1.elb.amazonaws.com` it will be a tedious task to add a new cluster and migrate all the services one by one. While using a self-hosted url For eg- `clear.example.com` you can just point to the new cluster's server url in DNS manager and update the new cluster token and sync all the deployments.

**\(b\) Easy cluster migrations -** Cluster url in case of managed Kubernetes clusters (like EKS, AKS, GKE etc) is cloud provider specific, so migrating your cluster from one provider to another will result in waste of time and effort. On the other hand, if using a self-hosted url migrations will be easy as the url is of single hosted domain independent of the cloud provider.

#### Bearer token

Enter your kubernetes cluster’s Bearer token for authentication purposes so that Devtron is able to communicate with your kubernetes cluster and can deploy your application in your kubernetes cluster.

### Configure Prometheus (Enable Applications Metrics)

If you want to see application metrics against the applications deployed in the  cluster, Prometheus should be deployed in the cluster. Prometheus is a powerful tool to provide graphical insight into your application behavior.

>**Note:** Make sure that you install `Monitoring (Grafana)` from the Devtron Stack Manager to configure prometheus.
If you do not install `Monitoring (Grafana)`, then the option to configure prometheus will not be available. 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/enable-app-metrics.png)

| Field | Description |
| :--- | :--- |
| `Prometheus Endpoint` | Provide the URL of your prometheus. |
| `Authentication Type` | Prometheus supports two authentication types:<ul><li>**Basic:** If you select the `Basic` authentication type, then you have to provide the `Username` and `Password` of prometheus for authentication.</li></ul> <ul><li>**Anonymous:** If you select the `Basic` authentication type, then you have to provide the `Username` and `Password` of prometheus for authentication.</li></ul> |
| `TLS Key` & `TLS Certificate` | `TLS Key` and `TLS Certificate` are optional, these options are used when you use a custom URL. |

Now click `Save Cluster` to save your cluster over Devtron.

### Installing Devtron agent

Your Kubernetes cluster gets mapped with the Devtron when you save the cluster configurations. Now the Devtron agent needs to be installed on the cluster that you have added on Devtron so that you're able to deploy on the added cluster.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/install-devtron-agent.png)

When the Devtron agent starts installing, you can check the installation status over the Cluster & Environment tab.

![Install Devtron Agent](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/gc-cluster-agents.jpg)

Click `Details` to check what got installed inside the agents. A new window will be popped up displaying all the details about these agents.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/cluster_gc5.jpg)

## Add Environment

Once you have added your cluster in the `Cluster & Environments`, you can add the environment also.

1.  Click `Add Environment`.

2. On the `New Environment` window, enter a name of your environment in the `Environment Name` field.

3.  Enter a namespace corresponding to your environment in the `Enter Namespace` field. (If this namespace doesn't already exist in your cluster, devtron will create it. If it already exists, Devtron will map the env to the existing namespace)

4. Select the Environment type:

Devtron shows Deployment metrics for environments tagged as Production only.

     -  `Production `

     -  `Non-production`


5. Click Save and your environment will be created. 


![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/gc-cluster-add-environment.jpg)


> **Note:** You can also update an environment by clicking the environment.
You can change `Production` and `Non-Production` options only.
You cannot change the `Environment Name` and `Namespace Name`.
Make sure to click **Update** to update your environment.

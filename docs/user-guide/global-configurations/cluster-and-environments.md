# Cluster And Environments

You can add your existing Kubernetes clusters and environments here.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/cluster-and-environments.png)

## Add Cluster:

To add a cluster on devtron, you must have superadmin access.

Navigate to the `Global Configurations` → `Clusters and Environments` on devtron and click on `Add Cluster`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/add-clusters.png)

 Provide the informations mentioned below to add your kubernetes cluster:

### 1. Name

Give a name to your cluster inside the name box.

### 2. Kubernetes Cluster Credentials

Provide your kubernetes cluster’s credentials i.e server url and bearer token.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/server-url-and-token.png)

#### Get Cluster Credentials

You can get the **`Server URL`** & **`Bearer Token`** by running the following command. But before that, please ensure that kubectl and jq are installed on the bastion on which you are running the command.

```bash
curl -O https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/kubernetes_export_sa.sh && bash kubernetes_export_sa.sh cd-user devtroncd https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/clusterrole.yaml
```
If you are using a **`microk8s cluster`**, run the following command to generate the server url and bearer token:

```bash
curl -O https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/kubernetes_export_sa.sh && sed -i 's/kubectl/microk8s kubectl/g' kubernetes_export_sa.sh && bash kubernetes_export_sa.sh cd-user devtroncd https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/clusterrole.yaml
```
![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/generate-cluster-credentials.png)

#### Server URL

Provide the server URL of your kubernetes cluster. It is recommended to use a self-hosted URL instead of cloud hosted. Self-hosted URL will provide the following benefits:

**\(a\) Disaster Recovery -** It is not possible to edit the server-url of a cluster. So if you're using an eks url, For eg- ` *****.eu-west-1.elb.amazonaws.com` it will be a tedious task to add a new cluster and migrate all the services one by one. While using a self-hosted url For eg- `clear.example.com` you can just point to the new cluster's server url in DNS manager and update the new cluster token and sync all the deployments.

**\(b\) Easy cluster migrations -** Cluster url is given in the name of the cloud provider used, so migrating your cluster from one provider to another will result in waste of time and effort. On the other hand, if using a self-hosted url migrations will be easy as the url is of single hosted domain independent of the cloud provider.

#### Bearer token

Provide your kubernetes cluster’s Bearer token for authentication purposes so that Devtron is able to communicate with your kubernetes cluster and can deploy your application in your kubernetes cluster.

### 3. Enable applications metrics for the cluster

If you want to see application metrics against the applications deployed in the  cluster, Prometheus should be deployed in the cluster. Prometheus is a powerful tool to provide graphical insight into your application behavior. The below inputs are required to configure Prometheus over Devtron.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/enable-app-metrics.png)

* **Prometheus endpoint**

Provide the URL of your prometheus. 
Prometheus supports two types of authentication `Basic` and `Anonymous`. Select the authentication type for your Prometheus setup.

* **Basic**

If you select the `basic` type of authentication then you have to provide the `Username` and `Password` of prometheus for authentication.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/basic-auth.png)

* **Anonymous**

If you select `Anonymous` then you do not have to provide any username and password for authentication.

* **TLS Key & TLS Certificate**

TLS key and TLS certificate both options are optional, these options are used when you use a custom URL, in that case, you can pass your TLS key and TLS certificate.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/anonymous-auth.png)

Now click on `Save Cluster` to save your cluster over Devtron.

### Installing Devtron agent

Your Kubernetes cluster gets mapped with the Devtron when you save the cluster configurations. Now the Devtron agent needs to be installed on the cluster that you added over Devtron. So that the components of Devtron can communicate with your cluster. 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/install-devtron-agent.png)

When the Devtron agent starts installing, you can check the installation status over the Cluster & Environment tab.

![Install Devtron Agent](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/gc-cluster-agents.jpg)

Click on `Details` to check what got installed inside the agents. A new window will be popped up displaying all the details about these agents.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/cluster_gc5.jpg)

## Add Environment

Once you have added your cluster in Cluster & Environment, you can add the environment also. Click on `Add Environment`, a window will be opened. Give a name to your environment in the `Environment Name` box and provide a namespace corresponding to your environment in the `Namespace` input box. Now choose if your environment is for Production purposes or for Non-production purposes. Production and Non-production options are only for tagging purposes. Click on `Save` and your environment will be created.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/gc-cluster-add-environment.jpg)

You can update an environment by clicking on the environment. You can only change Production and Non-production options here.

**Note**

You can not change the Environment name and Namespace name.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/gc-cluster-update-environment.jpg)

Click on `Update` to update your environment.


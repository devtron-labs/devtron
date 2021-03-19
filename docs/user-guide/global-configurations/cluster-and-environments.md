# Cluster And Environments

The Global configuration provides a feature of `Cluster & Environment` in which you can add your Kubernetes clusters and environment.

Select the Cluster & Environment section of global configuration and click on `Add Cluster` to add your cluster.

## Add Cluster:

Provide the below information to add your kubernetes cluster:

1. Name
2. Kubernetes Cluster Info
   * Server URL
   * Bearer token
3. Prometheus Info
   * Prometheus endpoint
     * Basic
       * Username
       * Password
     * Anonymous
   * TLS Key
   * TLS Certificate

![](../../.gitbook/assets/cluster_gc1.png)

### 1. Name

Give a name to your cluster inside the name box.

### 2. Kubernetes Cluster Info

Provide your kubernetes cluster’s credentials.

* **Server URL**

Provide the endpoint/URL of your kubernetes cluster.

* **Bearer token**

Provide your kubernetes cluster’s Bearer token for authentication purposes so that the Devtron tool will be able to talk to your kubernetes cluster and can deploy your application in your kubernetes cluster.

### 3. Prometheus Info

Prometheus is a powerful solution to provide graphical insight into your application behavior. If you want to see your application matrix against your applications deployed in kubernetes, install Prometheus in your kubernetes cluster. The below inputs are required to configure your prometheus into Devtron’s tool.

* **Prometheus endpoint**

Provide the URL of your prometheus. Prometheus supports two types of authentication `Basic` and `Anonymous`. Select the authentication type for your Prometheus setup.

* **Basic**

If you select the `basic` type of authentication then you have to provide the `Username` and `Password` of prometheus for authentication.

* **Anonymous**

If you select `Anonymous` then you do not have to provide any username and password for authentication.

* **TLS Key & TLS Certificate**

TLS key and TLS certificate both options are optional, these options are used when you use a custom URL, in that case, you can pass your TLS key and TLS certificate.


### K8s Version
on saving or update a cluster there is a call to fetch k8s version, it will store corresponding to cluster on db. used in listing api's and app detail page for grafana url.


Check the below screenshots to know how it looks like If you select the `Basic` authentication type

![](../../.gitbook/assets/cluster_gc2%20%283%29.png)

If you select the `Anonymous` authentication type

![](../../.gitbook/assets/cluster_gc3%20%283%29.png)

Now click on `Save Cluster` to save your cluster information.

### Note:

Your kubernetes cluster gets mapped with the Devtron when you save your kubernetes cluster Configuration. Now the agents of devtron will be installed on your cluster so that the components of devtron can communicate to your cluster. When the agent starts installing on your cluster, you can check the status of the agents in the Cluster & Environment tab also.

![](../../.gitbook/assets/gc-cluster-agents.png)

Click on `Details` to check what got installed inside the agents. A new window will be popped up displaying all the details about these agents.

![](../../.gitbook/assets/cluster_gc5.png)

## Add Environment

Once you have added your cluster in Cluster & Environment, you can add the environment also. Click on `Add Environment`, a window will be opened. Give a name to your environment in the `Environment Name` box and provide a namespace corresponding to your environment in the `Namespace` input box. Now choose if your environment is for Production purposes or for Non-production purposes. Production and Non-production options are only for tagging purposes. Click on `Save` and your environment will be created.

![](../../.gitbook/assets/gc-cluster-add-environment%20%283%29.png)

You can update an already created environment, Select and click on the environment which you want to update. You can only change Production and Non-production options here.

**Note**

You can not change the Environment name and Namespace name.

![](../../.gitbook/assets/gc-cluster-update-environment%20%281%29.png)

Click on `Update` to update your environment.


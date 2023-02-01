# 集群和环境
您可以在`集群和环境`添加现有的Kubernetes集群和环境。您必须有一个[超级管理员](https://docs.devtron.ai/global-configurations/authorization/user-access#assign-super-admin-permissions)访问权限才能添加集群。

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/cluster-and-environments.png)
## 添加集群：
要添加集群，请进入`全局配置`的`集群和环境`部分。点击**添加集群**。

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/add-clusters.png)

提供以下字段中的信息以添加kubernetes集群：

|字段|说明|
| :- | :- |
|`名称`|输入群集的名称。|
|`服务器URL`|集群的服务器URL。<br>注意：我们建议使用[自托管URL](#benefits-of-self-hosted-url)而不是云托管URL。</br>|
|`承载令牌`|集群的承载令牌。|

### 获取群集凭据
> **先决条件：**`kubectl`和`jq`必须安装在bastion上。

**注意**：我们建议使用自托管URL而不是云托管URL。参考[自托管URL](#benefits-of-self-hosted-url)。

您可以按照您的群集供应者，通过此命令获取**`服务器URL`**以及**`承载令牌`**：

{% tabs %}
{% tab title="k8s Cluster Providers" %}
如果您使用的是 EKS、AKS、GKE、Kops、Digital Ocean管理的Kubernetes，请运行以下命令以生成服务器URL和承载令牌：
~~~ bash
curl -O https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/kubernetes_export_sa.sh \
&& bash kubernetes_export_sa.sh cd-user devtroncd \
https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/clusterrole.yaml
~~~

{% endtab %}
{% tab title="Microk8s Cluster" %}
如果您使用**`microk8s集群`**，请运行以下命令以生成服务器URL和承载令牌：
~~~ bash
curl -O https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/kubernetes_export_sa.sh && sed -i 's/kubectl/microk8s kubectl/g' \
kubernetes_export_sa.sh && bash kubernetes_export_sa.sh cd-user \
devtroncd https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/clusterrole.yaml
~~~

{% endtab %}
{% endtabs %}

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/generate-cluster-credentials.png)
### 自托管URL的好处
* 灾难恢复：
  * 无法编辑特定于云供应商的服务器URL。如果您使用的是EKS URL（例如`*****.eu-west-1.elb.amazonaws.com`），添加新集群并逐个迁移所有服务将是一项繁琐的任务。
  * 但在使用自托管URL的情况下（例如`clear.example.com`），您只需在DNS管理器中指向新群集的服务器URL并更新新群集令牌并同步所有部署即可。
* 简单的集群迁移：
  * 对于特定于云供应商的托管Kubernetes集群（如EKS、AKS、GKE等），将集群从一个供应商迁移到另一个供应商将导致浪费时间和精力。
  * 另一方面，自托管URL的迁移很容易，因为URL是独立于云供应商的单个托管域。
### 配置Prometheus（启用应用程序指标）
如果要查看针对集群中部署的应用程序的应用程序指标，则必须在集群中部署Prometheus。Prometheus是一个强大的工具，它能为您的应用程序行为提供图形见解。
> **注意：**确保您从`Devtron堆栈管理器`安装`监测（Grafana）`以配置Prometheus。
> 如果不安装`监测（Grafana）`，那么配置prometheus的选项将不可用。

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/enable-app-metrics.png)

启用应用程序指标以配置prometheus并在以下字段中提供信息：

|字段|说明|
| :- | :- |
|`Prometheus端点`|提供您的Prometheus的URL。|
|`身份验证类型`|Prometheus支持两种身份验证类型：<ul><li>**基本：**如果您选择`基本`身份验证类型，则必须提供`用户名称`和`密码`Prometheus的身份验证。</li></ul><ul><li>**匿名：**如果您选择`匿名`身份验证类型，则不需要提供`用户名称`和`密码`。<br>注：`用户名称`和`密码`字段默认情况下将不可用。</li></ul>|
|`TLS密钥`和`TLS证书`|`TLS密钥`和`TLS证书`是非必填的。当您使用自定义URL时才会使用。|

现在，点击`保存集群`以在Devtron上保存您的集群。
### 安装Devtron代理
当您保存集群配置时，您的Kubernetes集群将与Devtron映射。现在，必须在添加的群集上安装Devtron代理，以便您可以在该群集上部署应用程序。

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/install-devtron-agent.png)

当Devtron代理开始安装时，点击`详情`来检查安装状态。

![Install Devtron Agent](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/gc-cluster-agents.jpg)

将弹出一个新视窗，显示有关Devtron代理的所有详细信息。

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/cluster\_gc5.jpg)
## 添加环境
一旦您在`集群和环境`，您可以通过点击添加环境`添加环境`。

将弹出一个新的环境视窗。

|字段|说明|
| :- | :- |
|`环境名称`|输入环境的名称。|
|`输入命名空间`|输入与您的环境相对应的命名空间。<br>**注意**：如果您的集群中还不存在此命名空间，Devtron将创建它。如果它已经存在，Devtron会将环境映射到现有的命名空间。</br>|
|`环境类型`|选择您的环境类型：<ul><li>`生产`</li></ul> <ul><li>`非生产`</li></ul>注意：Devtron 仅显示标记为`生产`的环境的部署指标（DORA 指标）。|

点击`保存`以创建您的环境。

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/gc-cluster-add-environment.jpg)
## 更新环境
* 您还可以通过点击某个环境来更新它。
* 您仅可更改`生产`和`非生产`选项。
* 您不能更改`环境名称`和`命名空间名称`。
* 确保点击**更新**以更新您的环境。

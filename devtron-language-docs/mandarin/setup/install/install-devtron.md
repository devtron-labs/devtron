# 安装Devtron
在本节中，我们将介绍如何在没有任何集成的情况下通过Devtron安装Helm Dashboard。稍后可以使用[Devtron Stack Manager](https://docs.devtron.ai/v/v0.6/usage/integrations)添加集成。

如果你想在Minikube、Microk8s、K3s、Kind上安装Devtron，请参考[本节](./Install-devtron-on-Minikube-Microk8s-K3s-Kind.md)。
## 开始之前：
如果您未安装[Helm](https://helm.sh/docs/intro/install/)请先安装。
## 添加Helm Repo
~~~ bash
helm repo add devtron https://helm.devtron.ai
~~~
## 安装Devtron的Helm Dashboard
**注意**：此安装命令不会安装CI/CD集成。对于CI/CD，请参阅[安装带有CI/CD的Devtron](https://docs.devtron.ai/install/install-devtron-with-cicd)。

运行以下命令以安装Devtron的Helm Dashboard：
~~~ bash
helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd
~~~
## 安装多架构节点（ARM和AMD）
要在具有多架构节点（ARM和AMD）的集群上安装Devtron，请在Devtron安装命令后附加`--set installer.arch=multi-arch`。

[//]: #（如果您计划使用 Hyperion 进行`生产部署`，请参阅我们对于[Devtron安装]&#40;override-default-devtron-installation-configs.md&#41;.）的推荐的覆盖。

[//]: # (## 安装状态)

[//]: # ()
[//]: # (运行以下命令)

[//]: # ()
[//]: # (```bash)

[//]: # (kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.status}')

[//]: # (```)
## Devtron仪表板
运行以下命令以获取仪表板URL：
~~~ text
kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}'
~~~

您会得到如下所示的结果：
~~~ text
[test2@server ~]$ kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}'
[map[hostname:aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com]]
~~~

如上所述的主机名`aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com`是Loadbalancer URL，您可以通过它访问Devtron仪表板。
> 您还可以执行与您的域/子域对应的CNAME条目，以指向此Loadbalancer URL以在自定义域访问它。

|> 主机|> 类型|> 指向|
| :- | :- | :- |
|> devtron.yourdomain.com|> CNAME|> aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com|
>
## Devtron管理员（Admin）凭据
### 对于Devtron v0.6.0及更高版本
**用户名称**: `admin` <br>
**密码**：运行以下命令获取管理员（admin）密码：
~~~ bash
kubectl -n devtroncd get secret devtron-secret \
-o jsonpath='{.data.ADMIN_PASSWORD}' | base64 -d
~~~
### 对于小于v0.6.0的Devtron版本
**用户名称**: `admin` <br>
**密码**：运行以下命令获取管理员（admin）密码：
~~~ bash
kubectl -n devtroncd get secret devtron-secret \
-o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
~~~

**注意事项**：如果您想卸载Devtron或清理Devtron的helm安装程序，请参阅[卸载Devtron](https://docs.devtron.ai/install/uninstall-devtron)。
## 升级
要在Devtron中使用CI/CD功能，您可以安装[带有CI/CD的Devtron](https://docs.devtron.ai/install/install-devtron-with-cicd)或[带有CI/CD和GitOps（Argo CD）的Devtron](https://docs.devtron.ai/install/install-devtron-with-cicd-with-gitops)。

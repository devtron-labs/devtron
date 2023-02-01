# 在Minikube、Microk8s、K3s、Kind上安装Devtron
您可以在高端机器或云VM上安装并试用Devtron。 如果您将其安装在笔记本电脑/电脑上，它可能会开始反应缓慢，因此建议在将其关闭电脑之前从系统中卸载Devtron。
## Devtron安装的系统配置
1. 2 vCPU
1. 4GB+可用内存
1. 20GB+可用磁盘空间
## 开始之前：
在我们开始安装Devtron之前，您必须在服务器中设置集群并安装先决条件：

* 使用以下方法使用[Minikube](https://minikube.sigs.k8s.io/docs/start/)、[Kind](https://kind.sigs.k8s.io/docs/user/quick-start/)或[K3s](https://rancher.com/docs/k3s/latest/en/installation/)创建集群。
* 安装[Helm3](https://helm.sh/docs/intro/install/)。
* 安装 [kubectl](https://kubernetes.io/docs/tasks/tools/)。
## 安装Devtron
{% tabs %}

{% tab title=" Minikube/Kind集群 "%}

要在``Minikube/kind``集群上安装Devtron，请运行以下命令：
~~~ bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set components.devtron.service.type=NodePort
~~~

{% endtab %}

{% tab title="k3s Cluster" %}
要在``k3s``集群上安装Devtron，请运行以下命令：
~~~ bash
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml

helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set components.devtron.service.type=NodePort
~~~

{% endtab %}

{% endtabs %}
## Devtron仪表板
若要在使用``Minikube``为集群时访问Devtron仪表板，请运行以下命令：
~~~ bash
minikube service devtron-service --namespace devtroncd
~~~

若要在使用``Kind/k3s`` 为时访问Devtron仪表板，请运行以下命令以将Devtron服务转发到端口8000：
~~~ bash
kubectl -ndevtroncd port-forward service/devtron-service 8000:80
~~~

**仪表板**：<http://127.0.0.1:8000>。
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
## 在云VM（AWS ec2、Azure VM、GCP VM）上安装Devtron
我们建议使用具有2vCPU+、4GB+可用内存、20GB+存储空间、计算优化VM类型以及Ubuntu操作系统的云VM。
### 创建Microk8s集群
~~~ bash
sudo snap install microk8s --classic --channel=1.22
sudo usermod -a -G microk8s $USER
sudo chown -f -R $USER ~/.kube
newgrp microk8s
microk8s enable dns storage helm3
echo "alias kubectl='microk8s kubectl '" >> .bashrc
echo "alias helm='microk8s helm3 '" >> .bashrc
source .bashrc
~~~
### 安装Devtron
~~~ bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set components.devtron.service.type=NodePort
~~~
### 运行以下命令以获取devtron-service端口号：
~~~ bash
kubectl get svc -n devtroncd devtron-service -o jsonpath='{.spec.ports[0].nodePort}'
~~~

确保运行devtron-service的端口在VM的安全组或网络安全组中保持开启状态。

**注意事项**：如果您想卸载Devtron或清理Devtron的helm安装程序，请参阅[卸载Devtron](https://docs.devtron.ai/install/uninstall-devtron)。

如果您有任何问题，请在我们的Discord频道联系我们。![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)[](https://discord.gg/jsRG5qx2gp)

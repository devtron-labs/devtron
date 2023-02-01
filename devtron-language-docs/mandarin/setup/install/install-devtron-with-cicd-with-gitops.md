# 安装带有CI/CD和GitOps（Argo CD）的Devtron
在本节中，我们将详细介绍如何通过在安装期间启用GitOps来安装带有CI/CD的Devtron的步骤。
## 开始之前：
如果您未安装[Helm](https://helm.sh/docs/intro/install/)请先安装。
## 安装带有CI/CD和GitOps（Argo CD）的Devtron
运行以下命令以安装带有CI/CD和GitOps（Argo CD）模块的最新版本Devtron：
~~~ bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set argo-cd.enabled=true
~~~

**注意事项**：如果要在安装过程中配置Blob存储，请参阅 [在安装时配置blob存储](#configure-blob-storage-duing-installation)。
## 安装多架构节点（ARM和AMD）
要在具有多架构节点（ARM和AMD）的集群上安装Devtron，请在Devtron安装命令后附加`--set installer.arch=multi-arch`。

**注意事项**：

* 要安装特定版本的Devtron，请在指令后加上`--set installer.release="vX.X.X"`，其中`vx.x.x`为[发布标签](https://github.com/devtron-labs/devtron/releases)。
* 如果你想为`生产部署`安装Devtron，请参阅我们对于[Devtron安装](override-default-devtron-installation-configs.md)的推荐复盖 。
## 在安装时配置Blob存储
在Devtron环境中配置Blob存储允许您存储构建日志和缓存。
如果您不配置Blob存储，则：

- 一小时后，您将无法访问构建和部署日志。
- 缓存不可用会造成提交散列的构建需要更长。
- 无法在构建前/构建后和部署阶段生成工件报告。

选择以下其中一个选项以配置blob存储：

{% tabs %}

{% tab title="MinIO Storage" %}

运行以下命令安装带有MinIO的Devtron，其用于存储日志和缓存。
~~~ bash
helm repo add devtron https://helm.devtron.ai 

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set minio.enabled=true \
--set argo-cd.enabled=true
~~~

**注意事项**：MinIO与AWS S3 Bucket、Azure Blob Storage和Google Cloud Storage等全球云供应商不同，因为它也可以在本地托管。

{% endtab %}

{% tab title="AWS S3 Bucket" %}

请参阅[日志和缓存的存储](./installation-configuration.md#aws-specific)页面上的`AWS特定`参数 。

运行以下命令以安装Devtron以及用于存储构建日志和缓存的AWS S3 Bucket：

* 使用S3 IAM策略安装。
> 注意：如果您使用以下命令，请确保S3对连接到集群节点的IAM角色的权限策略。
~~~ bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set configs.BLOB_STORAGE_PROVIDER=S3 \
--set configs.DEFAULT_CACHE_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CACHE_BUCKET_REGION=us-east-1 \
--set configs.DEFAULT_BUILD_LOGS_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CD_LOGS_BUCKET_REGION=us-east-1 \
--set argo-cd.enabled=true
~~~

* 使用access-key和secret-key进行AWS S3身份验证：
~~~ bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set configs.BLOB_STORAGE_PROVIDER=S3 \
--set configs.DEFAULT_CACHE_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CACHE_BUCKET_REGION=us-east-1 \
--set configs.DEFAULT_BUILD_LOGS_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CD_LOGS_BUCKET_REGION=us-east-1 \
--set secrets.BLOB_STORAGE_S3_ACCESS_KEY=<access-key> \
--set secrets.BLOB_STORAGE_S3_SECRET_KEY=<secret-key> \
--set argo-cd.enabled=true
~~~

* 使用S3兼容存储进行安装：
~~~ bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set configs.BLOB_STORAGE_PROVIDER=S3 \
--set configs.DEFAULT_CACHE_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CACHE_BUCKET_REGION=us-east-1 \
--set configs.DEFAULT_BUILD_LOGS_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CD_LOGS_BUCKET_REGION=us-east-1 \
--set secrets.BLOB_STORAGE_S3_ACCESS_KEY=<access-key> \
--set secrets.BLOB_STORAGE_S3_SECRET_KEY=<secret-key> \
--set configs.BLOB_STORAGE_S3_ENDPOINT=<endpoint> \
--set argo-cd.enabled=true
~~~

{% endtab %}

{% tab title="Azure Blob Storage" %}

请参阅[日志和缓存的存储](./installation-configuration.md#azure-specific)页面上的`Azure特定`参数。

运行以下命令以安装Devtron以及用于存储构建日志和缓存的Azure Blob存储：
~~~ bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set secrets.AZURE_ACCOUNT_KEY=xxxxxxxxxx \
--set configs.BLOB_STORAGE_PROVIDER=AZURE \
--set configs.AZURE_ACCOUNT_NAME=test-account \
--set configs.AZURE_BLOB_CONTAINER_CI_LOG=ci-log-container \
--set configs.AZURE_BLOB_CONTAINER_CI_CACHE=ci-cache-container \
--set argo-cd.enabled=true
~~~

{% endtab %}

{% tab title="Google Cloud Storage" %}

请参阅[日志和缓存的存储](./installation-configuration.md#google-cloud-storage-specific)页面上的`Google Cloud特定`参数。

运行以下命令以安装Devtron以及用于存储构建日志和缓存的Google Cloud存储：
~~~ bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set configs.BLOB_STORAGE_PROVIDER=GCP \
--set secrets.BLOB_STORAGE_GCP_CREDENTIALS_JSON=eyJ0eXBlIjogInNlcnZpY2VfYWNjb3VudCIsInByb2plY3RfaWQiOiAiPHlvdXItcHJvamVjdC1pZD4iLCJwcml2YXRlX2tleV9pZCI6ICI8eW91ci1wcml2YXRlLWtleS1pZD4iLCJwcml2YXRlX2tleSI6ICI8eW91ci1wcml2YXRlLWtleT4iLCJjbGllbnRfZW1haWwiOiAiPHlvdXItY2xpZW50LWVtYWlsPiIsImNsaWVudF9pZCI6ICI8eW91ci1jbGllbnQtaWQ+IiwiYXV0aF91cmkiOiAiaHR0cHM6Ly9hY2NvdW50cy5nb29nbGUuY29tL28vb2F1dGgyL2F1dGgiLCJ0b2tlbl91cmkiOiAiaHR0cHM6Ly9vYXV0aDIuZ29vZ2xlYXBpcy5jb20vdG9rZW4iLCJhdXRoX3Byb3ZpZGVyX3g1MDlfY2VydF91cmwiOiAiaHR0cHM6Ly93d3cuZ29vZ2xlYXBpcy5jb20vb2F1dGgyL3YxL2NlcnRzIiwiY2xpZW50X3g1MDlfY2VydF91cmwiOiAiPHlvdXItY2xpZW50LWNlcnQtdXJsPiJ9Cg== \
--set configs.DEFAULT_CACHE_BUCKET=cache-bucket \
--set configs.DEFAULT_BUILD_LOGS_BUCKET=log-bucket \
--set argo-cd.enabled=true
~~~

{% endtab %}
{% endtabs %}
## 检查Devtron安装状态
**注意事项**：安装大约需要15到20分钟才能逐个启动所有Devtron微服务。

运行以下命令检查安装状态：
~~~ bash
kubectl -n devtroncd get installers installer-devtron \
-o jsonpath='{.status.sync.status}'
~~~

该命令执行时带有以下输出消息之一，指示安装状态：

|状态|说明|
| :- | :- |
|`已下载`|安装程序已下载所有清单，安装正在进行中。|
|`已应用`|安装程序已成功应用所有清单，并且安装完成。|

## 检查安装程序日志
运行以下命令以检查安装程序日志：
~~~ bash
kubectl logs -f -l app=inception -n devtroncd
~~~
## Devtron仪表板
运行以下命令以获取Devtron仪表板URL：
~~~ bash
kubectl get svc -n devtroncd devtron-service \
-o jsonpath='{.status.loadBalancer.ingress}'
~~~

您将得到类似以下示例的输出：
~~~ bash
[map[hostname:aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com]]
~~~

使用主机名`aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com`（Loadbalancer URL）以访问Devtron仪表板。

**注意事项**：如果您没有收到主机名或收到"服务不存在"的信息，这意味着Devtron仍在安装。
请等待安装完成。

**注意事项**：你也可以使用一个与您的域/子域对应的`CNAME`条目指向要在自定义域访问的Loadbalancer URL。

|主机|类型|指向|
| :- | :- | :- |
|devtron.yourdomain.com|CNAME|aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com|

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

* 注意事项：如果您想卸载Devtron或清理Devtron的helm安装程序，请参阅[卸载Devtron](https://docs.devtron.ai/install/uninstall-devtron)。
* 有关安装，请也参阅[常见问题](https://docs.devtron.ai/install/faq-on-installation)。

**注意事项**：如果您有任何问题，请在我们的Discord频道上联系我们。![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)[](https://discord.gg/jsRG5qx2gp)

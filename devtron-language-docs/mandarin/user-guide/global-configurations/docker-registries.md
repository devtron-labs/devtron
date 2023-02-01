#容器注册表

容器注册表用于存储CI管道构建的映像。您可以使用您选择的任何容器注册表供应者设置容器注册表。它允许您使用易于使用的UI构建、部署和管理容器映像。

设置应用程序时，您可以在应用程序设置 > [构建设置]（https://docs.devtron.ai/usage/applications/creating-application/docker-build-configuration）部分中选择特定的容器注册表和存储库。


##添加容器注册表：

要添加容器注册表，请进入`全局设置`的`容器注册表`部分。点击\*\*添加容器注册表\*\*。

![]（https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/add-container-registry.jpg）

提供以下字段中的信息以添加容器注册表。

|字段|描述|

\| --- | --- |

| \*\*名称\*\*|为您的注册表提供一个名称，此名称将在[构建设置]（https://docs.devtron.ai/usage/applications/creating-application/docker-build-configuration）下拉列表中。|

|\*\*注册表类型\*|从下拉列表中选择注册表类型：<br><ul><li>[ECR]（#registry-type-ecr）</li></ul><ul><li>[Docker]（#registry-type-docker）</li></ul><ul><li>[Azure]（#registry-type-azure）</li></ul><ul><li>[工件注册表（GCP）]（#registry-type-artifact-registry-gcp）</li></ul><ul><li>[GCR]（#registry-type-google-container-registry-gcr）</li></ul><ul><li>[Quay]（#registry-type-quay）</li></ul><ul><li>[其他]（#registry-type-other）</li></ul>`注意`：每个\*\*注册表类型\*\*的凭据输入字段都不同。|

|\*\*注册表URL\*\*|提供您的注册表的URL。|

|\*\*设置为默认注册表\*\*|启用此字段以设置为图像的默认注册表中心。|



###注册表类型：ECR

Amazon ECR是一种AWS托管的容器映像注册服务。

ECR使用AWS Identity and Access Management（IAM）为私有存储库提供基于资源的权限。ECR允许基于密钥和基于角色的身份验证。

在开始之前，创建一个[IAM用户]（https://docs.aws.amazon.com/AmazonECR/latest/userguide/get-set-up-for-amazon-ecr.html）并根据认证类型附加ECR策略。

如果您选择注册表类型为`ECR`，请提供以下信息。

|字段|描述|

\| --- | --- |

|\*\*名称\*\*|Devtron中注册表的用户定义名称。|

|\*\*注册表类型\*\*|选择\*\*ECR\*\*。|

|\*\*注册表URL\*\*|这是您在AWS中的私有注册表的URL。<br>例如，URL格式为：`https://xxxxxxxxxxxx.dkr.ecr.<region>.amazonaws.com'.'xxxxxxxx'是您的12位AWS账户ID。</br>|

|\*\*身份验证类型\*\*|选择其中一种身份验证类型：<ul><li>\*\*EC2 IAM角色\*\*：使用workernode IAM角色进行身份验证，并将ECR策略（AmazonEC2ContainerRegistryFullAccess）附加到Kubernetes集群的集群工作节点IAM角色。</li></ul><ul><li>\*\*用户身份验证\*\*：它是基于密钥的身份验证，并将ECR策略（AmazonEC2ContainerRegistryFullAccess）附加到[IAM用户]（https://docs.aws.amazon.com/AmazonECR/latest/userguide/get-set-up-for-amazon-ecr.html）.<ul><li>`访问钥匙ID`：您的AWS访问钥匙</li></ul><ul><li>`访问密钥'：您的AWS密钥ID</li></ul>|

|\*\*设置为默认注册表\*\*|启用此字段以将`ECR`设置为图像的默认注册表中心。|

点击\*\*保存\*\*。


###登记类别：Docker

如果您选择注册表类型为`Docker`，请提供以下信息。

|字段|描述|

\| --- | --- |

|\*\*名称\*\*|Devtron中注册表的用户定义名称。|

|\*\*注册表类型\*\*|选择\*\*Docker\*\*。|

|\*\*注册表URL\*\*|这是Docker中您的私人注册表的URL。例如`docker.io`|

|\*\*用户名\*\*|提供您用于创建注册表的docker hub帐户的用户名。|

|\*\*密码/令牌\*\*|提供对应于您的docker hub帐户的密码/[令牌]（https://docs.docker.com/docker-hub/access-tokens/）。出于安全目的，建议使用`令牌`。|

|\*\*设置为默认注册表\*\*|启用此字段将`Docker`设置为映像的默认注册表中心。|

点击\*\*保存\*\*。

###注册表类型：Azure

对于注册表类型：Azure，服务主体身份验证方法可用于使用用户名和密码进行身份验证。请跟随[link]（https://docs.microsoft.com/en-us/azure/container-registry/container-registry-auth-service-principal）获取此注册表的用户名和密码。

如果选择注册表类型为`Azure`，请提供以下信息。

|字段|描述|

\| --- | --- |

|\*\*名称\*\*|Devtron中注册表的用户定义名称。|

|\*\*注册表类型\*\*|选择\*\*Azure\*\*。|

|\*\*注册表URL/登录服务器\*\*|这是Azure中私有注册表的URL。例如`xxx.azurecr.io`|

|\*\*用户/注册表名称\*\*|提供Azure容器注册表的用户名。|

|\*\*密码\*\*|提供Azure容器注册表的密码。|

|\*\*设置为默认注册表\*\*|启用此字段以将`Azure`设置为图像的默认注册表中心。|

点击\*\*保存\*\*。


###注册表类型：工件注册表（GCP）

JSON密钥文件身份验证方法可用于使用用户名和服务帐户JSON文件进行身份验证。请按[连结]（https://cloud.google.com/artifact-registry/docs/docker/authentication#json-key）获取此注册表的用户名和服务帐户JSON文件。

\*\*注意\*\*：请从json键中删除所有空格，并将其包装在单引号中，同时将其放入`服务帐户JSON文件`字段。


如果选择注册表类型为`工件注册表（GCP）`，请提供以下信息。

|字段|描述|

\| --- | --- |

|\*\*名称\*\*|Devtron中注册表的用户定义名称。|

|\*\*注册表类型\*\*|选择\*\*工件注册表（GCP）\*\*。|

|\*\*注册表URL\*\*|这是您在Artifact Registry（GCP）中的私有注册表的URL。例如`region-docker.pkg.dev`|

|\*\*Username\*\*|提供工件注册表（GCP）帐户的用户名。|

|\*\*服务账户JSON文件\*\*|提供弓箭注册表（GCP）的服务账户JSON文件。|

|\*\*设置为默认注册表\*\*|启用此字段以将`工件注册表（GCP）`设置为图像的默认注册表中心。|

点击\*\*保存\*\*。

###注册表类型：Google Container Registry（GCR）

JSON密钥文件身份验证方法可用于使用用户名和服务帐户JSON文件进行身份验证。请按[连结]（https://cloud.google.com/container-registry/docs/advanced-authentication#json-key）获取此注册表的用户名和服务帐户JSON文件。请从json键中删除所有空格，并将其包装在单引号中，同时放入`服务帐户JSON文件`字段。

如果您选择注册表类型为`GCR`，请提供以下信息。

|字段|描述|

\| --- | --- |

|\*\*名称\*\*|Devtron中注册表的用户定义名称。|

|\*\*注册表类型\*\*|选择\*\*GCR\*\*。|

|\*\*注册表URL\*\*|这是您在GCR中的私人注册表的URL。例如`gcr.io`|

|\*\*用户名\*\*|提供您的GCR帐户的用户名。|

|\*\*服务账户JSON文件\*\*|提供GCR的服务账户JSON文件。|

|\*\*设置为默认注册表\*\*|启用此字段以将`GCR`设置为图像的默认注册表中心。|

点击\*\*保存\*\*。

###注册表类型：Quay

如果选择注册表类型为`Quay`，请提供以下信息。

|字段|描述|

\| --- | --- |

|\*\*名称\*\*|Devtron中注册表的用户定义名称。|

|\*\*注册表类型\*\*|选择\*\*\*\*\*\*。|

|\*\*注册表URL\*\*|这是您在码头的私人注册表的URL。例如`quay.io`|

|\*\*Username\*\*|提供码头帐户的用户名。|

|\*\*密码/令牌\*\*|提供您的码头帐户的密码。|

|\*\*设置为默认注册表\*\*|启用此字段以将`Quay`设置为图像的默认注册表中心。|

点击\*\*保存\*\*。


###注册表类型：其他


如果您选择注册表类型为`其他`，请提供以下信息。

|字段|描述|

\| --- | --- |

|\*\*名称\*\*|Devtron中注册表的用户定义名称。|

|\*\*注册表类型\*\*|选择\*\*其他\*\*。|

|\*\*注册表URL\*\*|这是您的私人注册表的URL。|

|\*\*用户名\*\*|提供您创建注册表的帐户的用户名。|

|\*\*密码/令牌\*\*|提供与注册表用户名对应的密码/令牌。|

|\*\*设置为默认注册表\*\*|启用此字段以设置为图像的默认注册表中心。|

点击\*\*保存\*\*。

####高级注册表URL连接选项：

* 如果启用`仅允许安全连接`选项，则此注册表仅允许安全连接。
* 如果您启用`允许与CA证书的安全连接`选项，那么您必须上传/提供私有CA证书（ca.crt）。
* 如果容器注册表不安全（例如：SSL证书已过期），则启用`允许不安全连接`选项。

\*\*注意\*\*：您可以使用任何可以使用`docker login-u<username>-p<password><registry-url>`进行身份验证的注册表。但是，这些注册管理机构可能会为身份验证提供更安全的方式，我们将在稍后提供支持。


##从私有注册表中提取图像

您可以创建一个Pod，使用`秘密`从私有容器注册表中提取映像。您可以使用您选择的任何私有容器注册表。作为例子：[Docker Hub]（https://www.docker.com/products/docker-hub）。

超级管理员用户可以决定是要自动注入注册表凭据，还是使用密钥提取映像以部署到特定群集上的环境。

要管理注册表凭据的访问请点击\*\*管理\*\*。

有两个选项可以管理注册表凭据的访问：

|字段|描述|

\| --- | --- |

|\*\*不要向集群注入凭据\*\*|选择不希望为其注入凭据的集群。|

|\*\*向集群自动注入凭据\*\*|选择要为其注入凭据的集群。|

您可以选择用于定义凭据的两个选项之一：

* [用户注册表凭据]（#use-registry-credentials）
* [指定图像拉密]（#specify-image-pull-secret）

###使用注册表凭据

如果您选择\*\*使用注册表凭据\*\*，集群将自动注入您的注册表类型的注册表凭据。例如，如果您选择`Docker`作为注册表类型，那么集群将自动注入您在Docker Hub帐户上使用的`用户名`和`密码/令牌`。

点击\*\*保存\*\*。

![]（https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/use-registry-credentials.jpg）


###指定图片拉密

您可以通过在命令行上提供凭据来创建密钥。

![]（https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/specify-image-pull-secret-latest.png）

创建这个秘密，将其命名为`regcred`：

\```bash

kubectl create-n<namespace>secret docker-registry regcred--docker-server=<your-registry-server>--docker-username=<your-name>--docker-password=<your-pword>--docker-email=<your-email>

\```

其中：

* <namespace>是您的虚拟集群。例如，devtron-demo
* <your-registry-server>是您的私有Docker注册表FQDN对于DockerHub使用https://index.docker.io/v1/。
* <your-name>是您的Docker用户名。
* <your-pword>是您的Docker密码。
* <your-email>是您的Docker电子邮件。

您已经成功地将集群中的Docker凭据设置为一个名为`regcred`的秘密。

\*\*注意\*\*：在命令行中键入秘密可能会不受保护地将它们存储在shell历史记录中，并且在kubectl运行时，这些secrets也可能对PC上的其他用户可见。

在字段中输入`秘密`名称，然后点击\*\*保存\*\*。























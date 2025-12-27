# Install Devtron on AWS EKS

Devtron can be installed on any Kubernetes cluster. This cluster can use upstream Kubernetes, or it can be a managed Kubernetes cluster from a cloud provider such as [AWS EKS](https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html).

In this section, we will walk you through the steps of installing Devtron on [AWS EKS](https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html). To install the Devtron v6.0 on AWS EKS, the `EKS version` must not be higher then `v1.22`.

For installing AWS EKS `v1.23`, you must run the additional command provided in [step 6]()

If you already have an [AWS EKS](https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html) Kubernetes cluster, go to the step on installing an [ingress](https://docs.devtron.ai/v/v0.6/getting-started/install/ingress-setup). Then, install the `Helm chart` following the instructions on this [page](https://helm.sh/docs/intro/install/).

## Install the AWS CLI

First we need to install the AWS CLI. Below is an example for macOS and please refer to [Getting Started EKS](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-console.html) for other operating systems.

```bash
pip3 install awscli --upgrade --user
```
Check the installation with `aws --version`.

## Create an AWS EKS Cluster

A standard Kubernetes cluster in AWS is a prerequisite for installing Devtron. 

1. On the AWS Management Console.
2. On the `Search` bar, type `Elastic Kubernetes Service`.
3. Select `Add Cluster` and then click `Create` to create EKS cluster.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/install-devtron-on-AWS-EKS/aws-eks-add-cluster.jpg)

4. On the **Configure cluster** page, provide the information in the following fields:

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/install-devtron-on-AWS-EKS/configure-cluster.jpg)

| Fields | Description |
| --- | --- |
| **Name** | A unique name for your cluster. E.g., `ks-install`.|
| **Kubernetes version** | The version of Kubernetes to use for your cluster. Default value: 1.23. <br>**Note**: To install Devtron on Kubernetes, your Kubernetes version must be higher than `v1.16` but must be lesser than `v1.23`.</br> |
| **Cluster Service role** | Select the IAM role that you created with [Create your Amazon EKS cluster IAM role](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-console.html#role-create). |

5. You can leave the remaining settings at their default values and click **Next**.
6. On the **Specify networking** page, provide the information in the following fields:
 
 ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/install-devtron-on-AWS-EKS/specify-networking.jpg)

| Fields | Description |
| --- | --- |
| **VPC** | The VPC that you created previously in [Create your Amazon EKS cluster VPC](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-console.html#vpc-create). You can find the name of your VPC in the drop-down list. E.g., `vpc-44ffe12` which is a default value.|
| **Subnets** | By default, the available subnets in the VPC specified in the previous field are preselected. Select any subnet that you do not want to host cluster resources, such as worker nodes or load balancers. |
| **Choose cluster IP address family** | Specify the IP address type for pods and services in your cluster: `IPv4` or `IPv6`.<br> **Note**: Enable Configure Kubernetes service IP address range to enter the CIDR block, if required.</br> |
7. You can leave the remaining settings at their default values and click **Next**.
8. On the **Cluster endpoint access**, choose one of the following options and click **Next**:
    * `Public`
    * `Public and private`
    * `Private`

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/install-devtron-on-AWS-EKS/cluster-endpoint-access-aws-eks.jpg)

9. If you want to configure add-ons that provide advanced networking functionalities on the cluster, you can configure them on the **Networking add-ons** page and click **Next**. Or you can skip this step.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/install-devtron-on-AWS-EKS/networking-add-ons.jpg)

10. On the **Configure logging** page, click **Next**.<br>By default, each log type is Disabled. You can optionally choose which log types that you want to enable. For more information, see [Amazon EKS control plane logging](https://docs.aws.amazon.com/eks/latest/userguide/control-plane-logs.html).</br>

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/install-devtron-on-AWS-EKS/configure-logging.jpg)

11. On the **Review and create** page, review the information that you entered or selected on the previous pages. Click **Edit** if you need to make changes to any of your selections. Once you verify your settings, click **Create**. 

1
![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/install-devtron-on-AWS-EKS/cluster-creating-status.jpg)

 **Note**: 
 * The **Status** field shows as **Creating** for several minutes until the cluster provisioning process completes. Do not continue to the next step until the status is **Active**. 
 * If you want to modify the endpoint access for an existing cluster, refer [Modifying cluster endpoint access](https://docs.aws.amazon.com/eks/latest/userguide/cluster-endpoint.html#modify-endpoint-access).

12. Once the cluster provisioning is completed (status: `Active`), save the `API server endpoint` and `Certificate authority` values from the **Overview** details. These are used in your kubectl configuration.

13. Go to the **Compute** section, on the **Node groups**, click **Add node group** to define a minimum of 2 nodes in your cluster.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/install-devtron-on-AWS-EKS/add-node-group1.jpg)

14. On the **Configure node group** page, provide the information in the following fields and click **Next**.

| Fields | Description |
| --- | --- |
| **Name** | Enter a unique name for this node group. E.g., ks-nodes.|
| **Node IAM role** | Select the IAM role that will be used by the nodes. To create a new role, go to the [IAM console](https://us-east-2.console.aws.amazon.com/iam/home?#roles). |

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/install-devtron-on-AWS-EKS/configure-node-group.png)

15. On the **Set compute and scaling configuration**, you can leave settings at their default values and click **Next**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/install-devtron-on-AWS-EKS/set-compute-and-scaling-config-node.jpg)

16. On the **Specify networking** page, you can leave the settings at their default values and click **Next**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/install-devtron-on-AWS-EKS/specify-networking-node.jpg)

17. On the **Review and Create** page, verify the information that you entered or selected on the previous pages and click **Create**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/install-devtron-on-AWS-EKS/review-and-create-node.jpg)

18. You will see the message **The Node group creation in progress**. As soon as the nodes creation are completed, the EKS cluster is ready and you can connect to the cluster with `kubectl`.



## Configure your computer to communicate with your cluster

In this section, you create a kubeconfig file for your cluster. The settings in this file enable the kubectl CLI to communicate with your cluster.

1. Install AWS CLI Command by running the following command:

```bash
curl "https://awscli.amazonaws.com/AWSCLIV2.pkg" -o "AWSCLIV2.pkg"
sudo installer -pkg AWSCLIV2.pkg -target /
```

2. Next to configure AWS, run the following command:

```bash
aws configure
```
```
AWS Access Key ID [None]: AKIAIOSFODNN7EXAMPLE
AWS Secret Access Key [None]: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
Default region name [None]: region-code
Default output format [None]: json
```

3. Create or update a `kubeconfig` file for your cluster. Replace `region-code` with the AWS Region that you created your cluster in. Replace `my-cluster` with the name of your cluster.

```bash
aws eks update-kubeconfig --region region-code --name my-cluster
```

4. To install `kubectl`, run the following command: 
   Below is an example for macOS.

```bash
brew install kubectl
```

5. To install `helm`, run the following command:
    Below is an example for macOS.

```bash
brew install helmq
```

6. If you have installed the Kubernetes version higher than `v1.22`, then you must run the following command to get the CSI driver:

```bash
helm repo add aws-ebs-csi-driver https://kubernetes-sigs.github.io/aws-ebs-csi-driver
helm repo update
helm upgrade --install aws-ebs-csi-driver --namespace kube-system aws-ebs-csi-driver/aws-ebs-csi-driver
kube-system
```

**Note**: You can also EKS cluster using CLI command `eksctl`. Refer [here](https://github.com/devtron-labs/utilities/tree/main/eksctl-configs) for more detail.


## Install Devtron with CI/CD on AWS EKS 

* Install Devtron with CI/CD with the following command:

```bash
helm repo add devtron https://helm.devtron.ai
```
```bash
helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd}
```

* To check the status of pods, run the following command:

```bash
kubectl get po -n devtroncd 
```

* Or, to track the progress of Devtron microservices installation, run the following command:

```bash
kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.status}'
```
The command executes with one of the following output messages, indicating the status of the installation:

| Status | Description |
| :--- | :--- |
| `Downloaded` | The installer has downloaded all the manifests, and the installation is in progress. |
| `Applied` | The installer has successfully applied all the manifests, and the installation is complete. |

Once the status is `Downloaded`, you can access the Devtron dashoboard URL.

* To get the Devtron dashboard URL, run the following command:

 ```bash
  kubectl get svc -n devtroncd devtron-service \
 -o jsonpath='{.status.loadBalancer.ingress}'
 ```

* You will get an output as shown below (example):

```bash
[map[hostname:aaff16e9760594a92afa0140dbfd99f7-305259315. \
 us-east-1.elb.amazonaws.com]]
 ```

Use the hostname `aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com` to access the Devtron dasboard.

* To get the password for the default admin user, run the following command:

```bash
 kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ADMIN_PASSWORD}' | base64 -d
 ```


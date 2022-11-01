# Install Devtron on AWS EKS

Devtron can be installed on any Kubernetes cluster. This cluster can use upstream Kubernetes, or it can be a managed Kubernetes cluster from a cloud provider such as [AWS EKS](https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html).

In this section, we will walk you through the steps of installing Devtron on [AWS EKS](https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html).

If you already have an [AWS EKS](https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html) Kubernetes cluster, go to the step on installing an [ingress](https://docs.devtron.ai/v/v0.6/getting-started/install/ingress-setup). Then, install the `Helm chart` following the instructions on this [page](https://helm.sh/docs/intro/install/).

## Install the AWS CLI

First we need to install the AWS CLI. Below is an example for macOS and please refer to [Getting Started EKS](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-console.html) for other operating systems.

```bash
pip3 install awscli --upgrade --user
```
Check the installation with `aws --version`.

## Create an AWS EKS Cluster

A standard Kubernetes cluster in AWS is a prerequisite of installing Devtron. 

1. On the AWS Management Console.
2. On the `Search` bar, type `Elastic Kubernetes Service`.
3. Select `Add Cluster` and then click `Create` to create EKS cluster.
4. On the **Configure cluster** page, provide the information in the following fields:

| Fields | Description |
| --- | --- |
| **Name** | A unique name for your cluster. E.g., `my-cluster`.|
| **Kubernetes version** | The version of Kubernetes to use for your cluster. |
| **Cluster Service role** | Select the IAM role that you created with [Create your Amazon EKS cluster IAM role](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-console.html#role-create). |

5. You can leave the remaining settings at their default values and click **Next**.
6. On the **Specify networking** page, provide the information in the following field:
 * **VPC**:  The VPC that you created previously in [Create your Amazon EKS cluster VPC](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-console.html#vpc-create). You can find the name of your VPC in the drop-down list. E.g., `vpc-00x0000x000x0x000 | my-eks-vpc-stack-VPC`.
7. You can leave the remaining settings at their default values and click **Next**.
8. On the **Cluster endpoint access**, choose one of the following options and click **Next**.
    * `Public`
    * `Public and private`
    * `Private`

9. On the **Configure logging** page, click **Next**. By default, each log type is Disabled. For more information, see [Amazon EKS control plane logging](https://docs.aws.amazon.com/eks/latest/userguide/control-plane-logs.html).
10. On the **Review and create** page, review the information that you entered or selected on the previous pages. Click **Edit** if you need to make changes to any of your selections. Once you verify your settings, click **Create**. 

 **Note**: To the right of the cluster's name, the cluster status will display as **Creating** for several minutes until the cluster provisioning process completes. Do not continue to the next step until the status is **Active**.

11. Once the cluster provisioning is completed (status: `Active`), save the API server endpoint and Certificate authority values. These are used in your kubectl configuration.

12. You can create a cluster with one of the following node types:
    * `Fargate – Linux`: Select this type of node if you want to run Linux applications on AWS Fargate. Fargate is a serverless compute engine that lets you deploy Kubernetes pods without managing Amazon EC2 instances.
    * `Managed nodes – Linux`: Select this type of node if you want to run Amazon Linux applications on Amazon EC2 instances.

To learn more about each type, see [Amazon EKS nodes](https://docs.aws.amazon.com/eks/latest/userguide/eks-compute.html). After your cluster is deployed, you can add other node types.
13. When the EKS cluster is ready, you can connect to the cluster with `kubectl`.

## Configure your computer to communicate with your cluster

In this section, you create a kubeconfig file for your cluster. The settings in this file enable the kubectl CLI to communicate with your cluster.

1. Create or update a `kubeconfig` file for your cluster. Replace `region-code` with the AWS Region that you created your cluster in. Replace `my-cluster` with the name of your cluster.

```bash
aws eks update-kubeconfig --region region-code --name my-cluster
```
By default, the configuration file is created in `~/.kube` or the new cluster configuration is added to an existing `config` file in `~/.kube`.

2. Test your configuration with the following command:

```bash
kubectl get svc
```



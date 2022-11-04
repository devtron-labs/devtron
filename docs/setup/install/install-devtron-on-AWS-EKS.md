# Install Devtron on AWS EKS

Devtron can be installed on any Kubernetes cluster. This cluster can use upstream Kubernetes, or it can be a managed Kubernetes cluster from a cloud provider such as [AWS EKS](https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html).

In this section, we will walk you through the steps of installing Devtron on [AWS EKS](https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html). To install the Devtron 6.0 on AWS EKS, the EKS must not be higher then `v1.22`.

For installing AWS EKS `v1.23`, you must run the additional command provided in [step 6]()

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
| **Name** | A unique name for your cluster. E.g., `ks-install`.|
| **Kubernetes version** | The version of Kubernetes to use for your cluster. Default value: 1.23 |
| **Cluster Service role** | Select the IAM role that you created with [Create your Amazon EKS cluster IAM role](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-console.html#role-create). |

5. You can leave the remaining settings at their default values and click **Next**.
6. On the **Specify networking** page, provide the information in the following fields:
 
 | Fields | Description |
| --- | --- |
| **VPC** | The VPC that you created previously in [Create your Amazon EKS cluster VPC](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-console.html#vpc-create). You can find the name of your VPC in the drop-down list. E.g., `vpc-00x0000x000x0x000 | my-eks-vpc-stack-VPC`.|
| **Subnets** | By default, the available subnets in the VPC specified in the previous field are preselected. Select any subnet that you do not want to host cluster resources, such as worker nodes or load balancers. |
| **Choose cluster IP address family** | Specify the IP address type for pods and services in your cluster: IPv4 IPv6. **Note**: Enable Configure Kubernetes service IP address range to enter the CIDR block, if required. |
7. You can leave the remaining settings at their default values and click **Next**.
8. On the **Cluster endpoint access**, choose one of the following options:
    * `Public`
    * `Public and private`
    * `Private`

9. If you want to configure add-ons that provide advanced networking functionalities on the cluster, you can configure them on the **Networking add-ons** page and click **Next**. Or you can skip this step.

10. On the **Configure logging** page, click **Next**. By default, each log type is Disabled. You can optionally choose which log types that you want to enable. For more information, see [Amazon EKS control plane logging](https://docs.aws.amazon.com/eks/latest/userguide/control-plane-logs.html).

11. On the **Review and create** page, review the information that you entered or selected on the previous pages. Click **Edit** if you need to make changes to any of your selections. Once you verify your settings, click **Create**. 

 **Note**: 
 * The **Status** field shows as **Creating** for several minutes until the cluster provisioning process completes. Do not continue to the next step until the status is **Active**. 
 * If you want to modify the endpoint access for an existing cluster, refer [Modifying cluster endpoint access](https://docs.aws.amazon.com/eks/latest/userguide/cluster-endpoint.html#modify-endpoint-access).

12. Once the cluster provisioning is completed (status: `Active`), save the `API server endpoint` and `Certificate authority` values from the **Overview** details. These are used in your kubectl configuration.

13. Go to the **Compute** section, on the **Node groups**, click **Add node group** to define a minimum of 2 nodes in your cluster.

14. On the **Configure node group** page, provide the information in the following fields and click **Next**.

| Fields | Description |
| --- | --- |
| **Name** | Enter a unique name for this node group. E.g., ks-nodes.|
| **Node IAM role** | Select the IAM role that will be used by the nodes. To create a new role, go to the [IAM console](https://us-east-2.console.aws.amazon.com/iam/home?#roles). |

15. On the **Set compute and scaling configuration**, you can leave settings at their default values and click **Next**.

16. On the **Specify networking** page, you can leave the settings at their default values and click **Next**.

17. On the **Review and Create** page, verify the information that you entered or selected on the previous pages and click **Create**.

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

* You will get an output similar to the example as shown below:

```bash
[test2@server ~]$ kubectl get svc -n devtroncd devtron-service \
-o jsonpath='{.status.loadBalancer.ingress}'
[map[hostname:aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com]]
```

You can access the Devtron dasboard URL using the hostname `aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com` as shown in the above example.

* To get the password for the default admin user, run the following command:

```bash
 kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ADMIN_PASSWORD}' | base64 -d
 ```


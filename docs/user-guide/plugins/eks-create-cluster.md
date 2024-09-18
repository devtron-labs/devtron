# EKS Create Cluster

## Introduction
The **EKS Create Cluster** plugin streamlines the creation of a Amazon Elastic Kubernetes Service (EKS) cluster through the eksctl config file or by providing some required input parameters. This plugin automates the provisioning of EKS by which you can save time, ensures consistency in cluster setup, maintain security standards, and provides a Kubernetes-ready environment for deploying your containerized applications.

### Prerequisites
Before integrating the **EKS Create Cluster** plugin make sure that you have a AWS account with valid permissions to provision EKS.

---

## Steps
1. Navigate to the **Jobs** section, click **Create**, and choose **Job**.
2. In the 'Create job' window, enter **Job Name** and choose a target project.
3. Click **Create Job**.
4. In the 'Configurations' tab, fill the required fields under the 'Source code' section and click **Save**.
5. In Workflow Editor, click **+ Job Pipeline**.
6. Give a name to the workflow and click **Create Workflow**.
7. Click **Add job pipeline to this workflow**.
8. Fill the required fields in ‘Basic configuration’ tab.
9. Go to the ‘Tasks to be executed’ tab.
10. Under ‘Tasks’, click the **+ Add task** button.
11. Select the **EKS Create Cluster** plugin.
12. Enter the following [user inputs](#user-inputs) with appropriate values.
---

## User Inputs

### Task Name
Enter the name of your task 

e.g., `EKS Create Cluster`

### Description
Add a brief explanation of the task and the reason for choosing the plugin. Include information for someone else to understand the purpose of the task.

e.g., `The EKS Create Cluster plugin is integrated for provisioning of EKS cluster.`

### Input Variables

| Variable                 | Format       | Description | Sample Value |
| ------------------------ | ------------ | ----------- | ------------ |
|   MinNodes               | STRING       | Minimum number of nodes in the EKS NodeGroup.          |      1        |
|   MaxNodes               | STRING       | Maximum number of nodes in the EKS NodeGroup.          |      1        |
|   UseEKSConfigFile       | BOOL         | Flag to use a config file for EKS cluster creation (true/false).    |      false        |
|   EKSConfigFilePath      | STRING       | Path to the EKS cluster configuration file (required if `UseEKSConfigFile` flag is true).|      ~/.kube/config        |
|   EnablePlugin           | BOOL         | Flag to enable the **EKS Create Cluster** plugin.      |      true        |
|   AutomatedName          | BOOL         | Flag to enable naming for the cluster (true/false).                  |    false          |
|   UseIAMNodeRole         | BOOL         | Flag to use IAM Node Role for EKS cluster creation.    |    false          |
|   AWSAccessKeyId         | STRING       | Valid AWS access key ID for authentication.            |   VtbXliYXNlNjR2YWx1           |
|   AWSSecretAccessKey     | STRING       | AWS secret access key for authentication.              |   Njknsdcwjnchwjn34nk          |
|   ClusterName            | STRING       | Name for the EKS cluster.                              |   plugin-test-2           |
|   Version                | STRING       | Kubernetes version to use for the EKS cluster.         |    1.30          |
|   Region                 | STRING       | AWS region where the EKS cluster will be provisioned.  |    ap-south-1          |
|   Zones                  | STRING       | Availability zone for the EKS cluster.                 |    ap-south-1a,ap-south-1b          | 
|   NodeGroupName          | STRING       | Name for the EKS cluster's NodeGroup.                  |   plugin-test-1           |
|   NodeType               | STRING       | EC2 instance type for EKS worker nodes.                |   t3.medium           |
|   DesiredNodes           | STRING       | Desired number of nodes in the EKS cluster.            |     1         |


### Trigger/Skip Condition
Here you can set conditions to execute or skip the task. You can select `Set trigger conditions` for the execution of a task or `Set skip conditions` to skip the task.

### Output Variables
| Variable                 | Format       | Description | 
| ------------------------ | ------------ | ----------- |
| CreatedClusterName       | STRING       | Name of the created EKS cluster. |   
| EKSKubeConfigPath        | STRING       | File path of the generated EKS cluster kubeconfig   |


Click **Update Pipeline**.



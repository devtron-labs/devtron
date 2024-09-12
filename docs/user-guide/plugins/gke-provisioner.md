# GKE Provisioner

## Introduction
This plugin streamlines the creation and configuration of a Google Kubernetes Engine (GKE) cluster on your Google Cloud Platform (GCP). It automates the provisioning process while implementing essential security measures, including a preconfigured firewall that allows access to SSH, HTTP (port 80), 8080, and Kubernetes NodePorts. By automating the GKE provisioning process through this plugin, you can save time, ensure consistency in cluster setup, maintain security standards, and provide a Kubernetes-ready environment for deploying your containerized applications. 

{% hint style="warning" %}
The GKE Provisioner plugin, while primarily recommended for Job Pipelines, but can also be effectively integrated into the PRE and POST stages of CI/CD pipelines.
{% endhint %}

### Prerequisites
Before integrating the **GKE Provisioner** plugin make sure that you have an GCP account with valid permissions to provision GKE.

---

## Steps
1. Go to Applications → **Devtron Apps**.
2. Click on your application.
3. Go to **App Configuration** → **Workflow Editor**.
4. Click **New Workflow** and navigate to the **CREATE JOB PIPELINE**.
5. Enter the required fields in the **Basic configuration** window.
6. Under 'TASKS', click the **+ Add task** button.
7. Click the **GKE Provisioner**.
8. Enter the following [user inputs](#user-inputs) with appropriate values.
---

## User Inputs

### Task Name
Enter the name of your task 

e.g., `GKE Provisioner`

### Description
Add a brief explanation of the task and the reason for choosing the plugin. Include information for someone else to understand the purpose of the task.

e.g., `The GKE Provisioner plugin is integrated for provisioning of GKE cluster.`

### Input Variables

| Variable                 | Format       | Description | Sample Value |
| ------------------------ | ------------ | ----------- | ------------ |
|   GcpServiceAccountEncodedCredential  | STRING       | GCP Service Account credentials (base64 encoded) for GKE cluster creation.      | ZHVtbXliYXNlNjR2YWx1ZQ== |
|   GkeMinNodes                         | STRING       |  Minimum node count for the GKE cluster (default: 1)           | 2 |
|   DisplayGkeKubeConfig                | BOOL         |  Flag to determine if the GKE Kubeconfig should be displayed.  | true |
|   Identifier                          | STRING       |  Brief description of the GKE cluster's purpose or characteristics | plugin-demo-test |
|   GkeMaxNodes                         | STRING       | Maximum node count for the GKE cluster (default: 3).| 4 |
|   GkeNodeServiceAccountName           | STRING       | Custom GCP service account name for node VMs (uses project default if not specified) | gke-node-service-account-xyz123 |
|   GkeRegion                           | STRING       | GCP region for cluster provisioning (default: us-central1).| us-central1  |
|   GkeMachineType                      | STRING       |  Machine type for GKE nodes (default: n1-standard-4).| e2-medium |
|   GkeImageType                        | STRING       | OS image for GKE nodes (default: COS_CONTAINERD).| COS_CONTAINERD  |
|   GcpProjectId                        | STRING       | GCP project ID where the GKE cluster will be created.| gepton-393706 |
|   GkeClusterVersion                   | STRING       | Kubernetes version for the GKE cluster.               | 1.30.2-gke.1587003 |

### Trigger/Skip Condition
Here you can set conditions to execute or skip the task. You can select `Set trigger conditions` for the execution of a task or `Set skip conditions` to skip the task.

### Output Variables
| Variable                 | Format       | Description | 
| ------------------------ | ------------ | ----------- |
|   GkeKubeconfigFilePath | STRING        | File path of the generated GKE cluster kubeconfig |   

Click **Update Pipeline**.



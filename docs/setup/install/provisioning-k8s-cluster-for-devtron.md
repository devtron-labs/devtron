# provision Kubernetes Cluster For Devtron

To set up Devtron, a Kubernetes cluster is required. This documentation covers the process of creating production greade EKS, GKE, and AKS clusters.

## AWS EKS Cluster

To create an EKS cluster, you'll need the eksctl cluster configuration file. You can download it from [here](https://github.com/devtron-labs/utilities/blob/main/eksctl-configs/eksctl-devtron-prod-configs.yaml). 

For a smooth setup process and to understand the prerequisites, refer to the step-by-step guidance provided in the README file, accessible [here](https://github.com/devtron-labs/utilities/blob/main/eksctl-configs/README.md).

## GCP GKE Cluster

To create a GKE cluster, we have a terraform script that you van use. You can download the script file from [here](https://github.com/devtron-labs/utilities/gke-terraform/gke-cluster-terraform.tf).

Creating a Google Kubernetes Engine (GKE) cluster using Terraform involves several steps and prerequisites. Here's a general guide to help you set up a GKE cluster using Terraform:

### Prerequisites:

1. **Install Google Cloud SDK**:

    Install the [Google Cloud SDK](https://cloud.google.com/sdk/docs/install) on your local machine.

2. **Configure Google Cloud SDK**:

    Run gcloud auth login to authenticate and set up your default GCP project.

3. **Create a GCP Project**:

    If you don't have a project, create one using the [Google Cloud Console](https://console.cloud.google.com/).

4. **Enable the Kubernetes Engine API**:

    Enable the Kubernetes Engine API for your GCP project through the [Google Cloud Console](https://console.cloud.google.com/).


### Terraform Setup:

1. **Install Terraform:**

    Install [Terraform](https://developer.hashicorp.com/terraform/downloads) on your local machine.

2. **Configure Google Cloud Provider in Terraform script file:**

    ```yaml
    provider "google" {
        credentials = file("path/to/your/credentials.json")
        project     = "your-gcp-project-id"
        region      = "your-gcp-region"
    }
    ```


    Replace placeholders like `path/to/your/credentials.json`, `your-gcp-project-id`, and `your-gcp-region` with your actual values.

    To obtain credentials for the Google Cloud provider to use in this Terraform script, you can follow the steps mentioned below.

    **Create a Service Account**:

    * Go to the Google Cloud Console.
    * Navigate to the `IAM & Admin` page.
    * Click on `Service accounts` and create a new service account.
    * Assign the necessary roles to the service account from `IAM`.

        * **Compute Admin**

            This role is required to create and manage the Google Compute Engine resources associated with your GKE cluster.

        * **Compute Storage Admin**

            Your GKE cluster uses Google Cloud Storage (for example, to store cluster logs), you may need this role.

        * **Compute Network Admin**

            This role is required for managing and maintaining the network infrastructure for the GKE cluster.

        * **Create Service Accounts**

            Required to create service accounts.

    **Download JSON Key File**:

    * After creating the service account, click on it in the `Service accounts` page.
    * Navigate to the `Keys` tab.
    * Click `Add Key` and choose JSON. This will download a JSON key file containing the credentials.

    **Place JSON Key File**:

    * Place the downloaded JSON key file in a secure location on your local machine.
    * Update the path in your Terraform configuration to point to this JSON key file.


3. **Initialize and Apply Terraform Configuration:**

    ```bash
    terraform init
    terraform apply
    ```

### Get Credentials for GKE Cluster:

Use the following command to get credentials for your GKE cluster. Replace `your-gke-cluster` with the name of your GKE cluster, and `your-gcp-region` with the GCP region where your cluster is located:

```bash
gcloud container clusters get-credentials your-gke-cluster --region your-gcp-region
```

This command updates your kubeconfig file with the credentials for the specified GKE cluster, allowing you to use kubectl to interact with the cluster.

### Install Kubectl Authentication Plugin

You will need to install the gke-gcloud-auth-plugin binary on all systems where kubectl or Kubernetes custom clients are used. To install the binary using "gcloud components install", run the following command:

```bash
gcloud components install gke-gcloud-auth-plugin
```

To verify the installation, run:

```bash
gke-gcloud-auth-plugin --version
```

For more installation methods, you can visit [here](https://cloud.google.com/blog/products/containers-kubernetes/kubectl-auth-changes-in-gke).

To read more about 


### Verify Connection:

Run the following command to verify that kubectl is now configured to use your GKE cluster:

```bash
kubectl get nodes
```

This should show you the list of nodes in your GKE cluster. But if you st

Now you have the necessary credentials to manage your GKE cluster using `kubectl`.



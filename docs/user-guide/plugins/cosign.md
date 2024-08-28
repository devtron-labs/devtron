# Cosign

## Introduction
The **Cosign** plugin by Devtron enables secure signing of your container images, enhancing supply chain security. It authenticates your identity as the creator and ensures image integrity, allowing users to verify the source and detect any tampering. This provides greater assurance to developers incorporating your artifacts into their workflows.

### Prerequisites
Before integrating the Cosign plugin, ensure that you have configured the [Cosign](https://github.com/sigstore/cosign) and have a set of private and public keys to sign the container images.

---

## Steps
1. Go to **Applications** → **Devtron Apps**.
2. Click your application.
3. Go to **App Configuration** → **Workflow Editor**.
4. Click **New Workflow** and navigate to the **Build and Deploy from Source Code**.
5. Fill the required fields in the **Create build pipeline** window and navigate to the **Post-build stage**.

{% hint style="warning" %}
If you have already configured workflow, edit the build pipeline, and navigate to **Post-build stage**.
{% endhint %}

6. Under 'TASKS', click the **+ Add task** button.
7. Click the **Cosign** plugin.
8. Enter the following [user inputs](#user-inputs) with appropriate values.
---

## User Inputs

### Task Name
Enter the name of your task

e.g., `Signing of container images`

### Description
Add a brief explanation of the task and the reason for choosing the plugin. Include information for someone else to understand the purpose of the task.

e.g., `The Cosign plugin is integrated for  ensuring the authenticity of container images.`

### Input Variables

| Variable                 | Format       | Description | Sample Value |
| ------------------------ | ------------ | ----------- | ------------ |
|   PrivateKeyFilePath     |    STRING    | Path of private key file in Git repo           |    cosign/cosign.key                                      |
|   PostCommand            |    STRING    | Command to run after image is signed by Cosign |    cosign verify $DOCKER_IMAGE                            |
|   ExtraArguments         |    STRING    | Arguments for Cosign command                   |    --certificate-identity=name@example.com                                      | 
|   CosignPassword         |    STRING    | Password for Cosign private key                |   S3cur3P@ssw0rd123!                   |
|   VariableAsPrivateKey   |    STRING    | base64 encoded private-key                     |   @{{COSIGN_PRIVATE_KEY}}   |
|   PreCommand             |    STRING    | Command to get the required conditions to execute Cosign command | curl -sLJO https://raw.githubusercontent.com/devtron-labs/sampleRepo/branchName/private             |

### Trigger/Skip Condition
Here you can set conditions to execute or skip the task. You can select `Set trigger conditions` for the execution of a task or `Set skip conditions` to skip the task.

### Output Variables
Cosign will not be generating an output variable.

Click **Update Pipeline**.




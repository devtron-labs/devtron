# Build Infra

## Introduction

The [CI process](../creating-application/workflow/ci-pipeline.md) involves activities that require infra resources such as CPU, memory (RAM), and many more. The amount of resources required depends on the complexity of the application. In other words, large applications require more resources compared to small applications.

Therefore, applying a common infra configuration to all applications is not optimal. Since resources incur heavy costs, it's wise to efficiently allocate resources (*not more, not less*). 

With the 'Build Infra' feature, Devtron makes it possible for you to tweak the resources as per the needs of your applications. The build (ci-runner) pod will be scheduled on an available node (considering applied taints and tolerations) in the cluster on which 'Devtron' is installed.

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to configure build Infra.
{% endhint %}

---

## Steps to Configure Build Infra

From the left sidebar, go to **Global Configurations** → **Build Infra**. 

![Figure 1: Global Configurations - Build Infra](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/build-infra/gc-build-infra.jpg)

You will see the [Default Profile](#default-profile) and a list of [Custom Profiles](#custom-profile) (if they exist). Setting up profiles makes it easier for you to manage the build infra configurations, ensuring its reusability in the long term.

### Default Profile

This contains the default infra configuration applicable to all the applications, be it large or small.

![Figure 2: Default Profile](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/build-infra/default-profile.jpg)

You may click it to modify the following:
* **CPU** - Processor core allocated to the build process. See [CPU units](#cpu-units).
* **Memory** - RAM allocated to the build process. See [memory units](#memory-units).
* **Build Timeout** - Max. time limit allocated to the build process. See [timeout units](#timeout-units).

![Figure 3: Editing Default Profile](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/build-infra/default-infra-config.jpg)

Furthermore, CPU and Memory have 2 fields each:

* **Request** - Use this field to specify the minimum guaranteed amount of CPU/Memory resources your application needs for its CI build. In our example, we required 1500m or 1.5 cores CPU along with 6 GB of RAM. 

* **Limit** - Use this field to set the maximum amount of CPU/Memory resources the build process can use, even if there is a lot available in the cluster. 


### Creating Profile

Instead of default profile, you can create custom profiles having different infra configurations. Example: One profile for Python apps, a second profile for large apps, and a third profile for small apps, and many more.

1. Click **Create Profile**.

    ![Figure 4: Creating Custom Profile](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/build-infra/create-new-profile.jpg)

2. Give a name to the profile along with a brief description, and select the configurations to specify the values.

    ![Figure 5a: Empty Profile](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/build-infra/new-profile-fields.jpg)

    ![Figure 5b: Filled Profile](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/build-infra/filled-profile-fields.jpg)

3. Click **Save**. Your custom profile will appear under the list of custom profiles as shown below.

    ![Figure 6: Listed Profile](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/build-infra/new-profile-listed.jpg)

### Attaching Profile

Once you create a profile, attach it to the intended applications, or else the [default profile](#default-profile) will remain applied. 

1. Go to the **Applications** tab.

    ![Figure 7: Applications Tab](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/build-infra/applications-tab.jpg)

2. Choose an application and click the dropdown below it.

    ![Figure 8: Profile Dropdown](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/build-infra/profile-dropdown.jpg)

3. Choose the profile you wish to apply from the dropdown.

    ![Figure 9: Selecting a Profile](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/build-infra/profile-selection.jpg)

4. Click **Change** to apply the profile to your application.

    ![Figure 10: Confirming Profile Change](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/build-infra/confirm-profile-change.jpg)

**Tip**: If you missed creating a profile but selected your application(s), you can use the 'Create Profile' button. This will quickly open a new tab for creating a profile. Once done, you can return and click the refresh icon as shown below.

![Figure 11: Quick Profile Creation](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/build-infra/quick-profile-creation.jpg)

#### Performing Bulk Action

If you wish to apply a profile to multiple applications at once, you can do that too.

Simply use the checkboxes to select the applications. You can do this even if there are many applications spanning multiple pages. You will see a draggable floating widget as shown below.

![Figure 12: Floating Widget](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/build-infra/floating-widget.jpg)

Select the profile you wish to apply from the dropdown and confirm the changes.

![Figure 13: Selecting a Profile](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/build-infra/profile-selection-bulk.jpg)

Once you apply a profile, it will show the count of applications attached to it.

![Figure 14: Count of Applications](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/build-infra/profile-applicable-count.jpg)


### Editing or Deleting Profile

You can edit or delete a custom profile using the respective icons as shown below.

![Figure 15: Edit and Delete Icons](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/build-infra/edit-delete-icons-v2.jpg)

If you delete a profile attached to one or more applications, the [default profile](#default-profile) will apply from the next build.

![Figure 16: Confirm Profile Deletion](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/build-infra/delete-dialog.jpg)


### Need More Options?

If you need extra control on the build infra configuration apart from CPU, memory, and build timeout, feel free to open a [GitHub issue](https://github.com/devtron-labs/devtron/issues) for us to help you.

---

## Extras

### CPU Units

CPU resources are measured in millicore. 1000m or 1000 millicore is equal to 1 core. If a node has 4 cores, the node's CPU capacity would be represented as 4000m.

### Memory Units

Memory is measured in bytes. You can enter memory with suffixes (E, P, T, G, M, K, and Ei, Pi, Ti, Gi, Mi, Ki).

| Symbol | Prefix | Value (Bytes)                   |
| ------ | ------ | ------------------------------- |
| m      | -      | 0.001 byte                      |
| byte   | -      | 1 byte                          |
| k      | Kilo   | 1,000 bytes                     |
| Ki     | Kibi   | 1,024 bytes                     |
| M      | Mega   | 1,000,000 bytes                 |
| Mi     | Mebi   | 1,048,576 bytes                 |
| G      | Giga   | 1,000,000,000 bytes             |
| Gi     | Gibi   | 1,073,741,824 bytes             |
| T      | Tera   | 1,000,000,000,000 bytes         |
| Ti     | Tebi   | 1,099,511,627,776 bytes         |
| P      | Peta   | 1,000,000,000,000,000 bytes     |
| Pi     | Petabi | 1,125,899,906,842,624 bytes     |
| E      | Exa    | 1,000,000,000,000,000,000 bytes |
| Ei     | Exabi  | 1,152,921,504,606,846,976 bytes |

### Timeout Units

You can specify timeouts in the following units, beyond which the build process would be marked as failed:

* seconds
* minutes
* hours



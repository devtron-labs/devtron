# Custom charts

Devtron includes predefined helm charts that cover the majority of use cases.
For any use case not addressed by the default helm charts, you can upload your own helm chart and use it as a custom chart in Devtron.

* Who can upload a custom chart - Super admins
* Who can use the custom chart - All users

> A super admin can upload multiple versions of a custom helm chart.

![Custom charts](https://devtron-public-asset.s3.us-east-2.amazonaws.com/custom-charts/custom-charts-lists.png)

## Prerequisites

1. A valid helm chart, which contains `Chart.yaml` file with name and version fields.
2. Image descriptor template file - `.image_descriptor_template.json`.
3. Custom chart packaged in the `*.tgz` format.

### 1. How to create a helm chart

`Chart.yaml` is the metadata file that gets created when you create a [helm chart](https://helm.sh/docs/helm/helm_create/).

```bash
helm create my-custom-chart
```

| Field | Description |
| --- | --- |
| Name | Required. Name of the helm chart. |
| Version | Required. This is the chart version. Update this value for each new version of the chart. |
| Description | Optional. Description of the chart. |

![Chart.yaml file](https://devtron-public-asset.s3.us-east-2.amazonaws.com/custom-charts/chart-yaml-file.png)

### 2. Create the image descriptor template file - `.image_descriptor_template.json`

It's a GO template file that should produce a valid `JSON` file upon rendering. This file is passed as the last argument in
`helm install -f myvalues.yaml -f override.yaml ` command.

Place the `.image_descriptor_template.json` file in the root directory of your chart.

You can use the following variables in the helm template (all the placeholders are optional):

> The values from the CD deployment pipeline are injected at the placeholder specified in the `.image_descriptor_template.json` template file.

```bash
{
    "server": {
        "deployment": {
            "image_tag": "{{.Tag}}"
            "image": "{{.Name}}"
        }
    },
    "pipelineName": "{{.PipelineName}}",
    "releaseVersion": "{{.ReleaseVersion}}",
    "deploymentType": "{{.DeploymentType}}", ?
    "app": "{{.App}}",
    "env": "{{.Env}}",
    "appMetrics": {
        {.AppMetrics
        }
    }
}
```

| Field | Description |
| --- | --- |
| **image_tag** | The build image tag |
| **image** | Repository name |
| **pipelineName** | The CD pipeline name created in Devtron |
| **releaseVersion** | Devtron's internal release number |
| **deploymentType** | Deployment strategy used in the pipeline |
| **app** | Application's ID within the Devtron ecosystem |
| **env** | Environment used to deploy the chart |
| **appMetrics** | For the App metrics UI feature to be effective, include the `appMetrics` placeholder. |

> **For example**:
> 
> To create a template file to allow Devtron to only render the repository name and the tag from the CI/CD pipeline that you created, edit the `.image_descriptor_template.json` file as:
> ```bash
> {
>     "image": {
>	        "repository": "{{.Name}}",
>	        "tag": "{{.Tag}}"
>     }
> }
> ```

### 3. Package the custom chart in the `*.tgz` format

> Before you begin, ensure that your helm chart includes both `Chart.yaml` (with `name` and `version` fields) and `.image_descriptor_template.json` files.

The helm chart to be uploaded must be packaged as a versioned archive file in the format - `<helm-chart-name>-vx.x.x.tgz`.

```
helm package my-custom-chart
```

The above command will create a `my-custom-chart-0.1.0.tgz` file.

## Uploading a custom chart

> A custom chart can only be uploaded by a super admin.

* On the Devtron dashboard, select **Global Configurations > Custom charts**.
* Select **Import Chart**.
* Choose **Select tar.gz file...** and upload the packaged custom chart in the `*.tgz` format.

![Selecting custom chart](https://devtron-public-asset.s3.us-east-2.amazonaws.com/custom-charts/Chart+pre-requisites.png)

The chart is being uploaded and validated. You may also **Cancel upload** if required.

![Uploading custom chart](https://devtron-public-asset.s3.us-east-2.amazonaws.com/custom-charts/List+-+Empty-4.png)

### Validation

The uploaded archive will be validated against:

- Supported archive template should be in `*.tgz` format.
- `Chart.yaml` must include the name and the version number.
- `image_descriptor_template.json` file should be present and the field format must match the format listed in the image builder template section.

The following are the validation results:

| Validation status | Description | User action |
| :--- | :--- | :--- |
| **Success** | The files uploaded are validated. | Enter a description for the chart and select **Save** or **Cancel upload**. |
| **Unsupported template** | The archive file do not match the [required template](#prerequisites). | **Upload another chart** or **Cancel upload**. |
| **New version detected** | You are uploading a newer version of an existing chart | Enter a **Description** and select **Save** to continue uploading, or **Cancel upload**. |
| **Already exists** | There already exists a chart with the same version. | <ul><li>Edit the version and re-upload the same chart using **Upload another chart**.</li><li>Upload a new chart with a new name using  **Upload another chart**.</li><li>**Cancel upload**.</li></ul> |

![Chart validated](https://devtron-public-asset.s3.us-east-2.amazonaws.com/custom-charts/List+-+Empty-2.png)

![Unsupported template](https://devtron-public-asset.s3.us-east-2.amazonaws.com/custom-charts/List+-+Empty.png)

![New version detected](https://devtron-public-asset.s3.us-east-2.amazonaws.com/custom-charts/List+-+Empty-3.png)

![Already exists](https://devtron-public-asset.s3.us-east-2.amazonaws.com/custom-charts/List+-+Empty-1.png)

## View the custom charts

> All users can view the custom charts.

To view a list of available custom charts, go to  **Global Configurations > Custom charts** page.

* The charts can be searched with their name, version, or description.
* New [custom charts can be uploaded](#uploading-a-custom-chart) by selecting **Upload chart**.

![Custom charts](https://devtron-public-asset.s3.us-east-2.amazonaws.com/custom-charts/custom-charts-lists.png)

## Use the custom chart in an application

The custom charts can be used from the [Deployment Template](../creating-application/deployment-template.md) section.

> **Info**:
>
> The deployment strategy for a custom chart is fetched from the custom chart template and cannot be configured in the [CD pipeline](../creating-application/workflow/cd-pipeline.md#deployment-strategies).

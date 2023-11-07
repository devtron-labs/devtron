# Filter Condition

[![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/elements/EnterpriseTag.svg)](https://devtron.ai/pricing)

## Introduction

As you know, Devtron provides you the option to create [workflows](../creating-application/workflow/) for managing the CI-CD of your application. The events flowing from CI stage (i.e., build pipeline) are eligible to go to the CD stage (i.e., deployment pipeline). This can be either manual or automatic depending on how you configure the workflow.

However, in real-world scenarios, all events need not flow through the stages. For e.g., events in CI stage are images, and it's not necessary for all the images to be eligible for deployment. Therefore, instead of creating multiple pipelines to cater to complex requirements, Devtron provides you the option of creating filters to make your workflow efficient. 

Using filter conditions, you can control the flow of events. Here are a few general examples:
* Images containing the label "test" should not be eligible for deployment in production environment
* Only images having tag versions greater than v0.7.4 should be eligible for deployment
* Images hosted on Docker Hub should be eligible but not the rest

---

## Steps to Create a Filter

{% hint style="info" %}
**Pre-requisite**: You must have application(s) with CI-CD workflow(s) configured
{% endhint %}

1. From the left sidebar, go to **Global Configurations** → **Filter Condition**.

2. Click **Add Filter Condition** (visible only when no filters exist).

    ![Figure 1: Creating Our First Filter](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/filters/add-filter.jpg)

3. In the **Define Filter condition** section, you get the following fields:
    * **Filter For**: Choose the pipeline upon which the filter should apply. Currently, you can use filter conditions for CD pipelines only. Support for CI pipelines is underway.

    ![Figure 2: 'Define Filter Condition' section](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/filters/define-filter-page.jpg)

    * **Filter Name**: Give a name to the filter.
    * **Description**: (Optional) Add a description to the filter, preferably explaining what it does.
    * **Filter Condition**: You get 2 options:
        * **Pass Condition**: Only those events that satisfy the condition you define become eligible to flow through your pipeline.
        * **Fail Condition**: Events that fail the condition will not flow through your pipeline, the rest events will flow.
    * **Use CEL Expression**: You can use `Common Expression Language` (CEL) to define the conditions. Currently, you can create conditions around the following criteria:
        * containerImage
        * containerRepository
        * containerImageTag
        * imageLabels

        Click **View filter criteria** to check the supported criteria. You get a copy button and a description of each criterion upon hovering. Moreover, you can go to **CEL expression** to learn more about the rules and supported syntax. Check [Examples](#examples) to know more.

        ![Figure 3: List of Supported Values](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/filters/filter-criteria.jpg)

    {% hint style="info" %}
        Example of containerImage → gcr.io/k8s-minikube/kicbase:v0.0.39
        where containerRepository = kicbase
        where containerImageTag = v0.0.39 
        and imageLabels = the label you assign to an image in CD pipeline
    {% endhint %}


4. Click **Next**.

5. In the **Apply to** section, you get the following fields:
    * **Application**: Choose one or more applications to which your filter condition must apply.

    ![Figure 4: Selecting Application(s)](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/filters/app-selection.jpg)

    * **Environment**: Choose one or more environments to which your filter condition must apply.

    ![Figure 5: Selecting Environment(s) from Cluster(s)](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/filters/environment-selection.jpg)

    {% hint style="info" %}
    Since an application can have more than one environment, the filter conditions apply only to the environment chosen in the **Apply to** section. In other words, events will flow restricted for the chosen environment(s), and unrestricted for other environment(s).
    {% endhint %}

    {% hint style="warning" %}
    If you create a filter condition without choosing an application or environment, it will not apply to any of your pipelines
    {% endhint %}

6. Click **Save**. You have successfully created a filter.

    ![Figure 6: Success Toast](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/filters/filter-created.jpg)

    {% hint style="warning" %}
    If you create a pass filter and a fail filter having the same CEL expression, the fail filter will be taken into consideration
    {% endhint %}

---

## Examples

Here's a sample pipeline we will be using for our explanation of [pass condition](#pass-condition) and [fail condition](#fail-condition).

![Figure 7: Sample Pipeline](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/filters/sample-workflow.jpg)


### Pass Condition

Consider a scenario where you wish to make an image eligible for deployment only if its tag contains the word `prod`

The CEL Expression should be `containerImageTag.contains("prod")`

![Figure 8: CEL Expression for Pass Condition](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/filters/pass-condition.jpg)

Go to the **Build & Deploy** tab. The filter condition was created specifically for `test` environment, therefore the filter condition would be evaluated only at the relevant CD pipeline, i.e., `test`

![Figure 9: Build & Deploy tab](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/filters/pass-env.jpg)

Click **Select Image** for the `test` CD pipeline. The first tab **Eligible images** shows the list and count of images that have satisfied the pass condition since their tags had the word `prod`. Hence, they are marked eligible for deployment.

![Figure 10: List of Eligible Images](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/filters/eligible-list.jpg)

The second tab **Latest images** shows the latest builds (up to 15 images) irrespective of whether they have satisfied the filter condition(s) or not. The ones that have not satisfied the filter conditions get marked as `Excluded`. In other words, they are not eligible for deployment. 

![Figure 11: List of Latest Images](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/filters/all-images.jpg)

Clicking the döner menu at the top-left shows the filter condition(s) applied to the `test` CD pipeline.

![Figure 12a: Döner Menu Icon](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/filters/doner-menu-pass.jpg)

![Figure 12b: Conditions Applied](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/filters/conditions-applied.jpg)


### Fail Condition

Consider a scenario where you wish to exclude an image from deployment if its label contains either the word `test` or `testing`

The CEL Expression should be `containerImageTag.contains("test") || containerImageTag.contains("testing")`

![Figure 13: CEL Expression for Fail Condition](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/filters/test-filter.jpg)

Go to the **Build & Deploy** tab. The filter condition was created specifically for `devtron-demo` environment, therefore the filter condition would be evaluated only at the relevant CD pipeline, i.e., `devtron-demo`

![Figure 14: Build & Deploy tab](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/filters/fail-filter.jpg)

Click **Select Image** for the `devtron-demo` CD pipeline. The first tab **Eligible images** shows the list and count of images that have not met the fail condition. Hence, they are marked eligible for deployment.

![Figure 15: List of Eligible Images](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/filters/eligible-list-2.jpg)

The second tab **Latest images** shows the latest builds (up to 15 images) irrespective of whether they have satisfied the filter condition(s) or not. The ones that have satisfied the filter conditions get marked as `Excluded`. In other words, they are not eligible for deployment. 

![Figure 16: List of Latest Images](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/filters/all-list-2.jpg)

Clicking the döner menu at the top-left shows the filter condition(s) applied to the `devtron-demo` CD pipeline.

![Figure 17a: Döner Menu Icon](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/filters/doner-menu-fail.jpg)

![Figure 17b: Conditions Applied](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/filters/conditions-applied-2.jpg)


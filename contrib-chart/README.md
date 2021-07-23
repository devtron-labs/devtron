# Bring Your Own Chart

We at Devtron strongly believe that a product is always incomplete without the contribution of it's community and the users. What else can be better than devtron users becoming a part of devtron team, the product they love the most and here we are giving you an opportunity to contribute to an Open Source product by sharing your own helm charts with us.

## Introduction

Very soon we are coming up with a new feature in the charts section of devtron which will be called `Community Charts` where we'll upload the charts submitted by you for all the users who love to deploy and build pipelines on devtron. **All of your charts will have your name associated with it for all users to know who made their life easier by uploading that chart**. So what are you waiting for? Now is the time to do that

## Use dynamic devrtron variables in your template

Devtron injects [some variables](https://github.com/devtron-labs/devtron/blob/main/contrib-chart/reference-chart_3-11-0/.image_descriptor_template.json) in values.yaml in order to fully integrate with automatic CD capability. In Template you can use these variables wherever required.

| Syntax | Output |
| ----------- | ----------- |
| server.deployment.image | Image Name |
| server.deployment.image_tag | Image Tag |
| pipelineName | Pipeline Name |
| releaseVersion | Release Version |
| deploymentType | Deployment Type |
| app | App Name |
| env | Env Name |
| appMetrics | App Metrics |

## How to Share Chart

1. Fork the [devtron-labs/devtron](https://github.com/devtron-labs/devtron) repository
2. Upload your chart directory inside `chart-contrib` directory
4. Add a README file inside your chart describing it's use case and how to use
5. Verify everything works fine by using the command `helm template --debug <template-dir-name>`
6. Generate a Pull Request on the branch `contribute-chart` of [devtron-labs/devtron](https://github.com/devtron-labs/devtron)
7. Hold on till we verify everything and upload on the community charts section

## Communications

The project uses discord for communication:

To join the conversation, simply join the **[discord](https://discord.gg/jsRG5qx2gp)**  and use the __#contrib__ channel.
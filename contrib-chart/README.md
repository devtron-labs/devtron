# Bring Your Own Chart

We at Devtron strongly believe that a product is always incomplete without the contribution of it's community and the users. What else can be better than devtron users becoming a part of devtron team, the product they love and through this upcoming feature you can help other community members by sharing your own helm charts with devtron.

## Introduction

Very soon we are coming up with a new feature in the charts section of devtron which will be called `Community Charts` where we'll upload the charts submitted by you for all the users who love to deploy and build pipelines on devtron. **All of your charts will have your name associated with it for all users to know who made their life easier by uploading that chart**. So what are you waiting for? Now is the time to do that

## Use dynamic devtron variables in your template

Devtron injects [some variables](https://github.com/devtron-labs/devtron/blob/main/contrib-chart/reference-chart_3-11-0/.image_descriptor_template.json) in values.yaml in order to fully integrate with automatic CD capability. In Template you can use these variables wherever required.

| Syntax | Output |
| ----------- | ----------- |
| server.deployment.image | Image repo populated by Devtron CI |
| server.deployment.image_tag | Image tag populated by Devtron CI |
| pipelineName | Devtron pipeline name |
| releaseVersion | Release version of the chart |
| deploymentType | Deployment strategy used for deployments |
| app | Devtron App ID |
| env | Devtron Environment ID |
| appMetrics | Configuration for the metrics exposed by the app |

## How to use dynamic devtron variables

This is an example of deployment.yaml which will create a deployment named `httpd-web` and it will identify the app to deploy using the appId and environment in which to deploy using envId both of which values will be taken from devtron dynamic variables and same is the case for image repo and tag to be used for deployment. This deployment will create a pod named `myweb-pod` where our application will be up and running.

    apiVersion: apps/v1
    kind: Deployment
    metadata:
        name: httpd-web
        labels:
            app: myweb
    spec:
        replicas: 1
        selector:
            matchLabels:
                app: myweb
        template:
            metadata:
                labels:
                    app: myweb
                    appId: {{ $.Values.app | quote }}
                    envId: {{ $.Values.env | quote }}
            spec:
                containers:
                - name: myweb-pod
                image: "{{ $.Values.server.deployment.image }}:{{ $.Values.server.deployment.image_tag }}"
                ports:
                - containerPort: 80


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
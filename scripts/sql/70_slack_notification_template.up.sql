---- update notification template for CI trigger slack
UPDATE notification_templates
set template_payload = '{
    "text": ":arrow_forward: Build pipeline Triggered |  {{#ciMaterials}} Branch > {{branch}} {{/ciMaterials}} | Application > {{appName}}",
    "blocks": [{
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "\n"
            }
        },
        {
            "type": "divider"
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": ":arrow_forward: *Build Pipeline triggered*\n<!date^{{eventTime}}^{date_long} {time} | \"-\"> \n Triggered by {{triggeredBy}}"
            },
            "accessory": {
                "type": "image",
                "image_url": "https://github.com/devtron-labs/notifier/assets/image/img_build_notification.png",
                "alt_text": "calendar thumbnail"
            }
        },
        {
            "type": "section",
            "fields": [{
                    "type": "mrkdwn",
                    "text": "*Application*\n{{appName}}"
                },
                {
                    "type": "mrkdwn",
                    "text": "*Pipeline*\n{{pipelineName}}"
                }
            ]
        },
        {{#ciMaterials}}
        {{^webhookType}}
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Branch*\n`{{appName}}/{{branch}}`"
            },
            {
            "type": "mrkdwn",
            "text": "*Commit*\n<{{& commitLink}}|{{commit}}>"
            }
        ]
        },
        {{/webhookType}}
        {{#webhookType}}
        {{#webhookData.mergedType}}
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Title*\n{{webhookData.data.title}}"
            },
            {
            "type": "mrkdwn",
            "text": "*Git URL*\n<{{& webhookData.data.giturl}}|View>"
            }
        ]
        },
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Source Branch*\n{{webhookData.data.sourcebranchname}}"
            },
            {
            "type": "mrkdwn",
            "text": "*Source Commit*\n<{{& webhookData.data.sourcecheckoutlink}}|{{webhookData.data.sourcecheckout}}>"
            }
        ]
        },
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Target Branch*\n{{webhookData.data.targetbranchname}}"
            },
            {
            "type": "mrkdwn",
            "text": "*Target Commit*\n<{{& webhookData.data.targetcheckoutlink}}|{{webhookData.data.targetcheckout}}>"
            }
        ]
        },
        {{/webhookData.mergedType}}
        {{^webhookData.mergedType}}
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Target Checkout*\n{{webhookData.data.targetcheckout}}"
            }
        ]
        },
        {{/webhookData.mergedType}}
        {{/webhookType}}
        {{/ciMaterials}}
        {
            "type": "actions",
            "elements": [{
                "type": "button",
                "text": {
                    "type": "plain_text",
                    "text": "View Details"
                }
                {{#buildHistoryLink}}
                    ,
                    "url": "{{& buildHistoryLink}}"
                {{/buildHistoryLink}}
            }]
        }
    ]
}'
where channel_type = 'slack'
and node_type = 'CI'
and event_type_id = 1;


---- update notification template for CI success slack
UPDATE notification_templates
set template_payload = '{
  "text": ":tada: Build pipeline Successful |  {{#ciMaterials}} Branch > {{branch}} {{/ciMaterials}} | Application > {{appName}}",
  "blocks": [
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "\n"
      }
    },
    {
      "type": "divider"
    },
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": ":tada: *Build Pipeline successful*\n<!date^{{eventTime}}^{date_long} {time} | \"-\"> \n Triggered by {{triggeredBy}}"
      },
      "accessory": {
        "type": "image",
        "image_url": "https://github.com/devtron-labs/notifier/assets/image/img_build_notification.png",
        "alt_text": "calendar thumbnail"
      }
    },
    {
      "type": "section",
      "fields": [
        {
          "type": "mrkdwn",
          "text": "*Application*\n{{appName}}"
        },
        {
          "type": "mrkdwn",
          "text": "*Pipeline*\n{{pipelineName}}"
        }
      ]
    },
    {{#ciMaterials}}
    {{^webhookType}}
    {
    "type": "section",
    "fields": [
        {
          "type": "mrkdwn",
           "text": "*Branch*\n`{{appName}}/{{branch}}`"
        },
        {
          "type": "mrkdwn",
          "text": "*Commit*\n<{{& commitLink}}|{{commit}}>"
        }
    ]
    },
    {{/webhookType}}
    {{#webhookType}}
    {{#webhookData.mergedType}}
    {
    "type": "section",
    "fields": [
        {
        "type": "mrkdwn",
        "text": "*Title*\n{{webhookData.data.title}}"
        },
        {
        "type": "mrkdwn",
        "text": "*Git URL*\n<{{& webhookData.data.giturl}}|View>"
        }
    ]
    },
    {
    "type": "section",
    "fields": [
        {
        "type": "mrkdwn",
        "text": "*Source Branch*\n{{webhookData.data.sourcebranchname}}"
        },
        {
        "type": "mrkdwn",
        "text": "*Source Commit*\n<{{& webhookData.data.sourcecheckoutlink}}|{{webhookData.data.sourcecheckout}}>"
        }
    ]
    },
    {
    "type": "section",
    "fields": [
        {
        "type": "mrkdwn",
        "text": "*Target Branch*\n{{webhookData.data.targetbranchname}}"
        },
        {
        "type": "mrkdwn",
        "text": "*Target Commit*\n<{{& webhookData.data.targetcheckoutlink}}|{{webhookData.data.targetcheckout}}>"
        }
    ]
    },
    {{/webhookData.mergedType}}
    {{^webhookData.mergedType}}
    {
    "type": "section",
    "fields": [
        {
        "type": "mrkdwn",
        "text": "*Target Checkout*\n{{webhookData.data.targetcheckout}}"
        }
    ]
    },
    {{/webhookData.mergedType}}
    {{/webhookType}}
    {{/ciMaterials}}
    {
      "type": "actions",
      "elements": [
        {
          "type": "button",
          "text": {
            "type": "plain_text",
            "text": "View Details"
          }
          {{#buildHistoryLink}}
            ,
            "url": "{{& buildHistoryLink}}"
          {{/buildHistoryLink}}
        }
      ]
    }
  ]
}'
where channel_type = 'slack'
and node_type = 'CI'
and event_type_id = 2;



---- update notification template for CI fail slack
UPDATE notification_templates
set template_payload = '{
    "text": ":x: Build pipeline Failed |  {{#ciMaterials}} Branch > {{branch}} {{/ciMaterials}} | Application > {{appName}}",
    "blocks": [{
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "\n"
            }
        },
        {
            "type": "divider"
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": ":x: *Build Pipeline failed*\n<!date^{{eventTime}}^{date_long} {time} | \"-\"> \n Triggered by {{triggeredBy}}"
            },
            "accessory": {
                "type": "image",
                "image_url": "https://github.com/devtron-labs/notifier/assets/image/img_build_notification.png",
                "alt_text": "calendar thumbnail"
            }
        },
        {
            "type": "section",
            "fields": [{
                    "type": "mrkdwn",
                    "text": "*Application*\n{{appName}}"
                },
                {
                    "type": "mrkdwn",
                    "text": "*Pipeline*\n{{pipelineName}}"
                }
            ]
        },
        {{#ciMaterials}}
        {{^webhookType}}
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Branch*\n`{{appName}}/{{branch}}`"
            },
            {
            "type": "mrkdwn",
            "text": "*Commit*\n<{{& commitLink}}|{{commit}}>"
            }
        ]
        },
        {{/webhookType}}
        {{#webhookType}}
        {{#webhookData.mergedType}}
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Title*\n{{webhookData.data.title}}"
            },
            {
            "type": "mrkdwn",
            "text": "*Git URL*\n<{{& webhookData.data.giturl}}|View>"
            }
        ]
        },
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Source Branch*\n{{webhookData.data.sourcebranchname}}"
            },
            {
            "type": "mrkdwn",
            "text": "*Source Commit*\n<{{& webhookData.data.sourcecheckoutlink}}|{{webhookData.data.sourcecheckout}}>"
            }
        ]
        },
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Target Branch*\n{{webhookData.data.targetbranchname}}"
            },
            {
            "type": "mrkdwn",
            "text": "*Target Commit*\n<{{& webhookData.data.targetcheckoutlink}}|{{webhookData.data.targetcheckout}}>"
            }
        ]
        },
        {{/webhookData.mergedType}}
        {{^webhookData.mergedType}}
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Target Checkout*\n{{webhookData.data.targetcheckout}}"
            }
        ]
        },
        {{/webhookData.mergedType}}
        {{/webhookType}}
        {{/ciMaterials}}
        {
            "type": "actions",
            "elements": [{
                "type": "button",
                "text": {
                    "type": "plain_text",
                    "text": "View Details"
                }
                  {{#buildHistoryLink}}
                    ,
                    "url": "{{& buildHistoryLink}}"
                   {{/buildHistoryLink}}
            }]
        }
    ]
}'
where channel_type = 'slack'
and node_type = 'CI'
and event_type_id = 3;


---- update notification template for CD trigger slack
UPDATE notification_templates
set template_payload = '{
    "text": ":arrow_forward: Deployment pipeline Triggered |  {{#ciMaterials}} Branch > {{branch}} {{/ciMaterials}} | Application > {{appName}}",
    "blocks": [{
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "\n"
            }
        },
        {
            "type": "divider"
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": ":arrow_forward: *Deployment Pipeline triggered on {{envName}}*\n<!date^{{eventTime}}^{date_long} {time} | \"-\"> \n by {{triggeredBy}}"
            },
            "accessory": {
                "type": "image",
                "image_url":"https://github.com/devtron-labs/notifier/assets/image/img_deployment_notification.png",
                "alt_text": "Deploy Pipeline Triggered"
            }
        },
        {
            "type": "divider"
        },
        {
            "type": "section",
            "fields": [{
                    "type": "mrkdwn",
                    "text": "*Application*\n{{appName}}\n*Pipeline*\n{{pipelineName}}"
                },
                {
                    "type": "mrkdwn",
                    "text": "*Environment*\n{{envName}}\n*Stage*\n{{stage}}"
                }
            ]
        },
        {{#ciMaterials}}
        {{^webhookType}}
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
             "text": "*Branch*\n`{{appName}}/{{branch}}`"
            },
            {
            "type": "mrkdwn",
            "text": "*Commit*\n<{{& commitLink}}|{{commit}}>"
            }
        ]
        },
        {{/webhookType}}
        {{#webhookType}}
        {{#webhookData.mergedType}}
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Title*\n{{webhookData.data.title}}"
            },
            {
            "type": "mrkdwn",
            "text": "*Git URL*\n<{{& webhookData.data.giturl}}|View>"
            }
        ]
        },
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Source Branch*\n{{webhookData.data.sourcebranchname}}"
            },
            {
            "type": "mrkdwn",
            "text": "*Source Commit*\n<{{& webhookData.data.sourcecheckoutlink}}|{{webhookData.data.sourcecheckout}}>"
            }
        ]
        },
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Target Branch*\n{{webhookData.data.targetbranchname}}"
            },
            {
            "type": "mrkdwn",
            "text": "*Target Commit*\n<{{& webhookData.data.targetcheckoutlink}}|{{webhookData.data.targetcheckout}}>"
            }
        ]
        },
        {{/webhookData.mergedType}}
        {{^webhookData.mergedType}}
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Target Checkout*\n{{webhookData.data.targetcheckout}}"
            }
        ]
        },
        {{/webhookData.mergedType}}
        {{/webhookType}}
        {{/ciMaterials}}
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "*Docker Image*\n`{{dockerImg}}`"
            }
        },
        {
            "type": "actions",
            "elements": [{
                    "type": "button",
                    "text": {
                        "type": "plain_text",
                        "text": "View Pipeline",
                        "emoji": true
                    }
                    {{#deploymentHistoryLink}}
                    ,
                    "url": "{{& deploymentHistoryLink}}"
                      {{/deploymentHistoryLink}}
                },
                {
                    "type": "button",
                    "text": {
                        "type": "plain_text",
                        "text": "App details",
                        "emoji": true
                    }
                    {{#appDetailsLink}}
                    ,
                    "url": "{{& appDetailsLink}}"
                      {{/appDetailsLink}}
                }
            ]
        }
    ]
}'
where channel_type = 'slack'
and node_type = 'CD'
and event_type_id = 1;



---- update notification template for CD success slack
UPDATE notification_templates
set template_payload = '{
    "text": ":tada: Deployment pipeline Successful |  {{#ciMaterials}} Branch > {{branch}} {{/ciMaterials}} | Application > {{appName}}",
    "blocks": [{
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "\n"
            }
        },
        {
            "type": "divider"
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": ":tada: *Deployment Pipeline successful on {{envName}}*\n<!date^{{eventTime}}^{date_long} {time} | \"-\"> \n by {{triggeredBy}}"
            },
            "accessory": {
                "type": "image",
                "image_url":"https://github.com/devtron-labs/notifier/assets/image/img_deployment_notification.png",
                "alt_text": "calendar thumbnail"
            }
        },
        {
            "type": "divider"
        },
        {
            "type": "section",
            "fields": [{
                    "type": "mrkdwn",
                    "text": "*Application*\n{{appName}}\n*Pipeline*\n{{pipelineName}}"
                },
                {
                    "type": "mrkdwn",
                    "text": "*Environment*\n{{envName}}\n*Stage*\n{{stage}}"
                }
            ]
        },
        {{#ciMaterials}}
        {{^webhookType}}
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
             "text": "*Branch*\n`{{appName}}/{{branch}}`"
            },
            {
            "type": "mrkdwn",
            "text": "*Commit*\n<{{& commitLink}}|{{commit}}>"
            }
        ]
        },
        {{/webhookType}}
        {{#webhookType}}
        {{#webhookData.mergedType}}
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Title*\n{{webhookData.data.title}}"
            },
            {
            "type": "mrkdwn",
            "text": "*Git URL*\n<{{& webhookData.data.giturl}}|View>"
            }
        ]
        },
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Source Branch*\n{{webhookData.data.sourcebranchname}}"
            },
            {
            "type": "mrkdwn",
            "text": "*Source Commit*\n<{{& webhookData.data.sourcecheckoutlink}}|{{webhookData.data.sourcecheckout}}>"
            }
        ]
        },
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Target Branch*\n{{webhookData.data.targetbranchname}}"
            },
            {
            "type": "mrkdwn",
            "text": "*Target Commit*\n<{{& webhookData.data.targetcheckoutlink}}|{{webhookData.data.targetcheckout}}>"
            }
        ]
        },
        {{/webhookData.mergedType}}
        {{^webhookData.mergedType}}
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Target Checkout*\n{{webhookData.data.targetcheckout}}"
            }
        ]
        },
        {{/webhookData.mergedType}}
        {{/webhookType}}
        {{/ciMaterials}}
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "*Docker Image*\n`{{dockerImg}}`"
            }
        },
        {
            "type": "actions",
            "elements": [{
                    "type": "button",
                    "text": {
                        "type": "plain_text",
                        "text": "View Pipeline",
                        "emoji": true
                    }
                    {{#deploymentHistoryLink}}
                    ,
                    "url": "{{& deploymentHistoryLink}}"
                      {{/deploymentHistoryLink}}
                },
                {
                    "type": "button",
                    "text": {
                        "type": "plain_text",
                        "text": "App details",
                        "emoji": true
                    }
                    {{#appDetailsLink}}
                    ,
                    "url": "{{& appDetailsLink}}"
                      {{/appDetailsLink}}
                }
            ]
        }
    ]
}'
where channel_type = 'slack'
and node_type = 'CD'
and event_type_id = 2;


---- update notification template for CD fail slack
UPDATE notification_templates
set template_payload = '{
    "text": ":x: Deployment pipeline Failed |  {{#ciMaterials}} Branch > {{branch}} {{/ciMaterials}} | Application > {{appName}}",
    "blocks": [{
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "\n"
            }
        },
        {
            "type": "divider"
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": ":x: *Deployment Pipeline failed on {{envName}}*\n<!date^{{eventTime}}^{date_long} {time} | \"-\"> \n by {{triggeredBy}}"
            },
            "accessory": {
                "type": "image",
                "image_url":"https://github.com/devtron-labs/notifier/assets/image/img_deployment_notification.png",
                "alt_text": "calendar thumbnail"
            }
        },
        {
            "type": "divider"
        },
        {
            "type": "section",
            "fields": [{
                    "type": "mrkdwn",
                    "text": "*Application*\n{{appName}}\n*Pipeline*\n{{pipelineName}}"
                },
                {
                    "type": "mrkdwn",
                    "text": "*Environment*\n{{envName}}\n*Stage*\n{{stage}}"
                }
            ]
        },
        {{#ciMaterials}}
        {{^webhookType}}
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Branch*\n`{{appName}}/{{branch}}`"
            },
            {
            "type": "mrkdwn",
            "text": "*Commit*\n<{{& commitLink}}|{{commit}}>"
            }
        ]
        },
        {{/webhookType}}
        {{#webhookType}}
        {{#webhookData.mergedType}}
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Title*\n{{webhookData.data.title}}"
            },
            {
            "type": "mrkdwn",
            "text": "*Git URL*\n<{{& webhookData.data.giturl}}|View>"
            }
        ]
        },
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Source Branch*\n{{webhookData.data.sourcebranchname}}"
            },
            {
            "type": "mrkdwn",
            "text": "*Source Commit*\n<{{& webhookData.data.sourcecheckoutlink}}|{{webhookData.data.sourcecheckout}}>"
            }
        ]
        },
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Target Branch*\n{{webhookData.data.targetbranchname}}"
            },
            {
            "type": "mrkdwn",
            "text": "*Target Commit*\n<{{& webhookData.data.targetcheckoutlink}}|{{webhookData.data.targetcheckout}}>"
            }
        ]
        },
        {{/webhookData.mergedType}}
        {{^webhookData.mergedType}}
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Target Checkout*\n{{webhookData.data.targetcheckout}}"
            }
        ]
        },
        {{/webhookData.mergedType}}
        {{/webhookType}}
        {{/ciMaterials}}
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "*Docker Image*\n`{{dockerImg}}`"
            }
        },
        {
            "type": "actions",
            "elements": [{
                    "type": "button",
                    "text": {
                        "type": "plain_text",
                        "text": "View Pipeline",
                        "emoji": true
                    }
                    {{#deploymentHistoryLink}}
                    ,
                    "url": "{{& deploymentHistoryLink}}"
                      {{/deploymentHistoryLink}}
                },
                {
                    "type": "button",
                    "text": {
                        "type": "plain_text",
                        "text": "App details",
                        "emoji": true
                    }
                    {{#appDetailsLink}}
                    ,
                    "url": "{{& appDetailsLink}}"
                      {{/appDetailsLink}}
                }
            ]
        }
    ]
}'
where channel_type = 'slack'
and node_type = 'CD'
and event_type_id = 3;
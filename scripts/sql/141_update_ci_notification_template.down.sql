---- revert notification template for CI fail ses/smtp
UPDATE notification_templates
set template_payload = '{"from": "{{fromEmail}}",
 "to": "{{toEmail}}",
 "subject": "CI failed for app: {{appName}}",
 "html": "<h2 style=\"color:#f33e3e;\">Build Pipeline Failed</h2><span>{{eventTime}}</span><br><span>Triggered by <strong>{{triggeredBy}}</strong></span><br><br><a href=\"{{& buildHistoryLink }}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:none;min-width:64px;text-transform:capitalize;text-align:center;background:#0066cc;color:#fff;border:1px solid transparent;cursor:pointer;\">View Pipeline</a><br><br><hr><br><span>Application: <strong>{{appName}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Pipeline: <strong>{{pipelineName}}</strong></span><br><br><hr><h3>Source Code</h3>{{#ciMaterials}}{{^webhookType}}<span>Branch: <strong>{{appName}}/{{branch}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Commit: <a href=\"{{& commitLink }}\"><strong>{{commit}}</strong></a></span><br><br>{{/webhookType}}{{#webhookType}}{{#webhookData.mergedType}}<span>Title: <strong>{{webhookData.data.title}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Git URL: <a href=\"{{& webhookData.data.giturl}}\"><strong>View</strong></a></span><br><br><span>Source Branch: <strong>{{webhookData.data.sourcebranchname}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Source Commit: <a href=\"{{& webhookData.data.sourcecheckoutlink}}\"><strong>{{webhookData.data.sourcecheckout}}</strong></a></span><br><br><span>Target Branch: <strong>{{webhookData.data.targetbranchname}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Target Commit: <a href=\"{{& webhookData.data.targetcheckoutlink}}\"><strong>{{webhookData.data.targetcheckout}}</strong></a></span><br><br>{{/webhookData.mergedType}}{{^webhookData.mergedType}}<span>Target Checkout: <strong>{{webhookData.data.targetcheckout}}</strong></span><br>{{/webhookData.mergedType}}{{/webhookType}}{{/ciMaterials}}<br>"}'
where channel_type = 'ses'
  and node_type = 'CI'
  and event_type_id = 3;


---- revert notification template for CI fail slack
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
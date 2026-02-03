/*
 * Copyright (c) 2024. Devtron Inc.
 */

-- Slack template for deployment approval (event_type_id = 4)
INSERT INTO "public"."notification_templates" (channel_type, node_type, event_type_id, template_name, template_payload)
SELECT 'slack', 'CD', 4, 'CD approval slack template', '{
  "text": "🛎️ Image approval requested for {{appName}}",
  "blocks": [
                {
                  "type": "header",
                  "text": {
                    "type": "plain_text",
                    "text": "🛎️ Image Approval Request"
                  }
                },
                {
                  "type": "section",
                  "fields": [
                    {
                      "type": "mrkdwn",
                      "text": "*Application:*\n{{appName}}"
                    },
                    {
                      "type": "mrkdwn",
                      "text": "*Environment:*\n{{envName}}"
                    },
                    {
                      "type": "mrkdwn",
                      "text": "*Pipeline:*\n{{pipelineName}}"
                    },
                    {
                      "type": "mrkdwn",
                      "text": "*Requested by:*\n{{triggeredBy}}"
                    }
                  ]
                },
                {
                  "type": "section",
                  "text": {
                    "type": "mrkdwn",
                    "text": "*Image Tag:* `{{imageTag}}`\n*Time:* <!date^{{eventTime}}^{date_long} {time} | \"-\">{{#comment}}\n*Comment:* {{comment}}{{/comment}}{{#tags}}\n*Tags:* {{tags}}{{/tags}}"
                  }
                },
                {
                  "type": "actions",
                  "elements": [
                    {
                      "type": "button",
                      "text": {
                        "type": "plain_text",
                        "text": "View Request"
                      },
                      "url": "{{{imageApprovalLink}}}",
                      "style": "primary"
                    }
                  ]
                }
              ]
}'
WHERE NOT EXISTS (
    SELECT 1 FROM "public"."notification_templates"
    WHERE channel_type = 'slack'
      AND node_type = 'CD'
      AND event_type_id = 4
      AND template_name = 'CD approval slack template'
);

-- Slack template for config approval (event_type_id = 5)
INSERT INTO "public"."notification_templates" (channel_type, node_type, event_type_id, template_name, template_payload)
SELECT 'slack', 'CD', 5, 'Config approval slack template', '{
  "text": "🛎️ Config change approval requested for {{appName}}",
  "blocks": [
                {
                  "type": "header",
                  "text": {
                    "type": "plain_text",
                    "text": "🛎️ Config Change Approval Request"
                  }
                },
                {
                  "type": "section",
                  "fields": [
                    {
                      "type": "mrkdwn",
                      "text": "*Application:*\n{{appName}}"
                    },
                    {
                      "type": "mrkdwn",
                      "text": "*Environment:*\n{{envName}}"
                    },
                    {
                      "type": "mrkdwn",
                      "text": "*Requested by:*\n{{triggeredBy}}"
                    },
                    {
                      "type": "mrkdwn",
                      "text": "*Time:*\n<!date^{{eventTime}}^{date_long} {time} | \"-\">"
                    }
                  ]
                },
                {
                  "type": "section",
                  "fields": [
                    {
                      "type": "mrkdwn",
                      "text": "*File Type:*\n{{protectConfigFileType}}"
                    },
                    {
                      "type": "mrkdwn",
                      "text": "*File Name:*\n{{protectConfigFileName}}"
                    }
                  ]
                },
                {
                  "type": "section",
                  "text": {
                    "type": "mrkdwn",
                    "text": "{{#protectConfigComment}}*Comment:*\n{{protectConfigComment}}{{/protectConfigComment}}{{^protectConfigComment}}_No comment provided_{{/protectConfigComment}}"
                  }
                },
                {
                  "type": "actions",
                  "elements": [
                    {
                      "type": "button",
                      "text": {
                        "type": "plain_text",
                        "text": "View Request"
                      },
                      "url": "{{{protectConfigLink}}}",
                      "style": "primary"
                    }
                  ]
                }
              ]
}'
WHERE NOT EXISTS (
    SELECT 1 FROM "public"."notification_templates"
    WHERE channel_type = 'slack'
      AND node_type = 'CD'
      AND event_type_id = 5
      AND template_name = 'Config approval slack template'
);


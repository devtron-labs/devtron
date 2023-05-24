---- revert notification template update for CI trigger ses/smtp
UPDATE notification_templates
set template_payload = '{"from": "{{fromEmail}}",
 "to": "{{toEmail}}",
 "subject": "CI triggered for app: {{appName}}",
 "html": "<b>CI triggered on pipeline: {{pipelineName}}</b>"}'
where channel_type = 'ses'
and node_type = 'CI'
and event_type_id = 1;


---- revert notification template update for CI success ses/smtp
UPDATE notification_templates
set template_payload = '{"from": "{{fromEmail}}",
 "to": "{{toEmail}}",
 "subject": "CI success for app: {{appName}}",
 "html": "<b>CI success on pipeline: {{pipelineName}}</b><br><b>docker image: {{{dockerImageUrl}}}</b><br><b>Source: {{source}}</b><br>"}'
where channel_type = 'ses'
and node_type = 'CI'
and event_type_id = 2;



---- revert notification template update for CI fail ses/smtp
UPDATE notification_templates
set template_payload = '{"from": "{{fromEmail}}",
 "to": "{{toEmail}}",
 "subject": "CI failed for app: {{appName}}",
 "html": "<b>CI failed on pipeline: {{pipelineName}}</b><br><b>build name: {{buildName}}</b><br><b>Pod status: {{podStatus}}</b><br><b>message: {{message}}</b>"}'
where channel_type = 'ses'
and node_type = 'CI'
and event_type_id = 3;


---- revert notification template update for CD trigger ses/smtp
UPDATE notification_templates
set template_payload = '{"from": "{{fromEmail}}",
 "to": "{{toEmail}}",
 "subject": "CD triggered for app: {{appName}} on environment: {{envName}}",
 "html": "<b>CD triggered for app: {{appName}} on environment: {{envName}}</b> <br> <b>Docker image: {{{dockerImageUrl}}}</b> <br> <b>Source snapshot: {{source}}</b> <br> <b>pipeline: {{pipelineName}}</b>"}'
where channel_type = 'ses'
and node_type = 'CD'
and event_type_id = 1;



---- revert notification template update for CD success ses/smtp
UPDATE notification_templates
set template_payload = '{"from": "{{fromEmail}}",
 "to": "{{toEmail}}",
 "subject": "CD success for app: {{appName}} on environment: {{envName}}",
 "html": "<b>CD success for app: {{appName}} on environment: {{envName}}</b>"}'
where channel_type = 'ses'
and node_type = 'CD'
and event_type_id = 2;


---- revert notification template update for CD fail ses/smtp
UPDATE notification_templates
set template_payload = '{"from": "{{fromEmail}}",
 "to": "{{toEmail}}",
 "subject": "CD failed for app: {{appName}} on environment: {{envName}}",
 "html": "<b>CD failed for app: {{appName}} on environment: {{envName}}</b>"}'
where channel_type = 'ses'
and node_type = 'CD'
and event_type_id = 3;
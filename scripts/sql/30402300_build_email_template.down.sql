/*
 * Copyright (c) 2024. Devtron Inc.
 */

---- update notification template for CI trigger ses/smtp
UPDATE notification_templates
set template_payload = '{"from": "{{fromEmail}}",
 "to": "{{toEmail}}",
 "subject": "CI triggered for app: {{appName}}",
 "html": "<h2 style=\"color:#767d84;\">Build Pipeline Triggered</h2><span>{{eventTime}}</span><br><span>Triggered by <strong>{{triggeredBy}}</strong></span><br><br>{{#buildHistoryLink}}<a href=\"{{& buildHistoryLink }}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:none;min-width:64px;text-transform:capitalize;text-align:center;background:#0066cc;color:#fff;border:1px solid transparent;cursor:pointer;\">View Pipeline</a><br><br>{{/buildHistoryLink}}<hr><br><span>Application: <strong>{{appName}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Pipeline: <strong>{{pipelineName}}</strong></span><br><br><hr><h3>Source Code</h3>{{#ciMaterials}}{{^webhookType}}<span>Branch: <strong>{{appName}}/{{branch}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Commit: <a href=\"{{& commitLink }}\"><strong>{{commit}}</strong></a></span><br><br>{{/webhookType}}{{#webhookType}}{{#webhookData.mergedType}}<span>Title: <strong>{{webhookData.data.title}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Git URL: <a href=\"{{& webhookData.data.giturl}}\"><strong>View</strong></a></span><br><br><span>Source Branch: <strong>{{webhookData.data.sourcebranchname}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Source Commit: <a href=\"{{& webhookData.data.sourcecheckoutlink}}\"><strong>{{webhookData.data.sourcecheckout}}</strong></a></span><br><br><span>Target Branch: <strong>{{webhookData.data.targetbranchname}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Target Commit: <a href=\"{{& webhookData.data.targetcheckoutlink}}\"><strong>{{webhookData.data.targetcheckout}}</strong></a></span><br><br>{{/webhookData.mergedType}}{{^webhookData.mergedType}}<span>Target Checkout: <strong>{{webhookData.data.targetcheckout}}</strong></span><br>{{/webhookData.mergedType}}{{/webhookType}}{{/ciMaterials}}<br>"}'
where node_type = 'CI'
and event_type_id = 1;


---- update notification template for CI success ses/smtp
UPDATE notification_templates
set template_payload = '{"from": "{{fromEmail}}",
 "to": "{{toEmail}}",
 "subject": "CI success for app: {{appName}}",
 "html": "<h2 style=\"color:#1dad70;\">Build Pipeline Successful</h2><span>{{eventTime}}</span><br><span>Triggered by <strong>{{triggeredBy}}</strong></span><br><br><a href=\"{{& buildHistoryLink }}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:none;min-width:64px;text-transform:capitalize;text-align:center;background:#0066cc;color:#fff;border:1px solid transparent;cursor:pointer;\">View Pipeline</a><br><br><hr><br><span>Application: <strong>{{appName}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Pipeline: <strong>{{pipelineName}}</strong></span><br><br><hr><h3>Source Code</h3>{{#ciMaterials}}{{^webhookType}}<span>Branch: <strong>{{appName}}/{{branch}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Commit: <a href=\"{{& commitLink }}\"><strong>{{commit}}</strong></a></span><br><br>{{/webhookType}}{{#webhookType}}{{#webhookData.mergedType}}<span>Title: <strong>{{webhookData.data.title}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Git URL: <a href=\"{{& webhookData.data.giturl}}\"><strong>View</strong></a></span><br><br><span>Source Branch: <strong>{{webhookData.data.sourcebranchname}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Source Commit: <a href=\"{{& webhookData.data.sourcecheckoutlink}}\"><strong>{{webhookData.data.sourcecheckout}}</strong></a></span><br><br><span>Target Branch: <strong>{{webhookData.data.targetbranchname}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Target Commit: <a href=\"{{& webhookData.data.targetcheckoutlink}}\"><strong>{{webhookData.data.targetcheckout}}</strong></a></span><br><br>{{/webhookData.mergedType}}{{^webhookData.mergedType}}<span>Target Checkout: <strong>{{webhookData.data.targetcheckout}}</strong></span><br>{{/webhookData.mergedType}}{{/webhookType}}{{/ciMaterials}}<br>"}'
where node_type = 'CI'
and event_type_id = 2;



---- update notification template for CI fail ses/smtp
UPDATE notification_templates
set template_payload = '{"from": "{{fromEmail}}",
 "to": "{{toEmail}}",
 "subject": "CI failed for app: {{appName}}",
 "html": "<h2 style=\"color:#f33e3e;\">Build Pipeline Failed</h2><span>{{eventTime}}</span><br><span>Triggered by <strong>{{triggeredBy}}</strong></span><br><br><a href=\"{{& buildHistoryLink }}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:none;min-width:64px;text-transform:capitalize;text-align:center;background:#0066cc;color:#fff;border:1px solid transparent;cursor:pointer;\">View Pipeline</a><br><br><hr><br><span>Application: <strong>{{appName}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Pipeline: <strong>{{pipelineName}}</strong></span><br><br><hr><h3>Source Code</h3>{{#ciMaterials}}{{^webhookType}}<span>Branch: <strong>{{appName}}/{{branch}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Commit: <a href=\"{{& commitLink }}\"><strong>{{commit}}</strong></a></span><br><br>{{/webhookType}}{{#webhookType}}{{#webhookData.mergedType}}<span>Title: <strong>{{webhookData.data.title}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Git URL: <a href=\"{{& webhookData.data.giturl}}\"><strong>View</strong></a></span><br><br><span>Source Branch: <strong>{{webhookData.data.sourcebranchname}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Source Commit: <a href=\"{{& webhookData.data.sourcecheckoutlink}}\"><strong>{{webhookData.data.sourcecheckout}}</strong></a></span><br><br><span>Target Branch: <strong>{{webhookData.data.targetbranchname}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Target Commit: <a href=\"{{& webhookData.data.targetcheckoutlink}}\"><strong>{{webhookData.data.targetcheckout}}</strong></a></span><br><br>{{/webhookData.mergedType}}{{^webhookData.mergedType}}<span>Target Checkout: <strong>{{webhookData.data.targetcheckout}}</strong></span><br>{{/webhookData.mergedType}}{{/webhookType}}{{/ciMaterials}}<br>"}'
where  node_type = 'CI'
and event_type_id = 3;

UPDATE "public"."notification_templates" SET template_payload = '{"from": "{{fromEmail}}",
 "to": "{{toEmail}}",
 "subject": "CD triggered for app: {{appName}} on environment: {{envName}}",
 "html": "<h2 style=\"color:#767d84;\">{{stage}} Pipeline Triggered</h2><span>{{eventTime}}</span><br><span>Triggered by <strong>{{triggeredBy}}</strong></span><br><br>{{#deploymentHistoryLink}}<a href=\"{{& deploymentHistoryLink}}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:none;min-width:64px;text-transform:capitalize;text-align:center;background:#0066cc;color:#fff;border:1px solid transparent;cursor:pointer;\">View Pipeline</a>{{/deploymentHistoryLink}}&nbsp;&nbsp;&nbsp;{{#appDetailsLink}}<a href=\"{{& appDetailsLink}}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:none;min-width:64px;text-transform:capitalize;text-align:center;background:#fff;color:#3b444c;border:1px solid #d0d4d9;cursor:pointer;\">App Details</a><br><br>{{/appDetailsLink}}<hr><br><span>Application: <strong>{{appName}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Pipeline: <strong>{{pipelineName}}</strong></span><br><br><span>Environment: <strong>{{envName}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Stage: <strong>{{stage}}</strong></span><br><br><hr><h3>Source Code</h3>{{#ciMaterials}}{{^webhookType}}<span>Branch: <strong>{{appName}}/{{branch}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Commit: <a href=\"{{& commitLink}}\"><strong>{{commit}}</strong></a></span><br>{{/webhookType}}{{#webhookType}}{{#webhookData.mergedType}}<span>Title: <strong>{{webhookData.data.title}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Git URL: <a href=\"{{& webhookData.data.giturl}}\"><strong>View</strong></a></span><br><span>Source Branch: <strong>{{webhookData.data.sourcebranchname}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Source Commit: <a href=\"{{& webhookData.data.sourcecheckoutlink}}\"><strong>{{webhookData.data.sourcecheckout}}</strong></a></span><br><span>Target Branch: <strong>{{webhookData.data.targetbranchname}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Target Commit: <a href=\"{{& webhookData.data.targetcheckoutlink}}\"><strong>{{webhookData.data.targetcheckout}}</strong></a></span><br>{{/webhookData.mergedType}}{{^webhookData.mergedType}}<span>Target Checkout: <strong>{{webhookData.data.targetcheckout}}</strong></span><br>{{/webhookData.mergedType}}{{/webhookType}}{{/ciMaterials}}<br><br><hr><h3>Image</h3><span>Docker Image: <strong>{{dockerImg}}</strong></span><br>"}' WHERE event_type_id=1 AND  node_type='CD' AND channel_type!='slack';
UPDATE "public"."notification_templates" SET template_payload = '{"from": "{{fromEmail}}",
 "to": "{{toEmail}}",
 "subject": "CD success for app: {{appName}} on environment: {{envName}}",
 "html": "<h2 style=\"color:#1dad70;\">{{stage}} Pipeline Successful</h2><span>{{eventTime}}</span><br><span>Triggered by <strong>{{triggeredBy}}</strong></span><br><br>{{#deploymentHistoryLink}}<a href=\"{{& deploymentHistoryLink}}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:none;min-width:64px;text-transform:capitalize;text-align:center;background:#0066cc;color:#fff;border:1px solid transparent;cursor:pointer;\">View Pipeline</a>{{/deploymentHistoryLink}}&nbsp;&nbsp;&nbsp;{{#appDetailsLink}}<a href=\"{{& appDetailsLink}}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:none;min-width:64px;text-transform:capitalize;text-align:center;background:#fff;color:#3b444c;border:1px solid #d0d4d9;cursor:pointer;\">App Details</a><br><br>{{/appDetailsLink}}<hr><br><span>Application: <strong>{{appName}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Pipeline: <strong>{{pipelineName}}</strong></span><br><br><span>Environment: <strong>{{envName}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Stage: <strong>{{stage}}</strong></span><br><br><hr><h3>Source Code</h3>{{#ciMaterials}}{{^webhookType}}<span>Branch: <strong>{{appName}}/{{branch}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Commit: <a href=\"{{& commitLink}}\"><strong>{{commit}}</strong></a></span><br>{{/webhookType}}{{#webhookType}}{{#webhookData.mergedType}}<span>Title: <strong>{{webhookData.data.title}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Git URL: <a href=\"{{& webhookData.data.giturl}}\"><strong>View</strong></a></span><br><span>Source Branch: <strong>{{webhookData.data.sourcebranchname}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Source Commit: <a href=\"{{& webhookData.data.sourcecheckoutlink}}\"><strong>{{webhookData.data.sourcecheckout}}</strong></a></span><br><span>Target Branch: <strong>{{webhookData.data.targetbranchname}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Target Commit: <a href=\"{{& webhookData.data.targetcheckoutlink}}\"><strong>{{webhookData.data.targetcheckout}}</strong></a></span><br>{{/webhookData.mergedType}}{{^webhookData.mergedType}}<span>Target Checkout: <strong>{{webhookData.data.targetcheckout}}</strong></span><br>{{/webhookData.mergedType}}{{/webhookType}}{{/ciMaterials}}<br><br><hr><h3>Image</h3><span>Docker Image: <strong>{{dockerImg}}</strong></span><br>"}' WHERE event_type_id=2 AND  node_type='CD' AND channel_type!='slack';
UPDATE "public"."notification_templates" SET template_payload = '{"from": "{{fromEmail}}",
 "to": "{{toEmail}}",
 "subject": "CD failed for app: {{appName}} on environment: {{envName}}",
 "html": "<h2 style=\"color:#f33e3e;\">{{stage}} Pipeline Failed</h2><span>{{eventTime}}</span><br><span>Triggered by <strong>{{triggeredBy}}</strong></span><br><br>{{#deploymentHistoryLink}}<a href=\"{{& deploymentHistoryLink}}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:none;min-width:64px;text-transform:capitalize;text-align:center;background:#0066cc;color:#fff;border:1px solid transparent;cursor:pointer;\">View Pipeline</a>{{/deploymentHistoryLink}}&nbsp;&nbsp;&nbsp;{{#appDetailsLink}}<a href=\"{{& appDetailsLink}}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:none;min-width:64px;text-transform:capitalize;text-align:center;background:#fff;color:#3b444c;border:1px solid #d0d4d9;cursor:pointer;\">App Details</a><br><br>{{/appDetailsLink}}<hr><br><span>Application: <strong>{{appName}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Pipeline: <strong>{{pipelineName}}</strong></span><br><br><span>Environment: <strong>{{envName}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Stage: <strong>{{stage}}</strong></span><br><br><hr><h3>Source Code</h3>{{#ciMaterials}}{{^webhookType}}<span>Branch: <strong>{{appName}}/{{branch}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Commit: <a href=\"{{& commitLink}}\"><strong>{{commit}}</strong></a></span><br>{{/webhookType}}{{#webhookType}}{{#webhookData.mergedType}}<span>Title: <strong>{{webhookData.data.title}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Git URL: <a href=\"{{& webhookData.data.giturl}}\"><strong>View</strong></a></span><br><span>Source Branch: <strong>{{webhookData.data.sourcebranchname}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Source Commit: <a href=\"{{& webhookData.data.sourcecheckoutlink}}\"><strong>{{webhookData.data.sourcecheckout}}</strong></a></span><br><span>Target Branch: <strong>{{webhookData.data.targetbranchname}}</strong></span>&nbsp;&nbsp;|&nbsp;&nbsp;<span>Target Commit: <a href=\"{{& webhookData.data.targetcheckoutlink}}\"><strong>{{webhookData.data.targetcheckout}}</strong></a></span><br>{{/webhookData.mergedType}}{{^webhookData.mergedType}}<span>Target Checkout: <strong>{{webhookData.data.targetcheckout}}</strong></span><br>{{/webhookData.mergedType}}{{/webhookType}}{{/ciMaterials}}<br><br><hr><h3>Image</h3><span>Docker Image: <strong>{{dockerImg}}</strong></span><br>"}' WHERE event_type_id=3 AND  node_type='CD' AND channel_type!='slack';
delete from "public"."notification_templates" where event_type_id=6;
delete from notifier_event_log where event_type_id=6;
delete from public.event where event_type='BLOCKED';
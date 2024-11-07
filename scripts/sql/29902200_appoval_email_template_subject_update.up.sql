
---- update notification template for CD approval smtp
UPDATE notification_templates
set template_payload = '{"from": "{{fromEmail}}", "to": "{{toEmail}}","subject": "🛎️ Image approval requested | Application > {{appName}} | Environment > {{envName}} at {{eventTime}}","html": "<table cellpadding=\"0\" style=\"font-family:Arial,Verdana,Helvetica;width:600px;height:485px;border-collapse:inherit;border-spacing:0;border:1px solid #d0d4d9;border-radius:8px;padding:16px 20px;margin:20px auto;box-shadow:0 0 8px 0 rgba(0,0,0,.1)\"><tr><td colspan=\"3\"><div style=\"padding-bottom:16px;margin-bottom:20px;border-bottom:1px solid #edf1f5;max-width:600px\"><img style=\"max-width:122px\" src=\"https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/devtron/devtron-logo.png\" alt=\"ci-triggered\"></div></td></tr><tr><td colspan=\"3\"><div style=\"background-color:#e5f2ff;border-top-left-radius:8px;border-top-right-radius:8px;padding:20px 20px 16px 20px;display:flex;justify-content:space-between\"><div style=\"width:90%\"><div style=\"font-size:16px;line-height:24px;font-weight:600;margin-bottom:6px;color:#000a14\">Image approval request</div><span style=\"font-size:14px;line-height:20px;color:#000a14\">{{eventTime}}</span><br><div><span style=\"font-size:14px;line-height:20px;color:#000a14\">by</span><span style=\"font-size:14px;line-height:20px;color:#06c;margin-left:4px\">{{triggeredBy}}</span></div></div><div><img src=\"https://cdn.devtron.ai/images/img_build_notification.png\" style=\"height:72px;width:72px\"></div></div></td></tr><tr><td colspan=\"3\"><div style=\"display:flex\"><div style=\"background-color:#e5f2ff;border-bottom-left-radius:8px;padding:0 0 20px 20px\"><a href=\"{{&approvalLink}}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:0;min-width:64px;text-transform:capitalize;text-align:center;background:#06c;color:#fff;border:1px solid transparent;cursor:pointer\">Approve</a></div><div style=\"width:90%;background-color:#e5f2ff;border-bottom-right-radius:8px;padding:0 0 20px 8px\">{{#imageApprovalLink}}<a href=\"{{&imageApprovalLink}}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:0;min-width:64px;text-transform:capitalize;text-align:center;border:1px solid #d0d4d9;color:#06c;cursor:pointer; background-color: #fff\">View request</a>{{/imageApprovalLink}}</div></div></td></tr><tr><tr><td><br></td></tr><td><div style=\"color:#3b444c;font-size:13px;\">Application</div></td><td colspan=\"2\"><div style=\"color:#3b444c;font-size:13px;\">Environment</div></td><tr><td><div style=\"color:#000a14;font-size:14px\">{{appName}}</div></td><td><div style=\"color:#000a14;font-size:14px\">{{envName}}</div></td></tr><tr><td colspan=\"3\"><div style=\"font-weight:600;margin-top:20px;width:100%;border-top:1px solid #edf1f5;padding:16px 0 12px;font-size:14px\">Image Details</div></td></tr><tr><td><div style=\"color:#3b444c;font-size:13px;\">Image tag</div></td></tr><tr><td><div style=\"color:#000a14;font-size:14px\">{{imageTag}}</div></td><tr><td><br></td></tr><tr><td colspan=\"3\"><div style=\"color:#3b444c;font-size:13px;display:{{commentDisplayStyle}};\">Comment</div></td></tr><tr><td colspan=\"3\"><div style=\"color:#000a14;font-size:14px;display:{{commentDisplayStyle}};\">{{comment}}</div></tr><tr><td><br></td></tr><tr><td colspan=\"3\"><div style=\"color:#3b444c;font-size:13px; display: {{tagDisplayStyle}};\">Tags</div></td></tr><tr><td colspan=\"3\"><div style=\"color:#000a14;font-size:14px;display: {{tagDisplayStyle}};\">{{tags}}</div></tr><tr><td colspan=\"3\"><div style=\"border-top:1px solid #edf1f5;margin:20px 0 16px 0;height:1px\"></div></td></tr><tr><td colspan=\"2\" style=\"display:flex\"><span><a href=\"https://twitter.com/DevtronL\" target=\"_blank\" style=\"cursor:pointer;text-decoration:none;padding-right:12px;display:flex\"><div><img style=\"width:20px\" src=\"https://cdn.devtron.ai/images/twitter_social_dark.png\"></div></a></span><span><a href=\"https://www.linkedin.com/company/devtron-labs/mycompany/\" target=\"_blank\" style=\"cursor:pointer;text-decoration:none;padding-right:12px;display:flex\"><div><img style=\"width:20px\" src=\"https://cdn.devtron.ai/images/linkedin_social_dark.png\"></div></a></span><span><a href=\"https://devtron.ai/blog/\" target=\"_blank\" style=\"color:#000a14;font-size:13px;line-height:20px;cursor:pointer;text-decoration:underline;padding-right:12px\">Blog</a></span><span><a href=\"https://devtron.ai/\" target=\"_blank\" style=\"color:#000a14;font-size:13px;line-height:20px;cursor:pointer;text-decoration:underline\">Website</a></span></td><td colspan=\"2\" style=\"text-align:right\"><div style=\"color:#767d84;font-size:13px;line-height:20px\">&copy; Devtron Labs 2024</div></td></tr></table>"}'
where channel_type = 'smtp'
and node_type = 'CD'
and event_type_id = 4;

---- update notification template for CD approval ses
UPDATE notification_templates
set template_payload = '{"from": "{{fromEmail}}", "to": "{{toEmail}}","subject": "🛎️ Image approval requested | Application > {{appName}} | Environment > {{envName}} at {{eventTime}}","html": "<table cellpadding=\"0\" style=\"font-family:Arial,Verdana,Helvetica;width:600px;height:485px;border-collapse:inherit;border-spacing:0;border:1px solid #d0d4d9;border-radius:8px;padding:16px 20px;margin:20px auto;box-shadow:0 0 8px 0 rgba(0,0,0,.1)\"><tr><td colspan=\"3\"><div style=\"padding-bottom:16px;margin-bottom:20px;border-bottom:1px solid #edf1f5;max-width:600px\"><img style=\"max-width:122px\" src=\"https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/devtron/devtron-logo.png\" alt=\"ci-triggered\"></div></td></tr><tr><td colspan=\"3\"><div style=\"background-color:#e5f2ff;border-top-left-radius:8px;border-top-right-radius:8px;padding:20px 20px 16px 20px;display:flex;justify-content:space-between\"><div style=\"width:90%\"><div style=\"font-size:16px;line-height:24px;font-weight:600;margin-bottom:6px;color:#000a14\">Image approval request</div><span style=\"font-size:14px;line-height:20px;color:#000a14\">{{eventTime}}</span><br><div><span style=\"font-size:14px;line-height:20px;color:#000a14\">by</span><span style=\"font-size:14px;line-height:20px;color:#06c;margin-left:4px\">{{triggeredBy}}</span></div></div><div><img src=\"https://cdn.devtron.ai/images/img_build_notification.png\" style=\"height:72px;width:72px\"></div></div></td></tr><tr><td colspan=\"3\"><div style=\"display:flex\"><div style=\"background-color:#e5f2ff;border-bottom-left-radius:8px;padding:0 0 20px 20px\"><a href=\"{{&approvalLink}}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:0;min-width:64px;text-transform:capitalize;text-align:center;background:#06c;color:#fff;border:1px solid transparent;cursor:pointer\">Approve</a></div><div style=\"width:90%;background-color:#e5f2ff;border-bottom-right-radius:8px;padding:0 0 20px 8px\">{{#imageApprovalLink}}<a href=\"{{&imageApprovalLink}}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:0;min-width:64px;text-transform:capitalize;text-align:center;border:1px solid #d0d4d9;color:#06c;cursor:pointer; background-color: #fff\">View request</a>{{/imageApprovalLink}}</div></div></td></tr><tr><tr><td><br></td></tr><td><div style=\"color:#3b444c;font-size:13px;\">Application</div></td><td colspan=\"2\"><div style=\"color:#3b444c;font-size:13px;\">Environment</div></td><tr><td><div style=\"color:#000a14;font-size:14px\">{{appName}}</div></td><td><div style=\"color:#000a14;font-size:14px\">{{envName}}</div></td></tr><tr><td colspan=\"3\"><div style=\"font-weight:600;margin-top:20px;width:100%;border-top:1px solid #edf1f5;padding:16px 0 12px;font-size:14px\">Image Details</div></td></tr><tr><td><div style=\"color:#3b444c;font-size:13px;\">Image tag</div></td></tr><tr><td><div style=\"color:#000a14;font-size:14px\">{{imageTag}}</div></td><tr><td><br></td></tr><tr><td colspan=\"3\"><div style=\"color:#3b444c;font-size:13px;display:{{commentDisplayStyle}};\">Comment</div></td></tr><tr><td colspan=\"3\"><div style=\"color:#000a14;font-size:14px;display:{{commentDisplayStyle}};\">{{comment}}</div></tr><tr><td><br></td></tr><tr><td colspan=\"3\"><div style=\"color:#3b444c;font-size:13px; display: {{tagDisplayStyle}};\">Tags</div></td></tr><tr><td colspan=\"3\"><div style=\"color:#000a14;font-size:14px;display: {{tagDisplayStyle}};\">{{tags}}</div></tr><tr><td colspan=\"3\"><div style=\"border-top:1px solid #edf1f5;margin:20px 0 16px 0;height:1px\"></div></td></tr><tr><td colspan=\"2\" style=\"display:flex\"><span><a href=\"https://twitter.com/DevtronL\" target=\"_blank\" style=\"cursor:pointer;text-decoration:none;padding-right:12px;display:flex\"><div><img style=\"width:20px\" src=\"https://cdn.devtron.ai/images/twitter_social_dark.png\"></div></a></span><span><a href=\"https://www.linkedin.com/company/devtron-labs/mycompany/\" target=\"_blank\" style=\"cursor:pointer;text-decoration:none;padding-right:12px;display:flex\"><div><img style=\"width:20px\" src=\"https://cdn.devtron.ai/images/linkedin_social_dark.png\"></div></a></span><span><a href=\"https://devtron.ai/blog/\" target=\"_blank\" style=\"color:#000a14;font-size:13px;line-height:20px;cursor:pointer;text-decoration:underline;padding-right:12px\">Blog</a></span><span><a href=\"https://devtron.ai/\" target=\"_blank\" style=\"color:#000a14;font-size:13px;line-height:20px;cursor:pointer;text-decoration:underline\">Website</a></span></td><td colspan=\"2\" style=\"text-align:right\"><div style=\"color:#767d84;font-size:13px;line-height:20px\">&copy; Devtron Labs 2024</div></td></tr></table>"}'
where channel_type = 'ses'
and node_type = 'CD'
and event_type_id = 4;

---- update notification template for Config approval smtp
UPDATE notification_templates
set template_payload = '{"from": "{{fromEmail}}", "to": "{{toEmail}}","subject": "🛎️ Config Change approval requested | Application > {{appName} | Change Impacts > {{envName}}  at {{eventTime}}","html": "<table cellpadding=\"0\" style=\"font-family:Arial,Verdana,Helvetica;width:600px;height:485px;border-collapse:inherit;border-spacing:0;border:1px solid #d0d4d9;border-radius:8px;padding:16px 20px;margin:20px auto;box-shadow:0 0 8px 0 rgba(0,0,0,.1)\"><tr><td colspan=\"3\"><div style=\"padding-bottom:16px;margin-bottom:20px;border-bottom:1px solid #edf1f5;max-width:600px\"><img style=\"max-width:122px\" src=\"https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/devtron/devtron-logo.png\" alt=\"ci-triggered\"></div></td></tr><tr><td colspan=\"3\"><div style=\"background-color:#e5f2ff;border-top-left-radius:8px;border-top-right-radius:8px;padding:20px 20px 16px 20px;display:flex;justify-content:space-between\"><div style=\"width:90%\"><div style=\"font-size:16px;line-height:24px;font-weight:600;margin-bottom:6px;color:#000a14\">Config Change approval request</div><span style=\"font-size:14px;line-height:20px;color:#000a14\">{{eventTime}}</span><br><div><span style=\"font-size:14px;line-height:20px;color:#000a14\">by</span><span style=\"font-size:14px;line-height:20px;color:#06c;margin-left:4px\">{{triggeredBy}}</span></div></div><div><img src=\"https://cdn.devtron.ai/images/img_build_notification.png\" style=\"height:72px;width:72px\"></div></div></td></tr><tr><td colspan=\"3\"><div style=\"display:flex\"><div style=\"background-color:#e5f2ff;border-bottom-left-radius:8px;padding:0 0 20px 20px\">{{#approvalLink}}<a href=\"{{&approvalLink}}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:0;min-width:64px;text-transform:capitalize;text-align:center;background:#06c;color:#fff;border:1px solid transparent;cursor:pointer\">Approve</a>{{/approvalLink}}</div><div style=\"width:90%;background-color:#e5f2ff;border-bottom-right-radius:8px;padding:0 0 20px 8px\">{{#protectConfigLink}}<a href=\"{{&protectConfigLink}}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:0;min-width:64px;text-transform:capitalize;text-align:center;border:1px solid #d0d4d9;color:#06c;cursor:pointer; background-color: #fff\">View request</a>{{/protectConfigLink}}</div></div></td></tr><tr><tr><td><br></td></tr><td><div style=\"color:#3b444c;font-size:13px;\">Application</div></td><td colspan=\"2\"><div style=\"color:#3b444c;font-size:13px;\">Change Impacts</div></td><tr><td><div style=\"color:#000a14;font-size:14px\">{{appName}}</div></td><td><div style=\"color:#000a14;font-size:14px\">{{envName}}</div></td></tr><tr><td colspan=\"3\"><div style=\"font-weight:600;margin-top:20px;width:100%;border-top:1px solid #edf1f5;padding:16px 0 12px 0;font-size:14px\">File Details</div></td></tr><tr><td><div style=\"color:#3b444c;font-size:13px;\">File Type</div></td><td><div style=\"color:#3b444c;font-size:13px;\">File Name</div></td></tr><tr></tr><tr><td><div style=\"color:#000a14;font-size:14px\">{{protectConfigFileType}}</div></td><td><div style=\"color:#000a14;font-size:14px\">{{protectConfigFileName}}</div></td><tr><td><br></td></tr><tr><td colspan=\"3\"><div style=\"color:#3b444c;font-size:13px;display: {{commentDisplayStyle}}\">Comment</div></td></tr><tr><td colspan=\"3\"><div style=\"color:#000a14;font-size:14px; display: {{commentDisplayStyle}}\">{{protectConfigComment}}</div></tr><tr><td colspan=\"3\"><div style=\"border-top:1px solid #edf1f5;margin:20px 0 16px 0;height:1px\"></div></td></tr><tr><td colspan=\"2\" style=\"display:flex\"><span><a href=\"https://twitter.com/DevtronL\" target=\"_blank\" style=\"cursor:pointer;text-decoration:none;padding-right:12px;display:flex\"><div><img style=\"width:20px\" src=\"https://cdn.devtron.ai/images/twitter-new.png\"></div></a></span><span><a href=\"https://www.linkedin.com/company/devtron-labs/mycompany/\" target=\"_blank\" style=\"cursor:pointer;text-decoration:none;padding-right:12px;display:flex\"><div><img style=\"width:20px\" src=\"https://cdn.devtron.ai/images/linkedin_social_dark.png\"></div></a></span><span><a href=\"https://devtron.ai/blog/\" target=\"_blank\" style=\"color:#000a14;font-size:13px;line-height:20px;cursor:pointer;text-decoration:underline;padding-right:12px\">Blog</a></span><span><a href=\"https://devtron.ai/\" target=\"_blank\" style=\"color:#000a14;font-size:13px;line-height:20px;cursor:pointer;text-decoration:underline\">Website</a></span></td><td colspan=\"2\" style=\"text-align:right\"><div style=\"color:#767d84;font-size:13px;line-height:20px\">&copy; Devtron Labs 2024</div></td></tr></table>"}'
where channel_type = 'smtp'
and node_type = 'CD'
and event_type_id = 5;

---- update notification template for Config approval ses
UPDATE notification_templates
set template_payload = '{"from": "{{fromEmail}}", "to": "{{toEmail}}","subject": "🛎️ Config Change approval requested | Application > {{appName} | Change Impacts > {{envName}} at {{eventTime}}","html": "<table cellpadding=\"0\" style=\"font-family:Arial,Verdana,Helvetica;width:600px;height:485px;border-collapse:inherit;border-spacing:0;border:1px solid #d0d4d9;border-radius:8px;padding:16px 20px;margin:20px auto;box-shadow:0 0 8px 0 rgba(0,0,0,.1)\"><tr><td colspan=\"3\"><div style=\"padding-bottom:16px;margin-bottom:20px;border-bottom:1px solid #edf1f5;max-width:600px\"><img style=\"max-width:122px\" src=\"https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/devtron/devtron-logo.png\" alt=\"ci-triggered\"></div></td></tr><tr><td colspan=\"3\"><div style=\"background-color:#e5f2ff;border-top-left-radius:8px;border-top-right-radius:8px;padding:20px 20px 16px 20px;display:flex;justify-content:space-between\"><div style=\"width:90%\"><div style=\"font-size:16px;line-height:24px;font-weight:600;margin-bottom:6px;color:#000a14\">Config Change approval request</div><span style=\"font-size:14px;line-height:20px;color:#000a14\">{{eventTime}}</span><br><div><span style=\"font-size:14px;line-height:20px;color:#000a14\">by</span><span style=\"font-size:14px;line-height:20px;color:#06c;margin-left:4px\">{{triggeredBy}}</span></div></div><div><img src=\"https://cdn.devtron.ai/images/img_build_notification.png\" style=\"height:72px;width:72px\"></div></div></td></tr><tr><td colspan=\"3\"><div style=\"display:flex\"><div style=\"background-color:#e5f2ff;border-bottom-left-radius:8px;padding:0 0 20px 20px\">{{#approvalLink}}<a href=\"{{&approvalLink}}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:0;min-width:64px;text-transform:capitalize;text-align:center;background:#06c;color:#fff;border:1px solid transparent;cursor:pointer\">Approve</a>{{/approvalLink}}</div><div style=\"width:90%;background-color:#e5f2ff;border-bottom-right-radius:8px;padding:0 0 20px 8px\">{{#protectConfigLink}}<a href=\"{{&protectConfigLink}}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:0;min-width:64px;text-transform:capitalize;text-align:center;border:1px solid #d0d4d9;color:#06c;cursor:pointer; background-color: #fff\">View request</a>{{/protectConfigLink}}</div></div></td></tr><tr><tr><td><br></td></tr><td><div style=\"color:#3b444c;font-size:13px;\">Application</div></td><td colspan=\"2\"><div style=\"color:#3b444c;font-size:13px;\">Change Impacts</div></td><tr><td><div style=\"color:#000a14;font-size:14px\">{{appName}}</div></td><td><div style=\"color:#000a14;font-size:14px\">{{envName}}</div></td></tr><tr><td colspan=\"3\"><div style=\"font-weight:600;margin-top:20px;width:100%;border-top:1px solid #edf1f5;padding:16px 0 12px 0;font-size:14px\">File Details</div></td></tr><tr><td><div style=\"color:#3b444c;font-size:13px;\">File Type</div></td><td><div style=\"color:#3b444c;font-size:13px;\">File Name</div></td></tr><tr></tr><tr><td><div style=\"color:#000a14;font-size:14px\">{{protectConfigFileType}}</div></td><td><div style=\"color:#000a14;font-size:14px\">{{protectConfigFileName}}</div></td><tr><td><br></td></tr><tr><td colspan=\"3\"><div style=\"color:#3b444c;font-size:13px;display: {{commentDisplayStyle}}\">Comment</div></td></tr><tr><td colspan=\"3\"><div style=\"color:#000a14;font-size:14px; display: {{commentDisplayStyle}}\">{{protectConfigComment}}</div></tr><tr><td colspan=\"3\"><div style=\"border-top:1px solid #edf1f5;margin:20px 0 16px 0;height:1px\"></div></td></tr><tr><td colspan=\"2\" style=\"display:flex\"><span><a href=\"https://twitter.com/DevtronL\" target=\"_blank\" style=\"cursor:pointer;text-decoration:none;padding-right:12px;display:flex\"><div><img style=\"width:20px\" src=\"https://cdn.devtron.ai/images/twitter-new.png\"></div></a></span><span><a href=\"https://www.linkedin.com/company/devtron-labs/mycompany/\" target=\"_blank\" style=\"cursor:pointer;text-decoration:none;padding-right:12px;display:flex\"><div><img style=\"width:20px\" src=\"https://cdn.devtron.ai/images/linkedin_social_dark.png\"></div></a></span><span><a href=\"https://devtron.ai/blog/\" target=\"_blank\" style=\"color:#000a14;font-size:13px;line-height:20px;cursor:pointer;text-decoration:underline;padding-right:12px\">Blog</a></span><span><a href=\"https://devtron.ai/\" target=\"_blank\" style=\"color:#000a14;font-size:13px;line-height:20px;cursor:pointer;text-decoration:underline\">Website</a></span></td><td colspan=\"2\" style=\"text-align:right\"><div style=\"color:#767d84;font-size:13px;line-height:20px\">&copy; Devtron Labs 2024</div></td></tr></table>"}'
where channel_type = 'ses'
and node_type = 'CD'
and event_type_id = 5;

---- update notification template for Config approval smtp auto
UPDATE notification_templates
set template_payload = '{"from": "{{fromEmail}}", "to": "{{toEmail}}","subject": "🚫  Auto-deployment blocked:| Application: {{appName}} | Environment:  {{envName}} at {{eventTime}}","html":"<table cellpadding=0 style=\"font-family:Arial,Verdana,Helvetica;width:600px;height:485px;border-collapse:inherit;border-spacing:0;border:1px solid #d0d4d9;border-radius:8px;padding:16px 20px;margin:20px auto;box-shadow:0 0 8px 0 rgba(0,0,0,.1)\"><tr><td colspan=3><div style=\"padding-bottom:16px;margin-bottom:20px;border-bottom:1px solid #edf1f5;max-width:600px\"><img src=https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/devtron/devtron-logo.png style=max-width:122px alt=ci-triggered></div><tr><td colspan=3><div style=\"background-color:#FDE7E7;border-top-left-radius:8px;border-top-right-radius:8px;padding:20px 20px 16px 20px;display:flex;justify-content:space-between\"><div style=width:90%><div style=font-size:16px;line-height:24px;font-weight:600;margin-bottom:6px;color:#000a14>🚫  Auto-deployment blocked</div><span style=font-size:14px;line-height:20px;color:#000a14>{{eventTime}}</span><br><div><span style=font-size:14px;line-height:20px;color:#000a14>by</span><span style=font-size:14px;line-height:20px;color:#06c;margin-left:4px>{{triggeredBy}}</span></div></div><div><img src=https://cdn.devtron.ai/images/image_deploy_notification.png style=height:72px;width:72px></div></div><tr><td colspan=3><div style=display:flex><div style=\"width:124px;background-color:#FDE7E7;border-bottom-left-radius:8px;padding:0 0 20px 20px\">{{#deploymentHistoryLink}}<a href={{&deploymentHistoryLink}} style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:0;min-width:64px;text-transform:capitalize;text-align:center;background:#06c;color:#fff;border:1px solid transparent;cursor:pointer\">View Pipeline</a>{{/deploymentHistoryLink}}</div><div style=\"width:90%;background-color:#FDE7E7;border-bottom-right-radius:8px;padding:0 0 20px 20px\">{{#appDetailsLink}}<a href={{&appDetailsLink}} style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:0;min-width:64px;text-transform:capitalize;text-align:center;border:1px solid #d0d4d9;color:#06c;cursor:pointer;width:90%\">App Details</a>{{/appDetailsLink}}</div></div><tr><td><br><tr><td colspan=3><div style=\"display:flex;font-weight:600;border-radius:4px;border:1px solid #fde7e7;background:#fff;padding:8px 16px;margin-bottom:16px;display:{{deploymentWindowCommentStyle}};\"><img src=https://cdn.devtron.ai/images/shield-image.png style=margin-right:4px;height:16px;width:16px>{{deploymentWindowComment}}</div><tr><td><div style=color:#3b444c;font-size:13px>Application</div><td colspan=2><div style=color:#3b444c;font-size:13px>Environment</div><tr><td><div style=color:#000a14;font-size:14px>{{appName}}</div><td><div style=color:#000a14;font-size:14px>{{envName}}</div><tr><td><div style=color:#3b444c;font-size:13px;margin-top:12px>Stage</div><tr><td><div style=color:#000a14;font-size:14px>{{stage}}</div><tr><tr><td colspan=3><div style=\"font-weight:600;margin-top:20px;width:100%;border-top:1px solid #edf1f5;padding:16px 0 12px;font-size:14px\">Source Code</div></tr>{{#ciMaterials}} {{^webhookType}}<tr><td><div style=color:#3b444c;font-size:13px>Branch</div><td colspan=2><div style=color:#3b444c;font-size:13px>Commit</div><tr><tr><td><div style=color:#000a14;font-size:14px>{{appName}}/{{branch}}</div><td><div style=color:#000a14;font-size:14px>{{commit}}</div></tr>{{/webhookType}} {{/ciMaterials}}<tr><td colspan=3><div style=\"font-weight:600;margin-top:20px;width:100%;border-top:1px solid #edf1f5;padding:16px 0 12px;font-size:14px\">Image Details</div><tr><td><div style=color:#3b444c;font-size:13px>Image tag</div><tr><td><div style=color:#000a14;font-size:14px>{{dockerImg}}</div><tr><td><br><tr><td colspan=3><div style=\"border-top:1px solid #edf1f5;margin:20px 0 16px 0;height:1px\"></div><tr><td colspan=2 style=display:flex><span><a href=https://twitter.com/DevtronL style=cursor:pointer;text-decoration:none;padding-right:12px;display:flex target=_blank><div><img src=https://cdn.devtron.ai/images/twitter_social_dark.png style=width:20px></div></a></span><span><a href=https://www.linkedin.com/company/devtron-labs/mycompany/ style=cursor:pointer;text-decoration:none;padding-right:12px;display:flex target=_blank><div><img src=https://cdn.devtron.ai/images/linkedin_social_dark.png style=width:20px></div></a></span><span><a href=https://devtron.ai/blog/ style=color:#000a14;font-size:13px;line-height:20px;cursor:pointer;text-decoration:underline;padding-right:12px target=_blank>Blog</a></span><span><a href=https://devtron.ai/ style=color:#000a14;font-size:13px;line-height:20px;cursor:pointer;text-decoration:underline target=_blank>Website</a></span><td colspan=2 style=text-align:right><div style=color:#767d84;font-size:13px;line-height:20px>© Devtron Labs 2024</div></table>"}'
where channel_type = 'smtp'
and node_type = 'CD'
and event_type_id = 6;

---- update notification template for Config approval ses auto
UPDATE notification_templates
set template_payload = '{"from": "{{fromEmail}}", "to": "{{toEmail}}","subject": "🚫  Auto-deployment blocked:| Application: {{appName}} | Environment:  {{envName}} at {{eventTime}}","html":"<table cellpadding=0 style=\"font-family:Arial,Verdana,Helvetica;width:600px;height:485px;border-collapse:inherit;border-spacing:0;border:1px solid #d0d4d9;border-radius:8px;padding:16px 20px;margin:20px auto;box-shadow:0 0 8px 0 rgba(0,0,0,.1)\"><tr><td colspan=3><div style=\"padding-bottom:16px;margin-bottom:20px;border-bottom:1px solid #edf1f5;max-width:600px\"><img src=https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/devtron/devtron-logo.png style=max-width:122px alt=ci-triggered></div><tr><td colspan=3><div style=\"background-color:#FDE7E7;border-top-left-radius:8px;border-top-right-radius:8px;padding:20px 20px 16px 20px;display:flex;justify-content:space-between\"><div style=width:90%><div style=font-size:16px;line-height:24px;font-weight:600;margin-bottom:6px;color:#000a14>🚫  Auto-deployment blocked</div><span style=font-size:14px;line-height:20px;color:#000a14>{{eventTime}}</span><br><div><span style=font-size:14px;line-height:20px;color:#000a14>by</span><span style=font-size:14px;line-height:20px;color:#06c;margin-left:4px>{{triggeredBy}}</span></div></div><div><img src=https://cdn.devtron.ai/images/image_deploy_notification.png style=height:72px;width:72px></div></div><tr><td colspan=3><div style=display:flex><div style=\"width:124px;background-color:#FDE7E7;border-bottom-left-radius:8px;padding:0 0 20px 20px\">{{#deploymentHistoryLink}}<a href={{&deploymentHistoryLink}} style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:0;min-width:64px;text-transform:capitalize;text-align:center;background:#06c;color:#fff;border:1px solid transparent;cursor:pointer\">View Pipeline</a>{{/deploymentHistoryLink}}</div><div style=\"width:90%;background-color:#FDE7E7;border-bottom-right-radius:8px;padding:0 0 20px 20px\">{{#appDetailsLink}}<a href={{&appDetailsLink}} style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:0;min-width:64px;text-transform:capitalize;text-align:center;border:1px solid #d0d4d9;color:#06c;cursor:pointer;width:90%\">App Details</a>{{/appDetailsLink}}</div></div><tr><td><br><tr><td colspan=3><div style=\"display:flex;font-weight:600;border-radius:4px;border:1px solid #fde7e7;background:#fff;padding:8px 16px;margin-bottom:16px;display:{{deploymentWindowCommentStyle}};\"><img src=https://cdn.devtron.ai/images/shield-image.png style=margin-right:4px;height:16px;width:16px>{{deploymentWindowComment}}</div><tr><td><div style=color:#3b444c;font-size:13px>Application</div><td colspan=2><div style=color:#3b444c;font-size:13px>Environment</div><tr><td><div style=color:#000a14;font-size:14px>{{appName}}</div><td><div style=color:#000a14;font-size:14px>{{envName}}</div><tr><td><div style=color:#3b444c;font-size:13px;margin-top:12px>Stage</div><tr><td><div style=color:#000a14;font-size:14px>{{stage}}</div><tr><tr><td colspan=3><div style=\"font-weight:600;margin-top:20px;width:100%;border-top:1px solid #edf1f5;padding:16px 0 12px;font-size:14px\">Source Code</div></tr>{{#ciMaterials}} {{^webhookType}}<tr><td><div style=color:#3b444c;font-size:13px>Branch</div><td colspan=2><div style=color:#3b444c;font-size:13px>Commit</div><tr><tr><td><div style=color:#000a14;font-size:14px>{{appName}}/{{branch}}</div><td><div style=color:#000a14;font-size:14px>{{commit}}</div></tr>{{/webhookType}} {{/ciMaterials}}<tr><td colspan=3><div style=\"font-weight:600;margin-top:20px;width:100%;border-top:1px solid #edf1f5;padding:16px 0 12px;font-size:14px\">Image Details</div><tr><td><div style=color:#3b444c;font-size:13px>Image tag</div><tr><td><div style=color:#000a14;font-size:14px>{{dockerImg}}</div><tr><td><br><tr><td colspan=3><div style=\"border-top:1px solid #edf1f5;margin:20px 0 16px 0;height:1px\"></div><tr><td colspan=2 style=display:flex><span><a href=https://twitter.com/DevtronL style=cursor:pointer;text-decoration:none;padding-right:12px;display:flex target=_blank><div><img src=https://cdn.devtron.ai/images/twitter_social_dark.png style=width:20px></div></a></span><span><a href=https://www.linkedin.com/company/devtron-labs/mycompany/ style=cursor:pointer;text-decoration:none;padding-right:12px;display:flex target=_blank><div><img src=https://cdn.devtron.ai/images/linkedin_social_dark.png style=width:20px></div></a></span><span><a href=https://devtron.ai/blog/ style=color:#000a14;font-size:13px;line-height:20px;cursor:pointer;text-decoration:underline;padding-right:12px target=_blank>Blog</a></span><span><a href=https://devtron.ai/ style=color:#000a14;font-size:13px;line-height:20px;cursor:pointer;text-decoration:underline target=_blank>Website</a></span><td colspan=2 style=text-align:right><div style=color:#767d84;font-size:13px;line-height:20px>© Devtron Labs 2024</div></table>"}'
where channel_type = 'ses'
and node_type = 'CD'
and event_type_id = 6;
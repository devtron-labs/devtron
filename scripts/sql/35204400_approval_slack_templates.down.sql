/*
 * Copyright (c) 2024. Devtron Inc.
 */

-- Remove Slack template for config approval (event_type_id = 5)
DELETE FROM "public"."notification_templates"
WHERE channel_type = 'slack' AND node_type = 'CD' AND event_type_id = 5 AND template_name = 'Config approval slack template';

-- Remove Slack template for deployment approval (event_type_id = 4)
DELETE FROM "public"."notification_templates"
WHERE channel_type = 'slack' AND node_type = 'CD' AND event_type_id = 4 AND template_name = 'CD approval slack template';


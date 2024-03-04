delete from "public"."notification_templates" where event_type_id=6;
delete from notifier_event_log where event_type_id=6;
delete from public.event where event_type='BLOCKED';
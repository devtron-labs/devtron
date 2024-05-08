delete from "public"."notification_templates" where event_type_id=5;
delete from notifier_event_log where event_type_id=5;
delete from public.event where event_type='CONFIG APPROVAL';
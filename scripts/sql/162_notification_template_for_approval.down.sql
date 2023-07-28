ALTER TABLE ONLY public.notifier_event_log DROP CONSTRAINT notifier_event_log_event_type_id_fkey;
ALTER TABLE ONLY public.notifier_event_log
    ADD CONSTRAINT notifier_event_log_event_type_id_fkey FOREIGN KEY (event_type_id) REFERENCES public.event(id) ON DELETE CASCADE;
delete from "public"."notification_templates" where event_type_id=4;
delete from public.event where event_type='APPROVAL';
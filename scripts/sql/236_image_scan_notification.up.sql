-- Migration of data type in webhook_config
ALTER TABLE webhook_config ALTER payload TYPE VARCHAR;

-- Migration for creating new event
INSERT INTO "public"."event" (id, event_type, description) VALUES (8, 'IMAGE SCANNING', '');

-- Create notification_rule sequence
CREATE SEQUENCE IF NOT EXISTS id_seq_notification_rule_sequence;

-- notification_rule table definition
CREATE TABLE IF NOT EXISTS "public"."notification_rule" (
                                                            "id"                        integer NOT NULL DEFAULT nextval('id_seq_notification_rule_sequence'::regclass),
    "expression"                VARCHAR(1000),
    "condition_type"            integer NOT NULL DEFAULT 0,
    "created_on"                timestamptz NOT NULL,
    "created_by"                int4        NOT NULL,
    "updated_on"                timestamptz,
    "updated_by"                int4,
    PRIMARY KEY ("id")
    );

-- Migration for creating new column in notification_settings_view
ALTER TABLE "public"."notification_settings_view"
    ADD COLUMN IF NOT EXISTS internal bool;

-- Migration for creating new column in notification_settings
ALTER TABLE "public"."notification_settings"
    ADD COLUMN IF NOT EXISTS notification_rule_id integer,
    ADD COLUMN IF NOT EXISTS additional_config_json VARCHAR;

-- Constraints for the new column in notification_settings
ALTER TABLE "public"."notification_settings"
    ADD CONSTRAINT notification_setting_notification_rule_id_fk
        FOREIGN KEY ("notification_rule_id")
            REFERENCES "public"."notification_rule"("id");

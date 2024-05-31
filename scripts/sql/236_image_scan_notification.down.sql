-- Create a new function for JSON validation
CREATE OR REPLACE FUNCTION jsonb_valid(_txt text)
  RETURNS bool
  LANGUAGE plpgsql IMMUTABLE STRICT AS
$func$
BEGIN
RETURN _txt::json IS NOT NULL;
EXCEPTION
   WHEN SQLSTATE '22P02' THEN  -- invalid_text_representation
      RETURN false;
END
$func$;

-- Add comment to the func
COMMENT ON FUNCTION jsonb_valid(text) IS 'Validate if input text is valid JSON.
Returns true, false, or NULL on NULL input.';

-- Down Migration
DELETE FROM webhook_config WHERE NOT jsonb_valid(payload);
ALTER TABLE webhook_config ALTER payload TYPE JSONB USING payload::JSONB;


-- Migration for creating new column in notification_settings
ALTER TABLE "public"."notification_settings"
    DROP COLUMN IF EXISTS notification_rule_id,
    DROP COLUMN IF EXISTS additional_config_json;

/*
 * Copyright (c) 2024. Devtron Inc.
 */

-- Drop notification_rule table
DROP TABLE IF EXISTS "public"."notification_rule";

-- Drop sequence notification_rule
DROP SEQUENCE IF EXISTS id_seq_notification_rule_sequence;

-- Drop internal column from table notification_settings_view
ALTER TABLE "public"."notification_settings_view"
    DROP COLUMN IF EXISTS internal;

-- Delete notification_settings
DELETE FROM "public"."notification_settings" WHERE event_type_id = 8;

-- Delete notification_settings_view
DELETE FROM "public"."notification_settings_view" WHERE config::jsonb @> '{"eventTypeIds": [8]}';

-- Delete event
DELETE FROM "public"."event" WHERE id = 8;

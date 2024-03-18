ALTER  TABLE global_policy_searchable_field ADD COLUMN IF NOT EXISTS field_name varchar;
ALTER  TABLE global_policy_searchable_field ADD COLUMN IF NOT EXISTS value_int integer;
ALTER  TABLE global_policy_searchable_field ADD COLUMN IF NOT EXISTS value_time_stamp timestamptz;
CREATE UNIQUE INDEX idx_unique_policy_name_policy_of
    ON global_policy (name,policy_of)
    WHERE deleted = false;
ALTER TABLE  resource_filter_evaluation_audit ADD COLUMN "filter_type" integer DEFAULT 1;

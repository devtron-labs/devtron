ALTER  TABLE global_policy_searchable_field ADD COLUMN IF NOT EXISTS field_name varchar;
ALTER  TABLE global_policy_searchable_field ADD COLUMN IF NOT EXISTS value_int integer;
ALTER  TABLE global_policy_searchable_field ADD COLUMN IF NOT EXISTS value_time_stamp timestamptz;

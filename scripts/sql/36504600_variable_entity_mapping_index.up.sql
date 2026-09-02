-- index for looking up live usages of a scoped variable by name (variable deletion protection)
CREATE INDEX IF NOT EXISTS idx_variable_entity_mapping_variable_name
    ON variable_entity_mapping (variable_name)
    WHERE is_deleted = false;

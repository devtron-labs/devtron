UPDATE cve_policy_control
SET deleted = true, updated_on = 'now()', updated_by = '1'
WHERE severity = '3' OR severity = '5';
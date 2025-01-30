-- Have deleted duplicate entries in role_group_role_mapping which got created on save/update of role group.
-- have partitioned by role_group_id, role_id and filtered only those rows which have row number >1 as first entry will have row number 1

DELETE FROM role_group_role_mapping
WHERE ctid IN (
    SELECT ctid
    FROM (
             SELECT ctid,
                    ROW_NUMBER() OVER (PARTITION BY role_group_id, role_id ORDER BY ctid) AS rn
             FROM role_group_role_mapping
         ) AS subquery
    WHERE rn > 1
);
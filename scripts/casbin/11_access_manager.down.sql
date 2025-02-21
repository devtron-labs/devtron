BEGIN;
DELETE FROM casbin_rule where v0='role:all-access-manager___' and v1='user/*/*';
DELETE FROM casbin_rule where v0='role:super-admin___' and v1='user/*/*';
COMMIT;
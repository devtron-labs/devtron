DROP TABLE IF EXISTS public.default_rbac_role_data;
DROP TABLE IF EXISTS public.resource_protection;
DROP TABLE IF EXISTS public.resource_protection_history;
DROP TABLE IF EXISTS public.draft;
DROP TABLE IF EXISTS public.draft_version;
DROP TABLE IF EXISTS public.draft_version_comment;

---- DROP sequence
DROP SEQUENCE IF EXISTS public.id_seq_default_rbac_role_data;
DROP SEQUENCE IF EXISTS public.id_seq_resource_protection;
DROP SEQUENCE IF EXISTS public.id_seq_resource_protection_history;
DROP SEQUENCE IF EXISTS public.id_seq_draft;
DROP SEQUENCE IF EXISTS public.id_seq_draft_version;
DROP SEQUENCE IF EXISTS public.id_seq_draft_version_comment;

DELETE from rbac_policy_resource_detail where resource = 'config';
DELETE from rbac_role_resource_detail where resource = 'approver';
DELETE from default_rbac_role_data where role = 'configApprover';
---- drop column git_host_id from git_provider
ALTER TABLE git_provider
    DROP COLUMN git_host_id;

---- drop table git_host
DROP TABLE git_host;

---- DROP sequence
DROP SEQUENCE IF EXISTS public.git_host_id_seq;

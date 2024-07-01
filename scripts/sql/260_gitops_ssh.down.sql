ALTER TABLE public.gitops_config DROP COLUMN IF EXISTS ssh_key;
ALTER TABLE public.gitops_config DROP COLUMN IF EXISTS auth_mode;
ALTER TABLE public.gitops_config ALTER COLUMN "token" SET NOT NULL;
ALTER TABLE public.gitops_config ALTER COLUMN "host" SET NOT NULL;
ALTER TABLE public.gitops_config DROP COLUMN IF EXISTS ssh_host;
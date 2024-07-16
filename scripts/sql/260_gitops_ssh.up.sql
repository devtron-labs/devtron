ALTER TABLE public.gitops_config ADD COLUMN IF NOT EXISTS ssh_key text;
ALTER TABLE public.gitops_config ADD COLUMN IF NOT EXISTS auth_mode text;
ALTER TABLE public.gitops_config ALTER COLUMN "token" DROP NOT NULL;
ALTER TABLE public.gitops_config ALTER COLUMN "host" DROP NOT NULL;
ALTER TABLE public.gitops_config ADD ssh_host varchar(250) NULL;



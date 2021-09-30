ALTER TABLE "public"."git_provider" ADD COLUMN "ssh_private_key" TEXT;

ALTER TABLE "public"."git_material" ADD COLUMN fetch_submodules BOOLEAN;
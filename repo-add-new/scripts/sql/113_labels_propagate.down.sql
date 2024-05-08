ALTER TABLE app_label DROP COLUMN IF EXISTS propagate;

ALTER TABLE "public"."app_label" ALTER COLUMN "key" SET DATA TYPE varchar(255);
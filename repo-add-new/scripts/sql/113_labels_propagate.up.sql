ALTER TABLE app_label ADD COLUMN IF NOT EXISTS propagate boolean NOT NULL DEFAULT true;

ALTER TABLE "public"."app_label" ALTER COLUMN "key" SET DATA TYPE varchar(317);
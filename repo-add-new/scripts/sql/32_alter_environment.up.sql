ALTER TABLE "public"."environment" ADD COLUMN IF NOT EXISTS "environment_identifier" varchar(250);
UPDATE environment SET environment_identifier = e2.environment_name FROM environment e2 WHERE e2.id = environment.id AND environment.environment_identifier IS NULL;
ALTER TABLE environment ALTER COLUMN environment_identifier SET NOT NULL;
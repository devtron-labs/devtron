ALTER TABLE "public"."gitops_config" ADD COLUMN "validated_on" TIMESTAMP;
ALTER TABLE "public"."gitops_config" ADD COLUMN "validation_errors" TEXT;
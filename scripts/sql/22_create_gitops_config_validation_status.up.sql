-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_gitops_config_validation_status;

-- Table Definition
CREATE TABLE "public"."gitops_config_validation_status" (
                                           "id" INT4 NOT NULL DEFAULT nextval('id_seq_gitops_config_validation_status'::regclass),
                                           "provider" TEXT,
                                           "validation_errors" TEXT,
                                           "validated_on" TIMESTAMP,
                                           PRIMARY KEY ("id")
);


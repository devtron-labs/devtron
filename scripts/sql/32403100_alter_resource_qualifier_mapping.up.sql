ALTER TABLE resource_qualifier_mapping ADD COLUMN IF NOT EXISTS global_policy_id INT;

ALTER TABLE  resource_qualifier_mapping
    ADD CONSTRAINT "resource_qualifier_mapping_global_policy_id_fkey" FOREIGN KEY ("global_policy_id") REFERENCES "public"."global_policy" ("id");

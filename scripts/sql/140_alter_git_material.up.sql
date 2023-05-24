ALTER TABLE "public"."git_material" ADD COLUMN "filter_pattern" json DEFAULT '[]';
ALTER TABLE "public"."git_material_history" ADD COLUMN "filter_pattern" json DEFAULT '[]';
ALTER TABLE "public"."roles"
    ADD COLUMN IF NOT EXISTS "access_type" character varying(100);

CREATE UNIQUE INDEX IF NOT EXISTS "role_unique" ON "public"."roles" USING BTREE ("role","access_type");
ALTER TABLE "public"."roles" DROP COLUMN IF EXISTS "access_type";

/*
 * Copyright (c) 2024. Devtron Inc.
 */

---- DROP Index
DROP INDEX IF EXISTS "public"."role_unique";
CREATE UNIQUE INDEX IF NOT EXISTS "role_unique" ON "public"."roles" USING BTREE ("role");
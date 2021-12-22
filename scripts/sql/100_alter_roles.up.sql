UPDATE "public"."roles" SET "access_type" = '';

DROP INDEX "public"."role_unique";

CREATE UNIQUE INDEX "role_unique" ON "public"."roles" USING BTREE ("role","access_type");
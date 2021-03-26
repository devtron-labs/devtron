CREATE UNIQUE INDEX "app_app_name_key" ON "public"."app" USING BTREE ("app_name");

ALTER TABLE "public"."deployment_status" DROP COLUMN "active";
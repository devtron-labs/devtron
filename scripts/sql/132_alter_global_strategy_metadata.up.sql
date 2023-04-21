ALTER TABLE "public"."global_strategy_metadata" ADD COLUMN "key" varchar(250);
COMMENT ON COLUMN "public"."global_strategy_metadata"."key" IS 'strategy json key';
UPDATE "public"."global_strategy_metadata" SET "key" = 'recreate' WHERE "name" = 'RECREATE';
UPDATE "public"."global_strategy_metadata" SET "key" = 'canary' WHERE "name" = 'CANARY';
UPDATE "public"."global_strategy_metadata" SET "key" = 'rolling' WHERE "name" = 'ROLLING';
UPDATE "public"."global_strategy_metadata" SET "key" = 'blueGreen' WHERE "name" = 'BLUE-GREEN';

ALTER TABLE "public"."global_strategy_metadata_chart_ref_mapping" ADD COLUMN "default" bool DEFAULT 'false';
UPDATE global_strategy_metadata_chart_ref_mapping set "default"=true WHERE global_strategy_metadata_id=(SELECT id from global_strategy_metadata WHERE name like 'ROLLING');

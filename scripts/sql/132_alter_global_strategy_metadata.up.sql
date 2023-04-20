ALTER TABLE "public"."global_strategy_metadata" ADD COLUMN "key" varchar(250);
COMMENT ON COLUMN "public"."global_strategy_metadata"."key" IS 'strategy json key';
UPDATE "public"."global_strategy_metadata" SET "key" = 'recreate' WHERE "name" = 'RECREATE';
UPDATE "public"."global_strategy_metadata" SET "key" = 'canary' WHERE "name" = 'CANARY';
UPDATE "public"."global_strategy_metadata" SET "key" = 'rolling' WHERE "name" = 'ROLLING';
UPDATE "public"."global_strategy_metadata" SET "key" = 'blueGreen' WHERE "name" = 'BLUE-GREEN';
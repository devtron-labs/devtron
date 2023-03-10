ALTER TABLE "chart_repo" ADD COLUMN "allow_insecure_connection" bool;
ALTER TABLE "chart_repo" ADD COLUMN "is_healthy" bool;

UPDATE chart_repo SET "allow_insecure_connection" = false;
UPDATE chart_repo SET "is_healthy" = true where deleted=false;


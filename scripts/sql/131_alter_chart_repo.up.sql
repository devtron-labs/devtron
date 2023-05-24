ALTER TABLE "chart_repo" ADD COLUMN "allow_insecure_connection" bool;
UPDATE chart_repo SET "allow_insecure_connection" = false;



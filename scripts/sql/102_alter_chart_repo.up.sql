ALTER TABLE "chart_repo" ADD COLUMN "secured" bool;

UPDATE chart_repo SET "secured" = true;
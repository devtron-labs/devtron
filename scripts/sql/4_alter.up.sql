ALTER TABLE "public"."chart_repo" ADD COLUMN "user_name" varchar(250);
ALTER TABLE "public"."chart_repo" ADD COLUMN "password" varchar(250);
ALTER TABLE "public"."chart_repo" ADD COLUMN "ssh_key" varchar(250);
ALTER TABLE "public"."chart_repo" ADD COLUMN "access_token" varchar(250);
ALTER TABLE "public"."chart_repo" ADD COLUMN "auth_mode" varchar(250);